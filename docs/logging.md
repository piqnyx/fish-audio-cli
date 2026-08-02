# Logging

> **Document status:** normative description of the current pre-release logging behavior.
>
> **Audience:** operators collecting logs, service integrators, maintainers changing lifecycle events, and developers reviewing privacy or observability boundaries.
>
> **Scope:** this document describes bootstrap and configured loggers, request correlation, destinations, file handling, levels, formats, lifecycle events, module logs, sensitive-text policy, failure behavior, rotation, and compatibility constraints. Configuration fields are documented in [`configuration.md`](configuration.md); command behavior in [`cli.md`](cli.md); pipeline execution in [`pipeline.md`](pipeline.md); overall ownership in [`architecture.md`](architecture.md).

---

## 1. Purpose

Logging provides structured diagnostics for one CLI invocation.

The current design has two phases:

```text
process starts
    ↓
bootstrap stderr logger
    ↓
request ID generation
    ↓
CLI, paths, config validation
    ↓
configured stderr + file logger
    ↓
modules, text processing, Fish synthesis
    ↓
log file close
```

The logging package owns:

- structured logger construction;
- request ID generation;
- text and JSON handlers;
- level parsing;
- log-file path resolution;
- parent-directory creation;
- append-mode file opening;
- file permission tightening;
- stderr/file fan-out.

The command owns:

- which lifecycle events are emitted;
- event severity;
- event field names;
- when the configured logger replaces the bootstrap logger;
- whether top-level text values are included;
- exit-code mapping.

The pipeline owns:

- standard per-module lifecycle logging through its decorator.

Modules do not receive a logger through the current module interface.

---

## 2. Two logging phases

The executable does not begin with the configured logger.

It first creates a minimal bootstrap logger, then opens the configured log destination only after configuration has been loaded and validated.

This separation is necessary because the configured path, level, and format do not exist until the configuration file is available.

---

## 3. Bootstrap logger

At process startup, the command creates:

```text
text format
INFO threshold
stderr only
```

Conceptually:

```go
logging.New(os.Stderr, slog.LevelInfo)
```

The bootstrap logger exists before:

- request ID generation;
- CLI parsing;
- project-path initialization;
- configuration loading;
- configuration validation;
- configured logger initialization.

### 3.1 Bootstrap destination

Bootstrap records are written only to:

```text
stderr
```

They are not copied into the persistent log file later.

The logger does not buffer early records for replay.

### 3.2 Bootstrap format

The bootstrap logger always uses the standard text handler.

It does not use:

- `logging.format`;
- `logging.level`;
- `logging.file`.

Those values are not available yet.

### 3.3 Bootstrap level

The bootstrap threshold is fixed at:

```text
INFO
```

Changing configured logging level does not change early startup logging.

### 3.4 Raw fallback

If bootstrap logger construction itself fails, the command writes a plain line directly with:

```go
fmt.Fprintf(os.Stderr, ...)
```

The message begins conceptually with:

```text
logging error:
```

This path:

- is not structured;
- has no request ID;
- uses exit status `1`.

It exists because a failed logger cannot report its own failure. Software does occasionally meet philosophy.

---

## 4. Request ID

After bootstrap logger construction, the command creates one request ID for the invocation.

The generator:

- reads 16 random bytes from `crypto/rand`;
- encodes them as hexadecimal;
- returns 32 hexadecimal characters.

Example shape:

```text
4f96f8fdbb3de79a2bc7ef2ac54876ac
```

### 4.1 Per-invocation identity

Each process invocation receives one request ID.

The same ID is attached to:

- early structured stderr records;
- configured stderr records;
- persistent file records;
- a deferred log-file close error.

### 4.2 No shared counter

The request ID does not depend on:

- process ID;
- timestamps;
- a global incrementing counter;
- a persistent state file.

Independent processes can therefore generate IDs without shared coordination.

### 4.3 Generation failure

If secure random generation fails:

- the bootstrap logger emits `request ID generation failed`;
- the error record has no request ID because generation did not succeed;
- the command exits with status `1`;
- configuration is not loaded.

### 4.4 Help still generates an ID

Request ID generation happens before CLI parsing.

Therefore:

```bash
fish-audio-cli --help
```

still performs request ID generation.

The normal help path emits no log record, prints usage to stdout, and exits `0`.

---

## 5. Early correlated logger

After request ID generation, the command derives:

```go
stderrLogger := logger.With("request_id", requestID)
```

This remains a text-format, stderr-only logger.

It handles failures in:

- option parsing;
- path initialization;
- configuration loading;
- configuration validation;
- configured logger initialization.

### 5.1 Early records are not persistent

If configuration loading fails, the persistent file is never opened.

The failure appears only on stderr.

Operators collecting only the configured file will miss early startup failures.

Service managers should capture stderr as well.

