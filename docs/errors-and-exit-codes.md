# Errors and exit codes

> **Document status:** normative description of the current pre-release command failure model.
>
> **Audience:** CLI users, shell-script authors, service integrators, operators diagnosing failed invocations, module authors preserving error identity, and maintainers changing command-stage boundaries.
>
> **Scope:** this document describes process exit codes, startup and runtime failure stages, structured log messages, stdout/stderr behavior, wrapped and joined errors, pipeline recovery, Fish API categories, cancellation, output publication edge cases, automation guidance, and compatibility constraints. Command syntax is documented in [`cli.md`](cli.md); configuration failures in [`configuration.md`](configuration.md); pipeline behavior in [`pipeline.md`](pipeline.md); module contracts in [`modules.md`](modules.md); Fish behavior in [`fish-audio.md`](fish-audio.md); logging in [`logging.md`](logging.md); secrets in [`secrets-and-paths.md`](secrets-and-paths.md); output publication in [`output-files.md`](output-files.md); architecture ownership in [`architecture.md`](architecture.md).

---

## 1. Purpose

`fish-audio-cli` returns a small stage-oriented exit-code set.

The codes identify the broad command phase that failed:

```text
0  success or help
1  bootstrap diagnostics or request identity
2  invocation, configuration, initialization, or input
3  text processing, Fish request preparation, secret, or client setup
4  synthesis, Fish API, response streaming, or output publication
```

The exit code is intentionally broader than the underlying error.

Detailed diagnosis comes from:

- the structured log message;
- the wrapped error chain;
- stage-specific fields;
- the filesystem state when output publication may already have occurred.

The command does not expose one exit code per error string.

---

## 2. Complete exit-code table

| Code | Meaning | Typical final message |
|---:|---|---|
| `0` | help displayed or invocation completed successfully | none for help; `synthesis completed` for synthesis |
| `1` | bootstrap logger or request ID generation failed | raw `logging error:` or `request ID generation failed` |
| `2` | CLI, paths, config, configured logger, module/app initialization, or input failed | stage-specific error |
| `3` | pipeline, Fish request preparation, secret loading, or Fish client setup failed | stage-specific error or warning |
| `4` | Fish synthesis or atomic output failed | `synthesis failed` |

No other application-defined numeric code is currently returned by `run()`.

Operating-system termination, shell behavior, runtime panics, and external supervisors can produce other observed statuses outside this application-defined table.

---

## 3. Process entry point

The executable entry point is conceptually:

```go
func main() {
    os.Exit(run())
}
```

`run()` selects the application-defined status.

### 3.1 Deferred work inside `run`

Deferred cleanup registered inside `run()` executes before `run()` returns.

This includes:

- signal-context cleanup;
- persistent log-file close.

### 3.2 Defers in `main`

`main` contains no deferred cleanup after `run()`.

`os.Exit` itself does not run defers, but the relevant `run()` defers have already executed before its integer result reaches `os.Exit`.

### 3.3 Panic behavior

The command does not recover arbitrary panics at the top level.

An uncaught panic:

- prints a runtime panic report to stderr;
- does not follow the normal structured error path;
- does not guarantee one of application codes `1` through `4`;
- may still run active defers during stack unwinding before the runtime terminates.

A panic is a defect or exceptional package misuse, not a normal documented operational result.

---

## 4. Exit code `0`

Code `0` has two normal meanings:

1. help was requested successfully;
2. the complete synthesis invocation succeeded.

These cases are distinguished by output and logs.

---

## 5. Help success

When CLI parsing returns the standard help sentinel:

```text
cli.ErrHelp
```

the command:

- writes usage to stdout;
- returns `0`;
- does not load configuration;
- does not open the persistent log file;
- does not read text;
- does not synthesize;
- does not create output.

Command:

```bash
fish-audio-cli --help
```

### 5.1 Help output channel

Usage is written to:

```text
stdout
```

not stderr.

### 5.2 Help and request ID

The bootstrap logger and request ID are created before option parsing.

Help therefore depends on successful bootstrap logging and request-ID generation.

A failure in either stage returns `1` instead of showing help.

### 5.3 No success log

The normal help path emits no structured “help displayed” event.

---

## 6. Synthesis success

A complete normal invocation returns `0` only after:

- options are valid;
- paths initialize;
- config loads and validates;
- configured logging opens;
- modules initialize;
- input validates;
- pipeline finishes without a returned error;
- Fish request validates;
- secret loads;
- Fish client initializes;
- Fish response streams successfully;
- temporary output synchronizes and closes;
- destination rename succeeds;
- containing directory synchronizes and closes.

The final normal event is:

```text
synthesis completed
```

### 6.1 Recovered pipeline still succeeds

A module can fail while pipeline policy recovers.

Policies:

```text
use_previous
use_original
skip
```

can allow the invocation to continue or stop the pipeline successfully.

A module-level error log does not necessarily imply a nonzero process status.

### 6.2 `skip` outcome

With `skip`, the failing step becomes a stopped outcome and the pipeline returns no error.

The application can continue to Fish synthesis and return `0`.

### 6.3 Success is stage-based

Code `0` does not mean:

- every module returned success;
- no warning was logged;
- no retry occurred;
- no previous destination existed.

It means the command’s configured recovery rules and final publication completed successfully.

---

## 7. Exit code `1`

Code `1` is reserved for failures before ordinary request-correlated command processing can be established.

Current causes:

- bootstrap logger creation failure;
- request ID generation failure.

This code is narrower and more severe than general setup code `2`.

---

## 8. Bootstrap logger creation failure

At startup, the command creates a text logger on stderr at INFO threshold.

If this fails, it cannot use structured logging.

Fallback output:

```text
logging error: <error>
```

Channel:

```text
stderr
```

Exit:

```text
1
```

### 8.1 No request ID

Request ID generation has not occurred.

### 8.2 No structured fields

The fallback is plain formatted text.

### 8.3 Practical rarity

The normal writer is `os.Stderr`, so failure is unusual.

The error path remains part of the command contract and is testable at package boundaries.

---

## 9. Request ID generation failure

The command requests 16 secure random bytes.

Failure event:

```text
request ID generation failed
```

Severity:

```text
ERROR
```

Fields:

```text
error
```

Exit:

