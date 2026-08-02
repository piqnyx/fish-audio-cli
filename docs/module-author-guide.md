# Module author guide

> **Document status:** normative implementation guide for the current pre-release module API.
>
> **Audience:** Go developers adding a new built-in text-processing module to `fish-audio-cli`.
>
> **Prerequisites:** read [`architecture.md`](architecture.md), [`pipeline.md`](pipeline.md), and [`modules.md`](modules.md) before implementing a production module.
>
> **Scope:** this guide shows the complete path from choosing a module boundary through package creation, strict config decoding, preparation, building, processing, tests, registry registration, documentation, and review.

---

## 1. The contract in one page

A module is a compiled-in text transformer.

The core gives it valid current text and expects either:

- valid replacement text; or
- an error.

The complete lifecycle is:

```text
JSON module instance
    ↓
core validates envelope
    ↓
registry selects type
    ↓
module Prepare validates private config
    ↓
Prepare returns instance-specific builder
    ↓
all module instances finish Prepare
    ↓
builders create processors in array order
    ↓
pipeline calls Process in array order
    ↓
processor leaves valid text or returns error
    ↓
pipeline applies rollback and configured error policy
```

A module author owns:

- the private config schema;
- module-specific defaults;
- semantic validation;
- path resolution for module-owned files;
- the transformation algorithm;
- instance-local runtime state;
- provider-specific request and response validation;
- module-focused tests;
- module documentation.

The core owns:

- module instance names;
- type selection;
- `onError`;
- preparation and build ordering;
- text snapshots and rollback;
- fallback policy;
- pipeline reports;
- standard lifecycle logging;
- Fish synthesis;
- secrets for Fish Audio;
- output publication.

A module must not duplicate those core responsibilities.

---

## 2. Decide whether the feature is a module

A feature belongs in a module when all of the following are true:

- its shared input is current text;
- its shared output is replacement text;
- it can report failure as an error;
- users may reasonably enable, disable, repeat, or reorder it;
- its config belongs to one module instance;
- it should run before Fish Audio synthesis.

Typical module features:

- punctuation cleanup;
- Unicode normalization;
- abbreviation expansion;
- emoji verbalization;
- pronunciation markup;
- text replacement rules;
- language-model emotion tagging;
- translation;
- remote text rewriting.

A feature usually does not belong in a module when it owns:

- CLI flags;
- top-level config loading;
- Fish request headers;
- Fish API authentication;
- audio streaming;
- output formats;
- final-file publication;
- process exit codes;
- global logging destinations.

A useful boundary test is:

```text
Can the feature be described as:
"valid text in, valid text or error out"?
```

If not, it probably belongs elsewhere.

---

## 3. Choose a stable type name

The type name becomes a public configuration identifier.

Use a name that is:

- lowercase;
- concise;
- descriptive;
- independent of one configured instance;
- unlikely to need renaming.

Good examples:

```text
replace
emoji
normalizer
pronunciation
llm-emotion
```

Poor examples:

```text
module2
my-module
new-normalizer
first-pass
temporary
```

Instance roles belong in `name`, not `type`.

Example:

```json
{
  "name": "replace-product-names",
  "type": "replace",
  "config": {}
}
```

Once released, changing the type name breaks existing configs because registry lookup is exact.

---

## 4. Create a dedicated package

Create one package per implementation:

```text
internal/modules/<type>/
```

For the worked example:

```text
internal/modules/replace/
├── replace.go
└── replace_test.go
```

Keep module-specific code inside that package.

Do not place private provider structs or config rules in:

- `internal/modules`;
- `internal/pipeline`;
- `internal/config`;
- `internal/app`.

The registry package dispatches modules. It should not become a warehouse for their algorithms.

---

## 5. Worked example: `replace`

This guide uses a complete example module named:

```text
replace
```

It replaces occurrences of one string with another.

Example configuration:

```json
{
  "name": "replace-product-name",
  "type": "replace",
  "onError": "abort",
  "config": {
    "from": "Old Product",
    "to": "New Product",
    "count": -1
  }
}
```

Behavior:

- `from` is required and must not be empty;
- `to` is required as a JSON field but may be an empty string;
- `count` is optional;
- omitted `count` means replace all occurrences;
- `count: -1` also means replace all;
- positive values limit the number of replacements;
- `0` and values below `-1` are rejected;
- excessively large positive values are rejected;
- final output must remain valid nonblank UTF-8 text.