---

## 6. Configured logger

After configuration validation, the command opens the configured logger.

It uses:

```text
logging.level
logging.format
logging.file
project path resolver
```

The configured logger writes every record to:

- stderr;
- one persistent log file.

The same request ID is attached to the configured logger.

### 6.1 Transition point

The transition occurs after:

1. options parse successfully;
2. project paths initialize;
3. configuration loads;
4. configuration validates;
5. log path resolves;
6. parent directory exists;
7. log file opens;
8. log file permissions are set;
9. configured handler is created.

The first normal configured lifecycle record is:

```text
config loaded
```

### 6.2 No replay

Earlier stderr records are not copied into the file.

The persistent file contains records only from the configured phase.

### 6.3 Same request ID

The configured logger is created independently, then enriched with the already generated request ID.

This allows an operator to correlate:

- an early stderr record;
- later stderr records;
- file records

from the same invocation.

---

## 7. Destination fan-out

The configured logger uses a fan-out writer with destinations in this order:

1. stderr;
2. persistent log file.

Each encoded log payload is offered to both destinations.

### 7.1 Continue after one failure

The fan-out writer attempts every destination even when an earlier destination fails.

Example:

```text
stderr write fails
    ↓
file write is still attempted
```

The reverse also holds because both destinations are always visited.

### 7.2 Short writes

A destination that writes fewer bytes without returning an error is converted into:

```text
io.ErrShortWrite
```

### 7.3 Joined errors

When multiple destinations fail, their errors are joined.

The writer reports:

- the minimum byte count written by any destination;
- the joined destination errors.

### 7.4 Runtime error visibility limitation

The writer can return an error to the `slog` handler.

Ordinary calls such as:

```go
logger.Info(...)
logger.Warn(...)
logger.Error(...)
```

do not return an error to application code.

Consequently, after successful logger initialization, a later destination failure such as:

- disk full;
- broken stderr pipe;
- filesystem I/O error;
- short write

does not currently alter the CLI exit status through the logging call.

The fan-out still attempts the other destination, but the command does not have a runtime log-write failure channel.

### 7.5 Initialization failures are different

Failures while:

- resolving the path;
- creating the directory;
- opening the file;
- changing file permissions;
- creating the handler

are returned normally and stop startup with exit status `2`.

---

## 8. Persistent log path

The configured path is:

```text
logging.file
```

Default configured value:

```json
"file": ""
```

An empty or whitespace-only value selects:

```text
logs/fish-audio-cli.log
```

### 8.1 Trimming

The logging package trims surrounding whitespace from `logging.file`.

Therefore:

```json
"file": "   "
```

selects the default path.

A path intentionally beginning or ending with spaces cannot be represented through this option.

### 8.2 Relative paths

Relative log paths are resolved from the project directory derived from the configuration path.

Example:

```text
config file:
    /opt/fish-audio-cli/config/config.json

logging.file:
    logs/application.log

resolved log:
    /opt/fish-audio-cli/logs/application.log
```

### 8.3 Configuration outside `config`

Example:

```text
config file:
    /etc/fish-audio-cli/settings.json

logging.file:
    logs/application.log

resolved log:
    /etc/fish-audio-cli/logs/application.log
```

### 8.4 Absolute paths

Absolute paths are cleaned and used without rebasing.

Example:

```json
"file": "/var/log/fish-audio-cli/application.log"
```

### 8.5 Working directory

The log path is not generally resolved from the current working directory.

It follows the project path resolver.

This differs from the CLI `--output` path, which follows normal process working-directory semantics when relative.

---

## 9. Persistent log directory

Before opening the file, the logging package runs conceptually:

```go
os.MkdirAll(directory, 0o750)
```

### 9.1 Missing directories

Missing directories are requested with mode:

```text
0750
```

The operating-system umask may remove permission bits during creation.

### 9.2 Existing directories

Existing parent-directory permissions are not changed.

The logger does not:

- tighten an existing directory to `0750`;
- reject a group-writable existing directory;
- reject an other-writable existing directory;
- verify directory ownership.

Operators must secure pre-existing log directories separately.

### 9.3 Directory path failure

Startup fails when the parent path:

- cannot be created;
- is not traversable;
- contains a non-directory component;
- is denied by filesystem permissions.

The configured logger is not installed.

The bootstrap stderr logger reports:

```text
logger initialization failed
```

---

## 10. Persistent log file

The file is opened with:

```text
create if missing
write only
append
```

Conceptually:

```go
os.O_CREATE | os.O_WRONLY | os.O_APPEND
```

### 10.1 Append behavior

Existing content is preserved.

New records are appended.

The file is not truncated at startup.

### 10.2 Creation mode

A new file is requested with mode:

```text
0640
```

The package then explicitly applies:

```text
0640
```