```text
1
```

### 9.1 No `request_id` field

The value could not be generated.

### 9.2 No configured logger

Configuration processing has not started.

The event goes only to bootstrap stderr logging.

---

## 10. Exit code `2`

Code `2` represents invocation and local setup failures before text processing begins.

Current stages:

```text
option parsing
project-path initialization
configuration loading
configuration validation
configured logger initialization
module registry/build initialization
module logging decorator initialization
application/pipeline initialization
input reading and validation
```

Some occur before persistent logging opens; others occur after.

---

## 11. Option parsing failure

Final event:

```text
option parsing failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
bootstrap stderr only
```

Fields:

```text
request_id
error
```

### 11.1 Unknown option

Example:

```bash
fish-audio-cli --unknown
```

The underlying flag parser error is wrapped with:

```text
parse arguments
```

### 11.2 Missing option value

Example:

```bash
fish-audio-cli --output
```

The parser reports a missing value, wrapped as argument parsing failure.

### 11.3 Unexpected positional arguments

Example:

```bash
fish-audio-cli extra \
  --format opus \
  --output speech.opus
```

Error shape:

```text
unexpected positional arguments: [...]
```

### 11.4 Missing output

Error:

```text
--output is required
```

### 11.5 Unsupported format

Supported input names:

```text
wav
mp3
opus
ogg
```

Unsupported values produce an error describing the expected set.

### 11.6 Format normalization

Format is lowercased before validation.

`ogg` becomes `opus`.

The error displays the lowercased value after normalization.

---

## 12. Option parser output behavior

The internal flag set discards its default parser output.

This prevents duplicate automatic usage/error text.

The command emits one structured stage error instead.

### 12.1 Help exception

The standard help sentinel is handled as successful help rather than option failure.

### 12.2 No automatic usage on ordinary parse error

Invalid arguments do not automatically print the full usage document.

The caller receives:

- exit `2`;
- structured stderr error.

---

## 13. Path initialization failure

Final event:

```text
path initialization failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
bootstrap stderr only
```

Fields:

```text
request_id
error
```

Typical cause:

```text
configuration path is empty
```

The default `--config` value prevents this unless the caller explicitly supplies a blank-like argument after shell processing.

---

## 14. Configuration loading failure

Final event:

```text
config loading failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
bootstrap stderr only
```

Possible causes include:

- config file missing;
- permission denied;
- config read failure;
- config close failure;
- file above 1 MiB;
- invalid UTF-8;
- malformed JSON;
- duplicate JSON keys;
- more than one JSON value;
- unknown fields;
- prohibited explicit `null`;
- Fish API key path resolution failure.

### 14.1 Wrapped path context

Loading errors include the absolute cleaned configuration path.

### 14.2 Joined read/close errors

A read failure and close failure can both be preserved.

### 14.3 No persistent record

The configured log file is not open yet.

Capture stderr in services.

---

## 15. Configuration validation failure

Final event:

```text
config validation failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
bootstrap stderr only
```

Possible categories:

- numeric range violation;
- missing pipeline array;
- invalid module name or type;
- duplicate module name;
- invalid module config object;
- unsupported error policy;
- invalid Fish endpoint;
- blank or padded model;
- invalid retry relation;
- invalid Fish request parameters;
- blank secret path;
- unsupported logging level;
- unsupported logging format.

### 15.1 Loading versus validation

Loading concerns:

- bytes;
- JSON structure;
- unknown keys;
- explicit-null rules;
- key-path resolution.

Validation concerns:

- semantic values;
- ranges;
- cross-field rules;
- supported enumerations.

Both return `2`.

---

## 16. Configured logger initialization failure

Final event:

```text
logger initialization failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
bootstrap stderr only
```

Possible causes:

- unsupported internal level or format;
- log-path resolution failure;
- log-directory creation failure;
- log-file open failure;
- chmod failure;
- handler construction failure.

### 16.1 Persistent file unavailable

The failure cannot be written to the configured file that failed to open.

### 16.2 Existing configuration was valid

This stage occurs only after config validation succeeded.

---

## 17. Module initialization failure

Final event:

```text
module initialization failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
stderr and persistent file
```

Possible causes:

- unknown module type;
- module config decode failure;
- module-specific preparation failure;
- module-specific builder failure;
- invalid prepared module identity;
- nil processor;
- registry contract violation.

### 17.1 No text input yet

Module construction happens before CLI text reading.

### 17.2 Partial module construction

The current module system has no cleanup lifecycle.

If an earlier builder created external resources before a later builder fails, cleanup depends on the module implementation and process termination.

Module authors should avoid uncontrolled side effects.

---

## 18. Module logging initialization failure

Final event:

```text
module logging initialization failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
stderr and persistent file
```

Possible causes:

- nil logger;
- invalid step identity;
- nil processor;
- logging decorator contract failure.

This stage wraps prepared processors before application construction.

---

## 19. Application initialization failure

Final event:

```text
application initialization failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
stderr and persistent file
```

The application constructor validates the final pipeline steps.

Possible causes:

- duplicate module identity;
- blank module metadata;
- nil processor;
- unsupported policy;
- pipeline construction invariant failure.

Many of these should already be caught earlier, but the final boundary validates again.

---

## 20. Input failure

Final event:

```text
input failed
```

Severity:

```text
ERROR
```

Exit:

```text
2
```

Destination:

```text
stderr and persistent file
```

Possible causes:

- nil stdin in package use;
- input read failure;
- input above `input.maxBytes`;
- invalid UTF-8;
- empty or whitespace-only text.

### 20.1 Text argument precedence

A non-empty `--text` argument wins over stdin.

Errors then use source context:

```text
text argument
```

Otherwise:

```text
stdin
```

### 20.2 Exact empty argument

An exact empty `--text` selects stdin.

A whitespace-only `--text` is selected as argument input and then fails text validation.

### 20.3 Text validation messages

Shared text contract errors include:

```text
text is not valid UTF-8
text is empty
```

Input wraps them with:

```text
input
```

---

## 21. Exit code `3`

Code `3` represents failures after input acceptance and local application setup but before the synthesis/output stage begins.

Current stages:

```text
text pipeline execution
Fish request construction
Fish API key loading
Fish client construction
```