This example is intentionally local and deterministic so the module architecture remains visible.

---

## 6. Define a private config type

Use an unexported config type unless another package genuinely needs it.

```go
type config struct {
    From  string  `json:"from"`
    To    *string `json:"to"`
    Count *int    `json:"count,omitempty"`
}
```

Why `Count` is a pointer:

- omitted field produces `nil`;
- explicit `-1` remains distinguishable;
- explicit `0` can be rejected;
- no decoder-side inheritance is needed.

Why `To` is a pointer:

- an omitted field produces `nil` and can be rejected;
- an explicitly configured empty string remains distinguishable and valid;
- the documented required-field contract is enforced rather than implied.

Do not use exported configuration types merely because exported names feel important. Public surface area is a maintenance cost, not a medal.

### 6.1 Required strings

Go’s JSON decoder cannot distinguish an omitted string from an explicit empty string when the field type is `string`.

For `from`, semantic validation handles both cases:

```go
if cfg.From == "" {
    return nil, fmt.Errorf("from must not be empty")
}
```

For `to`, omission and an explicit empty string have different meanings, so the field uses `*string`:

```go
if cfg.To == nil {
    return nil, fmt.Errorf("to must be present")
}
```

An explicit `""` is retained as a valid replacement value.

### 6.2 JSON field naming

Use the exact camel-case convention used by project configuration:

```go
`json:"maxReplacements"`
```

Do not rely on case-insensitive decoding.

`moduleconfig.Decode` rejects incorrectly capitalized field names.

---

## 7. Strictly decode config

Use:

```go
moduleconfig.Decode(raw, &cfg)
```

This enforces the common structural policy:

- config is present;
- config is a JSON object;
- valid UTF-8;
- one JSON value;
- no duplicate keys;
- exact field-name casing;
- no unknown fields.

Pattern:

```go
var cfg config

if err := moduleconfig.Decode(raw, &cfg); err != nil {
    return nil, fmt.Errorf(
        "replace config: %w",
        err,
    )
}
```

The module-specific prefix makes the final layered error readable.

Do not call `encoding/json.Unmarshal` directly unless the module has a documented reason to bypass project strictness. Ordinary modules do not.

---

## 8. Add semantic validation

Strict decoding proves that JSON matches the Go shape. It does not prove that values make sense.

For `replace`, define a limit:

```go
const maxReplacementCount = 10_000
```

Resolve `count` with a helper:

```go
func resolveCount(value *int) (int, error) {
    if value == nil || *value == -1 {
        return -1, nil
    }

    if *value <= 0 {
        return 0, fmt.Errorf(
            "count must be -1 or greater than zero",
        )
    }

    if *value > maxReplacementCount {
        return 0, fmt.Errorf(
            "count must be less than or equal to %d",
            maxReplacementCount,
        )
    }

    return *value, nil
}
```

Validate `from`:

```go
if cfg.From == "" {
    return nil, fmt.Errorf(
        "replace config: from must not be empty",
    )
}
```

Do not silently normalize values unless that normalization is part of the documented contract.

For example, trimming `from` would prevent users from intentionally replacing whitespace.

---

## 9. Prepare versus build

The distinction is mandatory.

### Prepare

Use `Prepare` for:

- strict config decoding;
- semantic validation;
- resolving paths;
- creating immutable prepared values;
- returning a builder.

### Build

Use the returned builder for:

- constructing one processor;
- construction deferred until every configured module has prepared successfully;
- runtime initialization that requires no explicit cleanup.

Do not invoke the builder inside `Prepare`.

Incorrect:

```go
func Prepare(...) (pipeline.ProcessorBuilder, error) {
    processor := newProcessor()

    return func() (pipeline.Processor, error) {
        return processor, nil
    }, nil
}
```

This creates runtime state before later module configs have been validated.

Preferred:

```go
func Prepare(...) (pipeline.ProcessorBuilder, error) {
    preparedValue := ...

    return func() (pipeline.Processor, error) {
        return &processor{
            value: preparedValue,
        }, nil
    }, nil
}
```

---

## 10. Complete implementation

Create:

```text
internal/modules/replace/replace.go
```

with:

```go
package replace

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/piqnyx/fish-audio-cli/internal/moduleconfig"
    "github.com/piqnyx/fish-audio-cli/internal/nilvalue"
    "github.com/piqnyx/fish-audio-cli/internal/pipeline"
    "github.com/piqnyx/fish-audio-cli/internal/projectpath"
    "github.com/piqnyx/fish-audio-cli/internal/textcontract"
)

const maxReplacementCount = 10_000

type config struct {
    From  string  `json:"from"`
    To    *string `json:"to"`
    Count *int    `json:"count,omitempty"`
}

type processor struct {
    from  string
    to    string
    count int
}

func resolveCount(value *int) (int, error) {
    if value == nil || *value == -1 {
        return -1, nil
    }

    if *value <= 0 {
        return 0, fmt.Errorf(
            "count must be -1 or greater than zero",
        )
    }

    if *value > maxReplacementCount {
        return 0, fmt.Errorf(
            "count must be less than or equal to %d",
            maxReplacementCount,
        )
    }

    return *value, nil
}

// Prepare validates one replace module instance and returns its processor
// builder.
func Prepare(
    _ projectpath.Resolver,
    raw json.RawMessage,
) (
    pipeline.ProcessorBuilder,
    error,
) {
    var cfg config

    if err := moduleconfig.Decode(
        raw,
        &cfg,
    ); err != nil {
        return nil, fmt.Errorf(
            "replace config: %w",
            err,
        )
    }

    if cfg.From == "" {
        return nil, fmt.Errorf(
            "replace config: from must not be empty",
        )
    }

    if cfg.To == nil {
        return nil, fmt.Errorf(
            "replace config: to must be present",
        )
    }

    count, err := resolveCount(cfg.Count)
    if err != nil {
        return nil, fmt.Errorf(
            "replace config: %w",
            err,
        )
    }

    from := cfg.From
    to := *cfg.To

    return func() (
        pipeline.Processor,
        error,
    ) {
        return &processor{
            from:  from,
            to:    to,
            count: count,
        }, nil
    }, nil
}

// Process replaces configured text occurrences in the current document.
func (p *processor) Process(
    ctx context.Context,
    document *pipeline.Document,
) error {
    if p == nil {
        return fmt.Errorf(
            "replace processor is nil",
        )
    }

    if nilvalue.IsNil(ctx) {
        return fmt.Errorf(
            "context is nil",
        )
    }

    if document == nil {
        return fmt.Errorf(
            "document is nil",
        )
    }

    if err := ctx.Err(); err != nil {
        return err
    }

    result := strings.Replace(
        document.Text,
        p.from,
        p.to,
        p.count,
    )

    if err := textcontract.Validate(result); err != nil {
        return fmt.Errorf(
            "replace output is invalid: %w",
            err,
        )
    }

    document.Text = result

    return nil
}
```

### 10.1 Why the path resolver is unused

This module has no paths.

Keep the parameter and name it `_`:

```go
_ projectpath.Resolver
```

The signature remains compatible with the registry.

### 10.2 Why values are copied before the closure

These lines:

```go
from := cfg.From
to := cfg.To
```

make the builder’s captured intent explicit.

Strings are immutable values, so no deep copy is needed.

For slices, maps, pointers, or byte buffers, consider defensive copies.

### 10.3 Why output is validated in the module

The pipeline validates successful output again.

The module-side check provides:

- module-local error context;
- safe direct processor tests;
- a clear statement that the module does not intentionally emit invalid text.

The pipeline remains the final shared boundary.

### 10.4 Why no fallback appears here

The module returns an error.

The pipeline applies:

- `use_previous`;
- `use_original`;
- `skip`;
- `abort`.

Embedding fallback inside `Process` would duplicate and conflict with the core policy.

---

## 11. Write focused tests

Create:

```text
internal/modules/replace/replace_test.go
```

with:

```go
package replace

import (
    "context"
    "encoding/json"
    "errors"
    "testing"

    "github.com/piqnyx/fish-audio-cli/internal/pipeline"
    "github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

func buildProcessor(
    t *testing.T,
    raw string,
) pipeline.Processor {
    t.Helper()

    builder, err := Prepare(
        projectpath.Resolver{},
        json.RawMessage(raw),
    )
    if err != nil {
        t.Fatalf(
            "Prepare() error = %v",
            err,
        )
    }

    if builder == nil {
        t.Fatal(
            "Prepare() builder = nil",
        )
    }

    processor, err := builder()
    if err != nil {
        t.Fatalf(
            "builder() error = %v",
            err,
        )
    }

    if pipeline.IsNilProcessor(processor) {
        t.Fatal(
            "builder() processor = nil",
        )
    }

    return processor
}

func processText(
    t *testing.T,
    processor pipeline.Processor,
    text string,
) (
    string,
    error,
) {
    t.Helper()

    document, err := pipeline.NewDocument(text)
    if err != nil {
        t.Fatalf(
            "pipeline.NewDocument() error = %v",
            err,
        )
    }

    err = processor.Process(
        context.Background(),
        document,
    )

    return document.Text, err
}

func TestPrepareRejectsUnknownField(
    t *testing.T,
) {
    t.Parallel()

    _, err := Prepare(
        projectpath.Resolver{},
        json.RawMessage(
            `{
                "from":"old",
                "to":"new",
                "inventMeaning":true
            }`,
        ),
    )
    if err == nil {
        t.Fatal(
            "Prepare() error = nil, want error",
        )
    }
}

func TestPrepareRejectsEmptyFrom(
    t *testing.T,
) {
    t.Parallel()

    _, err := Prepare(
        projectpath.Resolver{},
        json.RawMessage(
            `{"from":"","to":"new"}`,
        ),
    )
    if err == nil {
        t.Fatal(
            "Prepare() error = nil, want error",
        )
    }
}

func TestPrepareRejectsMissingTo(
    t *testing.T,
) {
    t.Parallel()

    _, err := Prepare(
        projectpath.Resolver{},
        json.RawMessage(
            `{"from":"old"}`,
        ),
    )
    if err == nil {
        t.Fatal(
            "Prepare() error = nil, want error",
        )
    }
}

func TestPrepareRejectsZeroCount(
    t *testing.T,
) {
    t.Parallel()

    _, err := Prepare(
        projectpath.Resolver{},
        json.RawMessage(
            `{
                "from":"old",
                "to":"new",
                "count":0
            }`,
        ),
    )
    if err == nil {
        t.Fatal(
            "Prepare() error = nil, want error",
        )
    }
}

func TestProcessReplacesAllByDefault(
    t *testing.T,
) {
    t.Parallel()

    processor := buildProcessor(
        t,
        `{"from":"old","to":"new"}`,
    )

    got, err := processText(
        t,
        processor,
        "old and old",
    )
    if err != nil {
        t.Fatalf(
            "Process() error = %v",
            err,
        )
    }

    const want = "new and new"

    if got != want {
        t.Fatalf(
            "text = %q, want %q",
            got,
            want,
        )
    }
}

func TestProcessRespectsCount(
    t *testing.T,
) {
    t.Parallel()

    processor := buildProcessor(
        t,
        `{
            "from":"old",
            "to":"new",
            "count":1
        }`,
    )

    got, err := processText(
        t,
        processor,
        "old and old",
    )
    if err != nil {
        t.Fatalf(
            "Process() error = %v",
            err,
        )
    }

    const want = "new and old"

    if got != want {
        t.Fatalf(
            "text = %q, want %q",
            got,
            want,
        )
    }
}

func TestProcessRejectsBlankOutput(
    t *testing.T,
) {
    t.Parallel()

    processor := buildProcessor(
        t,
        `{"from":"only","to":""}`,
    )

    got, err := processText(
        t,
        processor,
        "only",
    )
    if err == nil {
        t.Fatal(
            "Process() error = nil, want error",
        )
    }

    const want = "only"

    if got != want {
        t.Fatalf(
            "text = %q, want unchanged %q",
            got,
            want,
        )
    }
}

func TestProcessReturnsCanceledContext(
    t *testing.T,
) {
    t.Parallel()

    processor := buildProcessor(
        t,
        `{"from":"old","to":"new"}`,
    )

    document, err := pipeline.NewDocument(
        "old",
    )
    if err != nil {
        t.Fatalf(
            "pipeline.NewDocument() error = %v",
            err,
        )
    }

    ctx, cancel := context.WithCancel(
        context.Background(),
    )
    cancel()

    err = processor.Process(
        ctx,
        document,
    )
    if !errors.Is(
        err,
        context.Canceled,
    ) {
        t.Fatalf(
            "Process() error = %v, want %v",
            err,
            context.Canceled,
        )
    }

    if document.Text != "old" {
        t.Fatalf(
            "document.Text = %q, want unchanged",
            document.Text,
        )
    }
}

func TestInstancesAreIndependent(
    t *testing.T,
) {
    t.Parallel()

    first := buildProcessor(
        t,
        `{"from":"A","to":"first"}`,
    )
    second := buildProcessor(
        t,
        `{"from":"A","to":"second"}`,
    )

    firstText, err := processText(
        t,
        first,
        "A",
    )
    if err != nil {
        t.Fatalf(
            "first Process() error = %v",
            err,
        )
    }

    secondText, err := processText(
        t,
        second,
        "A",
    )
    if err != nil {
        t.Fatalf(
            "second Process() error = %v",
            err,
        )
    }

    if firstText != "first" {
        t.Fatalf(
            "first text = %q, want %q",
            firstText,
            "first",
        )
    }

    if secondText != "second" {
        t.Fatalf(
            "second text = %q, want %q",
            secondText,
            "second",
        )
    }
}
```