with `Chmod`.

### 10.3 Existing file permissions

An existing file is also changed to:

```text
0640
```

on every successful open.

Example:

```text
before: 0666
after:  0640
```

### 10.4 Ownership

The logger does not change file owner or group.

The process must have sufficient permission to open and chmod the file.

### 10.5 Special files and symlinks

The logging path is opened through ordinary `os.OpenFile`.

The logging package does not currently perform the secret-loader protections for:

- symlinks;
- regular-file verification;
- same-file race checks;
- device-file rejection;
- FIFO rejection.

Operators should point `logging.file` at an ordinary trusted file path.

### 10.6 Do not use `/dev/null`

Persistent logging currently has no supported disable switch.

Do not configure:

```json
"file": "/dev/null"
```

The logger opens the path and then attempts to change its mode to `0640`.

Possible outcomes include:

- logger initialization failure because the process cannot chmod the device;
- dangerous permission changes when running with excessive privilege.

`/dev/null` is not a supported persistent-logging disable mechanism.

### 10.7 No automatic file creation on every record

The file is opened once during configured logger initialization and remains open for the invocation.

---

## 11. Log file close

The command defers file close after configured logger creation.

### 11.1 Normal close

At the end of the invocation, the file descriptor is closed.

### 11.2 Close failure

A close failure emits:

```text
log file closing failed
```

with fields:

```text
path
error
request_id
```

### 11.3 Close-error destination

The close error is emitted through the bootstrap stderr-only logger, not through the configured logger whose file is being closed.

This avoids attempting to report a file-close failure through the same file destination.

### 11.4 Exit status

A deferred log-file close failure does not change the already selected process exit status.

It is diagnostic only.

### 11.5 No explicit sync

The logging lifecycle does not call `Sync` on the log file before close.

Normal writes and operating-system close semantics apply.

This differs from atomic audio output, which has explicit synchronization requirements.

---

## 12. Formats

Supported configured formats:

```text
text
json
```

The public configuration requires exact lowercase values.

### 12.1 Text format

Text format uses Go’s standard `slog.TextHandler`.

Typical shape:

```text
time=... level=INFO msg="config loaded" request_id=... path=...
```

The handler supplies standard fields such as:

- time;
- level;
- message.

Additional structured attributes follow.

### 12.2 JSON format

JSON format uses Go’s standard `slog.JSONHandler`.

Typical shape:

```json
{
  "time": "...",
  "level": "INFO",
  "msg": "config loaded",
  "request_id": "...",
  "path": "..."
}
```

The actual output is one JSON object per log record.

### 12.3 No source field

The handler options currently set only the level.

Source-code location is not enabled.

Records do not intentionally include:

- source filename;
- source line;
- function name.

### 12.4 No custom field rewriting

The project does not currently customize standard `slog` keys.

It does not rename:

```text
time
level
msg
```

### 12.5 Public configuration exactness

Configuration validation accepts exactly:

```text
text
json
```

Although the internal constructor trims and lowercases its direct argument, users should not rely on that through JSON configuration because validation runs first.

These configured values are rejected:

```text
JSON
Text
<leading-space>json
json<trailing-space>
```

---

## 13. Levels

Supported configured levels:

```text
debug
info
warn
error
```

Default:

```text
info
```

### 13.1 Threshold semantics

The selected level is the minimum emitted severity.

| Configured level | Emitted severities |
|---|---|
| `debug` | debug, info, warn, error |
| `info` | info, warn, error |
| `warn` | warn, error |
| `error` | error |

### 13.2 Public configuration exactness

Configuration validation requires exact lowercase values.

Although the package-level parser trims and lowercases direct calls, JSON values such as:

```text
INFO
<leading-space>warn
error<trailing-space>
```

are rejected before logger initialization.

### 13.3 Bootstrap exception

The bootstrap logger always uses `INFO`.

A configured `error` threshold cannot suppress an early option-parsing error.

A configured `debug` threshold cannot enable early debug records.

### 13.4 Current lifecycle severities

The command currently emits normal lifecycle progress primarily at `INFO`.

Operational conditions use:

- `WARN` for a newly created empty secret file;
- `WARN` for module interruption;
- `ERROR` for failed stages.

The current core may emit few or no `DEBUG` records.

Selecting `debug` preserves the ability to receive debug records but does not manufacture additional events.

---

## 14. Request correlation

Every structured record after successful request ID generation contains:

```text
request_id
```

### 14.1 Concurrent invocations

When multiple CLI processes write to the same stderr collector or file, request ID distinguishes their records.

### 14.2 No process metadata

The logger does not automatically add:

- PID;
- parent PID;
- hostname;
- user;
- working directory;
- executable version.

Request ID is the primary built-in correlation field.

### 14.3 No cross-invocation trace

A new invocation receives a new request ID.