The configured logger is active.

---

## 22. Text processing failure

Final event:

```text
text processing failed
```

Severity:

```text
ERROR
```

Exit:

```text
3
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

Possible causes:

- context already canceled;
- module interruption;
- module failure with `abort`;
- invalid module output combined with `abort`;
- unsupported runtime error policy;
- internal pipeline argument/invariant failure.

### 22.1 Rollback

A failing or interrupted module has its text changes rolled back before the pipeline returns.

### 22.2 No Fish request

Fish request construction and secret loading do not occur after a returned pipeline error.

---

## 23. Pipeline recovery does not return exit `3`

A module processor can return an error while the configured policy handles it.

### `use_previous`

- failing step output is discarded;
- previous text continues;
- report outcome becomes recovered;
- later modules continue;
- process can return `0`.

### `use_original`

- failing step output is discarded;
- original pipeline text is restored;
- later modules continue;
- process can return `0`.

### `skip`

- failing step output is discarded;
- later modules do not run;
- pipeline returns no error;
- current text continues to Fish;
- process can return `0`.

### `abort`

- pipeline returns an error;
- process returns `3`.

### 23.1 Error logs with successful exit

The logging decorator emits:

```text
module processing failed
```

before pipeline policy is applied.

Therefore a log stream can contain `ERROR` and still end with exit `0`.

Automation must use the process status and final lifecycle events, not a naive “any ERROR means failure” rule.

---

## 24. Pipeline interruption

Interruption occurs when:

- a processor returns `context.Canceled`;
- a processor returns `context.DeadlineExceeded`;
- the context becomes canceled during or immediately after a step.

Pipeline outcome:

```text
interrupted
```

Final event:

```text
text processing failed
```

Exit:

```text
3
```

### 24.1 Error policy ignored

Cancellation is not recovered through:

- `use_previous`;
- `use_original`;
- `skip`.

The pipeline stops.

### 24.2 Signal mapping

`SIGINT` or `SIGTERM` received after signal context creation and during text processing normally becomes context cancellation and exit `3`.

---

## 25. Fish request creation failure

Final event:

```text
Fish request creation failed
```

Severity:

```text
ERROR
```

Exit:

```text
3
```

Possible causes:

- final processed text invalid;
- unsupported normalized format at package boundary;
- format-specific request incompatibility;
- invalid reference or request fields;
- repeated defensive validation failure.

Configuration validation normally catches static parameter errors earlier.

Runtime format and final text are added later, so request construction validates the completed request.

---

## 26. Missing secret file created

Final event:

```text
empty secret file created
```

Severity:

```text
WARN
```

Exit:

```text
3
```

Fields:

```text
request_id
secret="Fish API key"
path
action
```

Stable category:

```text
secrets.ErrFileCreated
```

### 26.1 Not an ordinary success

The loader created the missing empty file securely.

The command intentionally stops so the operator can populate it.

### 26.2 Warning with failure status

Severity is WARN because this is an actionable initialization condition, but the process still returns nonzero.

Do not infer success from the absence of an ERROR record.

---

## 27. Fish API key loading failure

Final event:

```text
Fish API key loading failed
```

Severity:

```text
ERROR
```

Exit:

```text
3
```

Possible causes:

- unsafe final directory;
- non-regular secret leaf;
- symlink leaf;
- race detected;
- permission failure;
- chmod failure;
- file above configured limit;
- invalid UTF-8;
- empty value;
- multiple lines;
- surrounding whitespace;
- read failure;
- close failure.

Stable empty category:

```text
secrets.ErrEmpty
```

### 27.1 Secret contents omitted

The command logs the path and remediation but does not intentionally log the key value.

---

## 28. Fish client initialization failure

Final event:

```text
Fish client initialization failed
```

Severity:

```text
ERROR
```

Exit:

```text
3
```

Possible causes:

- invalid base URL;
- blank or padded API key;
- invalid API-key UTF-8;
- ASCII control in API key;
- blank or padded model;
- invalid model UTF-8;
- ASCII control in model;
- non-positive timeout;
- invalid error-body limit;
- invalid retry options;
- nil or invalid internal dependency.

### 28.1 Defensive repetition

Config and secret layers validate earlier.

The Fish client validates again because values cross the HTTP boundary.

---

## 29. Exit code `4`

Code `4` represents failure after synthesis stage begins.

The application wraps both of these boundaries in one atomic output operation:

```text
Fish client synthesis
+
temporary/final output publication
```

Any returned error produces:

```text
synthesis failed
```

and exit `4`.

---

## 30. Synthesis failure event

Final event:

```text
synthesis failed
```

Severity:

```text
ERROR
```

Exit:

```text
4
```

Fields:

```text
request_id
error
```

The event does not separately include:

- model;
- format;
- output path.

Those values were emitted in the preceding:

```text
synthesis started
```

record.

---

## 31. Fish request transport failure

Possible errors include:

- request construction failure inside client;
- network connection failure;
- DNS failure;
- TLS failure;
- timeout;
- context cancellation;
- request send failure.

These become part of the atomic writer callback error and therefore exit `4`.

### 31.1 No transport retry

The current retry loop does not retry transport errors.

### 31.2 Timeout identity

Go HTTP timeout errors remain wrapped in the error chain.

Package callers can inspect underlying interfaces and context categories when needed.

The CLI exposes only exit `4`.

---

## 32. Fish non-success HTTP response

A non-2xx HTTP response produces a typed:

```go
*fish.APIError
```

Fields:

```text
HTTPStatusCode
HTTPStatus
APIStatus
Message
```

The message is derived from a bounded response body.

### 32.1 JSON error body

When the body contains valid JSON with:

```json
{
  "status": 1008,
  "message": "insufficient credits"
}
```

the API status and message are captured.

### 32.2 Plain-text body

Malformed JSON or plain text is preserved as a trimmed bounded message.

### 32.3 Empty body

The error still includes the HTTP status.

### 32.4 CLI result

All API error categories return exit `4`.

The categories are primarily useful to package consumers and future policy code.

---

## 33. Fish API stable categories

| HTTP status | Stable category |
|---:|---|
| `400` | `fish.ErrValidation` |
| `401` | `fish.ErrAuthentication` |
| `402` | `fish.ErrPaymentRequired` |
| `403` | `fish.ErrPermission` |
| `404` | `fish.ErrNotFound` |
| `422` | `fish.ErrValidation` |
| `429` | `fish.ErrRateLimit` |
| `500`–`599` | `fish.ErrServer` |
| other | no stable category |

Package code can use:

```go
errors.Is(err, fish.ErrRateLimit)
```

and:

```go
var apiErr *fish.APIError
errors.As(err, &apiErr)
```

The CLI does not map these to distinct numeric exit codes.

---

## 34. Fish API retry and final error

Retry policy can delay the final exit.

Retryable responses include:

- `429`;
- optionally `5xx`.

After allowed attempts are exhausted, the final API error returns through synthesis failure.

Exit remains:

```text
4
```

### 34.1 Retry does not change category

A final `429` remains rate-limit category.

A final `503` remains server category.

### 34.2 Retry-After policy failure

A valid server-requested delay above the configured maximum causes retry to stop rather than clamp.

The final returned failure remains in the synthesis stage.

### 34.3 Cancellation during delay

Cancellation while waiting between attempts produces a wrapped context error and exit `4`.

---

## 35. Error response read failure

When the Fish API returns a non-success status but its bounded error body cannot be read, the client joins:

- typed API status error;
- error-response read failure.

Package callers can still match the HTTP category through `errors.Is`.

CLI result:

```text
4
```

This is a deliberate example of one operation preserving multiple simultaneous error facts.

---

## 36. Successful-response streaming failure

After a 2xx response, the body streams into the temporary output file.

Possible failures:

- response body read error;
- local file write error;
- disk full;
- quota exceeded;
- context cancellation;
- short or zero successful stream checks.

The Fish client returns an error to the output callback.

Before rename:

- existing destination remains;
- temporary cleanup is attempted.

Exit:

```text
4
```

---

## 37. Empty Fish response

A 2xx response that produces zero bytes is rejected.

The generic output writer can publish an empty callback success, but the Fish client does not treat a zero-byte synthesis stream as success.

Result:

- no final publication;
- temp cleanup;
- exit `4`.

---

## 38. Output temporary-file creation failure

Possible causes:

- parent missing;
- parent unwritable;
- read-only filesystem;
- invalid path;
- file-descriptor exhaustion;
- quota failure.

Error context:

```text
create temporary output file
```

Exit:

```text
4
```

Existing destination remains unchanged.

---

## 39. Output write failure

The callback error is wrapped with:

```text
write temporary output file
```

This can contain:

- Fish transport/API/stream error;
- local writer error;
- context cancellation.

Exit:

```text
4
```

Existing destination remains unchanged unless a concurrent process changes it.

---

## 40. Temporary file sync failure

Error context:

```text
sync temporary output file
```

Exit:

```text
4
```

Rename has not happened.

The old destination is preserved.

Cleanup attempts to remove the temp.

---

## 41. Temporary file close failure

Error context:

```text
close temporary output file
```

Exit:

```text
4
```

Rename is not attempted.

The old destination is preserved.

---

## 42. Destination rename failure

Error context:

```text
replace output file
```

Exit:

```text
4
```

Examples:

- destination is a directory;
- permission denied;
- filesystem-specific conflict;
- path changed concurrently.

The old destination remains.

Cleanup attempts temp removal.

---

## 43. Directory persistence failure

After rename, the output package opens, syncs, and closes the containing directory.

Error context:

```text
persist output replacement
```

Exit:

```text
4
```

### 43.1 Output already published

The new destination already exists.

The package does not remove it.

### 43.2 Old destination not restored

Rename already replaced the directory entry.

Deleting the new output would not restore the old file.

### 43.3 Automation warning

Exit `4` does not guarantee that `--output` is absent.

This is the most important exception to simple success/failure filesystem assumptions.

---

## 44. Cleanup failure

Before publication, output cleanup can fail while:

- closing the temp during cleanup;
- removing the temp.

Cleanup errors are joined with the primary error.

Exit:

```text
4
```

Possible filesystem state:

- final destination preserved;
- stale hidden temp remains.

The error chain contains both primary and cleanup context.

---

## 45. Output success followed by log close failure

The persistent log file is closed in a deferred function after `run()` has selected its return value.

If close fails:

- bootstrap stderr logger emits `log file closing failed`;
- the previously selected exit status is unchanged.

Therefore:

```text
synthesis completed
log file closing failed
exit 0
```

is possible.

### 45.1 Diagnostic-only close failure

The command does not convert log close failure into code `1`, `2`, `3`, or `4`.

### 45.2 Other selected status preserved

A log close failure also does not replace an existing nonzero stage code.

---

## 46. Runtime log write failures

After configured logger initialization, ordinary `slog` calls do not return writer errors to command code.

A later log write failure can occur because of:

- disk full;
- broken stderr;
- filesystem error;
- short write.

The logger fan-out attempts both destinations, but the business exit code is unchanged.

### 46.1 Possible silent diagnostic loss

The command can return the correct stage code even when one or both log destinations failed to record it.

### 46.2 Capture limitations

Exit status remains the primary machine-readable result.

Logging is diagnostic, not a transaction participant.

---

## 47. Signal behavior

The command creates a signal-aware context after input has been read.

Handled signals:

```text
SIGINT
SIGTERM
```

### 47.1 During text processing

Cancellation normally causes:

```text
text processing failed
exit 3
```

### 47.2 During synthesis or retry delay

Cancellation normally causes:

```text
synthesis failed
exit 4
```

### 47.3 Before signal context creation

Before the handler is installed, default operating-system signal behavior can terminate the process.

Observed shell status may then be signal-derived rather than application-defined.

### 47.4 `SIGKILL`

`SIGKILL` cannot be handled.

It can:

- terminate immediately;
- bypass cleanup;
- leave temporary files;
- produce a supervisor-specific or shell-derived status.

---

## 48. Shell signal statuses

Many Unix shells report a signal termination as:

```text
128 + signal number
```

Common examples:

```text
130  SIGINT
143  SIGTERM
137  SIGKILL
```

These are shell conventions, not explicit values returned by `run()`.

When the application catches `SIGINT` or `SIGTERM` after context setup, it normally exits with stage code `3` or `4` instead.

---

## 49. Error wrapping

The project adds context with `%w` throughout package boundaries.

Conceptual chain:

```text
synthesis failed log event
    ↓