---

## 12. Understand one subtle test

`TestProcessRejectsBlankOutput` calls the processor directly.

The module computes an empty result, validates it before assignment, and returns an error.

Therefore the document remains unchanged.

When a processor mutates first and then fails, direct tests may observe partial mutation. The pipeline would roll it back, but module-local code is clearer when it computes and validates before assignment.

This is why the preferred pattern is:

```text
compute
validate
assign
return nil
```

rather than:

```text
assign partial result
discover error
return error
```

---

## 13. Register the type

Open:

```text
internal/modules/registry.go
```

Add the import:

```go
"github.com/piqnyx/fish-audio-cli/internal/modules/replace"
```

Change:

```go
var preparers = map[string]preparer{
    "passthrough": passthrough.Prepare,
}
```

to:

```go
var preparers = map[string]preparer{
    "passthrough": passthrough.Prepare,
    "replace":     replace.Prepare,
}
```

Registration is intentionally explicit.

Do not add package `init()` registration.

Explicit registry entries provide:

- deterministic availability;
- easy code review;
- visible type names;
- no import-order behavior;
- no hidden mutable global registration.

---

## 14. Add a registry integration test

Package-level module tests prove the implementation.

The registry should also prove that the public type is wired.

A focused test may configure:

```go
config.PipelineConfig{
    OnError: "use_previous",
    Modules: []config.ModuleConfig{
        {
            Name: "replace-test",
            Type: "replace",
            Config: json.RawMessage(
                `{"from":"old","to":"new"}`,
            ),
        },
    },
}
```

Then:

1. call `modules.Build`;
2. create a pipeline;
3. process `"old"`;
4. verify `"new"`;
5. verify step identity;
6. verify effective policy.

This catches the highly sophisticated failure mode where a perfect package exists and nobody registered it.

---

## 15. Test repeated instances through the registry

The module package test proves independent processors.

A registry-level test should prove independent raw configs are delivered in array order.

Example config:

```json
{
  "pipeline": {
    "onError": "use_previous",
    "modules": [
      {
        "name": "first-replace",
        "type": "replace",
        "config": {
          "from": "A",
          "to": "B"
        }
      },
      {
        "name": "second-replace",
        "type": "replace",
        "config": {
          "from": "B",
          "to": "C"
        }
      }
    ]
  }
}
```

Input:

```text
A
```

Final output:

```text
C
```

This simultaneously verifies:

- repeated type support;
- independent config;
- configured order;
- processor construction;
- pipeline text handoff.

---

## 16. Add pipeline failure-policy tests where relevant

A module does not implement fallback, but integration tests should verify its errors interact correctly with the pipeline.

For a module capable of failure, cover at least the policies relevant to real use.

### `use_previous`

```text
module error
previous valid text restored
later module runs
pipeline succeeds with recovered outcome
```

### `use_original`

```text
module error
original pipeline text restored
later module runs
pipeline succeeds with recovered outcome
```

### `skip`

```text
module error
previous text restored
later modules do not run
pipeline succeeds with stopped outcome
```

### `abort`

```text
module error
previous text restored
later modules do not run
pipeline returns error
```

Do not write four nearly identical integration tests for a module that cannot realistically fail. Test architecture where it provides evidence, not where it merely adds line count.

---

## 17. Document the config

Every registered module needs a complete configuration reference.

For `replace`, document:

| Field | Type | Required | Default | Accepted values |
|---|---|---:|---|---|
| `from` | string | yes | none | non-empty string |
| `to` | string | yes | none | any string; final text must remain valid |
| `count` | integer | no | `-1` | `-1` or `1` through `10000` |

Explain semantics:

- omitted `count` replaces all;
- `-1` replaces all;
- positive count replaces at most that many occurrences;
- no match is a successful unchanged result;
- removal that would produce blank output fails.

Include at least one complete instance example.

Do not make users infer defaults from Go source like medieval scholars reconstructing a lost civilization.