The CLI does not accept a caller-supplied correlation ID.

A wrapper needing end-to-end correlation must associate its own job identifier with:

- the process;
- destination path;
- surrounding service logs.

---

## 15. Character counts

Text lifecycle logs use:

```go
utf8.RuneCountInString(text)
```

Fields:

```text
input_chars
output_chars
```

### 15.1 Not byte counts

Character counts do not equal `input.maxBytes`.

Examples:

```text
ASCII: often 1 byte per rune
Cyrillic: commonly 2 bytes per rune
emoji: commonly 4 bytes per rune
```

### 15.2 Not grapheme counts

A user-perceived character may contain multiple Unicode code points.

The logged value counts runes, not grapheme clusters.

### 15.3 Valid text assumption

Input and pipeline output are validated UTF-8 before normal success logs.

---

## 16. Duration fields

Pipeline and module timings use integer milliseconds.

Fields include:

```text
duration_ms
pipeline_duration_ms
```

### 16.1 Truncation

Go’s `Duration.Milliseconds()` returns whole milliseconds.

A duration below one millisecond appears as:

```text
0
```

### 16.2 Monotonic measurement

Durations are measured inside process execution.

They are not wall-clock timestamps and should not be reconstructed by subtracting formatted record times when exact stage timing matters.

---

## 17. Sensitive-text policy

Configuration:

```text
logging.logText
```

Default:

```text
false
```

This flag controls top-level input and final processed-text values.

### 17.1 Disabled

When `false`, the command logs:

```text
input_chars
output_chars
```

but omits:

```text
input_text
output_text
```

### 17.2 Enabled

When `true`, the command adds:

```text
input_text
output_text
```

to the corresponding lifecycle records.

### 17.3 What is not logged

The core does not log intermediate module text values.

Even with `logText: true`, standard module logs contain counts, identity, duration, and errors, not every intermediate text snapshot.

### 17.4 No text transformation for logs

Logged text is the actual string selected at the top-level boundary.

The logger does not:

- trim it;
- redact it;
- truncate it;
- normalize it;
- remove newlines;
- mask names or numbers.

### 17.5 Structured escaping

Text and JSON handlers encode attribute strings according to `slog` behavior.

This prevents a newline inside the text value from becoming an unstructured raw write by the application, but log collectors must still treat the field as untrusted content.

### 17.6 Privacy recommendation

Keep:

```json
"logText": false
```

for:

- private messages;
- credentials accidentally present in text;
- personal data;
- regulated data;
- shared logging systems.

---

## 18. Secrets

The Fish API key value is never intentionally logged.

The application may log:

- secret description: `Fish API key`;
- secret file path;
- an action telling the operator to write one key line;
- a validation or I/O error.

It does not log:

- secret contents;
- Bearer header;
- raw authorization value.

### 18.1 Error-chain caution

Errors from external providers may contain bounded remote response text.

A custom endpoint could return sensitive text in its error body, which can then appear in:

```text
synthesis failed
```

Protect access to logs accordingly.

---

## 19. Configuration-loaded event

After configured logger initialization, the command emits:

```text
config loaded
```

Severity:

```text
INFO
```

Fields:

```text
request_id
path
pipeline_on_error
fish_model
pipeline_module_count
pipeline_module_names
pipeline_module_types
```

### 19.1 Path

`path` is the absolute cleaned configuration file path.

### 19.2 Pipeline metadata

Module names and types are emitted in configured order.

The record does not include:

- complete module configs;
- per-module `onError`;
- module-owned secrets;
- Fish request parameters.

### 19.3 Model disclosure

`fish_model` is logged.

Treat model identifiers as operational metadata.

### 19.4 No reference ID

The configured Fish reference ID is not included in this record.

---

## 20. Module initialization events

When module registry construction fails, the command emits:

```text
module initialization failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

When adding the logging decorator fails, it emits:

```text
module logging initialization failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

Application pipeline construction failure emits:

```text
application initialization failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

These configured-phase records go to stderr and the log file.

---

## 21. Text-processing start event

Before invoking the application pipeline, the command emits:

```text
text processing started
```

Severity:

```text
INFO
```

Always included:

```text
request_id
input_chars
```

Conditionally included when `logging.logText` is true:

```text
input_text
```

This record occurs after:

- configuration;
- logger setup;
- module initialization;
- input reading;
- signal-aware context creation.

---

## 22. Standard module start event

Each decorated module emits before calling its processor:

```text
module processing started
```

Severity:

```text
INFO
```

Fields:

```text
request_id
module_name
module_type
input_chars
```

It does not include:

- full text;
- configured error policy;
- module config;
- secrets.

---

## 23. Standard module completion event

After the processor returns successfully, context remains active, and output validates, the decorator emits:

```text
module processing completed
```

Severity:

```text
INFO
```

Fields:

```text
request_id
module_name
module_type
input_chars
output_chars
duration_ms
```

### 23.1 Successful log boundary

The decorator validates output before logging completion.

It does not emit a successful completion record for invalid output.

### 23.2 Pipeline validates again

The pipeline validates after the decorated processor returns as well.

The duplicate boundary prevents a false success log while preserving pipeline correctness without the decorator.

---

## 24. Standard module failure event

For an ordinary processor error or invalid output, the decorator emits:

```text
module processing failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
module_name
module_type
input_chars
duration_ms
error
```

The decorator does not apply fallback.

After it returns the error, the pipeline:

- rolls back text;
- applies `use_previous`, `use_original`, `skip`, or `abort`;
- records the step result;
- decides whether later modules run.

A module failure log can therefore be followed by an overall successful synthesis when the configured policy recovers.

---

## 25. Standard module interruption event

For context cancellation or deadline expiration, the decorator emits:

```text
module processing interrupted
```

Severity:

```text
WARN
```

Fields:

```text
request_id
module_name
module_type
input_chars
duration_ms
error
```

Cancellation is not converted into ordinary fallback.

The pipeline stops.

---

## 26. Text-processing failure event

When the application pipeline returns an error, the command emits:

```text
text processing failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
pipeline_outcome
steps_total
steps_executed
pipeline_duration_ms
error
```

No Fish request is created after this event.

### 26.1 Text omission

This failure record does not add `input_text` or a partially processed text field.

The earlier start record may contain `input_text` when text logging is enabled.

---

## 27. Text-processing completion event

When the pipeline completes without returning an error, the command emits:

```text
text processing completed
```

Severity:

```text
INFO
```

Always included:

```text
request_id
output_chars
pipeline_outcome
steps_total
steps_executed
pipeline_duration_ms
```

Conditionally included when `logging.logText` is true:

```text
output_text
```

### 27.1 Successful recovered outcomes

A completion record may contain outcomes such as:

```text
completed
recovered
stopped
```

depending on pipeline behavior.

A module error already logged at module level may coexist with successful top-level completion after fallback.

---

## 28. Fish request creation failure

When final text and format cannot form a valid Fish request, the command emits:

```text
Fish request creation failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

This can include a format/sample-rate incompatibility.

No API key is loaded after this failure.

---

## 29. Missing secret-file event

When the secret loader securely creates a missing empty Fish API key file, the command emits:

```text
empty secret file created
```

Severity:

```text
WARN
```

Fields:

```text
request_id
secret="Fish API key"
path
action="write exactly one API key line into this file"
```

The API key value is absent.

The command exits with status `3`.

---

## 30. Secret loading failure

For an existing but unreadable, empty, malformed, oversized, or insecure secret path, the command emits:

```text
Fish API key loading failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
path
error
action="write exactly one API key line into this file"
```

The command exits with status `3`.

---

## 31. Fish client initialization failure

When client construction fails, the command emits:

```text
Fish client initialization failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

Possible causes include invalid:

- API key header value;
- model header value;
- endpoint;
- timeout;
- error-body limit;
- retry options.

---

## 32. Synthesis start event

Before entering atomic output and Fish synthesis, the command emits:

```text
synthesis started
```

Severity:

```text
INFO
```

Fields:

```text
request_id
model
format
output_path
```

### 32.1 Output path disclosure

The destination path is logged.

Do not encode secrets into output filenames.

### 32.2 Model disclosure

The model name is logged again for direct synthesis-stage context.

### 32.3 No text field

Processed text is not included in this event.

---

## 33. Synthesis failure event

When Fish synthesis or atomic output publication fails, the command emits:

```text
synthesis failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

The current event does not include `output_path` as a separate field.

The error chain may itself contain path or remote API details.

The command exits with status `4`.

---

## 34. Synthesis completion event

After successful atomic publication, the command emits:

```text
synthesis completed
```

Severity:

```text
INFO
```

Fields:

```text
request_id
output_path
```

A successful completion record means the atomic output function returned successfully.

---

## 35. Early failure events

The following use the bootstrap stderr-only logger after request ID generation.

### Option parsing

```text
option parsing failed
```

Fields:

```text
request_id
error
```

Exit:

```text
2
```

### Path initialization

```text
path initialization failed
```

Fields:

```text
request_id
error
```

Exit:

```text
2
```

### Config loading

```text
config loading failed
```

Fields:

```text
request_id
error
```

Exit:

```text
2
```

### Config validation

```text
config validation failed
```

Fields:

```text
request_id
error
```

Exit:

```text
2
```

### Configured logger initialization

```text
logger initialization failed
```

Fields:

```text
request_id
error
```

Exit:

```text
2
```

