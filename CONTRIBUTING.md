# Contributing to fish-audio-cli

Thank you for contributing to `fish-audio-cli`.

The project is a small command-line program with deliberately explicit behavior around configuration, text processing, external HTTP requests, secrets, logs, errors, and atomic output files. A change that looks local can affect several public contracts at once.

This guide explains the current contribution workflow, required verification, design expectations, documentation obligations, and review criteria.

> [!NOTE]
> The project is in alpha development. Public interfaces can still change, but changes must remain deliberate, tested, documented, and reviewable.

## Table of contents

- [Before contributing](#before-contributing)
- [Project principles](#project-principles)
- [Ways to contribute](#ways-to-contribute)
- [Discussing changes](#discussing-changes)
- [Development environment](#development-environment)
- [Repository layout](#repository-layout)
- [Local setup](#local-setup)
- [Development workflow](#development-workflow)
- [Required verification](#required-verification)
- [Go code standards](#go-code-standards)
- [Error handling](#error-handling)
- [Context and cancellation](#context-and-cancellation)
- [Interface and nil safety](#interface-and-nil-safety)
- [Configuration changes](#configuration-changes)
- [Module changes](#module-changes)
- [Pipeline changes](#pipeline-changes)
- [Fish Audio changes](#fish-audio-changes)
- [Secrets and path changes](#secrets-and-path-changes)
- [Output-file changes](#output-file-changes)
- [Logging changes](#logging-changes)
- [CLI and exit-code changes](#cli-and-exit-code-changes)
- [Testing expectations](#testing-expectations)
- [Documentation changes](#documentation-changes)
- [Dependencies](#dependencies)
- [Security-sensitive changes](#security-sensitive-changes)
- [Commit messages](#commit-messages)
- [Pull requests](#pull-requests)
- [Review checklist](#review-checklist)
- [After review](#after-review)

---

## Before contributing

Read the documentation relevant to the change.

Start with:

- [`README.md`](README.md)
- [`docs/index.md`](docs/index.md)
- [`docs/architecture.md`](docs/architecture.md)
- [`docs/testing.md`](docs/testing.md)

Then read the subsystem document that owns the behavior being changed.

| Change area | Required reference |
|---|---|
| CLI flags or input selection | [`docs/cli.md`](docs/cli.md) |
| JSON fields, defaults, or validation | [`docs/configuration.md`](docs/configuration.md) |
| Pipeline execution or fallback | [`docs/pipeline.md`](docs/pipeline.md) |
| Module model or registry | [`docs/modules.md`](docs/modules.md) |
| New built-in module | [`docs/module-author-guide.md`](docs/module-author-guide.md) |
| Fish request, response, or retry | [`docs/fish-audio.md`](docs/fish-audio.md) |
| Logs or lifecycle events | [`docs/logging.md`](docs/logging.md) |
| Paths, secrets, or permissions | [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md) |
| Output publication or cleanup | [`docs/output-files.md`](docs/output-files.md) |
| Errors or process statuses | [`docs/errors-and-exit-codes.md`](docs/errors-and-exit-codes.md) |
| Failure diagnosis | [`docs/troubleshooting.md`](docs/troubleshooting.md) |

The specialized document is authoritative for its subsystem.

Do not begin a cross-cutting change from a single code file without first understanding the existing lifecycle and public contract.

---

## Project principles

Contributions should preserve the following principles.

### One-shot command

`fish-audio-cli` is currently a single-invocation executable.

It is not:

- a daemon;
- a local HTTP server;
- a long-running worker;
- a dynamic plugin host;
- a stateful service.

A proposal to change that model is architectural work, not a small feature patch.

### Explicit ownership

Every behavior should have a clear owner.

Examples:

- CLI package owns option parsing and input selection.
- Config package owns JSON defaults and semantic validation.
- Module packages own their instance configuration and transformation logic.
- Pipeline owns ordered execution, rollback, and error policy.
- Fish package owns HTTP protocol behavior.
- Secrets package owns secure key-file loading.
- Output package owns atomic publication.
- Logging package owns destinations and structured handlers.
- Command package owns orchestration and process exit codes.

Do not move behavior across boundaries merely because another package is convenient to edit.

### Strict inputs

The project prefers rejecting ambiguous input over silently guessing.

Examples include:

- strict JSON;
- duplicate-key rejection;
- unknown-field rejection;
- exact module type lookup;
- explicit output format;
- exact secret-line rules;
- bounded reads;
- UTF-8 validation;
- header control-character rejection.

New configuration or external input should follow the same philosophy.

### Narrow interfaces

Prefer small interfaces that describe the actual dependency.

Do not create broad abstractions that mirror entire packages or standard-library types without need.

### Deterministic behavior

Runtime and test behavior should be reproducible.

Avoid:

- hidden environment fallback;
- implicit path expansion;
- live network dependencies in normal tests;
- timing-based test synchronization;
- mutable shared global test state;
- undocumented default changes.

### Preserve error identity

Errors should carry human context while retaining machine-detectable causes.

Use wrapping and joining correctly.

### Side effects are part of the contract

A returned value is not the complete result of an operation.

Tests and documentation must consider:

- created files;
- retained files;
- permissions;
- cleanup;
- network attempts;
- log records;
- output publication;
- rollback;
- retry count.

### Documentation is part of the implementation

A behavior change is incomplete until every affected normative document agrees with code and tests.

---

## Ways to contribute

Useful contributions include:

- bug fixes;
- regression tests;
- input-validation improvements;
- error-context improvements;
- deterministic failure injection for tests;
- documentation corrections;
- documentation completeness improvements;
- security hardening;
- new built-in text-processing modules;
- Fish request support;
- portability fixes;
- CI improvements;
- performance improvements supported by measurement;
- cleanup of dead or misleading code.

A contribution does not need to be large.

A precise test that exposes a real lifecycle bug can be more valuable than an elaborate new abstraction.

---

## Discussing changes

### Small, isolated fixes

A focused bug fix or documentation correction can usually proceed directly to a pull request.

Examples:

- preserve a wrapped sentinel;
- reject one malformed value;
- correct one false documentation claim;
- add a missing boundary test;
- fix one cleanup path.

### Broad or compatibility-sensitive changes

Discuss the design before investing in a large implementation when the change affects:

- CLI flags;
- exit codes;
- configuration schema;
- defaults;
- module lifecycle;
- pipeline policies;
- Fish retry semantics;
- secret handling;
- logging destinations;
- output atomicity;
- public error categories;
- dependencies;
- daemon or service behavior;
- dynamic plugin loading.

A design discussion should answer:

1. What user problem is being solved?
2. Which package owns the behavior?
3. What is the public contract?
4. What existing behavior changes?
5. What failure states exist?
6. What side effects occur?
7. How is cancellation handled?
8. How is the behavior tested?
9. Which documents must change?
10. Is backward compatibility affected?

### No issue-first requirement

The repository does not currently require a public issue before every pull request.

Do not create process overhead for a self-contained correction.

Use prior discussion when it reduces wasted implementation or clarifies a compatibility decision.

### Security reports

Do not publish suspected vulnerabilities, leaked credentials, or exploitable filesystem races in a public issue.

Follow the repository security policy once `SECURITY.md` is present.

Until then, use a private contact method provided by the repository owner rather than public disclosure.

---

## Development environment

The module declares:

```text
Go 1.26.5
```

Use the declared version when practical.

A newer compatible toolchain can be useful for local development, but the required behavior must pass the repository CI configuration.

### Required tools

- Git
- Go toolchain
- a shell capable of running the documented commands
- standard Unix utilities for the examples, when developing on Unix-like systems

The current Go module has no third-party module requirements.

### Supported CI environment

Current CI uses:

```text
ubuntu-latest
```

This does not mean the program is intentionally Linux-only.

It means Linux is the environment currently enforced by automation.

A portability claim for another operating system should be supported by code review and testing rather than assumption.

---

## Repository layout

The main repository areas are:

```text
cmd/fish-audio-cli/
    command entry point and orchestration

internal/app/
    application composition and Fish request construction

internal/boundedio/
    bounded reads

internal/cli/
    command-line parsing and text input

internal/config/
    defaults, loading, and validation

internal/fish/
    Fish HTTP client, request validation, retries, and API errors

internal/logging/
    structured logging and persistent file handling

internal/moduleconfig/
    strict module-owned config decoding

internal/modules/
    module registry and built-in module packages

internal/output/
    atomic output publication

internal/pipeline/
    ordered text processing, rollback, policies, and reports

internal/projectpath/
    lexical project-path resolution

internal/secrets/
    secure secret-file loading

internal/strictjson/
    strict JSON validation and decoding

docs/
    normative and practical documentation

config/
    tracked example configuration

deploy/
    deployment support files

.github/workflows/
    CI workflow
```

Tests normally live beside the implementation:

```text
internal/fish/client.go
internal/fish/client_test.go
```

The project does not currently use a separate top-level Go test tree.

---

## Local setup

Clone the repository:

```bash
clear

git clone https://github.com/piqnyx/fish-audio-cli.git
cd fish-audio-cli
```

Check the toolchain:

```bash
clear

go version
cat go.mod
```

Run an initial clean verification:

```bash
clear

gofmt -l .
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...

go build \
  -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli
```

A clean checkout does not require:

- a real Fish API key;
- `config/config.json`;
- a built binary in `bin/`;
- a persistent log directory;
- external Fish network access.

Normal tests create their own temporary configuration, secret, HTTP server, log, and output resources.

---

## Ignored local files

The repository currently ignores:

```text
/bin/
/secrets/
/config/config.json
/logs/
/docs/maintainers/final-audit-anchor.md
```

Do not commit:

- real API keys;
- local runtime configuration;
- generated binaries;
- runtime logs;
- generated audio;
- diagnostic files containing private text;
- temporary output files.

A custom secret path outside the ignored `/secrets/` directory can still be committed accidentally.

Check staged files before every commit.

---

## Development workflow

A normal contribution follows this sequence.

### 1. Establish a clean baseline

```bash
clear

git status --short
git log -1 --oneline --decorate

go test -count=1 ./...
```

Do not begin from an unexplained dirty working tree.

### 2. Reproduce the problem

For a bug:

- identify the package and lifecycle stage;
- capture the exact error;
- record the exit code when relevant;
- inspect side effects;
- reduce the reproduction;
- avoid real credentials and paid provider calls.

### 3. Add or update a focused test

A bug fix should normally include a regression test that fails before the fix.

A feature should include contract tests before or with implementation.

### 4. Make the smallest coherent change

Avoid combining unrelated cleanup, renaming, formatting, and behavior changes in one patch.

A focused diff is easier to review and safer to revert.

### 5. Run package-local checks

Example:

```bash
clear

gofmt -w \
  internal/fish/client.go \
  internal/fish/client_test.go

go test \
  -count=1 \
  ./internal/fish
```

### 6. Update documentation

Review the documentation update matrix later in this guide.

### 7. Run full verification

Use the exact required commands.

### 8. Review the diff

```bash
clear

git diff --check
git status --short
git --no-pager diff --stat
git --no-pager diff
```

### 9. Commit intentionally

Use a focused message that describes the result.

### 10. Open a pull request

Explain the behavior, tests, compatibility impact, and documentation changes.

---

## Branches

External contributors will normally work in a fork or a dedicated branch and open a pull request against `main`.

The repository does not currently prescribe one mandatory branch-naming scheme.

Useful names include:

```text
fix/secret-close-error
feat/normalizer-module
docs/config-reference
test/fish-retry-lifecycle
```

Do not mix several unrelated changes merely because they are on one branch.

---

## Required verification

The GitHub Actions workflow runs:

```text
gofmt check
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
go build -trimpath ./cmd/fish-audio-cli
```

Before submitting a behavior change, run:

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

go build \
  -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli

git diff --check
```

### Why `-count=1`

The final test run must execute tests instead of reusing a cached successful result.

Plain:

```text
go test ./...
```

is useful during development but does not replace the uncached gate.

### Race detector

The race detector is a required CI step.

A patch that passes ordinary tests but fails under `-race` is not ready.

### Build target

The actual command target must build:

```text
./cmd/fish-audio-cli
```

A package-only test pass is insufficient.

### Documentation-only changes

Documentation-only commits currently pass the same Go CI.

Run the required checks unless the maintainer explicitly establishes a narrower workflow.

---

## Go code standards

### Formatting

All Go code must be formatted with `gofmt`.

Do not hand-align code against `gofmt` output.

### Standard library first

The project currently uses the standard library without third-party module dependencies.

Prefer standard-library facilities when they provide a clear, maintainable solution.

### Readability

Prefer:

- explicit control flow;
- small functions;
- narrow responsibilities;
- descriptive names;
- clear ownership;
- direct error context;
- table-driven tests where appropriate.

Avoid:

- unnecessary cleverness;
- large generic frameworks;
- hidden global state;
- speculative abstraction;
- reflection where typed code is sufficient;
- configuration magic;
- broad interfaces created only to support mocks.

### Exported identifiers

Export an identifier only when another package genuinely needs it.

Exported types, functions, variables, and constants should have clear doc comments.

Do not export internal implementation merely to make testing easier.

Same-package tests can reach unexported behavior.

### Comments

Comments should explain:

- ownership;
- invariants;
- security reason;
- non-obvious lifecycle;
- why a behavior must not change casually.

Do not narrate obvious syntax.

### Zero values

Decide whether the zero value is valid.

If it is invalid, validate early and return a clear error.

Do not allow an invalid zero-value object to panic deep in an operation.

### Copies and ownership

Be explicit about mutable ownership.

When constructors promise isolation, copy:

- slices;
- maps;
- pointer-backed defaults;
- raw JSON;
- other mutable caller-owned data.

Add tests that mutate the caller value after construction.

### Global state

Avoid new mutable package globals.

When a registry or hook must be package-global:

- make production initialization static;
- provide a narrow internal seam for tests;
- do not mutate it from parallel tests.

---

## Error handling

### Add context

Wrap errors at meaningful boundaries.

Preferred:

```go
return fmt.Errorf(
    "load module configuration: %w",
    err,
)
```

Avoid:

```go
return err
```

when the caller cannot identify the failed operation.

### Preserve identity

Use `%w` when wrapping a cause.

Incorrect:

```go
return fmt.Errorf(
    "load module configuration: %v",
    err,
)
```

Correct:

```go
return fmt.Errorf(
    "load module configuration: %w",
    err,
)
```

### Preserve simultaneous failures

Use `errors.Join` when more than one failure matters.

Examples:

- primary operation plus close failure;
- HTTP API error plus error-body read failure;
- output operation plus cleanup failure;
- directory sync plus close failure.

Tests should verify every cause with `errors.Is` or `errors.As`.

### Sentinel errors

Add a sentinel only when callers need a stable category.

Do not create a sentinel for every human-readable message.

A sentinel should have:

- clear ownership;
- documented meaning;
- tests;
- stable matching behavior;
- a reason callers need it.

### Typed errors

Use typed errors when callers need structured fields.

The Fish `APIError` is an example.

A typed error should support:

- useful `Error()` output;
- `errors.Is` where categories exist;
- `errors.As`;
- stable field meaning.

### Error strings

Human-readable error strings are diagnostic.

Do not make shell automation depend on exact nested English text unless no structured alternative exists and the limitation is documented.

### Cleanup errors

Do not discard close, sync, or remove failures when they affect correctness or diagnosis.

Do not overwrite the primary cause with cleanup failure.

Preserve both.

### No panic for normal invalid input

Invalid config, text, options, dependencies, files, and HTTP responses should return errors.

Panics are for programmer defects or impossible internal states, not ordinary user input.

---

## Context and cancellation

Long-running or external operations should accept `context.Context`.

### Nil context

Interface values can contain typed nil pointers.

Use the project’s nil-safety approach where interface dependencies can be typed nil.

Do not call methods on an unchecked interface merely because:

```go
ctx != nil
```

### Cancellation identity

Preserve:

```text
context.Canceled
context.DeadlineExceeded
```

through wrapping.

Callers and the pipeline use `errors.Is`.

### Pipeline cancellation

Cancellation is interruption, not a recoverable module failure.

Do not route cancellation through:

- `use_previous`;
- `use_original`;
- `skip`.

### Retry waits

Retry waits must stop on context cancellation.

Do not use uninterruptible `time.Sleep` for retry behavior.

### Tests

Test cancellation:

- before work;
- during work;
- after a callback returns;
- during retry delay;
- during HTTP;
- before publication where relevant.

Use deterministic synchronization instead of long sleeps.

---

## Interface and nil safety

Go interfaces can be non-nil while containing a nil pointer.

Public and pluggable interface boundaries should consider:

```text
untyped nil
typed nil
valid implementation
failing implementation
```

Current examples include:

- context;
- writer;
- processor;
- logging writer;
- HTTP-related dependencies;
- module builder results.

A typed-nil test double can panic if invoked.

The test should prove validation rejects it before method dispatch.

Do not add broad reflection-based nil handling to every value.

Use the existing narrow helper and apply it where the contract genuinely accepts interfaces.

---

## Configuration changes

Configuration is a public interface.

A new field or changed default must be treated as a compatibility change.

### Before adding a field

Define:

- owning package;
- JSON name;
- Go type;
- built-in default;
- omitted behavior;
- explicit `null` behavior;
- minimum;
- maximum;
- cross-field rules;
- whitespace rules;
- UTF-8 rules;
- path base when applicable;
- secret sensitivity;
- logging behavior;
- CLI interaction;
- documentation owner.

### Update all layers

A configuration change may require updates to:

- config struct;
- default constructor;
- strict loader;
- explicit-null checks;
- semantic validation;
- duration conversion;
- runtime options;
- tests;
- tracked example config;
- [`docs/configuration.md`](docs/configuration.md);
- owning subsystem documentation;
- [`docs/troubleshooting.md`](docs/troubleshooting.md);
- [`README.md`](README.md) when public quick-start behavior changes.

### Defaults

`config.Default()` must remain complete and valid.

Add or update tests that verify:

```text
Default().Validate() succeeds
```

Do not introduce a required field without a safe default unless the design explicitly changes omission behavior.

### Strict JSON

Do not replace strict decoding with permissive `json.Unmarshal`.

The project intentionally rejects:

- duplicate keys;
- unknown fields;
- multiple top-level values;
- invalid UTF-8;
- prohibited nulls.

### Field names

JSON field names are exact.

Changing case or spelling is a compatibility change.

### Arrays

Define whether an array:

- replaces defaults;
- merges;
- permits empty;
- permits null;
- permits duplicate elements.

Do not rely on incidental Go decoding behavior.

### Path fields

Document whether a path is:

- cwd-relative;
- project-relative;
- module-relative;
- absolute-only;
- lexically cleaned;
- symlink-resolved;
- confined.

Do not invent home or environment expansion unless it is explicitly implemented and secured.

### Example configuration

Update:

[`config/config.example.json`](config/config.example.json)

The example must remain valid under current validation.

Do not include:

- a real key;
- machine-specific private paths;
- inaccessible account identifiers presented as universally valid.

---

## Module changes

Read:

- [`docs/modules.md`](docs/modules.md)
- [`docs/module-author-guide.md`](docs/module-author-guide.md)
- [`docs/pipeline.md`](docs/pipeline.md)

before implementing or changing a module.

### Built-in, compiled-in model

Modules are currently compiled into the binary.

Configuration selects a registered type.

It does not load:

- shared libraries;
- scripts;
- arbitrary packages;
- runtime plugins.

Do not describe a new built-in module as dynamically installed.

### Instance identity

Each configured module instance has a unique `name`.

A type can appear more than once.

Do not use type as the only identity in logs or reports.

### Module-owned configuration

The module owns strict decoding and semantic validation of its `config` object.

Core should not know module-specific fields.

### Prepare before build

The registry must prepare every configured module before building any processor.

This prevents partial runtime initialization when a later module config is invalid.

Preserve this invariant.

### Preparation responsibilities

`Prepare` should:

- strictly decode configuration;
- validate semantic fields;
- resolve lexical project paths;
- construct immutable prepared state;
- return a builder.

It should avoid acquiring long-lived runtime resources unless the architecture requires it.

### Builder responsibilities

The builder should:

- create one independent processor instance;
- initialize runtime dependencies;
- return wrapped errors;
- never return a typed-nil processor as success.

### Processor responsibilities

A processor must:

- accept valid current text;
- respect context cancellation;
- return valid nonblank UTF-8 text on success;
- return an error on failure;
- avoid committing partial mutation on returned error where it controls mutation;
- avoid leaking secrets or full text through errors and logs.

### No cleanup lifecycle

The current module API has no universal cleanup hook.

A module that opens persistent resources must own safe failure and process-lifetime behavior.

Adding a general cleanup lifecycle is an architectural change.

### Registry update

A new type must be registered explicitly.

Add tests proving:

- lookup;
- configured order;
- repeated instances;
- independent config;
- prepare-before-build;
- unknown type behavior;
- builder behavior.

### Module documentation

Document:

- type name;
- purpose;
- exact config object;
- defaults;
- validation;
- side effects;
- external services;
- secret fields;
- logging;
- cancellation;
- error behavior;
- examples.

---

## Pipeline changes

Pipeline behavior is a public runtime contract.

Read [`docs/pipeline.md`](docs/pipeline.md).

### Preserve ordering

Modules execute in configured array order.

Do not parallelize transformations without a new semantic model.

### Preserve rollback

A failed or interrupted step must not commit its partial text mutation.

Tests must verify document state after failure.

### Error policies

Current policies are:

```text
use_previous
use_original
skip
abort
```

Changing their meaning is a compatibility change.

Adding a policy requires updates to:

- parsing;
- validation;
- pipeline execution;
- reports;
- logs;
- exit-code expectations;
- configuration docs;
- pipeline docs;
- troubleshooting;
- tests.

### Cancellation

Cancellation bypasses normal fallback.

Do not convert it into a successful recovered outcome.

### Reports

When argument validation succeeded, errors should still return a useful report where the current contract promises one.

Preserve:

- total steps;
- executed steps;
- outcome;
- step metadata;
- durations;
- text counts.

### Text contract

Pipeline input and successful output must be:

- valid UTF-8;
- nonblank under the shared text rule.

A module returning nil with invalid output is still a module failure.

---

## Fish Audio changes

Read [`docs/fish-audio.md`](docs/fish-audio.md).

### No live provider dependency in normal tests

Use:

- `httptest.Server`;
- custom `http.RoundTripper`;
- synthetic request and response data.

Do not require:

- real credentials;
- paid credits;
- network access;
- a currently available model.

### Endpoint construction

The client validates the base URL and joins:

```text
/v1/tts
```

A change to this behavior affects proxies and custom base paths.

### Headers

Protect:

- authorization;
- model;
- content type.

Reject invalid UTF-8 and ASCII controls in header-derived values.

Never log the full API key.

### Request mapping

A request-field change requires:

- config type/default when configurable;
- local validation;
- JSON tests;
- HTTP request tests;
- example config;
- configuration docs;
- Fish docs;
- troubleshooting where relevant.

### API errors

Preserve:

- HTTP status code;
- HTTP status text;
- Fish API status where present;
- bounded provider message;
- stable error category;
- body read failure when simultaneous.

### Retry behavior

Current internal retry categories are:

- `429`;
- optional `5xx`.

Transport errors are not currently retried.

A retry change must define:

- retryable categories;
- attempt counting;
- delay;
- maximum delay;
- `Retry-After`;
- cancellation;
- body closure;
- streaming behavior;
- idempotency risk;
- logging.

### No retry after streaming begins

The output writer is not rewindable.

Do not retry a successful-status response after partial audio has reached the writer.

### Response lifecycle

Every response body must close on:

- success;
- API error;
- retry;
- read failure;
- cancellation.

Tests must assert it.

### Empty success body

A 2xx response with zero audio bytes is an error.

Do not publish an empty successful output.

---

## Secrets and path changes

Read [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md).

### Secret contents

The Fish API key file contains:

- exactly one nonblank UTF-8 line;
- optional one final LF or CRLF;
- no other surrounding whitespace;
- no additional line.

Do not silently trim arbitrary whitespace.

### Secret file behavior

The loader:

- securely creates a missing file;
- reports `ErrFileCreated`;
- requires a regular leaf;
- rejects a leaf symlink;
- verifies the opened object did not change;
- forces mode `0600`;
- uses a bounded read;
- joins close errors.

A change to one of these behaviors is security-sensitive.

### Secret directory

The final secret directory must not be writable by group or others.

Do not replace this with blanket permission relaxation.

### Path domains

Do not collapse all paths into one base.

Current behavior deliberately differs:

| Path | Base |
|---|---|
| CLI config path | cwd |
| configured secret path | project directory |
| configured log path | project directory |
| module path | module contract |
| CLI output path | cwd |

### Lexical resolver

The project resolver is lexical.

It does not:

- expand `~`;
- expand environment variables;
- resolve real paths;
- confine `..`;
- inspect every ancestor.

Changing this requires a security and compatibility design.

### Tests

Filesystem tests should cover:

- regular file;
- missing file;
- directory mode;
- file mode;
- symlink;
- non-regular file;
- race;
- exact byte limit;
- UTF-8;
- line endings;
- whitespace;
- close failure.

---

## Output-file changes

Read [`docs/output-files.md`](docs/output-files.md).

### Publication sequence

The current success path is:

```text
create temp beside destination
write
sync temp
close temp
rename
sync directory
close directory
```

Do not reorder these steps casually.

### Same-directory temp

The temp file is created beside the destination to preserve same-filesystem rename semantics.

### Pre-rename failure

Before rename:

- existing destination is preserved;
- temp cleanup is attempted;
- cleanup errors are joined.

### Post-rename failure

After rename:

- new output remains published;
- directory-sync failure returns an error;
- old output is not restored.

This unusual state is deliberate and must remain documented and tested.

### Permissions

Final output mode is `0600`.

The old destination metadata is not preserved.

### Symlink behavior

A destination symlink is replaced as a leaf.

Its target is not modified.

Parent path components still follow ordinary filesystem resolution.

### Concurrency

There is no destination lock.

The last successful rename can win.

Do not claim safe same-path concurrent coordination.

### Testing

Cover:

- temp creation;
- callback failure;
- partial writes;
- sync failure;
- close failure;
- rename failure;
- directory sync/close failure;
- old destination preservation;
- new destination retention;
- symlink replacement;
- mode;
- stale temp absence;
- joined cleanup errors.

---

## Logging changes

Read [`docs/logging.md`](docs/logging.md).

### Two phases

Bootstrap logging is text on stderr.

Configured logging uses the configured format and writes to:

```text
stderr
persistent file
```

Early failures are not present in the configured file.

### Persistent file is mandatory

The current configured logger always opens a persistent file.

There is no supported stderr-only mode.

Do not use `/dev/null` as a disabling mechanism.

### File behavior

The logger:

- creates missing directories;
- opens in append mode;
- forces mode `0640`;
- returns a closer;
- reports deferred close failure through bootstrap stderr.

### Event compatibility

A lifecycle message or field can be consumed by operators and automation.

Changing:

- message;
- level;
- field name;
- field type;
- emission stage

can be a compatibility change.

### Privacy

Do not log:

- API key;
- authorization header;
- full module intermediate text by default;
- secret file contents;
- arbitrary environment dumps.

`logging.logText` controls documented top-level input and output text fields.

It is not permission for every module to log complete text.

### Runtime logging errors

Ordinary `slog` calls do not change command status when a writer later fails.

Do not document log delivery as transactional.

### Tests

Test both:

```text
text
json
```

Verify semantic fields rather than exact timestamps.

---

## CLI and exit-code changes

Read:

- [`docs/cli.md`](docs/cli.md)
- [`docs/errors-and-exit-codes.md`](docs/errors-and-exit-codes.md)

### CLI compatibility

Changing any of the following is public interface work:

- flag name;
- required status;
- default value;
- normalization;
- accepted value;
- positional-argument behavior;
- stdout/stderr behavior;
- help text;
- input precedence;
- output requirement.

### Output format

The CLI accepts:

```text
wav
mp3
opus
ogg
```

`ogg` normalizes to `opus`.

The filename extension does not infer format.

### Exit-code stages

Current application-defined statuses are:

| Code | Stage |
|---:|---|
| `0` | help or complete success |
| `1` | bootstrap logging or request ID |
| `2` | invocation, config, setup, modules, or input |
| `3` | pipeline, request, secret, or Fish client setup |
| `4` | Fish synthesis, streaming, or output publication |

A new return path must be classified deliberately.

### Error severity is not status

Preserve these distinctions:

- module ERROR can recover to exit `0`;
- missing secret WARN returns exit `3`;
- log close ERROR does not change selected status;
- exit `4` can coexist with a published output.

### Help

Help writes usage to stdout and returns `0`.

It does not load config or open configured logging.

---

## Testing expectations

Read [`docs/testing.md`](docs/testing.md).

### Test the contract

A test should protect observable behavior.

Depending on the subsystem, assert:

- return value;
- error identity;
- typed fields;
- file content;
- file type;
- file mode;
- cleanup;
- old destination;
- new destination;
- request method;
- request path;
- headers;
- JSON body;
- retry count;
- body close;
- cancellation;
- pipeline report;
- later-step calls;
- process exit code;
- logs.

### Regression tests

A bug fix should include a test that fails before the fix.

The test should reproduce the smallest meaningful scenario.

### Same-package tests

Current tests generally use the implementation package.

Use same-package tests when internal lifecycle details are part of the maintenance contract.

Do not export production symbols only for tests.

### Table-driven tests

Use table-driven subtests for shared contracts.

Give every case a descriptive name.

### Parallel tests

Use `t.Parallel()` only for isolated tests.

Do not parallelize tests that mutate:

- `os.Args`;
- cwd;
- environment variables;
- package globals;
- shared registry state;
- signal behavior.

### Temporary directories

Use `t.TempDir()`.

Do not write test artifacts into the repository.

### HTTP

Use local test servers or custom transports.

Never depend on the live Fish API in the normal suite.

### Timing

Avoid arbitrary sleeps.

Use:

- channels;
- cancellation;
- injected delay;
- synchronization;
- finite safety timeouts.

### Error assertions

Prefer:

```go
errors.Is(err, expected)
errors.As(err, &target)
```

Do not compare complete wrapped strings unless exact formatting is the intended contract.

### Typed nil

Test interface boundaries with typed-nil implementations where relevant.

### Fuzzing and coverage

Coverage, shuffle, fuzzing, benchmarks, and repetition are useful local tools.

They are not current CI gates unless the workflow is changed.

---

## Documentation changes

Documentation is written in English.

Use clear technical language and preserve project terminology.

### Documentation hierarchy

- `README.md` is the public overview and quick start.
- `docs/index.md` is the navigation and authority map.
- subsystem files are normative references.
- `docs/troubleshooting.md` is practical diagnosis.
- this file defines contribution workflow.

### Update matrix

| Code change | Documentation to review |
|---|---|
| CLI | README, CLI, errors, troubleshooting |
| Config field/default | config example, configuration, owning subsystem, troubleshooting |
| Pipeline | pipeline, modules, logging, errors, testing |
| Module lifecycle | architecture, modules, author guide, testing |
| New module | configuration, modules, author guide, testing, README when user-visible |
| Fish request/retry | configuration, Fish, errors, troubleshooting, testing |
| Secret/path | configuration, secrets and paths, troubleshooting, testing |
| Output | output files, errors, troubleshooting, testing |
| Logging | configuration, logging, errors, troubleshooting, testing |
| Exit code | CLI, errors, troubleshooting, README |
| CI | testing, CONTRIBUTING, README |
| Security behavior | owning subsystem, troubleshooting, tests, security policy |

### Keep summaries short

Do not copy entire normative references into README or the documentation index.

Link to the owning document.

### Exact claims

Every exact claim should be traceable to:

- code;
- test;
- config example;
- CI;
- another authoritative project artifact.

Do not document planned behavior as implemented.

### Provider facts

Do not promise current:

- model availability;
- price;
- quota;
- free access;
- account permission.

Those are provider-controlled and can change.

### Shell examples

Every shell code block in project documentation should begin with:

```text
clear
```

Commands should:

- avoid exposing credentials;
- use safe quoting;
- avoid destructive shortcuts;
- show expected paths clearly;
- not recommend `chmod -R 777`;
- not disable TLS verification;
- not use `/dev/null` as unsupported logger configuration.

### Links

Use relative links for repository files.

Verify every local target exists.

### Markdown quality

Before committing:

```bash
clear

git diff --check
```

Also inspect:

- heading hierarchy;
- balanced fences;
- trailing whitespace;
- duplicate sections;
- contradictory claims;
- stale examples.

### Documentation audit

Large documentation changes should be reviewed against code and tests, not only proofread for grammar.

---

## Dependencies

The current module has no third-party requirements.

Adding one is a design decision.

A dependency proposal should explain:

- why the standard library is insufficient;
- maintenance status;
- license;
- security history;
- transitive dependencies;
- binary-size impact;
- build impact;
- test impact;
- API stability;
- whether a small local implementation would be clearer.

Do not add a dependency solely to reduce a small amount of straightforward code.

### Updating Go version

A Go version change affects:

- `go.mod`;
- CI;
- README;
- testing docs;
- contribution docs;
- supported developer environment.

Run the full suite with the new toolchain.

---

## Security-sensitive changes

Security-sensitive areas include:

- secret file handling;
- path resolution;
- symlink behavior;
- permissions;
- HTTP authorization;
- header validation;
- log redaction;
- provider error bodies;
- output paths;
- temporary files;
- directory trust;
- configuration trust;
- dependency additions.

A security-sensitive change should include:

- threat or misuse description;
- explicit trust boundary;
- tests for failure paths;
- no real credentials;
- no sensitive logs;
- documentation update;
- backward-compatibility analysis.

### Credentials

Never commit:

- Fish API keys;
- access tokens;
- private provider identifiers that grant access;
- production secrets;
- logs containing credentials.

If a credential is exposed:

1. revoke or rotate it;
2. remove it from the working tree;
3. assess Git history and CI output;
4. do not treat deleting the latest file as sufficient.

### Public reports

Do not disclose an exploitable vulnerability in a public issue before the maintainer has a chance to assess it privately.

The dedicated repository security policy will define the reporting route.

---

## Commit messages

The repository uses short type-prefixed commit subjects.

Common prefixes include:

```text
feat:
fix:
refactor:
test:
docs:
```

Examples:

```text
fix: preserve directory sync errors
test: harden Fish HTTP lifecycle
docs: document output file behavior
refactor: secure secret lifecycle
```

### Subject guidance

Use:

```text
type: concise result
```

Prefer:

- imperative or result-oriented wording;
- one logical change;
- a specific subsystem or outcome;
- lowercase after the prefix unless a proper name requires capitals.

Avoid:

```text
updates
fix stuff
work
misc changes
final
try again
```

### Enforcement

The prefix style is current repository practice.

It is not currently enforced by a commit-lint tool.

A maintainer can request a clearer message before merge.

### Commit scope

Keep commits reviewable.

A good commit should have one coherent reason to exist.

Do not hide behavior changes inside:

- mass formatting;
- unrelated renaming;
- generated documentation churn;
- dependency updates.

---

## Pull requests

A pull request should be focused and self-contained.

### Title

Use the same concise style as commit subjects where practical.

Example:

```text
fix: preserve output after directory sync failure
```

### Description

Explain:

1. the problem;
2. current incorrect or missing behavior;
3. intended behavior;
4. implementation approach;
5. tests;
6. side effects;
7. compatibility impact;
8. documentation changes.

### Include evidence

Useful evidence includes:

- failing test before fix;
- passing focused test after fix;
- full suite result;
- race result;
- before/after error chain;
- filesystem state;
- synthetic HTTP request count;
- benchmark result for a performance claim.

Do not include:

- API keys;
- private input text;
- production audio;
- unreviewed environment dumps;
- provider account details;
- sensitive logs.

### Draft pull requests

A draft pull request is useful for:

- early architecture feedback;
- complex lifecycle changes;
- new module design;
- cross-platform filesystem work;
- dependency discussion.

Do not present an incomplete draft as ready for final review.

### Scope

Split unrelated work.

Examples that should usually be separate:

- behavior fix and broad naming cleanup;
- new module and CI redesign;
- docs rewrite and dependency migration;
- retry behavior and output-file refactor.

---

## Pull-request checklist

Before requesting review, verify:

### Repository state

- [ ] The change is based on a current `main`.
- [ ] No real secret, runtime config, log, binary, or generated audio is staged.
- [ ] The diff contains only intended files.
- [ ] `git diff --check` passes.

### Code

- [ ] Go code is formatted.
- [ ] Package ownership remains clear.
- [ ] Errors preserve identity.
- [ ] Cancellation remains detectable.
- [ ] Interface typed-nil cases are handled where relevant.
- [ ] Security-sensitive data is not logged.
- [ ] Side effects and cleanup are deliberate.

### Tests

- [ ] Focused package tests pass.
- [ ] A regression test exists for a bug fix.
- [ ] Boundary values are covered.
- [ ] Failure side effects are asserted.
- [ ] `go test -count=1 ./...` passes.
- [ ] `go test -race -count=1 ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] The command builds with `-trimpath`.

### Documentation

- [ ] Exact config changes are documented.
- [ ] The config example is updated when required.
- [ ] The owning subsystem document is updated.
- [ ] Cross-cutting error, logging, testing, and troubleshooting docs were reviewed.
- [ ] README remains a summary rather than a duplicate reference.
- [ ] Local links are valid.
- [ ] No planned behavior is described as implemented.

### Compatibility

- [ ] CLI impact is stated.
- [ ] Config impact is stated.
- [ ] Exit-code impact is stated.
- [ ] File/path/permission impact is stated.
- [ ] Retry or remote-duplicate risk is stated.
- [ ] Migration instructions are included when needed.

---

## Review checklist

Reviewers should evaluate more than whether the happy path works.

### Architecture

- Does the behavior belong in this package?
- Is the interface narrower than the implementation?
- Does core remain independent of module internals?
- Is lifecycle ownership clear?
- Is a new global state introduced?
- Is the abstraction solving a current problem?

### Correctness

- Are bounds correct at minimum and maximum?
- Are exact byte limits tested?
- Are UTF-8 and whitespace rules explicit?
- Are errors wrapped with `%w`?
- Are simultaneous failures joined?
- Are response bodies and files closed?
- Are rollback and cleanup correct?
- Is post-publication state handled correctly?

### Security

- Can the change leak the API key?
- Can it follow or replace an unexpected symlink?
- Can untrusted configuration escape a trusted directory?
- Are permissions enforced correctly?
- Does a log contain sensitive text?
- Does a provider body enter logs?
- Does a dependency increase risk?

### Concurrency

- Is shared state synchronized?
- Are tests race-clean?
- Are global mutations kept out of parallel tests?
- Can two invocations target the same resource?
- Is last-writer behavior documented?

### External HTTP

- Is the request path correct?
- Are headers validated?
- Are bodies closed?
- Are retries bounded?
- Is cancellation respected?
- Can retry duplicate provider work?
- Does partial streaming prevent retry?

### Filesystem

- What happens before rename?
- What happens after rename?
- Is the old file preserved?
- Is the new file retained when required?
- Are temp files removed?
- Are cleanup failures preserved?
- Are modes tested?
- Is the parent directory trusted?

### Public contract

- Did a default change?
- Did a JSON name change?
- Did a log event change?
- Did an exit status move?
- Did a CLI flag change?
- Did output behavior change?
- Are migrations documented?

### Documentation

- Do code, tests, config example, README, index, and subsystem docs agree?
- Does the documentation distinguish implemented behavior from plans?
- Are exact claims supported?
- Are troubleshooting instructions safe?

---

## After review

Address review feedback with focused follow-up commits.

Do not:

- rewrite published branch history without coordination;
- force-push over another contributor’s work;
- combine unrelated requested changes;
- mark a thread resolved without addressing the underlying point.

After all changes:

```bash
clear

gofmt -w .
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...

go build \
  -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli

git diff --check
git status --short
```

Review the final combined diff, not only the last commit.

---

## Maintainer merge considerations

Before merging, a maintainer should confirm:

- the branch is based on the intended target;
- CI is green;
- review comments are resolved;
- public behavior is deliberate;
- documentation agrees;
- no secret or private artifact is present;
- commit history is understandable;
- release notes are updated when the change is user-visible and the project release process requires it.

The repository currently has no documented requirement for:

- a particular merge strategy;
- signed commits;
- a mandatory issue;
- a mandatory branch name;
- a commit-lint gate;
- a coverage percentage.

Do not claim these controls exist unless repository configuration and documentation are updated.

---

## Contribution invariants

The following rules describe the current contribution contract.

1. Go `1.26.5` is declared by the module.
2. CI reads the version from `go.mod`.
3. CI checks `gofmt`.
4. CI runs `go vet ./...`.
5. CI runs uncached tests.
6. CI runs the race detector.
7. CI builds the real command target with `-trimpath`.
8. Normal tests do not require the live Fish API.
9. Real credentials never belong in tests or commits.
10. Bug fixes should include regression tests.
11. Errors should preserve causes.
12. Simultaneous relevant failures should be joined.
13. Cancellation identity should survive wrapping.
14. Interface boundaries should consider typed nil.
15. Filesystem tests should assert side effects.
16. HTTP tests should assert request and response lifecycle.
17. Module config belongs to the module.
18. Every module prepares before any processor is built.
19. Failed module mutations roll back.
20. Cancellation is not pipeline recovery.
21. Config decoding remains strict.
22. Defaults remain valid.
23. Secret leaf symlinks remain rejected.
24. Secret file mode remains `0600`.
25. Output publication remains atomic at rename.
26. Post-rename persistence failure retains the new output.
27. Configured logging targets stderr and file.
28. Persistent logging cannot currently be disabled.
29. CLI exit codes remain stage-oriented.
30. Documentation changes accompany behavior changes.
31. README remains a public summary.
32. Specialized docs own exact behavior.
33. Provider availability is not promised by the project.
34. Dependencies require explicit justification.
35. Pull requests remain focused and reviewable.

Changing one of these invariants requires deliberate code, test, and documentation updates.

---

## Questions

For an implementation question, begin with [`docs/index.md`](docs/index.md) and the owning subsystem reference.

For a failure reproduction, use [`docs/troubleshooting.md`](docs/troubleshooting.md).

For a new module, follow [`docs/module-author-guide.md`](docs/module-author-guide.md).

For a suspected security vulnerability, do not publish exploit details in a public issue. Use the private reporting process described by the repository security policy when available.
