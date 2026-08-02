# Testing

> **Document status:** normative description of the current pre-release test strategy and continuous-integration contract.
>
> **Audience:** contributors, module authors, maintainers reviewing changes, and operators reproducing failures locally.
>
> **Scope:** this document describes required commands, CI behavior, test organization, unit and integration boundaries, filesystem and HTTP test techniques, concurrency rules, error and side-effect assertions, module test expectations, optional diagnostics, and compatibility constraints. Architecture boundaries are documented in [`architecture.md`](architecture.md); pipeline behavior in [`pipeline.md`](pipeline.md); module contracts in [`modules.md`](modules.md) and [`module-author-guide.md`](module-author-guide.md); configuration in [`configuration.md`](configuration.md); CLI behavior in [`cli.md`](cli.md); Fish integration in [`fish-audio.md`](fish-audio.md); logging in [`logging.md`](logging.md); secret and path behavior in [`secrets-and-paths.md`](secrets-and-paths.md); atomic output in [`output-files.md`](output-files.md); errors and exit codes in [`errors-and-exit-codes.md`](errors-and-exit-codes.md).

---

## 1. Purpose

The test suite protects the command’s observable contracts.

The project is not only a request serializer.

A complete invocation crosses:

```text
CLI parsing
    ↓
project-path derivation
    ↓
strict configuration loading
    ↓
module preparation and construction
    ↓
bounded text input
    ↓
ordered pipeline execution
    ↓
secret loading
    ↓
Fish HTTP request and retry behavior
    ↓
streaming response
    ↓
atomic output publication
    ↓
structured logging
```

Testing must therefore verify:

- returned values;
- wrapped error identity;
- process exit code;
- emitted request shape;
- file content;
- file permissions;
- cleanup;
- ordering;
- rollback;
- retry timing decisions;
- cancellation;
- logging fields;
- absence of unintended side effects.

A test that checks only `err != nil` is sometimes sufficient for a narrow guard.

It is not sufficient for a lifecycle boundary where the filesystem, network, or pipeline state can also change.

---

## 2. Toolchain

The module declares:

```text
Go 1.26.5
```

The CI workflow reads the version from:

```text
go.mod
```

The repository currently has no third-party module requirements in `go.mod`.

Tests use the Go standard library, including:

- `testing`;
- `net/http/httptest`;
- `context`;
- `errors`;
- `bytes`;
- `strings`;
- `sync`;
- `sync/atomic`;
- `os`;
- `path/filepath`;
- `encoding/json`;
- `log/slog`.

No external assertion or mocking framework is required.

---

## 3. Required local verification

Before committing a normal change, run:

```bash
clear

gofmt -w .
go vet ./...
go test -count=1 ./...
```

Before pushing a change that affects runtime behavior, concurrency, I/O, HTTP, lifecycle, or shared infrastructure, also run:

```bash
clear

go test -race -count=1 ./...
go build -trimpath -o /tmp/fish-audio-cli ./cmd/fish-audio-cli
```

The repository’s CI runs the complete set for every push and pull request targeting `main`.

---

## 4. Why `-count=1` is required

Plain:

```bash
clear

go test ./...
```

may reuse cached successful results.

The CI command is:

```bash
clear

go test -count=1 ./...
```

`-count=1` forces execution rather than accepting a cached pass.

This matters for tests involving:

- temporary files;
- filesystem permissions;
- local HTTP servers;
- cancellation timing;
- retry state;
- process-global variables;
- cleanup behavior.

Use the uncached form for final verification.

A cached test run remains useful for rapid development feedback, but it is not the project’s release-quality check.

---

## 5. Continuous integration

The workflow file is:

```text
.github/workflows/ci.yml
```

Workflow name:

```text
CI
```

Triggers:

```text
push to main
pull request targeting main
manual workflow dispatch
```

Repository permissions:

```text
contents: read
```

Runner:

```text
ubuntu-latest
```

Job timeout:

```text
15 minutes
```

---

## 6. CI step order

The CI job performs:

```text
checkout
    ↓
set up Go from go.mod
    ↓
check gofmt
    ↓
go vet ./...
    ↓
go test -count=1 ./...
    ↓
go test -race -count=1 ./...
    ↓
go build -trimpath
```

The exact build target is:

```text
./cmd/fish-audio-cli
```

The output binary is written to:

```text
/tmp/fish-audio-cli
```

A change is not CI-clean unless every step succeeds.

---

## 7. Formatting gate

CI checks formatting with:

```bash
clear

unformatted="$(gofmt -l .)"

if [[ -n "$unformatted" ]]; then
    printf 'Files are not formatted:\n%s\n' "$unformatted"
    exit 1
fi
```

CI does not rewrite files.

It reports every path that differs from `gofmt`.

Local correction:

```bash
clear

gofmt -w .
```

### 7.1 Generated or copied Go files

A generated or copied `.go` file is still subject to the same formatting gate.

Do not rely on an editor’s visual indentation.

Run `gofmt`.

---

## 8. Vet gate

CI runs:

```bash
clear

go vet ./...
```

`go vet` checks suspicious constructs in all module packages.

It is not a replacement for tests.

It can identify classes of problems such as:

- invalid formatting directives;
- unreachable or suspicious code patterns supported by analyzers;
- incorrect struct-tag syntax;
- misuse of selected standard-library APIs.

A change must pass both vet and tests.

---

## 9. Standard test gate

CI runs:

```bash
clear

go test -count=1 ./...
```

This command:

- discovers every package beneath the module;
- compiles package and test code;
- executes ordinary tests;
- disables result caching for that run;
- fails on compile errors, test failures, panics, or package setup failures.

It does not run sustained fuzzing.

It does not create a coverage threshold.

---

## 10. Race detector gate

CI runs:

```bash
clear

go test -race -count=1 ./...
```

The race detector is a required gate, not an occasional suggestion.

It is particularly relevant because tests and production code use:

- concurrent HTTP behavior;
- cancellation;
- retry coordination;
- shared counters in test servers;
- logging fan-out;
- parallel tests;
- independent CLI process assumptions.

A test passing without `-race` does not establish race safety.

---

## 11. Build gate

CI runs:

```bash
clear

go build -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli
```

The build gate verifies the actual executable target.

`-trimpath` removes local filesystem paths from recorded build metadata where supported.

A package-only test pass is not enough if the command cannot build.

---

## 12. Current CI limitations

The current workflow verifies one environment:

```text
ubuntu-latest
```

It does not currently provide a matrix for:

- macOS;
- Windows;
- multiple Go patch versions;
- multiple CPU architectures;
- alternative filesystems.

The workflow also has no mandatory:

- coverage percentage;
- benchmark threshold;
- fuzzing duration;
- live Fish API test;
- staticcheck step;
- golangci-lint step;
- mutation-testing step.

Do not describe these as existing gates.

They may be useful future additions, but adding them changes contributor and CI expectations.

---

## 13. No live Fish API in the normal suite

The standard tests must not require:

- a real Fish API key;
- external Fish availability;
- paid credits;
- Internet access to Fish;
- stable provider rate limits;
- a particular real voice.

HTTP behavior is tested with local test servers and custom transports.

This keeps tests:

- deterministic;
- fast;
- private;
- free of provider charges;
- runnable in CI;
- independent of remote service incidents.

---

## 14. Test placement

Tests live beside the code they exercise.

Examples:

```text
internal/fish/client.go
internal/fish/client_test.go

internal/pipeline/pipeline.go
internal/pipeline/pipeline_test.go

internal/output/atomic.go
internal/output/atomic_test.go
```

The project does not use a separate top-level `tests/` tree for the current Go suite.

Benefits:

- package-local helpers stay near the contract;
- internal implementation boundaries can be tested directly;
- changes and tests are reviewed together;
- package commands remain targeted.

---

## 15. Package choice

Current tests generally use the same package name as the implementation.

Example:

```go
package pipeline
```

rather than:

```go
package pipeline_test
```

This permits testing:

- unexported helpers;
- defensive internal branches;
- small package-level seams;
- exact lifecycle ordering.

Use same-package tests when the internal contract is intentionally part of the package’s maintenance surface.

A future external test package is still appropriate for a genuinely black-box public API.

Do not duplicate every test in both forms.

---

## 16. Test naming

Use descriptive Go test names:

```text
TestPipelineRunsProcessorsInOrder
TestLoadRejectsDuplicateFields
TestWriteAtomicPreservesExistingFileAfterFailure
TestClientReportsAPIError
```

A name should identify:

- subject;
- condition;
- expected behavior.

Avoid vague names such as:

```text
TestThing
TestError
TestWorks
```

Failure output should tell the maintainer what contract broke before he opens the test body.

---

## 17. Table-driven tests

Use table-driven tests when several values share one contract.

Example shape:

```go
tests := []struct {
    name  string
    input string
    want  string
}{
    {
        name:  "root URL",
        input: "https://api.example.com",
        want:  "https://api.example.com/v1/tts",
    },
    {
        name:  "trailing slash",
        input: "https://api.example.com/",
        want:  "https://api.example.com/v1/tts",
    },
}
```

Run each as a subtest:

```go
for _, test := range tests {
    test := test

    t.Run(test.name, func(t *testing.T) {
        t.Parallel()
        // assertions
    })
}
```

### 17.1 Map tables

Maps are useful when iteration order is irrelevant and cases are naturally keyed.

Slices are preferable when:

- order improves readability;
- the order itself matters;
- duplicate names are possible;
- deterministic display order is useful.

### 17.2 Per-case mutation

A table can store a mutator:

```go
map[string]func(*Config)
```

This is useful for validation tests that start from a complete known-good default.

---

## 18. Parallel tests

Use:

```go
t.Parallel()
```

when a test has no shared mutable process state.

Good candidates include tests using only:

- local values;
- independent `t.TempDir()` directories;
- independent `httptest.Server` instances;
- immutable package data;
- per-test buffers;
- per-test contexts.

Parallel execution reduces suite time and exposes accidental shared-state assumptions.

---

