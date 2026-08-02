# Changelog

All notable changes to `fish-audio-cli` will be documented in this file.

The project is currently in alpha development and has not published a tagged release.

The current development state is collected under [`Unreleased`](#unreleased).

This changelog follows the structure of [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and intends to use [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once versioned releases begin.

> [!NOTE]
> Git history records every implementation step. This file records user-visible behavior, compatibility changes, security-relevant changes, and major maintainer-facing contracts. It is not a second copy of `git log`, because humans deserve at least one summary that was designed for reading.

## Release status

| Release | Date | Status |
|---|---|---|
| Unreleased | Not released | Active alpha development |
| Tagged releases | None | No version has been published |

No commit in the current history should be interpreted as a released version merely because it introduced a complete feature.

## Links

- Project overview: [`README.md`](README.md)
- Contribution guide: [`CONTRIBUTING.md`](CONTRIBUTING.md)
- Security policy: [`SECURITY.md`](SECURITY.md)
- Documentation index: [`docs/index.md`](docs/index.md)
- Complete configuration reference: [`docs/configuration.md`](docs/configuration.md)
- Troubleshooting: [`docs/troubleshooting.md`](docs/troubleshooting.md)

---

## Unreleased

### Release state

- The project remains in alpha development.
- No stable compatibility guarantee has been declared.
- No release artifact, tag, or versioned migration boundary has been published.
- Current upstream security support applies to the `main` branch as described in [`SECURITY.md`](SECURITY.md).
- The currently registered built-in module type is `passthrough`.
- The current Go module declares Go `1.26.5`.

### Added

#### Command-line application

- Added a standalone, one-shot Fish Audio text-to-speech command.
- Added text input through `--text`.
- Added text input through standard input when `--text` is omitted or exactly empty.
- Added required caller-selected output through `--output`.
- Added explicit output selection through `--format`.
- Added support for:
  - WAV;
  - MP3;
  - Opus.
- Added `ogg` as a CLI alias that normalizes to Opus.
- Added help output through `--help`.
- Added rejection of positional arguments.
- Added stage-oriented process exit codes:
  - `0` for help or success;
  - `1` for bootstrap logger or request-ID failure;
  - `2` for invocation, configuration, initialization, module, or input failure;
  - `3` for pipeline, request construction, secret, or Fish client initialization failure;
  - `4` for Fish synthesis, response streaming, or output publication failure.

#### Configuration

- Added one JSON configuration file per invocation.
- Added built-in defaults for all core configuration groups.
- Added a tracked example configuration at `config/config.example.json`.
- Added strict UTF-8 configuration loading.
- Added strict single-value JSON decoding.
- Added duplicate-object-key rejection, including escaped-equivalent keys.
- Added unknown-field rejection.
- Added exact JSON field-name matching.
- Added explicit-null validation where omission and `null` have different meaning.
- Added semantic validation for:
  - input limits;
  - pipeline definitions;
  - module names;
  - module types;
  - module error policies;
  - Fish endpoint;
  - Fish request values;
  - retry settings;
  - secret settings;
  - logging settings.
- Added bounded configuration file reads.
- Added tests keeping the tracked example configuration aligned with built-in defaults.

#### Text input

- Added bounded text reads.
- Added UTF-8 validation.
- Added rejection of empty and whitespace-only text.
- Added a shared text contract used by CLI input, pipeline input, module output, and Fish request construction.
- Added byte-limit enforcement while reporting text lengths as UTF-8 rune counts where documented.

#### Pipeline

- Added an ordered text-processing pipeline.
- Added support for zero or more configured module instances.
- Added strict execution in configuration array order.
- Added independent module instance names.
- Added support for repeating one module type under different names and configurations.
- Added pipeline-wide default failure policy.
- Added per-instance failure-policy override.
- Added the following policies:
  - `use_previous`;
  - `use_original`;
  - `skip`;
  - `abort`.
- Added rollback of failed module changes.
- Added rollback of interrupted module changes.
- Added cancellation handling that bypasses successful fallback.
- Added validation of successful module output.
- Added pipeline execution reports.
- Added step counts, outcomes, durations, and text-length metadata.
- Added construction-time validation of pipeline steps and dependencies.

#### Module system

- Added a compiled-in module registry.
- Added separation between module type and configured module instance.
- Added module-owned strict configuration decoding.
- Added module preparation before processor construction.
- Added the invariant that every configured module prepares before any processor is built.
- Added independent processor builders.
- Added rejection of nil and typed-nil processors.
- Added the built-in `passthrough` module.
- Added support for empty pipelines that return input text unchanged.
- Added module lifecycle logging.
- Added a complete module-author guide.

#### Fish Audio request support

- Added configurable Fish API base URL.
- Added endpoint construction by joining the configured base path with `/v1/tts`.
- Added validation of endpoint:
  - scheme;
  - host;
  - user information;
  - query;
  - fragment.
- Added Fish API key authentication through a bearer authorization header.
- Added model selection through the Fish model header.
- Added optional reference voice selection.
- Added configurable request fields for:
  - temperature;
  - top-p;
  - prosody speed;
  - prosody volume;
  - loudness normalization;
  - chunk length;
  - normalization;
  - sample rate;
  - MP3 bitrate;
  - Opus bitrate;
  - latency;
  - maximum new tokens;
  - repetition penalty;
  - minimum chunk length;
  - previous-chunk conditioning;
  - early-stop threshold;
  - feature list.
- Added local validation of Fish request fields.
- Added rejection of non-finite floating-point request values.
- Added rejection of invalid synthesis text.
- Added validation of request strings, model, reference ID, and feature values.
- Added isolation of pointer-backed sample-rate request state.
- Added configurable HTTP timeout.
- Added timeout-overflow protection.

#### Fish API errors

- Added typed Fish API errors.
- Added stable categories for common HTTP outcomes.
- Added preservation of:
  - HTTP status code;
  - HTTP status;
  - Fish API status;
  - provider message.
- Added bounded non-success response bodies.
- Added JSON provider-error parsing.
- Added fallback to bounded plain-text provider errors.
- Added preservation of both API status and response-body read failure.
- Added detection of successful HTTP responses with empty audio bodies.
- Added response-body closure on success, API failure, retry, and stream failure.

#### Fish API retries

- Added configurable maximum attempts.
- Added configurable initial retry delay.
- Added configurable maximum retry delay.
- Added optional retry of Fish `5xx` server errors.
- Added unconditional retry classification for Fish `429` rate limits.
- Added exponential backoff.
- Added `Retry-After` parsing for:
  - integer seconds;
  - HTTP dates.
- Added context-aware retry waits.
- Added rejection of invalid retry configuration.
- Added bounded maximum attempts and delays.
- Added behavior that stops retry when a valid `Retry-After` exceeds the configured maximum delay.
- Added tests for retry count, delay selection, cancellation, and response closure.

#### Secrets

- Added Fish API key loading from a separate configured file.
- Added project-relative default secret path.
- Added secure creation of a missing secret file.
- Added `ErrFileCreated` signaling after a missing file is created.
- Added empty-file first-run provisioning behavior.
- Added parent directory creation for a missing secret path.
- Added final-directory write-bit checks.
- Added rejection of group-writable and other-writable final secret directories.
- Added regular-file enforcement.
- Added leaf-symlink rejection.
- Added same-file verification after opening the secret.
- Added forced secret file mode `0600`.
- Added bounded secret reads.
- Added UTF-8 validation.
- Added exactly-one-line validation.
- Added support for one final LF or CRLF.
- Added rejection of surrounding whitespace.
- Added preservation of close failures through joined errors.

#### Path resolution

- Added centralized lexical project-path resolution.
- Added project-directory derivation from the lexical configuration path.
- Added special handling when the immediate configuration parent directory is named `config`.
- Added project-relative resolution for:
  - secret file;
  - persistent log file;
  - module-owned paths that use the project resolver.
- Added direct current-working-directory semantics for CLI config and output paths.
- Added explicit behavior for absolute paths.
- Added documentation that project path resolution does not:
  - expand `~`;
  - expand environment variables;
  - resolve symlink targets;
  - confine `..`.

#### Logging

- Added bootstrap text logging to stderr.
- Added request IDs generated per invocation.
- Added structured configured logging.
- Added text log format.
- Added JSON log format.
- Added configurable thresholds:
  - debug;
  - info;
  - warn;
  - error.
- Added simultaneous configured output to:
  - stderr;
  - persistent log file.
- Added project-relative default log path.
- Added automatic log-directory creation.
- Added append-only log-file opening.
- Added forced persistent log mode `0640`.
- Added configurable top-level text logging.
- Added default omission of input and processed text.
- Added lifecycle events for configuration, input, pipeline, modules, synthesis, and failures.
- Added fallback reporting of deferred log close failures through bootstrap stderr.
- Added continued writes to remaining fan-out destinations when one writer fails.
- Added a logrotate template.

#### Output files

- Added callback-based Fish response streaming.
- Added same-directory temporary output files.
- Added unique temporary filename generation.
- Added temporary file mode `0600`.
- Added direct streaming into the temporary file.
- Added temporary-file synchronization.
- Added temporary-file close before publication.
- Added atomic rename to the final destination.
- Added containing-directory synchronization after rename.
- Added containing-directory close handling.
- Added preservation of an existing destination before successful rename.
- Added retention of newly published output after a post-rename directory persistence error.
- Added temporary-file cleanup on pre-publication failure.
- Added preservation of primary, close, and cleanup errors through joining.
- Added destination leaf-symlink replacement behavior.
- Added support for concurrent invocations using independent output paths.

#### Tests and CI

- Added package-local unit and integration tests.
- Added end-to-end command tests with:
  - temporary configuration;
  - temporary secret;
  - local Fish server;
  - output publication checks;
  - exit-code checks.
- Added exact boundary tests for bounded reads.
- Added strict JSON tests.
- Added module registry lifecycle tests.
- Added pipeline rollback and policy tests.
- Added Fish request and HTTP lifecycle tests.
- Added secret permission, type, symlink, race, line, and size tests.
- Added logging fan-out, mode, append, path, and close tests.
- Added atomic output failure-injection tests.
- Added typed-nil dependency tests.
- Added cancellation tests.
- Added response-body closure tests.
- Added race-detector coverage.
- Added GitHub Actions CI for:
  - formatting;
  - vet;
  - uncached tests;
  - race detector;
  - command build.
- Added build verification using `-trimpath`.

#### Documentation and project governance

- Added a normative architecture document.
- Added a normative pipeline document.
- Added a normative module-system document.
- Added a module-author guide.
- Rewrote the complete configuration reference.
- Added a command-line interface reference.
- Added a Fish Audio integration reference.
- Added a logging reference.
- Added a secrets and path-resolution reference.
- Added an atomic output reference.
- Added an errors and exit-codes reference.
- Added a complete testing strategy.
- Added a practical troubleshooting guide.
- Added a documentation index and authority map.
- Rewrote the project README as a public overview and quick start.
- Added a contribution guide.
- Added a security policy.
- Added MIT licensing.
- Added repository ignore rules for runtime artifacts.

### Changed

#### Architecture and lifecycle

- Centralized strict JSON behavior in a shared package.
- Centralized bounded reads in a shared package.
- Centralized lexical project-path resolution.
- Moved module support checks from generic configuration validation into the module registry.
- Separated module preparation from processor construction.
- Deferred Fish client initialization until after:
  - configuration;
  - modules;
  - input;
  - pipeline;
  - request construction;
  - secret loading.
- Restored module initialization to occur before input and secret side effects.
- Made final app and pipeline construction reject invalid and typed-nil dependencies.
- Removed unused LLM configuration and runtime initialization from the current core.
- Kept Fish request defaults in configuration rather than duplicating them in lower layers.
- Changed pipeline execution to return structured reports.
- Unified text validation across input, modules, pipeline, app, and Fish request boundaries.

#### Error behavior

- Classified Fish client initialization failures under exit `3`.
- Preserved context cancellation and deadline identity through wrapped errors.
- Preserved cleanup and close errors rather than silently discarding them.
- Preserved Fish API status together with bounded error-body read failure.
- Preserved output directory sync and close errors.
- Made module errors recoverable according to configured pipeline policy while retaining failure logs.
- Kept selected command exit status unchanged by deferred log close failure.
- Improved error context at package and lifecycle boundaries.

#### Filesystem behavior

- Changed secret path resolution to use the derived project directory.
- Hardened existing secret files to `0600`.
- Hardened existing persistent log files to `0640`.
- Changed output publication to persist the rename through directory synchronization.
- Changed output cleanup so a post-rename failure does not remove the newly published file.
- Changed pre-rename cleanup to preserve the primary error and all relevant cleanup failures.
- Changed destination-error logging so remaining log destinations continue receiving records.

#### Validation

- Changed JSON decoding to require exact field names.
- Changed Fish client configuration to reject surrounding whitespace.
- Changed Fish client headers to reject invalid UTF-8 and ASCII controls.
- Changed Fish request values to reject NaN and infinity.
- Changed configured duration conversion to reject overflow.
- Changed configured read-limit validation to align exactly with bounded-reader behavior.
- Changed input and final synthesis text validation to use the shared text contract.
- Changed CLI parsing to reject positional arguments.
- Changed module registry construction to reject typed-nil processors and builders.

#### Documentation

- Replaced incomplete subsystem notes with normative references.
- Corrected earlier persistent-logging guidance that suggested `/dev/null`.
- Corrected exit-code documentation so secret failures belong to exit `3`.
- Corrected module-author examples to include required replacement fields.
- Established specialized documents as authoritative for exact subsystem behavior.
- Established `docs/index.md` as the detailed documentation entry point.

### Fixed

#### Configuration and input

- Fixed configuration files being read without a hard byte limit.
- Fixed off-by-one behavior between configured limits and bounded reads.
- Fixed permissive JSON field-name matching.
- Fixed acceptance of duplicate JSON keys.
- Fixed acceptance of multiple top-level JSON values.
- Fixed missing validation for invalid UTF-8.
- Fixed non-finite Fish request parameters reaching JSON serialization.
- Fixed Fish timeout conversion overflow.
- Fixed blank input reaching later lifecycle stages.
- Fixed invalid final synthesis text reaching the Fish request.
- Fixed positional CLI arguments being silently accepted.

#### Pipeline and modules

- Fixed failed processors leaving mutated document text.
- Fixed interrupted processors leaving mutated document text.
- Fixed module construction beginning before every configuration prepared successfully.
- Fixed unsupported module checks occurring in the wrong architectural layer.
- Fixed invalid or duplicate pipeline construction state.
- Fixed typed-nil processors and dependencies passing ordinary nil checks.
- Fixed module initialization ordering relative to input and secret side effects.
- Fixed pipeline reports missing failure and recovery state.

#### Fish HTTP

- Fixed invalid Fish base URLs reaching request creation.
- Fixed invalid API key or model header values reaching the HTTP transport.
- Fixed surrounding whitespace in Fish client settings being silently accepted.
- Fixed request strings with controls or invalid boundaries reaching Fish.
- Fixed shared pointer-backed sample-rate state between requests.
- Fixed missing typed API error categories.
- Fixed response bodies not being verified closed on every lifecycle path.
- Fixed retry responses not being closed before another attempt.
- Fixed oversized provider error bodies escaping bounded reads.
- Fixed retry waits not preserving cancellation correctly.
- Fixed server-error retries lacking explicit configuration.
- Fixed rate-limit handling lacking configurable bounded retry.

#### Secrets and paths

- Fixed relative secret paths resolving from process working directory instead of project directory.
- Fixed missing security checks for secret directory write permissions.
- Fixed secret leaf symlinks being accepted.
- Fixed non-regular secret objects being accepted.
- Fixed secret file replacement races not being detected.
- Fixed existing secret modes not being hardened.
- Fixed secret close failures being lost.
- Fixed absent secret provisioning being mixed with unrelated startup behavior.
- Fixed path resolution logic being duplicated across packages.

#### Logging

- Fixed log close failures being silently discarded.
- Fixed one failed fan-out destination preventing writes to later destinations.
- Fixed existing log files retaining unsafe permissions.
- Fixed typed-nil log writers reaching handler construction.
- Fixed documentation claiming persistent logging could be disabled with `/dev/null`.

#### Output publication

- Fixed successful rename not being followed by directory synchronization.
- Fixed directory sync errors being discarded.
- Fixed directory close errors being discarded.
- Fixed post-rename failure cleanup removing a newly published output.
- Fixed pre-rename cleanup losing the primary failure.
- Fixed temporary-file close and remove failures being hidden.
- Fixed atomic output cleanup not distinguishing pre- and post-publication state.
- Fixed tests that did not assert all publication side effects.

#### Exit codes and orchestration

- Fixed Fish initialization failures being reported under the wrong stage.
- Fixed Fish client initialization occurring before prerequisites.
- Fixed command tests not covering initialization order.
- Fixed help, input, module, secret, synthesis, and output paths lacking complete exit-code coverage.
- Fixed top-level errors missing request correlation and lifecycle fields.

### Security

- Added strict separation of API key material from JSON configuration.
- Added secret leaf-symlink rejection.
- Added same-file race detection for opened secrets.
- Added secret directory write-permission validation.
- Added forced secret file mode `0600`.
- Added forced output file mode `0600`.
- Added forced persistent log file mode `0640`.
- Added bounded reads for configuration, input, secrets, and provider errors.
- Added invalid UTF-8 rejection.
- Added duplicate-key and unknown-field rejection.
- Added API key and model header control-character rejection.
- Added Fish request string validation.
- Added finite-number validation for request fields.
- Added secure same-directory temporary output creation.
- Added cleanup that does not remove the final destination after publication.
- Added privacy defaults that omit top-level input and processed text from logs.
- Added a security policy defining:
  - supported versions;
  - private reporting;
  - vulnerability scope;
  - trust boundaries;
  - credential response;
  - safe research rules.
- Removed unsupported advice to use `/dev/null` as a log-disabling mechanism.
- Documented trusted configuration control over:
  - Fish endpoint;
  - secret path;
  - log path;
  - modules;
  - output behavior.
- Documented that modules are compiled trusted code and are not sandboxed.

### Removed

- Removed unused LLM configuration from the current core.
- Removed unconditional LLM secret initialization.
- Removed duplicated Fish defaults outside the configuration layer.
- Removed permissive JSON decoding behavior.
- Removed implicit acceptance of positional CLI arguments.
- Removed undocumented or misleading persistent-logging disable guidance.
- Removed assumptions that module types are validated by generic config code.
- Removed output cleanup behavior that could delete a newly published file after rename.

### Known limitations

The following are current design limitations, not release promises.

- Only the `passthrough` built-in module is currently registered.
- The command is one-shot and does not provide a daemon or local HTTP server.
- Modules are compiled into the binary and cannot be loaded dynamically.
- Audio cannot be streamed to stdout.
- The API key cannot be supplied inline, through a CLI flag, or through an environment fallback.
- JSON configuration does not expand home-directory or environment placeholders.
- The output parent directory is not created automatically.
- Persistent file logging cannot be disabled.
- Log rotation is external.
- There is no destination lock for concurrent same-path writers.
- There is no provider idempotency key.
- Transport errors are not internally retried.
- `5xx` retry is disabled by default.
- Successful audio bytes are not decoded or container-validated before publication.
- Output files are permission-protected but not encrypted at rest.
- The project currently enforces CI only on Linux `ubuntu-latest`.
- No stable release or backward-compatibility window has been declared.

See [`README.md`](README.md) and [`docs/index.md`](docs/index.md) for the current public scope.

---

## Initial development history

The repository began with:

```text
2026-07-31
Initial standalone Fish Audio TTS CLI
```

All changes since that initial commit are still part of unreleased alpha development.

The current history includes iterative hardening of:

- text validation;
- configuration;
- module architecture;
- pipeline rollback;
- Fish HTTP handling;
- retry behavior;
- secret lifecycle;
- path resolution;
- structured logging;
- atomic output;
- exit-code classification;
- tests;
- CI;
- documentation;
- security policy.

The Git history remains the authoritative source for individual commits.

This changelog summarizes the resulting behavior.

---

## Changelog maintenance

### What belongs here

Add an entry when a change affects:

- users;
- operators;
- integration authors;
- module authors;
- documented behavior;
- configuration compatibility;
- CLI compatibility;
- exit codes;
- output state;
- security boundaries;
- migration requirements;
- supported environments;
- release packaging.

Examples:

- new CLI flag;
- changed default;
- new module;
- changed error policy;
- fixed file-permission bug;
- new retry behavior;
- removed config field;
- security fix;
- new release artifact.

### What usually does not belong here

Do not add entries for every:

- test refactor;
- comment correction;
- internal rename;
- formatting-only change;
- helper extraction;
- temporary development commit;
- documentation typo with no meaning change.

Include such work only when it materially changes a public contract or fixes a misleading security or operational statement.

### Categories

Use these categories where applicable:

```text
Added
Changed
Deprecated
Removed
Fixed
Security
```

Do not create a category merely to avoid deciding what the change means.

### Entry style

Write the user-visible result.

Preferred:

```text
- Fixed post-rename cleanup so a directory-sync failure no longer removes
  the newly published output.
```

Avoid:

```text
- Updated output.go.
- Refactored helper.
- Fixed tests.
```

### Unreleased entries

Add new work under:

```text
## Unreleased
```

Do not create a version heading until a release is actually prepared.

### Release headings

When releases begin, use:

```text
## [0.1.0] - 2026-08-03
```

Use the real release date in `YYYY-MM-DD`.

Do not use:

- commit date;
- merge date;
- planned release date;
- local timezone guess

as the release date unless the release occurred then.

### Semantic Versioning

After versioned releases begin:

- increment patch for backward-compatible fixes;
- increment minor for backward-compatible features;
- increment major for incompatible stable-interface changes.

During `0.y.z` development, compatibility can still change, but breaking changes must be identified clearly.

### Breaking changes

Mark breaking changes explicitly.

Include:

- affected interface;
- old behavior;
- new behavior;
- migration;
- removal date when deprecated first;
- related documentation.

Example:

```text
### Changed

- **Breaking:** Renamed `fish.oldField` to `fish.newField`.
  Existing configuration files must update the key before upgrading.
```

### Security entries

Security entries should state:

- affected boundary;
- impact in non-exploit detail;
- fixed versions or commits;
- required operator action;
- credential rotation when relevant.

Do not publish weaponized exploit instructions before coordinated disclosure.

### Historical accuracy

Do not rewrite a released section to describe current behavior.

A changelog entry describes what changed in that release.

Later corrections belong in a later section.

### Links

When tags exist, add comparison links at the bottom.

Example shape:

```text
[Unreleased]: https://github.com/piqnyx/fish-audio-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/piqnyx/fish-audio-cli/releases/tag/v0.1.0
```

Do not add links to tags that do not exist.

### Release checklist

Before converting `Unreleased` into a version:

1. choose the version;
2. confirm release scope;
3. run full CI-equivalent verification;
4. audit configuration defaults and example;
5. audit CLI help;
6. audit documentation;
7. identify breaking changes;
8. add migration instructions;
9. update supported-version policy when needed;
10. set the actual release date;
11. create the release commit;
12. create the signed or annotated tag according to the adopted release policy;
13. publish release artifacts if the project distributes them;
14. add comparison links;
15. create a fresh empty `Unreleased` section.

No tag or release process is currently declared beyond these future requirements.

---

## Maintainer verification

Before committing a changelog change:

```bash
clear

git diff --check
git status --short
```

When the changelog accompanies code, run the complete required suite:

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

See [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`docs/testing.md`](docs/testing.md).

---

## Summary

Current state:

```text
release: none
status: alpha
changelog section: Unreleased
supported upstream line: main
```

The project already has a functional and hardened core, but no tagged release has converted that work into a stable compatibility boundary.

Until the first release:

- add notable work to `Unreleased`;
- do not invent version numbers;
- do not assign release dates retroactively;
- distinguish user-visible changes from internal commit noise;
- keep security details responsible;
- keep documentation aligned with code and tests.