---

## 18. Update module documentation

Update:

```text
docs/modules.md
```

Add the registered type to the built-in module list.

Describe:

- purpose;
- config summary;
- whether it uses paths;
- whether it calls an external service;
- whether input text leaves the machine;
- error conditions;
- performance implications.

Do not duplicate the complete author guide or every config row there.

---

## 19. Update top-level configuration examples

When the project’s canonical example config should demonstrate the module, update:

```text
config.example.json
```

Only add it to the default example when it improves the normal first-run experience.

A module may be documented without being enabled by default.

Adding every module to the default chain would make the example executable perform surprising transformations, because apparently examples enjoy becoming production policy when nobody is watching.

---

## 20. Run formatting and tests

Minimum commands:

```bash
gofmt -w \
  internal/modules/replace/replace.go \
  internal/modules/replace/replace_test.go \
  internal/modules/registry.go \
  internal/modules/registry_test.go

go test -count=1 \
  ./internal/modules/replace \
  ./internal/modules \
  ./internal/pipeline

go test -count=1 ./...

go test -race -count=1 \
  ./internal/modules/replace \
  ./internal/modules \
  ./internal/pipeline

go vet ./...

git diff --check
```

A module using HTTP, files, goroutines, or shared state needs additional focused tests.

---

## 21. Check the actual diff

Before committing:

```bash
git status --short

git --no-pager diff --stat

git --no-pager diff -- \
  internal/modules/replace \
  internal/modules/registry.go \
  internal/modules/registry_test.go \
  docs/modules.md \
  docs/configuration.md \
  config.example.json
```

Expected scope for this example:

```text
new module package
registry import and entry
registry integration tests
module documentation
configuration reference
optional example config update
```

Unexpected changes should be removed or explained before commit.

---

## 22. Commit structure

A new module may justify more than one commit.

Reasonable sequence:

```text
feat: add replace text module
test: cover replace pipeline integration
docs: document replace module
```

A single coherent commit may also be acceptable when the change is small and reviewable.

Do not mix unrelated cleanup into the module commit.

The final audit will notice. It has developed trust issues for understandable reasons.

---

## 23. External-service module pattern

A module using an HTTP or LLM provider needs additional boundaries.

Conceptual prepared values:

```go
type preparedConfig struct {
    endpoint string
    model    string
    timeout  time.Duration
    keyPath  string
}
```

Preparation should validate:

- URL scheme and host;
- model string;
- timeout bounds;
- retry bounds;
- response-size limit;
- secret path;
- prompt or template syntax.

The builder may create:

- an HTTP client;
- immutable request templates;
- provider-specific client state that requires no explicit close.

`Process` should:

1. check context;
2. load or use the credential according to documented lifecycle;
3. build a bounded request;
4. send with context;
5. bound the response;
6. strictly decode it;
7. validate returned text;
8. assign only after validation;
9. return errors to the pipeline.

Document clearly that input text is sent to the external provider.

---

## 24. Secret handling for modules

Do not put API keys inline:

```json
{
  "apiKey": "secret"
}
```

Prefer:

```json
{
  "apiKeyFile": "secrets/provider-key"
}
```

Security requirements:

- resolve through the project path resolver;
- use a bounded read;
- require a documented single-value format;
- reject insecure file types and permissions;
- avoid symlink races;
- never log the value;
- keep secret lifetime short;
- do not include it in errors.

Before inventing a second secret loader, evaluate whether the existing secure secret boundary can be generalized without coupling unrelated providers.

Generalize only from real requirements.

---

## 25. Path-backed module pattern

A dictionary-backed module may define:

```go
type config struct {
    DictionaryFile string `json:"dictionaryFile"`
}
```

During `Prepare`:

```go
resolved, err := paths.Resolve(
    cfg.DictionaryFile,
)
if err != nil {
    return nil, fmt.Errorf(
        "resolve dictionary file: %w",
        err,
    )
}
```

Then validate:

- path is non-empty;
- expected file type;
- maximum size;
- UTF-8 where required;
- exact file format;
- duplicate entries;
- semantic conflicts.

Decide and document whether the file is:

- parsed during preparation;
- parsed during build;
- read for every process call.

For the current single-run CLI, parse-once initialization is usually clearest.

---

## 26. Large-output modules

The pipeline enforces valid text but does not currently apply a separate post-module byte limit.

A module that can greatly expand text should consider a documented bound.

Examples:

- abbreviation expansion loops;
- recursive templates;
- provider responses;
- generated markup;
- repeated substitutions.

Do not add a hidden truncation.

Possible outcomes must be explicit:

- reject oversized output;
- configure a maximum;
- return a typed error;
- leave fallback to the pipeline.

Silent truncation can change meaning and speech content.

---

## 27. Avoid global mutable state

Do not write:

```go
var activeConfig config
```

or:

```go
var client *Client
```

when values differ by instance.

Global mutable state breaks:

- repeated types;
- independent configs;
- race safety;
- test isolation;
- predictable ownership.

Package-level immutable constants are fine.

Shared immutable data may be acceptable when truly independent of instance config.

Shared caches require a deliberate concurrency and lifecycle design.

---

## 28. Copy mutable captured values

Strings, numbers, and booleans are values.

Slices, maps, pointers, and byte arrays may share mutable backing state.

Example config:

```go
type config struct {
    Words []string `json:"words"`
}
```

Before capturing:

```go
words := append(
    []string(nil),
    cfg.Words...,
)
```

For maps:

```go
replacements := make(
    map[string]string,
    len(cfg.Replacements),
)

for key, value := range cfg.Replacements {
    replacements[key] = value
}
```

Nested mutable structures need deeper copies.

The goal is clear ownership, not ritual cloning of immutable values.

---

## 29. Handle typed nil values

Go interfaces can contain typed nil pointers.

This value:

```go
var processor *processor
return processor, nil
```

produces a non-nil interface containing a nil pointer.

The registry rejects it through:

```go
pipeline.IsNilProcessor(processor)
```

Module tests should use the same helper when checking builder output.

For interface inputs such as `context.Context`, use project nil detection when direct invocation might receive typed nil.

Do not rely only on:

```go
ctx == nil
```

when a typed nil is possible.

---

## 30. Error wrapping style

Build errors in layers.

Inside the module:

```go
return nil, fmt.Errorf(
    "replace config: %w",
    err,
)
```

Registry layer adds:

```text
prepare module "replace-primary" of type "replace"
```

Pipeline layer later adds instance context to runtime failures.

Avoid:

```go
fmt.Errorf(
    "fish-audio-cli module replace module config error: %v",
    err,
)
```

Problems with that style:

- duplicates context;
- loses wrapping;
- becomes unreadable after outer layers add their own context;
- makes `errors.Is` and `errors.As` useless.

Use `%w` when propagating a cause.

---

## 31. Logging guidance

The core logging decorator already emits:

- module start;
- module completion;
- module failure;
- module interruption;
- instance name;
- module type;
- duration;
- character counts.

Do not add duplicate lifecycle logs inside `Process`.

A module-specific logger is not currently part of the preparer or processor contract.

Do not create a global logger to work around this.

Never log:

- API keys;
- authorization headers;
- complete provider credentials;
- full text outside the core privacy policy;
- unbounded remote error bodies.

When future modules need structured module-specific logs, extend the architecture deliberately rather than smuggling a logger through package globals.

---

## 32. Cancellation guidance

Check context before expensive work:

```go
if err := ctx.Err(); err != nil {
    return err
}
```

Pass it into:

- `http.NewRequestWithContext`;
- provider client methods;
- subprocess APIs supporting context;
- retry waits;
- blocking selects.

The pipeline checks context around the processor call, but it cannot preempt code that ignores context.

For CPU loops, check periodically when work can be long enough to matter.

Do not check on every byte merely to demonstrate spiritual devotion to cancellation.

---

## 33. Panic policy

Expected failures return errors.

Do not panic for:

- invalid config;
- missing files;
- malformed provider responses;
- unsupported values;
- cancellation;
- network failures;
- invalid transformed text.

The pipeline does not recover processor panics.

A panic may terminate the CLI and bypass fallback.

Use panic only for conditions that genuinely indicate an impossible programmer invariant, and even then prefer construction-time validation.

---

## 34. Resource lifecycle

Processors currently have no `Close`.

Do not leave them owning:

- open files;
- subprocesses;
- background goroutines;
- database handles requiring shutdown;
- temporary directories requiring cleanup;
- clients requiring explicit close.

If a real module requires mandatory cleanup, stop and design the full lifecycle before implementation.

Required questions include:

- cleanup after partial build;
- ownership after pipeline construction;
- reverse order;
- idempotency;
- concurrent processing and close;
- joined errors;
- exit-code behavior.