## 19. Tests that must not run in parallel

Do not call `t.Parallel()` when the test mutates process-global state such as:

- `os.Args`;
- current working directory;
- global registry maps;
- package-level hooks;
- shared environment variables;
- default global loggers;
- process signals.

Examples in the current suite include command tests that replace:

```go
os.Args
```

and logging tests that use:

```go
t.Chdir(...)
```

These tests intentionally run serially relative to other tests in the same package.

### 19.1 Parallel subtest trap

A parent that mutates global state cannot safely launch parallel subtests that depend on that state.

### 19.2 Environment variables

Use:

```go
t.Setenv(...)
```

for test-owned environment changes.

Do not combine it with parallel execution when the variable is process-global.

---

## 20. Cleanup

Prefer:

```go
t.Cleanup(...)
```

for state restoration tied to a test.

Example:

```go
previousArgs := os.Args

t.Cleanup(func() {
    os.Args = previousArgs
})
```

Use `defer` when lifecycle is simple and scoped to the current function.

Examples:

```go
defer server.Close()
defer cancel()
```

Cleanup must restore:

- process globals;
- opened local servers;
- temporary resources not managed by `testing`;
- package hooks;
- replaced function variables.

A passing assertion followed by leaked state is not a successful test.

---

## 21. Temporary directories

Use:

```go
directory := t.TempDir()
```

for filesystem tests.

`testing` removes the directory after the test completes.

Benefits:

- isolation;
- unique paths;
- no repository pollution;
- no dependence on `/tmp` naming;
- automatic cleanup.

Do not write test secrets, logs, configs, or output files into the repository tree.

---

## 22. Temporary directory permissions

Secret tests need a secure directory.

The suite explicitly applies:

```go
os.Chmod(directory, 0o700)
```

where the contract requires owner-only write safety.

Do not assume the initial mode from `t.TempDir()` satisfies every permission assertion on every supported platform.

### 22.1 Platform sensitivity

Permission-bit tests primarily exercise Unix-like semantics.

The current CI runner is Linux.

A future Windows CI matrix may need platform-specific expectations rather than pretending POSIX mode bits became universal by committee vote.

---

## 23. Filesystem assertions

A filesystem test should verify the full state transition.

Depending on the contract, check:

- path exists or does not exist;
- file type;
- content;
- permission bits;
- preserved old content;
- replaced symlink behavior;
- absence of temporary files;
- directory existence;
- error identity;
- cleanup errors.

Example checks:

```go
data, err := os.ReadFile(path)
info, err := os.Stat(path)
matches, err := filepath.Glob(pattern)
```

Do not stop after confirming the function returned an error.

---

## 24. Atomic output tests

The output suite must cover both sides of the rename boundary.

### Before rename

Verify that failure:

- preserves the existing destination;
- removes the temporary file when possible;
- returns the primary error;
- joins cleanup errors;
- does not publish partial bytes.

### After rename

Verify that directory persistence failure:

- returns an error;
- leaves the new destination published;
- does not attempt destructive rollback.

### Success

Verify:

- exact output content;
- final mode `0600`;
- replacement of existing regular file;
- replacement of destination symlink without changing its target;
- no stale temp file.

---

## 25. Secret tests

The secret loader is a security-sensitive filesystem boundary.

Required cases include:

- missing file creation;
- `ErrFileCreated`;
- mode `0600`;
- secure directory requirement;
- group/other writable directory rejection;
- existing file permission tightening;
- symlink rejection;
- non-regular file rejection;
- same-file race protection where injectable;
- exact byte limit;
- one byte above limit;
- empty value;
- LF;
- CRLF;
- multiple lines;
- surrounding whitespace;
- invalid UTF-8;
- close errors.

Use synthetic values such as:

```text
secret-value
test-key
```

Never use a real credential.

---

## 26. Configuration tests

Configuration testing has several distinct layers.

### Byte loading

Verify:

- exact maximum file size;
- maximum plus one;
- read errors;
- close errors.

### Strict JSON

Verify:

- malformed syntax;
- invalid UTF-8;
- duplicate keys;
- escaped duplicate keys;
- unknown fields;
- more than one JSON value;
- object requirements.

### Default overlay

Verify that:

- provided values replace defaults;
- omitted values retain defaults;
- arrays replace rather than merge;
- module entries do not inherit fields from default entries.

### Explicit null

Verify every field whose `null` behavior matters.

Do not assume Go’s zero value is equivalent to omitted JSON.

### Semantic validation

Verify:

- lower and upper bounds;
- accepted boundary values;
- cross-field relationships;
- supported enumerations;
- exact whitespace rules;
- duplicate module names;
- valid empty pipeline.

---

## 27. Boundary-value tests

For a bounded value, test at least:

```text
below minimum
minimum
representative valid value
maximum
above maximum
```

For byte limits, test:

```text
exactly max
max + 1
```

The suite already uses this pattern for:

- configuration bytes;
- input bytes;
- secret bytes;
- Fish error response bytes;
- numeric configuration ranges.

Boundary failures catch off-by-one errors that representative happy-path values do not.

---

## 28. Strict JSON tests

The strict JSON package tests syntax independently from configuration structs.

Important cases:

```text
one valid JSON value
empty input
malformed input
multiple top-level values
invalid UTF-8
duplicate root key
duplicate nested key
duplicate key inside array object
escaped duplicate key
same key in separate objects
unknown target field
nil decode target
```

Keeping these tests at the utility layer prevents every consumer from reinventing incomplete strictness tests.

Consumers still need integration tests proving they call the utility correctly.

---

## 29. Module configuration tests

`moduleconfig.Decode` must be tested as a strict object decoder.

Required cases:

- valid object;
- missing config;
- `null`;
- array;
- scalar string;
- unknown field;
- duplicate field;
- multiple values;
- invalid UTF-8;
- non-exact field-name case;
- nil target.

A module author must also test the module’s own semantic validation after decoding.

Strict JSON proves shape.

It does not prove domain correctness.

---

## 30. Module registry tests

Registry tests should use a local injected registry rather than mutating the production registry.

The current package exposes an internal `build` seam that accepts:

```go
map[string]preparer
```

This permits deterministic tests for:

- configured order;
- repeated module types;
- independent instance configs;
- default policy;
- per-instance policy override;
- empty pipeline;
- unknown type;
- nil preparer;
- preparation error;
- nil builder;
- builder error;
- typed-nil processor;
- prepare-all-before-build ordering;
- stop-after-build-failure behavior.

### 30.1 Do not mutate the global registry in parallel tests

The production registry is package-global.

Tests should inject an alternate map into an internal function rather than replace global contents.

---

## 31. Prepare-before-build invariant

The module system requires:

```text
prepare every module
    ↓
only then build processors
```

Tests must verify event order explicitly.

Example expected sequence:

```text
prepare first
prepare second
build first
build second
```

When second preparation fails:

```text
first builder must not run
```

This is a lifecycle contract, not merely an implementation detail.

---

## 32. Module processor tests

Every module should test:

- `Prepare` accepts valid config;
- `Prepare` rejects unknown fields;
- `Prepare` rejects malformed config;
- semantic validation;
- returned builder is non-nil;
- builder creates an independent processor;
- builder failure identity;
- processor behavior;
- context cancellation;
- invalid or nil arguments where relevant;
- text contract preservation;
- no unintended mutation after failure.

The `passthrough` module is the minimal reference.

Its core assertion is:

```text
output text equals input text
```

It must still validate its empty strict config object.

---

## 33. Pipeline test doubles

Pipeline tests use small handwritten processors.

Example shape:

```go
type testProcessor struct {
    name    string
    process func(*Document) error
}
```

This is preferred over a general mocking framework because the contract is small:

```text
Process(context.Context, *Document) error
```

A good test double makes the behavior visible in the test body.

Examples:

- append a suffix;
- return a sentinel error;
- cancel a context;
- produce invalid text;
- record whether it was called.

---

## 34. Pipeline tests

Required pipeline behavior includes:

- processors run in configured order;
- output of one becomes input of the next;
- original text remains available;
- failure rolls back mutation;
- `use_previous`;
- `use_original`;
- `skip`;
- `abort`;
- cancellation before first step;
- cancellation returned by processor;
- cancellation observed after processor returns nil;
- invalid UTF-8 output;
- blank output;
- invalid output under each policy;
- report outcome;
- step count;
- input/output rune counts;
- durations;
- duplicate names;
- nil and typed-nil processors;
- nil and typed-nil context;
- empty pipeline.

### 34.1 Error identity

Use a sentinel:

```go
expectedErr := errors.New("processor exploded")
```

Then assert:

```go
errors.Is(err, expectedErr)
```

Do not rely only on substring matching.

### 34.2 Side effects

Also assert:

- later processors were or were not called;
- document text was restored;
- report contains the failing step;
- outcome is correct.

---

## 35. Typed-nil tests

Go interfaces can be non-nil while containing a nil pointer.

The suite explicitly tests typed-nil values for contracts such as:

- `context.Context`;
- `io.Writer`;
- processors;
- HTTP dependencies.

Use a typed value whose methods panic if invoked.

Then verify validation rejects it before method dispatch.

This proves the nil guard handles interface representation rather than only:

```go
value == nil
```

Typed-nil tests are especially important at public or pluggable interface boundaries.

---

## 36. Context tests

Cancellation tests should cover timing.

### Before work

Cancel before calling the function.

Assert no processor or HTTP request runs.

### During work

The fake processor or transport cancels the context.

Assert:

- wrapped `context.Canceled`;
- rollback;
- no later work;
- correct report outcome.

### After a processor returns nil

A processor can cancel and return nil.

The pipeline must still detect cancellation before committing the step output.

### During retry delay

Use controllable timing or injected delay behavior where available.

Do not sleep for large real durations.

---

## 37. Avoid fragile sleeps

Tests should not depend on arbitrary delays such as:

```go
time.Sleep(500 * time.Millisecond)
```

Prefer:

- channels;
- context cancellation;
- injected delay functions;
- deterministic `Retry-After` values;
- atomic counters;
- explicit synchronization.