These records never reach the configured persistent file because it is not usable yet.

---

## 36. Input failure event

Input reading happens after configured logger and module initialization.

Failure emits:

```text
input failed
```

Severity:

```text
ERROR
```

Fields:

```text
request_id
error
```

Examples:

- empty stdin;
- whitespace-only selected input;
- invalid UTF-8;
- input above `input.maxBytes`.

This event goes to stderr and the persistent file.

---

## 37. Event ordering

A normal successful invocation produces a sequence resembling:

```text
config loaded
text processing started
module processing started
module processing completed
text processing completed
synthesis started
synthesis completed
```

With multiple modules, each module contributes its start and terminal event in configured order.

### 37.1 No universal fixed count

Record count depends on:

- module count;
- logging threshold;
- module failures;
- fallback policy;
- missing secret initialization;
- Fish retries and failures;
- whether help was requested.

### 37.2 Retry attempts

The current command does not emit a distinct structured lifecycle record for each Fish retry attempt.

Retry information may be visible only through final error context unless lower-level logging is added later.

---

## 38. Field catalog

Common fields:

| Field | Meaning |
|---|---|
| `request_id` | 32-character invocation correlation ID |
| `error` | wrapped error value |
| `path` | configuration, secret, or log-related path depending on event |
| `input_chars` | input rune count |
| `output_chars` | output rune count |
| `input_text` | original selected text when enabled |
| `output_text` | final pipeline text when enabled |
| `module_name` | configured module instance name |
| `module_type` | registered module implementation type |
| `duration_ms` | module duration in whole milliseconds |
| `pipeline_duration_ms` | pipeline duration in whole milliseconds |
| `pipeline_outcome` | pipeline result category |
| `steps_total` | configured pipeline step count |
| `steps_executed` | number of recorded executed steps |
| `model` | Fish model used for synthesis |
| `format` | normalized output format |
| `output_path` | caller-selected destination path |
| `action` | operator remediation hint |

### 38.1 Field meaning is event-specific

A generic key such as `path` does not identify its domain by itself.

Consumers should use both:

- `msg`;
- field values.

---

## 39. Text format parsing

Text logs are human-readable structured lines.

They are not a stable whitespace-delimited protocol.

Values may be quoted or escaped.

Do not parse them with assumptions such as:

```text
the fourth token is always request_id
```

Use:

- a proper slog-compatible parser;
- JSON logging for machine ingestion;
- the message and named keys.

---

## 40. JSON format parsing

JSON logs are better suited to machine ingestion.

Each record is a JSON object emitted by `slog.JSONHandler`.

Consumers should:

- parse each line independently;
- tolerate field order changes;
- treat additional fields as compatible;
- key behavior on `msg`, `level`, and named attributes;
- preserve the raw record for diagnosis.

Do not depend on Go map ordering or visual field order.

---

## 41. Message strings as event identifiers

Current message strings include:

```text
config loaded
module initialization failed
module logging initialization failed
application initialization failed
input failed
text processing started
module processing started
module processing completed
module processing failed
module processing interrupted
text processing failed
text processing completed
Fish request creation failed
empty secret file created
Fish API key loading failed
Fish client initialization failed
synthesis started
synthesis failed
synthesis completed
log file closing failed
```

These messages currently function as human-readable event identifiers.

They are not represented by a separate stable event-code field.

Changing message text can affect log consumers that key on `msg`.

Such changes should be reviewed as observability compatibility changes.

---

## 42. Level filtering consequences

When configured level is:

```text
warn
```

normal progress records such as:

```text
config loaded
text processing started
module processing started
module processing completed
text processing completed
synthesis started
synthesis completed
```

are suppressed.

Warnings and errors remain.

When configured level is:

```text
error
```

the missing-secret warning and module interruption warning are suppressed.

The command still exits with the same status.

Logging threshold affects visibility, not control flow.

---

## 43. File logging cannot currently be disabled

The configured logger always requires:

- stderr destination;
- persistent file destination.

There is no configuration field such as:

```text
enabled
fileEnabled
stderrOnly
```

An operator needing stderr-only behavior cannot request it through the current supported configuration contract.

### 43.1 Unsupported workarounds

Do not use:

- `/dev/null`;
- a FIFO without a guaranteed reader;
- an unwritable path;
- a special device;
- a directory path.

These may fail initialization, block, or create unsafe side effects.

A real stderr-only mode requires an explicit code and configuration change.

---

## 44. Rotation

The application does not rotate logs itself.

The repository provides:

[`deploy/logrotate/fish-audio-cli`](../deploy/logrotate/fish-audio-cli)

The template currently specifies:

```text
daily
rotate 3
maxsize 5M
compress
delaycompress
missingok
notifempty
dateext
nocreate
```

### 44.1 Placeholder path

The template contains a placeholder absolute path.