write temporary output file
    ↓
send Fish API request
    ↓
context canceled
```

Error wrapping supports:

- human-readable stage context;
- `errors.Is`;
- `errors.As`;
- preservation across package boundaries.

### 49.1 Do not parse strings when typed inspection exists

Package callers should prefer:

```go
errors.Is
errors.As
```

for stable categories.

CLI shell users receive only strings and numeric status.

### 49.2 Message text remains diagnostic

Formatted error strings are not a formal machine protocol.

---

## 50. Joined errors

The project uses `errors.Join` when more than one failure matters.

Current examples include:

- config read and close;
- secret primary and file close;
- secret primary and directory close;
- Fish API status and error-body read;
- output primary and cleanup;
- directory sync and close;
- logging initialization primary and close.

### 50.1 Matching joined errors

`errors.Is` and `errors.As` traverse joined children.

### 50.2 String formatting

A joined error can render multiple lines or combined messages depending on the underlying errors.

Log collectors must treat the error field as an opaque diagnostic value.

---

## 51. Stable sentinel errors

Current exported sentinels with operational meaning include:

### CLI

```text
cli.ErrHelp
```

### Secrets

```text
secrets.ErrFileCreated
secrets.ErrEmpty
```

### Fish API

```text
fish.ErrAuthentication
fish.ErrPaymentRequired
fish.ErrPermission
fish.ErrNotFound
fish.ErrValidation
fish.ErrRateLimit
fish.ErrServer
```

### 51.1 Exit codes remain broader

Several sentinel categories map to the same exit code.

Example:

```text
ErrAuthentication → 4
ErrPaymentRequired → 4
ErrRateLimit → 4
ErrServer → 4
```

---

## 52. Bounded I/O errors

Config, input, secrets, and Fish error bodies use bounded reading.

A limit violation carries structured maximum information in the internal error type.

The command maps it according to stage:

| Bounded content | Exit |
|---|---:|
| configuration | `2` |
| input text | `2` |
| secret file | `3` |
| Fish error body | `4` |

The same low-level error concept therefore does not imply one universal process code.

---

## 53. Strict JSON errors

Strict JSON decoding can reject:

- invalid UTF-8;
- malformed syntax;
- unknown fields;
- duplicate object keys;
- trailing second JSON value;
- explicit prohibited nulls through config-specific checks.

For the top-level configuration, these become:

```text
config loading failed
exit 2
```

For module nested config decoding during module build, they become:

```text
module initialization failed
exit 2
```

The stage owning the decode determines the final code.

---

## 54. Validation error boundaries

Validation occurs repeatedly.

Examples:

```text
config semantics
module configuration
pipeline step identity
text input
module output
Fish request
Fish client headers and retry options
```

The same conceptual invalid value can surface at different stages depending on when it becomes known.

### 54.1 Static invalid Fish parameter

Usually:

```text
config validation failed
exit 2
```

### 54.2 Final runtime Fish request invalid

Usually:

```text
Fish request creation failed
exit 3
```

### 54.3 Remote provider rejects valid local request

Usually:

```text
synthesis failed
exit 4
```

This progression is intentional:

```text
local static
local completed request
remote authority
```

---

## 55. Stage transition summary

```text
bootstrap logger
    failure → 1