A small timeout can be used as a safety bound around a channel protocol.

It should not be the mechanism that makes the expected ordering happen.

---

## 38. Fish HTTP tests

Use:

```go
httptest.NewServer(...)
```

for end-to-end HTTP client behavior.

The handler can verify:

- method;
- path;
- headers;
- authorization;
- model;
- JSON body;
- retry count;
- response status;
- response body.

It can return synthetic audio bytes such as:

```text
fake-audio
fake-opus-audio
```

No actual audio decoder is needed to test transport and streaming contracts.

---

## 39. Custom HTTP transports

For lower-level client behavior, a handwritten transport is useful:

```go
type roundTripFunc func(
    *http.Request,
) (*http.Response, error)
```

Implement:

```go
func (f roundTripFunc) RoundTrip(
    request *http.Request,
) (*http.Response, error) {
    return f(request)
}
```

This permits deterministic tests for:

- transport errors;
- response body close;
- typed-nil client dependencies;
- malformed responses;
- request cancellation;
- no real socket.

Use `httptest.Server` when HTTP serialization and server behavior are part of the contract.

Use a custom transport when the transport boundary itself is the subject.

---

## 40. HTTP request assertions

A Fish client success test should verify at least:

```text
POST
/v1/tts
Authorization: Bearer <test key>
model header
JSON request body
response bytes copied to writer
```

Do not verify only the returned audio bytes.

A client can produce bytes while sending the wrong model, path, or authorization header.

---

## 41. Typed Fish API errors

Non-2xx tests should assert both:

```go
errors.As(err, &apiErr)
```

and:

```go
errors.Is(err, fish.ErrAuthentication)
```

as appropriate.

Also verify fields:

```text
HTTPStatusCode
HTTPStatus
APIStatus
Message
```

Test bodies:

- valid Fish JSON error;
- plain text;
- malformed JSON;
- empty body;
- oversized body;
- body read error.

This protects both human diagnostics and stable programmatic categories.

---

## 42. Retry tests

Retry tests must verify:

- which statuses retry;
- maximum attempts;
- server-error option;
- backoff progression;
- maximum delay;
- `Retry-After`;
- cancellation during wait;
- no retry after response streaming begins;
- body closure between attempts;
- final typed error.

Use an atomic request counter when the test server can be called concurrently or race detection is enabled.

Example:

```go
var attempts atomic.Int32
attempts.Add(1)
```

---

## 43. Streaming failure tests

A reader can return data and an error in the same call.

A writer can accept partial data and then fail.

Use custom readers and writers to verify:

- partial bytes are handled correctly;
- error identity is preserved;
- no retry corrupts output;
- response body closes;
- caller sees the final error.

These cases are more realistic than a writer that always fails before accepting one byte.

---

## 44. Command integration tests

The command package tests:

```go
run()
```

rather than launching an external subprocess for every case.

A full local integration test builds:

- temporary config;
- temporary secret;
- local `httptest.Server`;
- output path;
- `os.Args`;
- invocation through `run()`.

Then it verifies:

- exit code;
- HTTP request;
- processed text;
- selected format;
- output content;
- temp cleanup.

This covers the complete in-process command wiring without a real Fish service.

---

## 45. Global state in command tests

`run()` reads:

```go
os.Args
os.Stdin
os.Stdout
os.Stderr
```

Current command tests that replace `os.Args` must:

- save the previous slice;
- restore it with cleanup;
- avoid `t.Parallel()`.

A future injectable command runner could reduce global-state coupling.

Until then, tests must respect the current boundary.

---

## 46. Exit-code tests

Command integration tests should cover each code where practical.

### `0`

- help;
- full synthesis success;
- recoverable module failure;
- stopped pipeline with `skip`.

### `1`

Bootstrap failures require an injectable seam because normal `os.Stderr` and secure randomness rarely fail predictably.

### `2`

- invalid option;
- config failure;
- logger open failure;
- module initialization failure;
- invalid input.

### `3`

- pipeline abort;
- pipeline cancellation;
- request creation failure;
- missing secret;
- invalid secret;
- Fish client validation.

### `4`

- Fish transport;
- typed API error;
- stream failure;
- output create/write/sync/rename/persistence failure.

A code assertion should be paired with side-effect assertions.

---

## 47. Ordering tests

Startup ordering is part of the contract.

Examples:

- unsupported module fails before secret creation;
- invalid input fails before secret creation;
- every module is prepared before any processor is built;
- Fish request is built before secret loading;
- output temp is not published before stream success.

Tests should verify absence of later-stage artifacts.

Example:

```text
Fish secret file does not exist
output file does not exist
HTTP server request count is zero
```

This proves the command stopped at the intended boundary.

---

## 48. Logging tests

Logging tests use buffers and temporary files.

Verify:

- text and JSON format;
- level parsing;
- request fields;
- stderr output;
- persistent file output;
- fan-out behavior;
- short writes;
- joined destination errors;
- default path;
- relative path;
- absolute path;
- append behavior;
- mode `0640`;
- directory creation;
- typed-nil writer rejection;
- close-error preservation.

### 48.1 JSON logs

Decode into:

```go
map[string]any
```

and assert keys.

Do not compare the entire serialized line when time and field ordering are not the intended contract.

### 48.2 Text logs

Use targeted containment assertions for:

- message;
- request ID;
- path;
- error text.

---

## 49. Fan-out writer tests

A fan-out writer must be tested with writers that can:

- succeed;
- fail;
- short-write without error;
- write different byte counts;
- be typed nil.

Verify:

- every destination is attempted;
- minimum byte count is returned;
- `io.ErrShortWrite` is generated;
- multiple errors are joined;
- zero destinations are rejected.

Race testing is particularly valuable when any future fan-out implementation adds concurrency.

The current contract does not require concurrent destination writes.

---

## 50. Path resolver tests

Project-path tests should verify:

- blank config rejection;
- relative config becomes absolute;
- exact parent basename `config`;
- other parent names;
- absolute configured path;
- relative configured path;
- lexical cleaning;
- blank resolved path;
- uninitialized resolver;
- zero-value resolver with absolute path;
- `..` escape;
- symlink path remains lexical.

Tests should not assume:

- executable-relative paths;
- realpath behavior;
- home expansion;
- environment expansion.

---

## 51. CLI parser tests

Parser tests should cover:

- defaults;
- `--help`;
- unknown flag;
- missing flag value;
- unexpected positional args;
- missing `--output`;
- supported formats;
- case normalization;
- `ogg` normalization to `opus`;
- unsupported format;
- exact preservation of output path.

The internal flag set discards default parser output.

Tests should verify returned errors and options, not rely on stderr produced by the standard flag package.

---

## 52. Input tests

Input tests should cover:

- `--text` precedence;
- stdin fallback;
- exact empty argument;
- whitespace-only argument;
- whitespace-only stdin;
- valid UTF-8;
- invalid UTF-8;
- exact byte limit;
- limit plus one;
- reader error;
- nil stdin;
- Unicode rune text whose bytes differ from rune count.

The limit is bytes.

The log character count is runes.

Tests should not confuse the two.

---

## 53. Error assertion style

Prefer:

```go
if !errors.Is(err, expectedErr) {
    t.Fatalf(
        "error = %v, want wrapped error %v",
        err,
        expectedErr,
    )
}
```

For typed errors:

```go
var apiErr *APIError

if !errors.As(err, &apiErr) {
    t.Fatalf(
        "error = %v, want APIError",
        err,
    )
}
```

Use string containment only for contextual text that has no typed identity.

---

## 54. Joined error tests

When code uses `errors.Join`, assert every cause.

Example:

```go
if !errors.Is(err, primaryErr) {
    t.Fatal("primary error not preserved")
}

if !errors.Is(err, closeErr) {
    t.Fatal("close error not preserved")
}
```

Also assert relevant context strings when operator diagnostics matter.

Do not compare a joined error’s complete string unless exact formatting is deliberately contractual.

---

## 55. Failure messages

Use failure messages that show:

```text
actual
expected
context
```

Example:

```go
t.Fatalf(
    "output permissions = %#o, want %#o",
    actual,
    expected,
)
```

For paths:

```go
t.Fatalf(
    "log path = %q, want %q",
    actual,
    expected,
)
```

Quoted values help expose whitespace and Unicode differences.

---

## 56. Helpers

A test helper should:

- call `t.Helper()`;
- reduce setup duplication;
- fail with useful context;
- avoid hiding the behavior under test;
- return the value needed by the caller.

Examples:

```text
writeTestConfig
loadTestConfig
newTestPipeline
secureTempDir
validClientOptions
validSynthesisRequest
```

### 56.1 Avoid giant test frameworks

A helper that creates config, server, client, request, output, and assertions for every test can obscure which contract each case exercises.

Prefer small composable helpers.

---

## 57. Test fixtures

Use inline JSON for small focused cases.

Example:

```go
json.RawMessage(`{"enabled":true}`)
```

Use a file fixture only when:

- content is large;
- formatting itself matters;
- binary bytes are required;
- reuse clearly improves readability.

A fixture must not contain:

- live secrets;
- user data;
- copyrighted provider samples without permission;
- environment-specific absolute paths.

---

## 58. Golden files

The project does not currently require golden-file testing for its core contracts.

Golden files can be useful for large stable serialized outputs.

They are less useful when the output includes:

- timestamps;
- request IDs;
- random temp names;
- platform-specific paths;
- map ordering;
- changing provider error messages.

Prefer semantic assertions for structured logs and JSON.

If golden files are introduced, updates must be explicit and reviewed.

---

## 59. Test secrets

Use fake credentials:

```text
test-key
secret-key
invalid-key
```

Never read a developer’s real default secret file.

Never make tests depend on:

```text
secrets/fish-api-key
```

in the repository.

A test must create its own secret inside `t.TempDir()`.

### 59.1 Secret logging

Test failures should not print real credentials.

Synthetic test values are safe to include in assertion context.

---

## 60. Test output files