Operators must replace it with the actual resolved log path.

### 44.2 `nocreate`

Logrotate does not create a new file after rotation.

The next CLI invocation opens or creates the configured path and applies `0640`.

### 44.3 Short-lived process assumption

`fish-audio-cli` is a single-invocation command.

A process already holding the old file descriptor during external rotation may continue writing to that opened file until it exits.

Normal invocations are expected to be short-lived compared with rotation intervals.

### 44.4 Rotation permissions

Ensure the log directory allows the CLI process to create the next file after rotation.

---

## 45. File growth

The application has no internal:

- maximum file size;
- retention limit;
- age limit;
- compression;
- cleanup;
- truncation.

Without external rotation, the log file grows by append.

---

## 46. Failure scenarios

### Log directory cannot be created

Result:

```text
bootstrap stderr error
exit status 2
no configured log file
```

### Log file cannot be opened

Result:

```text
bootstrap stderr error
exit status 2
```

### File chmod fails

Result:

- primary initialization error retained;
- close is attempted;
- close failure is joined if it also fails;
- bootstrap stderr reports initialization failure;
- exit status `2`.

### Stderr write fails after initialization

Result:

- file write is still attempted;
- command does not receive a logging-call error;
- normal business logic continues.

### File write fails after initialization

Result:

- stderr write is still attempted;
- command does not receive a logging-call error;
- normal business logic continues.

### File close fails

Result:

- error emitted to bootstrap stderr logger;
- selected exit status unchanged.

---

## 47. Security considerations

### 47.1 File permissions

The file is forced to:

```text
0640
```

This permits:

- owner read/write;
- group read;
- no permissions for others.

The group may therefore read logs.

Choose file ownership and group membership accordingly.

### 47.2 Directory permissions

Missing directories are requested as `0750`, but existing directories are not hardened.

### 47.3 Paths

Logs expose:

- configuration path;
- secret file path on secret failures;
- output path;
- log close path;
- model and module names.

### 47.4 Text

Full input and output text are available when explicitly enabled.

### 47.5 Remote errors

Bounded Fish API response messages may enter logs through error chains.

### 47.6 Log injection

Structured handlers escape string values, but consumers must still treat text and remote error messages as untrusted data.

### 47.7 Symlink boundary

The log file path lacks the hardened anti-symlink behavior used for secrets.

Run with least privilege and use a trusted log directory.

---

## 48. Operational recommendations

Use JSON format when logs are ingested by:

- journald adapters;
- Fluent Bit;
- Vector;
- Logstash;
- custom collectors;
- structured test harnesses.

Use text format for:

- direct terminal operation;
- simple human inspection;
- development.

Keep `logText` false unless full text is genuinely needed.

Capture stderr even when a persistent file is configured.

Configure external rotation.

Avoid shared writable log directories.

Use a unique output or surrounding job identifier when request ID must be associated with external workflows.

---

## 49. Service-manager integration

A service manager should capture stderr because it contains:

- all configured records;
- early startup failures;
- log-file close failures.

The persistent file contains configured-phase records only.

### 49.1 Duplicate collection

When both stderr and the file are collected centrally, configured-phase records appear twice.

Deduplicate using:

```text
request_id
time
msg
relevant event fields
```

or choose one configured-phase source while retaining stderr for early failures.

### 49.2 Journald

When stderr is captured by journald, text logs remain structured text inside the journal message unless a separate JSON parser is applied.

### 49.3 Container use

In containers, the mandatory persistent file can be inconvenient.

Mount a writable log directory or add a deliberate stderr-only feature in code.

Do not point the file at `/dev/null`.

---

## 50. Testing expectations

### Bootstrap

Test:

- logger creation;
- nil writer;
- typed-nil writer;
- text output;
- INFO threshold;
- raw fallback path where practical.

### Request ID

Test:

- 32-character length;
- hexadecimal decoding;
- independent values;
- random-source failure through an injectable boundary if later introduced.

### Configured logger

Test:

- default path;
- relative path;
- absolute path;
- config outside a `config` directory;
- uninitialized resolver;
- text format;
- JSON format;
- every level;
- unknown level;
- unknown format;
- typed-nil writer.

### File handling

Test:

- missing directory creation;
- append preserves existing content;
- new file mode `0640`;
- existing file tightened to `0640`;
- open failure;
- chmod failure;
- close failure joining;
- existing-directory permissions remain unchanged.

### Fan-out

Test:

- both destinations receive payload;
- second destination runs after first failure;
- short write becomes `io.ErrShortWrite`;
- multiple errors join;
- empty destination list fails;
- returned byte count is the minimum.

### Lifecycle events

Test:

- message strings;
- request ID on records;
- conditional text fields;
- rune counts;
- module fields;
- pipeline fields;
- synthesis fields;
- early stderr-only behavior;
- close failure uses stderr-only logger.