request ID
    failure → 1

CLI / paths / config / logger / modules / input
    failure → 2

pipeline / request / secret / Fish client
    failure → 3

Fish HTTP / streaming / output publication
    failure → 4

complete publication
    success → 0
```

---

## 56. Logging destination by failure stage

| Failure stage | stderr | persistent file |
|---|---:|---:|
| bootstrap logger | raw fallback attempt | no |
| request ID | yes | no |
| option parsing | yes | no |
| path initialization | yes | no |
| config loading | yes | no |
| config validation | yes | no |
| configured logger initialization | yes | no |
| module initialization | yes | yes |
| input | yes | yes |
| pipeline | yes | yes |
| secret/client setup | yes | yes |
| synthesis/output | yes | yes |
| log close | yes | no |

Persistent-file absence does not imply a failure happened before logging; runtime file writes can also fail silently from the command’s perspective.

---

## 57. Stdout behavior

Normal structured diagnostics are not written to stdout.

Stdout is used for:

```text
--help usage
```

Audio is not written to stdout.

### 57.1 Successful synthesis

A successful synthesis does not print a plain success line to stdout.

The success signal is:

- exit `0`;
- final file;
- structured logs.

### 57.2 Script consequence

Command substitution does not return the output filename.

The wrapper already knows the path it supplied.

---

## 58. Stderr behavior

Stderr receives:

- all bootstrap structured logs;
- all configured structured logs;
- raw bootstrap logging fallback;
- deferred log-close failure.

### 58.1 Invalid arguments

An invalid invocation produces a structured error line, not ordinary usage text.

### 58.2 Format

Early stderr is always text format.

Configured stderr follows `logging.format`.

A single invocation can therefore emit early text and later JSON only if an early record exists before successful configured logger creation.

---

## 59. Exit code and log severity are independent

Examples:

| Log severity | Exit | Meaning |
|---|---:|---|
| none | `0` | help |
| INFO | `0` | successful synthesis |
| ERROR present | `0` | module failed but pipeline recovered |
| WARN | `3` | missing secret file created |
| ERROR | `2` | invalid invocation/setup |
| ERROR | `3` | pipeline or pre-synthesis preparation |
| ERROR | `4` | synthesis/output |
| ERROR | unchanged | deferred log close failure |

Do not map log level directly to process status.

---

## 60. Error count is not status

A pipeline with two recoverable module failures can log multiple errors and still return `0`.

A missing secret can log only a warning and return `3`.

A runtime log destination can fail without changing status.

Use:

1. process exit status;
2. final lifecycle event;
3. error chain and fields;
4. filesystem state where relevant.

---

## 61. Automation decision table

### Exit `0`

For help:

- no output expected;
- usage printed to stdout.

For synthesis:

- output publication completed;
- final `synthesis completed` expected unless logging failed.

### Exit `1`

- retrying is rarely useful without fixing runtime entropy or diagnostics;
- persistent logs do not exist for this attempt.

### Exit `2`

- fix invocation, configuration, setup, or input;
- retry after local correction;
- Fish service may not have been contacted.

### Exit `3`

- inspect pipeline, request, secret, or Fish client setup;
- Fish synthesis request was not sent;
- local modules may already have run.

### Exit `4`

- Fish request may have been sent;
- provider quota may have been consumed;
- output may or may not exist;
- inspect error and destination before retry.

---

## 62. Retry safety by exit code

| Exit | Remote Fish synthesis may have occurred | Output may exist |
|---:|---:|---:|
| `1` | no | no new output |
| `2` | no | no new output |
| `3` | no Fish synthesis | no new output |
| `4` | yes | yes |
| `0` | yes | yes |

### 62.1 Module remote side effects

Exit `3` can still follow module processing.

A future module may call a remote service before failing.

The table’s “no Fish synthesis” statement applies specifically to the Fish client stage.

### 62.2 No idempotency key

Fish requests currently have no idempotency header.

Blind retry after exit `4` can generate another synthesis and consume additional quota.

---

## 63. Output inspection after exit `4`

A robust wrapper can inspect the destination:

```bash
clear