Use a temporary destination:

```go
outputPath := filepath.Join(
    t.TempDir(),
    "speech.opus",
)
```

Verify:

- bytes;
- permissions;
- publication;
- cleanup.

Do not write:

```text
speech.opus
```

into the repository working directory during tests.

---

## 61. No external clock assumptions

Retry and duration code can become flaky if tests assert exact wall-clock time.

Prefer assertions such as:

```text
duration is non-negative
attempt count equals expected
delay sequence equals injected values
```

Avoid exact millisecond equality unless time is injected.

Pipeline report durations come from the real clock.

Tests should verify sensible presence rather than scheduler precision.

---

## 62. Randomness

Request IDs use secure randomness.

A shape test can verify:

- length;
- hexadecimal form;
- independent calls usually differ.

Do not make a probabilistic uniqueness test that can theoretically fail under a valid implementation without a controlled source.

For deterministic failure testing, randomness should be injectable rather than depending on system entropy failure.

---

## 63. Request ID tests

Verify:

```text
32 lowercase hexadecimal characters
```

because the implementation encodes 16 random bytes.

Logging integration should verify the same ID appears across correlated records where practical.

Do not assert a fixed generated value without an injected random reader.

---

## 64. Permission assertions

Use:

```go
info.Mode().Perm()
```

and compare with:

```go
os.FileMode(0o600)
os.FileMode(0o640)
```

Permission tests should distinguish:

- requested create mode;
- explicit chmod enforcement;
- existing directory mode left unchanged;
- umask effects where applicable.

The secret file and output file require `0600`.

The log file requires `0640`.

---

## 65. Symlink tests

Symlink behavior is security-relevant.

Required distinctions:

### Secret leaf

Must be rejected.

### Output leaf

Must be replaced without following the target.

### Log leaf

Currently uses ordinary open behavior and does not receive secret-style hardening.

### Parent components

Ordinary path resolution can follow them unless the subsystem explicitly anchors operations.

Skip symlink tests only on platforms where symlink creation is unavailable, and document the reason.

Do not silently convert a meaningful security test into a no-op.

---

## 66. File-type tests

Where practical, test rejection or behavior for:

- directory;
- regular file;
- symlink;
- FIFO;
- socket;
- device.

Some types require platform-specific setup and privileges.

Use build tags or explicit platform checks when necessary.

Do not make the ordinary Linux CI require privileged device creation.

---

## 67. Network isolation

`httptest.NewServer` listens on loopback.

The suite does not need Internet access, but it does need ordinary local networking.

A highly restricted sandbox that blocks loopback listeners can fail HTTP integration tests.

Such an environment differs from the current CI contract.

---

## 68. Port allocation

Never hard-code a local port in tests.

Use:

```go
httptest.NewServer
```

or:

```go
net.Listen("tcp", "127.0.0.1:0")
```

The operating system selects an available ephemeral port.

Hard-coded ports create collisions under parallel tests and shared CI runners.

---

## 69. HTTP server assertions

Inside an `httptest` handler, use:

```go
t.Errorf(...)
```

for request mismatches when the handler can still return a controlled response.

Use synchronization when handler results are read by the main test goroutine.

The server’s close operation should occur with `defer` or cleanup.

A handler should not call `t.Fatal` from an unrelated goroutine.

---

## 70. Shared counters

Use:

```go
sync/atomic
```

or a mutex when request handlers and test goroutines share counters.

The race detector should remain clean.

Do not assume the test server handler executes on the same goroutine as the test.

---

## 71. Response-body lifecycle

HTTP tests should verify response bodies close on:

- success;
- API error;
- retry;
- body read failure;
- cancellation.

Small custom `io.ReadCloser` implementations can record close state.

Resource lifecycle is part of the client contract.

---

## 72. Short read and short write tests

I/O code must handle legal combinations such as:

```text
n > 0 and err != nil
n < len(p) and err == nil
```

Use custom readers/writers to generate these cases.

Verify:

- consumed bytes;
- returned count;
- synthesized `io.ErrShortWrite`;
- error joining;
- cleanup.

Real disks and buffers rarely produce every edge condition on demand.

---

## 73. Nil dependency tests

For every interface dependency, consider:

```text
untyped nil
typed nil
valid implementation
implementation returning error
```

Examples:

- writer;
- context;
- processor;
- closer;
- HTTP client or transport;
- module preparer;
- processor builder.

Nil guards should fail predictably rather than panic later.

---

## 74. Panic tests

Use panic assertions only when panic is the intended contract.

The current project generally returns errors for invalid inputs.

A test double may panic if a typed-nil method is called, but the expected test outcome is that validation prevents the panic.

Do not normalize unexpected panics as ordinary error behavior.

---

## 75. Race-detector-friendly tests

To remain race-clean:

- do not mutate shared maps without locks;
- do not reuse buffers across parallel subtests;
- use per-test servers;
- protect handler state;
- restore globals serially;
- avoid package hooks shared by parallel tests;
- do not call `t.Parallel()` after mutating globals;
- keep cleanup ownership clear.

The race detector reports test-code races too.

That is useful.

A broken test harness can hide or invent production failures.

---

## 76. Test ordering independence

Tests must not require another test to run first.

Each test must create its own:

- config;
- secret;
- output;
- server;
- logger;
- pipeline;
- registry.

Go does not guarantee source-order execution.

Parallel tests make accidental dependencies fail more visibly.

---

## 77. Current-working-directory tests

Use:

```go
t.Chdir(directory)
```

when testing cwd-relative behavior.

Do not manually call `os.Chdir` without restoration.

A cwd-changing test must not run in parallel with tests in the same process that depend on cwd.

Use absolute paths in unrelated tests to avoid accidental coupling.

---

## 78. Process signal tests

Signal integration tests are delicate because they affect the test process.

Prefer testing cancellation through `context.Context` at package layers.

A true subprocess test is more appropriate for verifying actual `SIGINT` or `SIGTERM` command behavior.

Do not send a terminating signal to the `go test` process and hope the scheduler understands your artistic intent.

---

## 79. Subprocess tests

Use an external subprocess only when the behavior cannot be tested safely in-process.

Examples:

- real signal exit behavior;
- stdout/stderr descriptor behavior;
- shell-visible exit status;
- panic output;
- working-directory isolation;
- environment inherited at process start.

The normal command integration suite uses `run()` because it is faster and easier to inspect.

A subprocess suite should build or use a test binary deterministically and clean all artifacts.

---

## 80. Executable test pattern

A common Go subprocess pattern uses a test helper process selected by environment variable.

If introduced, it must:

- avoid recursive test execution;
- isolate environment;
- capture stdout and stderr separately;
- assert exit status;
- use temp config, secret, log, and output paths;
- never call the live Fish API.

This pattern is not currently required for the ordinary suite.

---

## 81. Local targeted tests

Run one package:

```bash
clear

go test -count=1 ./internal/fish
```

Run one named test:

```bash
clear

go test -count=1 -v \
  ./internal/fish \
  -run '^TestClientReportsAPIError$'
```

Run command integration:

```bash
clear

go test -count=1 -v \
  ./cmd/fish-audio-cli \
  -run '^TestRunSynthesisEndToEnd$'
```

Run output tests:

```bash
clear

go test -count=1 -v \
  ./internal/output
```

---

## 82. Regular-expression selection

`-run` accepts a regular expression over slash-separated test and subtest names.

Example:

```bash
clear

go test -count=1 -v \
  ./internal/config \
  -run 'TestLoadRejects'
```

Anchor an exact test:

```text
^TestName$
```

Targeted runs accelerate development.

They do not replace the full suite before push.

---

## 83. Verbose mode

Use:

```bash
clear

go test -count=1 -v ./...
```

when diagnosing:

- which test hangs;
- subtest ordering;
- skipped tests;
- package progression;
- server or filesystem failure context.

Verbose output is diagnostic.

CI currently uses normal non-verbose output and prints details when a test fails.

---

## 84. Shuffle mode

Optional local order check:

```bash
clear

go test -count=1 -shuffle=on ./...
```

This can expose hidden ordering dependencies.

Shuffle is not currently a CI gate.

When a failure reports a seed, rerun with that seed:

```bash
clear

go test -count=1 -shuffle=123456789 ./...
```

Do not dismiss an order-dependent failure because the default order passes.

---

## 85. Coverage

Optional local coverage:

```bash
clear

go test -count=1 -cover ./...
```

Detailed profile:

```bash
clear

go test -count=1 \
  -covermode=atomic \
  -coverprofile=/tmp/fish-audio-cli.cover \
  ./...

go tool cover \
  -func=/tmp/fish-audio-cli.cover
```

HTML view:

```bash
clear

go tool cover \
  -html=/tmp/fish-audio-cli.cover \
  -o /tmp/fish-audio-cli-cover.html
```

### 85.1 No current threshold

CI does not enforce a minimum percentage.

Do not optimize for a number by writing assertions that execute lines without verifying behavior.

Coverage is a map of executed code, not proof of correctness.

---

## 86. Coverage priorities

Prioritize behavior coverage for:

- exit-code branches;
- retry decisions;
- cancellation;
- output rename boundary;
- cleanup errors;
- secret hardening;
- strict JSON;
- module lifecycle;
- typed-nil guards;
- joined errors.

A defensive failure branch can be more valuable than several trivial getter lines.

---

## 87. Fuzzing

Go fuzzing can be valuable for parsers and normalizers.

Candidate targets:

- strict JSON validation;
- module config decoding;
- text contract;
- secret normalization;
- URL endpoint resolution;
- CLI option normalization where argument generation is constrained.

Example future command:

```bash
clear

go test ./internal/strictjson \
  -run '^$' \
  -fuzz '^FuzzValidate$' \
  -fuzztime=30s
```

No fuzz-duration stage is currently configured in CI.

A fuzz target must preserve useful crash inputs in the standard corpus when a bug is found.

---

## 88. Fuzz properties

Useful properties include:

- function never panics for arbitrary bytes;
- accepted JSON contains exactly one valid value;
- duplicate keys are never accepted;
- normalization output is valid UTF-8;
- secret success never contains CR or LF;
- endpoint success always has HTTP or HTTPS and `/v1/tts`;
- rejected input returns an error rather than hanging.

A fuzz target should assert invariants, not merely call the function.

---

## 89. Benchmarks

Benchmarks are optional and are not a CI gate.

Potential targets:

- bounded reads;
- strict JSON validation;
- large configuration decode;
- pipeline with many lightweight modules;
- logging fan-out;
- request serialization.

Run benchmarks with:

```bash
clear

go test -run '^$' -bench . -benchmem ./...
```

Do not add performance thresholds without stable hardware and a deliberate benchmark environment.

---

## 90. Test caching diagnostics

List package test cache behavior with normal Go tooling when investigating suspicious reuse.

The project’s final gate already avoids successful-result reuse through:

```text
-count=1
```

Deleting the entire Go cache is rarely the first appropriate response.

Prefer reproducing the exact package and command.

---

## 91. Repetition

To expose intermittent failures:

```bash
clear

go test -count=100 ./internal/pipeline
```

With race detection:

```bash
clear

go test -race -count=20 ./internal/fish
```

Use a focused package before repeating the full repository.

A high repeat count is a diagnostic tool, not a substitute for deterministic synchronization.

---

## 92. Timeout

Go’s test command has a package timeout.

For diagnosis:

```bash
clear

go test -count=1 -timeout=2m ./...
```

Do not solve a deadlock by increasing timeout indefinitely.

A test that requires a long real wait should usually gain an injectable clock or delay.

CI also has a whole-job timeout of 15 minutes.

---

## 93. CPU parallelism

To diagnose scheduling-sensitive behavior:

```bash
clear

go test -race -count=20 -cpu=1,2,4 ./...
```

This is optional and not a CI gate.

The suite must still pass under the default CI scheduler.

---

## 94. JSON and Unicode test data

Include cases with:

- Cyrillic text;
- emoji;
- multibyte UTF-8;
- invalid byte sequences;
- escaped JSON field names;
- Unicode whitespace.

The project’s input and logging contracts distinguish:

- bytes;
- runes;
- whitespace;
- UTF-8 validity.

ASCII-only tests cannot protect those distinctions.

---

## 95. Character and byte counts

For input limits:

```text
count bytes
```

For log character fields and pipeline report counts:

```text
count UTF-8 runes
```

A test should use text where these differ.

Example:

```text
Привет
```

or:

```text
🚀
```

Do not assert byte length as character count.

---

## 96. Testing log privacy

When `logging.logText` is false, verify text values are absent while counts remain.

When true, verify only intended top-level fields are added.

Also verify module lifecycle logs never include full intermediate text.

Use clearly identifiable synthetic text such as:

```text
Секретный текст
```

Then assert it is absent from the relevant buffer.

---

## 97. Testing error privacy

Error tests should verify that:

- API key is not included in logs;
- secret path may be included where documented;
- bounded Fish error body is included only within its limit;
- full input text is not accidentally added to generic errors;
- request authorization is not dumped.

Do not create tests that normalize leaking a credential as expected output.

---

## 98. Testing cleanup error paths

Cleanup failures are difficult to produce with ordinary files.

Use small injectable seams or controlled test doubles.

Required preservation pattern:

```text
primary error
+
cleanup error
```

The test should verify both with `errors.Is`.

Do not add broad dependency injection to every package merely to simulate one close error.

Use the narrowest seam that exposes the lifecycle boundary.

---

## 99. Testing close failures

A small closer is sufficient:

```go
type failingCloser struct {
    err error
}

func (c failingCloser) Close() error {
    return c.err
}
```

Use it to test:

- log close reporting;
- error joining;
- deferred diagnostic behavior.

For real `*os.File` close paths, use injection rather than platform-specific filesystem tricks when possible.

---

## 100. Testing partial writers

A custom writer can:

- accept a configured number of bytes;
- return a sentinel error;
- expose the accepted prefix.

This tests:

- propagation;
- partial output behavior;
- cleanup;
- no retry after streaming begins.

Use `bytes.Buffer` inside the writer for inspection.

---

## 101. Testing readers that return data and error

A custom reader can return:

```text
n > 0
err != nil
```

This is legal Go I/O behavior.

The client must not discard accepted bytes or lose the final error.

Such tests protect response streaming and bounded reads.

---

## 102. Testing HTTP body closure

Use a body type with:

```text
closed channel
atomic flag
```

Assert close occurs on:

- success;
- non-2xx;
- retry;
- copy failure.

A response body leak may not fail a short unit test without explicit observation.

---

## 103. Testing retry body reuse

Each retry must create a fresh HTTP request body.

A test server should decode every attempt.

Do not assume a previously consumed reader rewinds itself.

This is especially important when request serialization moves or streaming request bodies are introduced.

---

## 104. Testing `Retry-After`

Test supported forms separately:

```text
integer seconds
HTTP date
missing header
invalid header
delay above configured maximum
```

Use a controlled clock or delay seam where practical.

Do not make the suite wait until a real future HTTP date.

---

## 105. Testing no retry after output begins

The response writer is not rewindable.

A streaming failure after some bytes must return immediately.

Test:

- first attempt returns 2xx;
- some bytes reach writer;
- writer or body then fails;
- server attempt count remains one.

This prevents duplicated or concatenated audio.

---

## 106. Build-only packages

`go test ./...` compiles packages even when they contain no test functions.

A package reported as:

```text
[no test files]
```

still received a compilation check.

That is useful but not equivalent to behavioral coverage.

Add tests when a package owns a contract rather than relying on compile-only status.

---

## 107. Documentation changes

A documentation-only change currently still passes through:

- formatting check;
- vet;
- full tests;
- race tests;
- build.

The workflow does not skip Go checks based on changed paths.

This keeps documentation commits from accidentally landing beside unnoticed broken code, at the cost of additional CI time.

---

## 108. Test changes without production changes

A test-only change must pass the same full CI.

Review test-only changes for:

- weakened assertions;
- removed failure cases;
- accidental skips;
- excessive timing;
- global-state races;
- secret leakage;
- live-network dependencies.

A green test suite can be made meaningless more easily than software teams like to admit.

---

## 109. Skipped tests

Use:

```go
t.Skip(...)
```

only for an explicit environment limitation.

A skip message must explain:

- required capability;
- why it is unavailable;
- which contract is not being verified.

Do not skip a failing test merely because the failure is inconvenient.

CI output should make skipped security or filesystem behavior visible.

---

## 110. Build tags

Build tags may be appropriate for:

- Unix-only filesystem objects;
- Windows-specific rename behavior;
- integration tests requiring an external dependency;
- privileged tests.

A tagged test must not silently remove the only coverage of a cross-platform contract.

Document the command required to run the tagged suite.

No extra tagged suite is currently part of CI.

---

## 111. Live integration tests

A future live Fish suite must be opt-in.

It should require:

- explicit build tag or environment flag;
- dedicated test credential;
- dedicated account budget;
- non-sensitive text;
- unique output;
- conservative retry;
- cleanup;
- clear provider-cost warning.

It must not run in ordinary CI by default.

The standard suite remains authoritative for deterministic repository validation.

---

## 112. Live test separation

Do not make ordinary tests conditionally contact Fish when an environment variable happens to exist.

A developer’s shell can contain credentials unexpectedly.

Use a deliberate command such as:

```bash
clear

go test \
  -tags=live_fish \
  ./internal/fish
```

and an explicit documented credential source if such a suite is introduced.

---

## 113. Reproducing CI locally

Run:

```bash
clear

set -euo pipefail

unformatted="$(gofmt -l .)"

if [[ -n "$unformatted" ]]; then
    printf 'Files are not formatted:\n%s\n' "$unformatted"
    exit 1
fi

go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
go build -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli
```

This mirrors the current CI command order.

It does not reproduce:

- the exact GitHub-hosted runner image;
- clean checkout state;
- action cache state;
- runner kernel and filesystem;
- workflow timeout enforcement.

---

## 114. Clean-checkout verification

Before a release or final audit, test from a clean checkout or clean worktree.

This detects hidden dependencies on:

- ignored config;
- local secret files;
- stale generated files;
- untracked assets;
- previously built binaries;
- environment-specific paths.

The normal suite must pass without:

```text
config/config.json
secrets/fish-api-key
bin/fish-audio-cli
logs/
```

present in the repository checkout.

Tests create their own replacements.

---

## 115. Environment minimization

A useful diagnostic run uses a minimal environment.

Preserve only what Go and the shell require.

The project should not depend on:

- a developer Fish key;
- current username;
- home directory contents;
- editor settings;
- repository-local runtime config;
- a running local proxy.

When a test needs environment state, set it explicitly.

---

## 116. Time zone and locale

Tests should not depend on:

- local time zone;
- locale-specific error strings;
- localized month names;
- decimal separators.

HTTP date parsing tests should use fixed RFC-compliant strings.

Filesystem error text from the operating system can vary.

Assert wrapped identity or stable project context rather than complete platform error sentences.

---

## 117. Platform-dependent errors

Avoid exact equality against an entire `os.PathError.Error()` string.

Prefer:

```go
errors.Is(err, os.ErrNotExist)
errors.Is(err, os.ErrPermission)
```

and project context containment.

Paths and system messages vary across operating systems.

---

## 118. Testing file modes under umask

Creation modes can be reduced by umask.

The project explicitly chmods security-critical files after opening where the contract requires an exact mode.

Tests should verify the final mode.

For directories whose existing mode is intentionally not changed, tests should verify accepted or rejected write bits rather than assume forced `0700`.

---

## 119. Testing append behavior

Logging tests should pre-create a file with known content.

After writing a new record, assert both:

```text
old content remains
new content appended
```

Also assert final mode `0640`.

A test that verifies only the new message can miss accidental truncation.

---

## 120. Testing output replacement

Pre-create output with:

```text
old-audio
```

Then:

- successful write should replace it;
- pre-rename failure should preserve it.

Verify exact bytes and mode.

The two cases protect opposite sides of the atomic contract.

---

## 121. Testing output symlink replacement

Create:

```text
output -> target
```

Write known protected bytes to `target`.

After success, assert:

- target bytes unchanged;
- output is no longer a symlink;
- output is a regular file;
- output contains new bytes;
- output mode is `0600`.

This verifies the final leaf is replaced rather than followed.

---

## 122. Testing stale temp cleanup

After a controlled failure, glob:

```text
.<baseName>.*.tmp
```

and require zero matches when cleanup succeeds.

When testing cleanup failure, require the joined error and document whether a temp path can remain.

Do not rely only on `os.Stat(finalPath)`.

---

## 123. Testing command stage boundaries

The command’s ordering should be tested through artifacts.

Examples:

### Module failure before secret loading

- exit `2`;
- secret file absent;
- output absent;
- no Fish request.

### Invalid input before secret loading

- exit `2`;
- secret file absent;
- output absent.

### Missing secret after pipeline

- exit `3`;
- empty secret file exists;
- output absent;
- no Fish request.

### Fish failure

- exit `4`;
- secret exists;
- server request may exist;
- output state follows publication phase.

---

## 124. Testing help

Help test should verify:

- exit `0`;
- usage written to stdout;
- no config required;
- no output created;
- no persistent log file opened.

Because request ID generation happens before parsing, a fully injected startup test can separately verify its failure blocks help with exit `1`.

---

## 125. Testing log text option

For disabled text logging:

```text
input_chars present
input_text absent
```

For enabled text logging:

```text
input_chars present
input_text present
```

Repeat for processed output.

Do not require module logs to contain intermediate text.

---

## 126. Testing character counts

Use:

```go
utf8.RuneCountInString
```

as the expected calculation.

Test strings with multibyte runes.

Do not use:

```go
len(text)
```

for expected character fields.

---

## 127. Testing report durations

Pipeline duration and step durations should be:

```text
>= 0
```

A successful executed step should appear in the report.

Avoid requiring a positive number of whole milliseconds because a fast step can complete within the same millisecond.

---

## 128. Testing reports on failure

After argument validation succeeds, a pipeline failure should still return a meaningful report.

Verify:

- total steps;
- executed steps;
- outcome;
- failing step;
- error policy;
- input/output count;
- rolled-back document.

A returned error does not make the report disposable.

---

## 129. Testing empty pipeline

An empty module list is valid.

Verify:

- application builds;
- pipeline returns input unchanged;
- report has zero steps;
- Fish receives original text;
- command can complete successfully.

Do not conflate:

```text
nil module array
```

with:

```text
empty module array
```

Configuration treats them differently.

---

## 130. Testing repeated module types

Two instances of the same module type are valid when names differ.

Verify:

- order preserved;
- instance configs remain independent;
- names appear correctly;
- per-instance `onError` override works;
- processors are separately constructed.

Do not use module type as the unique identity.

---

## 131. Testing duplicate module names

Duplicate instance names must be rejected.

Test both:

- configuration validation;
- defensive pipeline construction.

The duplicate check protects logs, reports, and operator reasoning.

---

## 132. Testing unknown module fields

The outer module entry rejects unknown fields.

The inner `config` object belongs to the module and is decoded separately.

Tests should distinguish:

```text
unknown field beside name/type/config
```

from:

```text
unknown field inside module config
```

Both are rejected, but by different layers.

---

## 133. Testing null ownership

Top-level configuration rejects prohibited nulls.

Module config null semantics belong to the module.

Test that core leaves nested module values available for module-specific decoding rather than inventing universal meaning.

A module should explicitly decide whether each of its fields accepts null.

---

## 134. Testing defaults

`config.Default()` should remain a complete valid config.

At minimum verify:

- `Default().Validate()` succeeds;
- default pipeline is buildable;
- default Fish request parameters validate;
- default paths resolve;
- example config stays aligned through dedicated consistency tests where implemented.

A new required field must receive a valid default or the default-config test should fail.

---

## 135. Configuration example consistency

The example file is user-facing executable configuration.

Tests should compare semantically important values against defaults where the project intends alignment.

Do not require byte-for-byte equality when comments, formatting, or intentional example overrides differ.

The current planned work includes extending machine-checked consistency.

---

## 136. Testing logging formats

For JSON:

- decode one record;
- verify `msg`;
- verify level as appropriate;
- verify custom fields;
- avoid exact timestamp.

For text:

- verify required substrings;
- verify field presence;
- avoid exact timestamp layout unless the format becomes contractual.

---

## 137. Testing log levels

For each supported threshold:

```text
debug
info
warn
error
```

Verify lower-severity records are filtered and accepted severities remain.

Also test unsupported values.

Configuration validation accepts exact lowercase values.

Lower-level parsing helpers may normalize input differently; tests should preserve that distinction.

---

## 138. Testing request serialization

Fish request tests should verify:

- required text;
- format;
- reference ID omission/presence;
- pointer fields such as sample rate;
- zero versus omitted semantics;
- feature arrays;
- prosody;
- exact JSON field names.

Prefer decoding request JSON into a struct or map rather than comparing field order.

---

## 139. Testing header safety

Fish client construction should reject values containing:

- surrounding whitespace where forbidden;
- invalid UTF-8;
- ASCII control characters.

Test:

- API key;
- model;
- any future header-derived value.

The test must verify rejection occurs before an HTTP request is sent.

---

## 140. Testing endpoint resolution

Verify:

- root base URL;
- trailing slash;
- existing base path;
- surrounding whitespace behavior;
- HTTP;
- HTTPS;
- unsupported scheme;
- missing host;
- user info;
- query;
- fragment.

Expected endpoint suffix:

```text
/v1/tts
```

No live DNS lookup is necessary.

---

## 141. Testing bounded Fish error bodies

Test exact configured maximum and maximum plus one.

On overflow, verify:

- `boundedio.LimitError`;
- maximum value;
- typed API status still available when errors are joined;
- output writer remains empty;
- response body closes.

This protects both resource limits and classification.

---

## 142. Testing body close with retries

Every retry response body must close before the next attempt.

Use per-attempt close flags.

A test should fail if an earlier response remains open while a later attempt begins.

This prevents connection leaks.

---

## 143. Testing cancellation during HTTP

Use a context and controlled server/transport.

Assert:

- returned error matches `context.Canceled` or deadline;
- no extra retry;
- response body closes if acquired;
- output does not publish;
- command maps to exit `4`.

Do not depend on a real network timeout.

---

## 144. Testing timeout validation

Client timeout tests should include:

```text
zero
negative
maximum accepted
above maximum
```

A tiny valid timeout can be used for a controlled timeout integration test.

Do not make the suite depend on exact scheduler timing for validation tests.

---

## 145. Testing retry configuration

Verify:

- attempts positive and within maximum;
- initial delay positive and within maximum;
- max delay positive and within maximum;
- max delay not below initial delay;
- server-error flag behavior.

Use table mutators from a valid baseline.

---

## 146. Testing no mutable default aliasing

If defaults contain slices or pointers that callers can mutate, test whether independent `Default()` calls share backing state.

The intended contract should be explicit.

For module config raw messages and feature slices, accidental shared mutation can create order-dependent tests and runtime bugs.

Add a regression test when a mutable default field is introduced or changed.

---

## 147. Testing copy ownership

Constructors that promise private copies should be tested.

Examples:

- pipeline copies the step slice;
- document preserves original text;
- client retains required option values independently.

Mutate caller-owned input after construction and verify internal behavior remains stable where copying is promised.

---

## 148. Testing nil slices versus empty slices

JSON and Go distinguish:

```text
nil
[]
```

Relevant fields include:

- pipeline modules;
- features;
- report steps.

Test the intended serialization and validation behavior.

Do not let refactoring collapse a meaningful distinction silently.

---

## 149. Testing error context

An error should preserve the cause and add enough context to identify:

- path;
- module name;
- module type;
- operation;
- endpoint stage;
- output stage.

Test stable context fragments such as:

```text
prepare module
write temporary output file
read secret file
```

Avoid asserting incidental punctuation across an entire nested chain.

---

## 150. Testing error redaction

When an error includes request or response context, verify it does not include:

- authorization header;
- full API key;
- sensitive input unless explicitly configured;
- unrelated environment variables.

A synthetic unique marker makes accidental leakage easy to detect.

---

## 151. Testing log fan-out failure

Use two writers:

```text
one failing
one succeeding
```

Verify the succeeding destination is still attempted.

Reverse their order and verify the contract remains correct.

Test both failing and confirm joined errors.

The application’s `slog` calls do not surface runtime handler errors to `run`, but the writer package still must return them correctly.

---

## 152. Testing post-publication error

This branch is crucial:

```text
rename succeeded
directory sync or close failed
```

Use an injectable directory-sync boundary if the real filesystem cannot produce the error deterministically.

Assert:

- function returns error;
- destination contains new bytes;
- no temp remains;
- old content is not expected to return;
- error contains persistence context.

Do not fake this by returning an error before rename.

That tests the wrong state.

---

## 153. Test seam design

A good test seam is:

- narrow;
- package-private;
- typed;
- production-neutral;
- lifecycle-specific.

Examples:

- injected writer;
- injected closer;
- injected registry;
- custom `RoundTripper`;
- helper combining sync and close errors.

Avoid a global “mock everything” interface that mirrors the entire standard library.

---

## 154. Avoid production code written only for tests

Do not export internal functions solely to reach them from tests.

Same-package tests can access unexported functions.

If deterministic failure injection needs a seam, add the smallest internal abstraction that also clarifies lifecycle ownership.

---

## 155. Regression tests

Every bug fix should include a test that fails before the fix.

A regression test should:

- reproduce the smallest relevant scenario;
- assert the broken observable contract;
- avoid depending on unrelated implementation;
- use the original error boundary;
- keep a descriptive name.

Examples of current regression-style contracts include:

- duplicate JSON keys;
- typed-nil interfaces;
- output cleanup error preservation;
- directory sync and close error joining;
- exact read-limit boundaries.

---

## 156. Test review questions

For every changed test, ask:

- What observable contract does it protect?
- Could it pass when the production behavior is wrong?
- Does it assert side effects?
- Is error identity preserved?
- Is it deterministic?
- Is it isolated?
- Is parallelism safe?
- Does it use real external services?
- Does it leak secrets?
- Does it test the correct lifecycle phase?
- Will it pass under `-race`?
- Does it restore global state?

---

## 157. Adding a new module

A new module contribution should include tests for:

### Configuration

- valid minimal object;
- unknown field;
- duplicate field;
- invalid UTF-8;
- null behavior;
- every semantic bound.

### Preparation

- path resolution;
- no runtime resources acquired prematurely;
- returned builder non-nil;
- wrapped preparation errors.

### Builder

- independent instances;
- runtime dependency initialization;
- wrapped errors;
- typed-nil processor rejection.

### Processing

- normal transformation;
- valid UTF-8;
- nonblank output;
- cancellation;
- provider or I/O error;
- no partial mutation on returned error where module controls mutation.

### Registry

- type registration;
- repeated instances;
- order;
- policy override;
- prepare-before-build integration.

### Pipeline

- recovery policy behavior;
- abort behavior;
- invalid output handling.

---

## 158. Module HTTP tests

A remote module should follow the Fish client testing pattern:

- `httptest.Server`;
- custom transport for lower-level faults;
- no live provider;
- bounded response bodies;
- typed errors where useful;
- cancellation;
- retry policy;
- header safety;
- body closure;
- no text leakage in logs.

A module must not reuse Fish-specific test helpers when that creates inappropriate coupling.

Share only genuinely generic helpers.

---

## 159. Module filesystem tests

A filesystem module should test:

- resolver behavior;
- parent existence;
- permissions;
- symlinks;
- file types;
- atomicity if writing;
- cleanup;
- concurrent instances;
- path escape policy;
- absolute and relative paths.

Its contract may differ from secrets or output.

Do not copy their security claims without implementing and testing them.

---

## 160. Module LLM tests

A future LLM transformation module should test:

- exact request schema;
- strict response JSON;
- malformed model response;
- missing required fields;
- extra fields;
- invalid UTF-8;
- timeout;
- cancellation;
- token/output bounds;
- remote error classification;
- deterministic fallback through pipeline policy;
- no secret or full prompt logging by default.

Use a local HTTP server.

Do not make unit tests depend on model creativity, which is already enough of a production problem.

---

## 161. Module text tests

A text transformation module should include:

- ASCII;
- Cyrillic;
- emoji;
- punctuation;
- whitespace;
- empty-looking Unicode;
- invalid UTF-8 at boundary;
- idempotence where promised;
- ordering-sensitive examples;
- maximum expected size.

Expected output should be explicit.

Avoid tests that merely assert output differs from input unless difference itself is the contract.

---

## 162. CLI compatibility tests

When changing CLI behavior, protect:

- flag names;
- required flags;
- help text;
- accepted formats;
- `ogg` alias;
- stdin fallback;
- positional-argument rejection;
- exit codes;
- stdout/stderr destinations.

A flag rename or changed exit code is a compatibility change even if internal package tests remain green.

---

## 163. Documentation compatibility tests

Where documentation lists machine-readable values, add tests or generation checks when practical.

Candidates:

- format names;
- error policies;
- default model;
- default paths;
- exit codes;
- registered module types.

The current documentation build is manual Markdown validation.

Future consistency checks should compare authoritative code values rather than duplicate another hand-maintained list.

---

## 164. Markdown validation for documentation files

The documentation workflow currently validates generated files for:

- expected checksum;
- line count;
- byte count;
- heading count;
- balanced fenced blocks;
- trailing whitespace;
- tab characters;
- local link targets;
- `git diff --check`.

These checks protect installation integrity.

They do not replace human review of technical correctness.

---

## 165. `git diff --check`

Before committing any change, run:

```bash
clear

git diff --check
git diff --cached --check
```

This catches:

- trailing whitespace;
- whitespace errors in the diff.

It does not run Go formatting or tests.

Use it alongside the standard verification suite.

---

## 166. Focused pre-commit loop

During development:

```bash
clear

gofmt -w path/to/changed.go path/to/changed_test.go
go test -count=1 ./path/to/package
```

Before commit:

```bash
clear

gofmt -w .
go vet ./...
go test -count=1 ./...
git diff --check
```

Before push:

```bash
clear

go test -race -count=1 ./...
go build -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli
git diff --check
git diff --cached --check
```

---

## 167. Diagnosing a failing package

Run:

```bash
clear

go test -count=1 -v ./internal/package
```

Then target the test:

```bash
clear

go test -count=1 -v \
  ./internal/package \
  -run '^TestExactName$'
```

Then repeat:

```bash
clear

go test -count=50 \
  ./internal/package \
  -run '^TestExactName$'
```

Then add race detection:

```bash
clear

go test -race -count=20 \
  ./internal/package \
  -run '^TestExactName$'
```

Do not begin by editing production code until the failure is reproducible and the broken assertion is understood.

---

## 168. Diagnosing a race

A race report identifies:

- conflicting accesses;
- goroutine stacks;
- creation sites.

Fix the synchronization or ownership bug.

Do not suppress the test or serialize the entire suite unless serial execution is truly the production contract.

For tests, verify whether the race is in:

- test handler state;
- global mutation;
- production object;
- cleanup.

---

## 169. Diagnosing a hang

Use verbose output and a shorter package timeout:

```bash
clear

go test -count=1 -v \
  -timeout=30s \
  ./internal/fish
```

On timeout, inspect goroutine dumps.

Common causes:

- channel never closed;
- server waiting for body;
- body never closed;
- context not canceled;
- retry delay not injected;
- cleanup waiting on active handler;
- mutex misuse.

---

## 170. Diagnosing filesystem failures

Print metadata without dumping sensitive content:

```bash
clear

stat -c \
  'type=%F mode=%a owner=%U group=%G size=%s path=%n' \
  /path/to/test/file
```

For stale temp files:

```bash
clear

find /path/to/output \
  -maxdepth 1 \
  -type f \
  -name '.*.*.tmp' \
  -print
```

In tests, prefer assertions over ad hoc prints.

---

## 171. Diagnosing HTTP tests

Log only synthetic test values.

Useful observations:

- attempt count;
- method;
- URL path;
- selected headers excluding authorization value;
- decoded request fields;
- response status;
- body close state;
- context error.

Do not enable verbose dumping of full production-like headers.

---

## 172. Determinism requirements

The standard suite should be deterministic with respect to:

- remote services;
- credentials;
- wall clock;
- random ports;
- temporary paths;
- test order;
- current working directory;
- developer home;
- locale.

Randomly generated temp names and request IDs are acceptable when tests assert shape and behavior rather than fixed values.

---

## 173. Test independence from repository files

Tests must not require local runtime files ignored by Git.

Specifically avoid dependencies on:

```text
config/config.json
secrets/
logs/
bin/
```

Use:

- `config.Default()`;
- inline JSON;
- `t.TempDir()`;
- local HTTP servers;
- direct package constructors.

A clean checkout should be sufficient.

---

## 174. Test independence from command binary

Most package and command integration tests do not require a prebuilt:

```text
bin/fish-audio-cli
```

`go test` compiles its own test binaries.

The separate CI build step verifies the distributable command target.

Do not make unit tests depend on a stale manually built binary.

---

## 175. Test output discipline

Passing tests should produce little output.

Use `t.Log` for diagnostics that are useful with `-v`.

Do not print routine progress to stdout.

A noisy suite makes meaningful CI failures harder to identify and can expose test data.

---

## 176. Fatal versus nonfatal assertions

Use `t.Fatal` or `t.Fatalf` when later assertions cannot proceed safely.

Examples:

- constructor failed;
- file could not be read;
- server setup failed;
- type assertion required for later access.

Use `t.Error` inside HTTP handlers when the handler should continue enough to return a controlled response.

Avoid cascading panics after a failed prerequisite.

---

## 177. Assertions inside goroutines

The `testing.T` methods may be used carefully from goroutines while the test is active, but control flow such as `t.Fatal` only terminates the calling goroutine.

Prefer sending results through channels or using `t.Error` in handlers.

Ensure the main test waits for every goroutine before returning.

---

## 178. Resource ownership in tests

The code that creates a resource owns its cleanup unless ownership is explicitly transferred.

Examples:

- test creates server → test closes server;
- output package creates temp → output package cleans it;
- test creates context → test cancels it;
- logger open returns closer → test closes it;
- helper creates file in `t.TempDir()` → testing removes directory.

Unclear ownership produces leaks and flaky teardown.

---

## 179. Test helper errors

A helper should usually fail the test directly for setup errors:

```go
t.Fatalf(...)
```

A helper should return an error when the error itself is under test.

Example:

```text
loadTestConfig returns (Config, error)
```

because caller cases inspect success and failure.

Do not consume the error inside a helper when the test needs its identity.

---

## 180. Testing defensive branches

Some branches should be unreachable after earlier validation.

They still deserve tests when they protect package use outside the command.

Examples:

- unsupported pipeline policy;
- nil processor builder;
- typed-nil writer;
- uninitialized resolver;
- nil decode target.

Package contracts should remain safe when called directly.

---

## 181. Refactoring tests

A refactor must preserve observable behavior tests.

It may change:

- helper structure;
- internal function names;
- injected seams;
- package file layout.

It must not casually delete tests for:

- error identity;
- side effects;
- security modes;
- rollback;
- retry count;
- cleanup;
- exit codes.

If a test is too coupled, rewrite it around the contract before removing it.

---

## 182. Flaky test policy

A flaky test is a defect.

Do not:

- rerun CI until green and ignore it;
- add a large sleep;
- weaken the assertion;
- skip it permanently.

Investigate:

- unsynchronized state;
- timing;
- global mutation;
- server shutdown;
- filesystem assumptions;
- random order;
- leaked goroutines.

Record the reproducing seed or command.

---

## 183. CI timeout budget

The complete job must finish within:

```text
15 minutes
```

The required race run is the most expensive current step.

New tests should remain efficient.

A test that intentionally waits seconds per case can consume the budget quickly under:

- many subtests;
- race instrumentation;
- slower hosted runners.

Use controllable time.

---

## 184. Test package concurrency

Go may execute different packages concurrently.

Serializing tests inside one package does not protect shared external resources across packages.

Avoid shared:

- fixed ports;
- fixed `/tmp` filenames;
- repository output files;
- environment-dependent service names.

Use unique temp directories and ephemeral ports.

---

## 185. File descriptor discipline

HTTP and filesystem tests can exhaust descriptors when loops forget to close:

- response bodies;
- files;
- directories;
- servers.

A repeat run such as:

```bash
clear

go test -count=100 ./internal/fish
```

is useful for finding leaks.

Do not rely on garbage collection to close resources.

---

## 186. Goroutine discipline

A test should not leave background goroutines after completion.

Use:

- context cancellation;
- server close;
- channel close ownership;
- wait groups;
- finite retries.

The race detector does not automatically prove absence of goroutine leaks.

Repeated tests and package timeouts help expose them.

---

## 187. CI action versions

The current workflow uses:

```text
actions/checkout@v6
actions/setup-go@v6
```

`setup-go` enables caching.

Action-version changes should be reviewed as build-infrastructure changes.

The test commands remain the source of behavioral verification.

---

## 188. CI cache

Go setup caching can accelerate:

- module download;
- build cache.

The explicit:

```text
go test -count=1
```

still forces test execution.

Build artifacts and package compilation can remain cached according to Go tooling.

Do not confuse build cache with skipped tests.

---

## 189. Pull-request expectations

A pull request changing behavior should include:

- production change;
- focused regression or feature tests;
- documentation change where contract changes;
- green standard suite;
- green race suite;
- successful command build.

The review should identify the failure phase and side effects being protected.

---

## 190. Documentation-only pull requests

Documentation changes should still be checked for factual consistency with code.

The Go CI cannot detect a false sentence such as an unsupported filesystem workaround.

Documentation review must trace behavioral claims to:

- implementation;
- tests;
- existing normative documents.

Humans remain regrettably necessary.

---

## 191. Test-only APIs

Do not add exported production APIs named:

```text
ForTest
TestingOnly
Mock
```

without a genuine runtime contract.

Prefer:

- unexported injected functions;
- interfaces already required by architecture;
- same-package tests;
- custom standard-library implementations.

---

## 192. Mock framework policy

The project currently succeeds with handwritten fakes and standard-library test facilities.

A third-party mock framework should not be added merely for stylistic consistency.

It would need to justify:

- dependency cost;
- generated files;
- contributor workflow;
- readability;
- error quality;
- maintenance.

Small interfaces rarely need it.

---

## 193. Assertion library policy

The project uses standard `testing` assertions.

A third-party assertion library is not currently required.

Explicit checks make:

- actual values;
- expected values;
- error identity;
- control flow

visible in review.

Adding a library is a project dependency decision, not a shortcut for typing fewer `if` statements.

---

## 194. Code generation in tests

If generated mocks or fixtures are introduced, the repository must document:

- generator version;
- command;
- source;
- reproducibility;
- verification that generated files are current.

CI currently does not run a generation-diff gate.

Avoid generated complexity until it solves a real problem.

---

## 195. Snapshot update discipline

Any future snapshot or golden update command must be opt-in.

A normal test run must fail on mismatch rather than silently rewrite expected output.

Generated expected data should be reviewed like production code.

---

## 196. Security regression priority

Security-sensitive behavior deserves explicit regression tests:

- secret symlink rejection;
- secret directory write-bit rejection;
- secret mode enforcement;
- header control-character rejection;
- output symlink replacement;
- no credential logging;
- bounded error bodies;
- strict JSON duplicate rejection;
- trusted path semantics.

These tests should not be removed to improve portability without a replacement contract.

---

## 197. Durability regression priority

Output durability behavior deserves explicit tests:

- sync before rename;
- close before rename;
- directory sync after rename;
- preserve old destination before rename;
- retain new destination after post-rename failure;
- cleanup errors joined.

The exact system calls may be wrapped behind seams.

The state-machine assertions must remain.

---

## 198. Compatibility regression priority

User-facing compatibility tests should protect:

- flag names;
- format values;
- error policies;
- exit codes;
- config JSON names;
- default paths;
- module identity;
- output semantics;
- log event names where consumed.

A green internal refactor that changes one of these can still break users.

---

## 199. Test command quick reference

Fast package loop:

```bash
clear

go test -count=1 ./internal/package
```

Full tests:

```bash
clear

go test -count=1 ./...
```

Race:

```bash
clear

go test -race -count=1 ./...
```

Vet:

```bash
clear

go vet ./...
```

Formatting:

```bash
clear

gofmt -w .
```

Formatting check:

```bash
clear

gofmt -l .
```

Build:

```bash
clear

go build -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli
```

Coverage:

```bash
clear

go test -count=1 -cover ./...
```

Shuffle:

```bash
clear

go test -count=1 -shuffle=on ./...
```

---

## 200. Maintainer checklist

Before merging:

### Code quality

- `gofmt -l .` is empty.
- `go vet ./...` passes.
- command builds with `-trimpath`.

### Tests

- focused changed-package tests pass;
- full uncached suite passes;
- race suite passes;
- regression test exists;
- no skipped critical contract;
- no live external dependency.

### Isolation

- temp directories used;
- ports ephemeral;
- globals restored;
- parallelism safe;
- no repository-local runtime files;
- no real secrets.

### Assertions

- error identity checked;
- side effects checked;
- cleanup checked;
- permission checked where relevant;
- ordering checked;
- boundaries checked.

### Documentation

- new behavior documented;
- test description matches actual CI;
- optional tools are not described as required gates;
- no unsupported platform claim.

---

## 201. Testing invariants

The following rules are normative for the current repository.

1. The module declares Go `1.26.5`.
2. CI reads the Go version from `go.mod`.
3. CI runs on `ubuntu-latest`.
4. CI triggers on push to `main`.
5. CI triggers on pull requests targeting `main`.
6. CI supports manual dispatch.
7. CI has read-only contents permission.
8. CI job timeout is 15 minutes.
9. CI checks `gofmt -l .`.
10. CI runs `go vet ./...`.
11. CI runs `go test -count=1 ./...`.
12. CI runs `go test -race -count=1 ./...`.
13. CI builds `./cmd/fish-audio-cli`.
14. CI build uses `-trimpath`.
15. CI does not require a live Fish credential.
16. CI does not contact the live Fish API by design.
17. HTTP tests use local servers or custom transports.
18. Filesystem tests use temporary directories.
19. Test secrets are synthetic.
20. Tests live beside implementation packages.
21. Current suites generally use same-package tests.
22. Handwritten fakes are preferred for small interfaces.
23. `t.Parallel()` is used for isolated tests.
24. Tests mutating `os.Args` are not parallel.
25. Tests changing cwd are not parallel.
26. Global state must be restored with cleanup.
27. Error causes are checked with `errors.Is`.
28. Typed errors are checked with `errors.As`.
29. Joined errors preserve every relevant cause.
30. Filesystem tests assert side effects.
31. Output tests distinguish pre- and post-rename failures.
32. Secret tests verify security modes and file type.
33. Configuration tests verify exact byte boundaries.
34. Strict JSON tests verify duplicate keys.
35. Pipeline tests verify rollback and policy.
36. Registry tests verify prepare-before-build.
37. Fish tests verify request method, path, headers, and body.
38. Retry tests verify attempt count and final category.
39. Response bodies must close.
40. Typed-nil interface guards are tested.
41. No coverage threshold is currently enforced.
42. No fuzz-duration gate is currently enforced.
43. No benchmark threshold is currently enforced.
44. No cross-platform CI matrix currently exists.
45. Targeted tests do not replace the full suite.
46. Cached plain `go test` does not replace uncached verification.
47. Race success is required in CI.
48. A passing test must not leak resources or global state.
49. A bug fix should include a regression test.
50. A behavior change should update normative documentation.

Changing one of these rules is a testing or contributor-workflow compatibility change.

---

## 202. Non-goals

The current testing system does not provide:

- live Fish API certification;
- provider performance measurement;
- audio perceptual quality scoring;
- golden voice comparison;
- cross-platform CI;
- multi-architecture CI;
- coverage enforcement;
- mutation testing;
- sustained fuzzing in CI;
- benchmark regression enforcement;
- automatic test generation;
- third-party mock generation;
- automatic Markdown semantic verification;
- guaranteed simulation of every filesystem failure;
- privileged special-device tests;
- distributed-system end-to-end testing;
- production credential validation.

These require explicit infrastructure and cost decisions.

---

## 203. Summary

The required repository verification is:

```text
gofmt check
    ↓
go vet
    ↓
uncached full tests
    ↓
race detector
    ↓
trimmed command build
```

The test strategy is:

```text
pure contracts
    → table-driven unit tests

interfaces and lifecycle
    → handwritten fakes

HTTP
    → httptest server or custom transport

filesystem
    → t.TempDir plus content, type, mode, and cleanup assertions

command wiring
    → run() with temp config, fake secret, local Fish server, and output

concurrency
    → t.Parallel only with isolated state, then verify under -race
```

The most important contributor rules are:

- run the exact CI commands before push;
- use `-count=1` for final verification;
- keep tests independent from real credentials and services;
- assert error identity and side effects;
- test lifecycle boundaries, not only happy-path values;
- never parallelize tests that mutate process globals;
- verify filesystem cleanup and permissions;
- preserve prepare-before-build and pipeline rollback tests;
- distinguish optional coverage, shuffle, fuzzing, and benchmarks from required gates;
- treat flaky tests, races, and leaked resources as defects rather than weather.