---

## 51. Review checklist

### Destinations

- Does configured logging still write to stderr and file?
- Is fan-out order intentional?
- Are all destinations attempted after one fails?
- Are runtime write failures still non-fatal?
- Is a supported disable mode being added deliberately?

### Paths and files

- Is default path unchanged?
- Are relative paths resolved from project directory?
- Are missing directories created?
- Are existing directory permissions preserved?
- Is file append behavior preserved?
- Is mode `0640` preserved?
- Are special-file or symlink risks changing?

### Format and levels

- Do public values match config validation?
- Are standard slog handlers still used?
- Is source location still disabled?
- Are threshold semantics unchanged?
- Does bootstrap remain fixed text/INFO?

### Privacy

- Is `logText` still opt-in?
- Are only top-level input/output texts affected?
- Are intermediate module texts omitted?
- Are secrets absent?
- Can remote errors expose text?
- Are paths and model names acceptable metadata?

### Events

- Are message strings preserved or intentionally migrated?
- Are field names stable?
- Are severities stable?
- Is request ID attached consistently?
- Are character counts still rune counts?
- Are durations still whole milliseconds?

### Lifecycle

- Are early failures stderr-only?
- Does configured logging begin only after validation?
- Does close failure remain stderr-only?
- Does close failure affect exit status?
- Is external rotation documentation updated?

---

## 52. Logging invariants

The following rules are normative for the current implementation.

1. A bootstrap text logger is created before configuration.
2. Bootstrap logging writes only to stderr.
3. Bootstrap threshold is INFO.
4. One random request ID is generated per invocation.
5. The request ID contains 16 random bytes encoded as 32 hexadecimal characters.
6. Structured records after ID creation include `request_id`.
7. Request ID generation failure exits with status `1`.
8. Help does not open the persistent file.
9. Early CLI, path, and config failures are stderr-only.
10. Configured logging begins after configuration validation.
11. Configured logging writes to stderr and one file.
12. Empty or whitespace-only `logging.file` selects the default path.
13. Relative log paths resolve from the project directory.
14. Missing parent directories are requested with mode `0750`.
15. Existing directory permissions are not changed.
16. The file opens in append mode.
17. Existing content is preserved.
18. The file is forced to mode `0640`.
19. Existing file ownership is not changed.
20. The logging path has no secret-style symlink hardening.
21. `/dev/null` is not a supported disable mechanism.
22. Text and JSON use standard slog handlers.
23. Source-code location is not enabled.
24. Public level and format values are exact lowercase configuration values.
25. The fan-out writer attempts every destination.
26. Short writes become errors.
27. Destination errors are joined internally.
28. Ordinary logging calls do not propagate runtime write errors into exit status.
29. `logging.logText` defaults to false.
30. The flag controls top-level input and output text fields.
31. Module lifecycle logs do not contain full text.
32. Character counts are UTF-8 rune counts.
33. Durations are whole milliseconds.
34. The persistent file does not contain replayed bootstrap records.
35. A close failure is logged through stderr-only logging.
36. A close failure does not alter exit status.
37. The application does not rotate logs internally.
38. The application does not impose a log-file size limit.
39. The Fish API key is not intentionally logged.
40. Message strings currently serve as practical event identifiers.

Changing one of these rules is an observability compatibility change.

---

## 53. Non-goals

The current logging system does not provide:

- caller-supplied request IDs;
- distributed tracing;
- OpenTelemetry export;
- metrics;
- spans;
- remote log shipping;
- syslog output;
- journald-native fields;
- stderr-only configured mode;
- file-only mode;
- multiple configured files;
- size-based internal rotation;
- retention management;
- log compression;
- runtime reopen on signal;
- synchronous durability guarantees;
- runtime log-write failure exit codes;
- automatic redaction;
- text truncation;
- intermediate module text capture;
- PID or hostname enrichment;
- stable numeric event codes.

These may be added when a concrete operational requirement justifies the contract.

---

## 54. Summary

The logging lifecycle is:

```text
bootstrap text logger on stderr
    ↓
generate request ID
    ↓
report early startup failures
    ↓
open configured append file
    ↓
configured text or JSON logger
    ↓
fan out each record to stderr and file
    ↓
emit command and module lifecycle events
    ↓
close file
    ↓
report close failure on stderr only
```

The most important operational rules are:

- always capture stderr;
- expect the persistent file to begin only after configuration validation;
- correlate records with `request_id`;
- keep `logText` disabled for sensitive workloads;
- remember counts are runes, not bytes;
- secure existing parent directories yourself;
- do not use `/dev/null` as a disable trick;
- use external rotation;
- do not assume runtime log-write failures change the process status;
- treat message strings and field names as compatibility-sensitive observability surface.