output="/var/lib/voice-worker/output/message.opus"

fish-audio-cli \
  --config /opt/fish-audio-cli/config/config.json \
  --text "$TEXT" \
  --format opus \
  --output "$output"

status=$?

case "$status" in
  0)
    printf 'success: %s\n' "$output"
    ;;
  4)
    if [ -e "$output" ]; then
      printf \
        'synthesis/output failed, but destination exists: %s\n' \
        "$output" >&2
    else
      printf 'synthesis/output failed before publication\n' >&2
    fi
    ;;
  *)
    printf 'command failed with status %d\n' "$status" >&2
    ;;
esac

exit "$status"
```

Existence does not prove that this invocation created the file when a previous destination or concurrent writer exists.

---

## 64. Missing-secret automation

Exit `3` with log message:

```text
empty secret file created
```

means:

- the path now exists;
- it is empty;
- mode was set to `0600`;
- operator provisioning is required.

Do not continuously retry without populating the file.

A retry loop would merely rediscover an empty secret, because software cannot summon credentials from the void despite decades of configuration frameworks trying.

---

## 65. Rate-limit automation

A typed Fish `429` is:

```text
fish.ErrRateLimit
```

but the CLI still returns:

```text
4
```

A shell wrapper cannot distinguish it from other synthesis failures by numeric status alone.

Options:

- parse structured JSON logs carefully;
- use the Go package API and `errors.Is`;
- add a future machine-readable error classification contract.

Do not depend on fragile substring matching as a permanent interface.

---

## 66. Authentication automation

Fish `401` maps to:

```text
fish.ErrAuthentication
exit 4
```

Correct action:

- verify the secret file content;
- verify account/key validity;
- avoid repeated blind retries.

Exit code alone does not distinguish authentication from timeout or disk failure.

---

## 67. Payment and permission failures

Fish responses:

```text
402 → payment required
403 → permission denied
404 → resource not found
```

all return exit `4`.

Likely corrections differ:

- billing or credits;
- account/model/voice permission;
- model/reference/resource identifier.

Use the error detail, not only status `4`.

---

## 68. Server failures

Fish `5xx` maps to:

```text
fish.ErrServer
exit 4
```

When `retryServerErrors` is enabled, some attempts can occur before final failure.

A retry after command completion may be reasonable, but:

- quota semantics can be provider-specific;
- no idempotency key exists;
- output state must still be inspected.

---

## 69. Validation failures from Fish

Remote:

```text
400
422
```

map to:

```text
fish.ErrValidation
exit 4
```

This means local validation accepted the request shape, but the remote provider rejected it.

Possible reasons:

- provider-side constraints changed;
- model-specific parameter restriction;
- reference ID invalid;
- account-specific rule;
- undocumented remote validation.

Treat remote API as authoritative.

---

## 70. Unknown Fish HTTP status

An unclassified non-2xx response still produces `*fish.APIError`.

It has no stable sentinel category.

CLI result remains:

```text
4
```

Package callers can inspect:

```text
HTTPStatusCode
HTTPStatus
APIStatus
Message
```

---

## 71. Context identity

Wrapped cancellation preserves:

```text
context.Canceled
context.DeadlineExceeded
```

where applicable.

Package callers can use:

```go
errors.Is(err, context.Canceled)
```

The CLI does not allocate a dedicated cancellation exit code.

Stage determines:

- `3` during pipeline;
- `4` during synthesis/output.

---

## 72. Error policy and invalid module output

A module can return `nil` but leave blank or invalid UTF-8 text.

The pipeline converts this into:

```text
invalid text output
```

Then normal error policy applies.

Possible process outcomes:

| Policy | Process can continue |
|---|---:|
| `use_previous` | yes |
| `use_original` | yes |
| `skip` | yes |
| `abort` | no, exit `3` |

Validation failure is therefore a module failure, not always a process failure.

---

## 73. Unsupported runtime policy

Configuration validation should reject unknown policies before modules run.

The pipeline still has a defensive default branch.

If reached:

- report outcome becomes failed;
- pipeline returns an unsupported-policy error;
- command exits `3`.

This is an internal invariant failure, not an expected user-facing path after validated configuration.

---

## 74. Error identity through decorators

The module logging decorator wraps processor errors with logging context only where defined and returns the underlying failure.

The pipeline then wraps with module name/type context.

Module authors should wrap causes using `%w`.

Incorrect:

```go
return fmt.Errorf("remote failure: %v", err)
```

Correct:

```go
return fmt.Errorf("remote failure: %w", err)
```

Preserving identity allows cancellation and typed categories to remain detectable.

---

## 75. Errors that do not stop the process

Current examples:

- recoverable module processor error;
- module invalid output under recoverable policy;
- configured logger runtime write failure;
- deferred log file close failure;
- warnings;
- retryable Fish response before a later successful attempt.

These may appear in logs during a final exit `0`.

---

## 76. Errors that stop a stage but preserve artifacts

Examples:

### Missing secret

- process stops with `3`;
- newly created empty secret file remains.

### Published output directory-sync failure

- process stops with `4`;
- newly published output remains.

### Cleanup removal failure

- process stops with `4`;
- old destination remains;
- temporary file may remain.

Failure does not imply “nothing changed.”

---

## 77. Error message stability

Structured event messages currently function as practical event identifiers.

Examples:

```text
option parsing failed
config loading failed
text processing failed
synthesis failed
```

The nested error strings provide detail.

### 77.1 No separate error code field

Logs do not currently include:

```text
error_code
error_class
stage_code
```

### 77.2 Compatibility

Changing event messages or field names can break log consumers.

Changing nested human-readable error text should still not be treated as a formal wire-protocol migration unless a stable contract is explicitly introduced.

---

## 78. Recommended machine integration

Prefer:

- process exit status for broad branching;
- JSON log format for structured diagnosis;
- `request_id` for correlation;
- output-path inspection after `4`;
- Go package error inspection when fine categories are required.

Avoid:

- scraping text log token positions;
- branching solely on severity;
- assuming any nonzero status leaves no files;
- retrying every `4` identically;
- exposing full errors to untrusted users without reviewing path and provider details.

---

## 79. JSON log parsing

With:

```json
{
  "logging": {
    "format": "json"
  }
}
```

a failure record contains standard keys and fields such as:

```json
{
  "time": "...",
  "level": "ERROR",
  "msg": "synthesis failed",
  "request_id": "...",
  "error": "..."
}
```

The `error` representation is diagnostic text.

Typed Go identity is not serialized as a stable JSON class.

---

## 80. Error privacy

Errors can disclose:

- absolute config path;
- secret path;
- output path;
- module name/type;
- Fish model;
- remote API response message;
- filesystem details;
- network endpoint context.

The Fish API key itself is not intentionally logged.

Protect stderr and persistent logs.

---

## 81. Text privacy on failures

`logging.logText` controls top-level input/output text fields in normal processing events.

Failure errors can still contain module- or provider-generated details.

Module authors must not include full sensitive text in error messages unless explicitly justified.

Remote Fish error bodies are bounded but can contain provider-returned content.

---

## 82. Operator triage by code

### Code `1`

Check:

- stderr availability;
- runtime entropy source;
- severe process/runtime environment failures.

### Code `2`

Check:

- CLI syntax;
- config path;
- JSON validity;
- config ranges;
- log directory;
- module type/config;
- input source and size.

### Code `3`

Check:

- module failure and policy;
- cancellation during processing;
- final processed text;
- Fish request parameters;
- secret file and permissions;
- Fish endpoint/model/key header safety.

### Code `4`

Check:

- Fish HTTP category;
- network/TLS/timeouts;
- rate limits and billing;
- output parent and permissions;
- disk/quota;
- temp cleanup;
- destination state;
- directory sync.

---

## 83. Troubleshooting examples

### Missing output option

Observed:

```text
option parsing failed
exit 2
```

Correction:

```bash
--output speech.opus
```

### Empty stdin

Observed:

```text
input failed
exit 2
```

Correction:

- pipe nonblank UTF-8 text;
- or use `--text`.

### Unknown module type

Observed:

```text
module initialization failed
exit 2
```

Correction:

- use a registered module type;
- verify spelling and build contents.

### Module abort

Observed:

```text
module processing failed
text processing failed
exit 3
```

Correction:

- fix module cause;
- or deliberately choose a recovery policy.

### Missing key

Observed:

```text
empty secret file created
exit 3
```

Correction:

- write exactly one API key line.

### Fish 429

Observed:

```text
synthesis failed
exit 4
```

Correction:

- respect rate-limit timing;
- review retry settings.

### Missing output parent

Observed:

```text
synthesis failed
create temporary output file
exit 4
```

Correction:

- create parent directory.

---

## 84. Shell handling example

```bash
clear