Do not add an optional `Close` interface to one module and hope the rest of the application develops telepathy.

---

## 35. Review checklist before registration

### Boundary

- Is the feature text-to-text?
- Does it belong before Fish synthesis?
- Does the core remain unaware of module-specific fields?
- Can the module be reordered meaningfully?

### Config

- Is config private?
- Is decoding strict?
- Are defaults explicit?
- Are ranges bounded?
- Are unknown fields rejected?
- Are paths resolved consistently?
- Are mutable values copied?

### Initialization

- Does `Prepare` only prepare?
- Is the builder non-nil?
- Is runtime construction deferred?
- Does the processor avoid mandatory-close resources?
- Are multiple instances independent?

### Processing

- Is context honored?
- Is output valid?
- Is assignment delayed until validation?
- Are errors wrapped?
- Is fallback left to the pipeline?
- Are panics avoided?
- Are side effects minimized?

### Security

- Are secrets outside JSON?
- Are reads and responses bounded?
- Is external disclosure documented?
- Are secrets absent from logs and errors?
- Are URLs and headers validated?

### Tests

- Unknown fields?
- Invalid values?
- Defaults?
- Success?
- Cancellation?
- Invalid output?
- Independent instances?
- Registry wiring?
- Pipeline fallback where relevant?
- Race detector?

### Documentation

- Registered type listed?
- Every config field documented?
- Complete example included?
- Privacy and external service behavior disclosed?
- Troubleshooting updated?

---

## 36. Definition of done

A module is complete when:

1. its boundary is clearly text-to-text;
2. type name is stable;
3. package is isolated;
4. config is private and strict;
5. semantic validation is complete;
6. preparation returns a non-nil instance builder;
7. builder returns a non-nil processor;
8. repeated instances are independent;
9. context is honored;
10. valid output is guaranteed on success;
11. errors are contextual and wrapped;
12. fallback remains in the pipeline;
13. secrets and text privacy are respected;
14. resource ownership fits the current lifecycle;
15. focused tests pass;
16. integration tests pass;
17. race detector passes;
18. vet passes;
19. config documentation is complete;
20. module documentation is updated;
21. example configuration is accurate;
22. final diff contains no unrelated changes.

---

## 37. Compact implementation template

Use this when starting a real module:

```go
package example

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/piqnyx/fish-audio-cli/internal/moduleconfig"
    "github.com/piqnyx/fish-audio-cli/internal/pipeline"
    "github.com/piqnyx/fish-audio-cli/internal/projectpath"
)

type config struct {
    // Module-owned fields.
}

type processor struct {
    // Prepared instance-owned state.
}

// Prepare validates one example module instance and returns its processor
// builder.
func Prepare(
    paths projectpath.Resolver,
    raw json.RawMessage,
) (
    pipeline.ProcessorBuilder,
    error,
) {
    var cfg config

    if err := moduleconfig.Decode(
        raw,
        &cfg,
    ); err != nil {
        return nil, fmt.Errorf(
            "example config: %w",
            err,
        )
    }

    // Validate cfg.
    // Resolve module-owned paths through paths.
    // Copy mutable prepared values.

    return func() (
        pipeline.Processor,
        error,
    ) {
        return &processor{
            // Instance-owned state.
        }, nil
    }, nil
}

// Process transforms current document text.
func (p *processor) Process(
    ctx context.Context,
    document *pipeline.Document,
) error {
    if err := ctx.Err(); err != nil {
        return err
    }

    // Compute result.
    // Validate result.
    // Assign document.Text only after successful validation.

    return nil
}
```

This template is intentionally incomplete.

The worked `replace` module shows the required validation and tests. Copying a skeleton without designing the module’s actual contracts merely produces well-formatted uncertainty.

---

## 38. Summary

A production module follows one disciplined path:

```text
choose a text-to-text boundary
    ↓
define a stable type
    ↓
create an isolated package
    ↓
strictly decode private instance config
    ↓
validate semantics and resolve paths
    ↓
return an instance-specific builder
    ↓
build a non-nil processor after all preparation succeeds
    ↓
honor context
    ↓
compute and validate replacement text
    ↓
assign only on success
    ↓
return errors to the pipeline
    ↓
test repeated instances, cancellation, and failure
    ↓
register and document the type
```

The goal is not merely to make a module work.

The goal is to add behavior without weakening the stable core, leaking one instance into another, hiding failure policy, or forcing future maintainers to reverse-engineer intent from a pile of technically valid Go.