fish-audio-cli \
  --config /opt/fish-audio-cli/config/config.json \
  --text "$TEXT" \
  --format opus \
  --output "$OUTPUT"

status=$?

case "$status" in
  0)
    printf 'completed\n'
    ;;
  1)
    printf 'bootstrap failure\n' >&2
    ;;
  2)
    printf 'invocation or setup failure\n' >&2
    ;;
  3)
    printf 'processing or pre-synthesis failure\n' >&2
    ;;
  4)
    printf 'synthesis or output failure\n' >&2
    ;;
  *)
    printf \
      'external termination or unexpected status: %d\n' \
      "$status" >&2
    ;;
esac

exit "$status"
```

The `*` branch is necessary for signals, panics, wrapper failures, and supervisor behavior outside the defined table.

---

## 85. Systemd interpretation

A systemd unit should treat:

```text
0
```

as success.

Codes `1` through `4` are failures with different remediation.

Repeated restart may be inappropriate for:

- invalid config (`2`);
- missing or invalid secret (`3`);
- authentication/payment failure (`4`).

Use restart policy carefully.

A rate limit or transient server failure can justify delayed retry, while a malformed config will merely fail with impressive consistency.

---

## 86. Cron interpretation

Cron normally reports nonzero status through its configured notification mechanism.

Ensure stderr is captured because early failures never reach persistent logging.

Use absolute paths for:

- binary;
- config;
- output.

Check the code and avoid assuming output absence after `4`.

---

## 87. Container interpretation

Container platforms often expose the process exit code directly.

Mount or provision:

- valid config;
- writable log directory;
- writable/chmod-capable secret;
- existing writable output parent.

A container restart loop can repeatedly call Fish after an ambiguous exit `4`.

Use job-level retry policy with output inspection and provider awareness.

---

## 88. Package-level error handling example

```go
err := client.Synthesize(ctx, request, writer)
if err != nil {
    switch {
    case errors.Is(err, context.Canceled):
        // caller cancellation
    case errors.Is(err, fish.ErrRateLimit):
        // retry according to policy
    case errors.Is(err, fish.ErrAuthentication):
        // rotate or correct credential
    }

    var apiErr *fish.APIError
    if errors.As(err, &apiErr) {
        // inspect HTTPStatusCode and APIStatus
    }

    return err
}
```

The CLI intentionally compresses these distinctions into code `4`.

---

## 89. Test expectations for exit codes

Tests should verify every explicit return path in `run()`.

### Code `0`

- help;
- successful synthesis;
- recoverable module failure;
- skip policy.

### Code `1`

- bootstrap logger failure through injectable boundary;
- request ID generation failure through injectable boundary.

### Code `2`

- parse failure;
- path failure;
- config load failure;
- config validation failure;
- logger open failure;
- module build failure;
- logging decorator failure;
- application construction failure;
- input failure.

### Code `3`

- pipeline abort;
- pipeline interruption;
- Fish request build failure;
- missing secret creation;
- other secret failure;
- Fish client construction failure.

### Code `4`

- transport failure;
- typed Fish API errors;
- stream failure;
- empty response;
- temp create/write/sync/close failure;
- rename failure;
- directory sync/close failure;
- cancellation during synthesis.

---

## 90. Test expectations for artifacts

Tests should verify not only code but side effects.

### Code `2`

- persistent log absent for early failures;
- persistent log present for later setup/input failures;
- no output publication.

### Code `3`

- missing secret file remains;
- pipeline abort produces no output;
- local module side effects follow module contract.

### Code `4`

- pre-rename failure preserves old output;
- cleanup removes temp when possible;
- cleanup error preserves both causes;
- post-rename directory failure retains new output.

### Code `0`

- final output bytes correct;
- final mode correct;
- success log emitted where logging works.

---

## 91. Review checklist

### Numeric contract

- Are only codes `0` through `4` returned by `run()`?
- Is each return still assigned to the intended stage?
- Has a new stage been placed deliberately?
- Are help and synthesis both still `0`?

### Logging

- Is the final event message accurate?
- Is severity intentional?
- Does the failure occur before or after persistent logging?
- Are fields sufficient without leaking secrets?
- Can log failure affect status?

### Error identity

- Are causes wrapped with `%w`?
- Are simultaneous errors joined?
- Are sentinels preserved?
- Can cancellation still be detected?
- Are typed Fish errors still reachable?

### Pipeline

- Do recovery policies still suppress process failure?
- Does interruption bypass recovery?
- Can an ERROR log coexist with exit `0`?
- Are reports populated on failure?

### Output

- Does code `4` still cover both Fish and filesystem?
- Can published output coexist with failure?
- Are old destinations preserved before rename?
- Are cleanup errors retained?

### Automation

- Is retry safety documented?
- Can remote work have occurred?
- Can local artifacts remain?
- Are signal-derived statuses distinguished?
- Is a new machine-readable error class needed?

---

## 92. Exit-code invariants

The following rules are normative for the current implementation.

1. `main` exits with the integer returned by `run`.
2. Help returns `0`.
3. Complete synthesis returns `0`.
4. Recoverable module errors can return `0`.
5. `skip` can return `0`.
6. Bootstrap logger failure returns `1`.
7. Request ID generation failure returns `1`.
8. Option parsing failure returns `2`.
9. Path initialization failure returns `2`.
10. Config loading failure returns `2`.
11. Config validation failure returns `2`.
12. Configured logger initialization failure returns `2`.
13. Module initialization failure returns `2`.
14. Module logging initialization failure returns `2`.
15. Application initialization failure returns `2`.
16. Input failure returns `2`.
17. Pipeline returned error returns `3`.
18. Pipeline interruption returns `3`.
19. Fish request creation failure returns `3`.
20. Missing secret creation returns `3`.
21. Other secret loading failure returns `3`.
22. Fish client construction failure returns `3`.
23. Fish synthesis error returns `4`.
24. Fish non-success HTTP response returns `4`.
25. Fish response streaming failure returns `4`.
26. Atomic output pre-publication failure returns `4`.
27. Atomic output post-publication persistence failure returns `4`.
28. Code `4` can coexist with an existing new output.
29. Deferred log close failure does not change status.
30. Runtime log write failure does not change status.
31. A WARN record can accompany nonzero status.
32. An ERROR record can accompany final status `0`.
33. Early failures are stderr-only.
34. Configured-phase failures target stderr and file.
35. Fish API categories do not receive distinct CLI codes.
36. Cancellation has no dedicated code.
37. Cancellation during pipeline normally maps to `3`.
38. Cancellation during synthesis normally maps to `4`.
39. External signal termination can produce other observed statuses.
40. Panic is outside the normal numeric contract.
41. Error strings are diagnostic, not stable numeric subcodes.
42. Wrapped identity should be preserved with `%w`.
43. Simultaneous failures should be preserved with `errors.Join`.
44. Exit code identifies broad stage, not exact cause.
45. Retry decisions must inspect more than the numeric code when possible.
46. Exit `1` through `3` do not send a Fish synthesis request.
47. Exit `4` may follow a sent Fish request.
48. Exit `3` may follow module-side remote work.
49. Help writes usage to stdout.
50. Normal diagnostics use stderr and configured logging.

Changing one of these rules is a command-line compatibility change.

---

## 93. Non-goals

The current error system does not provide:

- one exit code per exact error;
- stable numeric subcodes;
- machine-readable error JSON on stdout;
- a dedicated cancellation code;
- a dedicated timeout code;
- distinct authentication, billing, permission, and rate-limit exit codes;
- automatic retry classification for shell callers;
- rollback of module side effects;
- rollback after output rename;
- guaranteed log delivery;
- panic recovery;
- automatic usage printing on every argument error;
- a success message on stdout;
- output existence guarantees for every failure code;
- provider idempotency;
- distributed tracing error status;
- localization of error messages.

These require explicit interface design.

---

## 94. Summary

The command’s error model is stage-oriented:

```text
0
    help or complete success

1
    bootstrap diagnostics / request identity

2
    invocation / config / initialization / input

3
    pipeline / Fish request / secret / client setup

4
    Fish synthesis / streaming / output publication
```

The most important operational rules are:

- treat the exit code as a broad stage, not an exact diagnosis;
- capture stderr because early failures never reach the persistent file;
- do not equate ERROR logs with process failure when pipeline recovery is enabled;
- do not equate WARN logs with success;
- use wrapped and typed errors in Go integrations;
- recognize all Fish API categories still map to `4`;
- inspect the output path after `4` because rename may already have succeeded;
- avoid blind retries after `4` because the provider may have processed the request;
- remember log write and close failures do not alter the selected code;
- handle statuses outside `0` through `4` as external termination, panic, or wrapper behavior.
