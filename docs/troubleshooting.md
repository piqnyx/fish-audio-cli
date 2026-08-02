# Troubleshooting

> **Document status:** practical troubleshooting guide for the current pre-release implementation.
>
> **Audience:** CLI users, operators, service integrators, module authors, and maintainers diagnosing failed or surprising invocations.
>
> **Scope:** this document maps symptoms, log messages, exit codes, filesystem state, network behavior, pipeline outcomes, and CI failures to concrete checks and corrective actions. It does not replace the normative contracts in [`architecture.md`](architecture.md), [`pipeline.md`](pipeline.md), [`modules.md`](modules.md), [`module-author-guide.md`](module-author-guide.md), [`configuration.md`](configuration.md), [`cli.md`](cli.md), [`fish-audio.md`](fish-audio.md), [`logging.md`](logging.md), [`secrets-and-paths.md`](secrets-and-paths.md), [`output-files.md`](output-files.md), [`errors-and-exit-codes.md`](errors-and-exit-codes.md), and [`testing.md`](testing.md).

---

## 1. Start with the exit status

`fish-audio-cli` uses stage-oriented exit codes:

| Code | Broad stage |
|---:|---|
| `0` | help or successful completion |
| `1` | bootstrap logging or request ID |
| `2` | arguments, paths, config, logging, modules, app setup, or input |
| `3` | pipeline, Fish request, secret, or Fish client setup |
| `4` | Fish HTTP, response streaming, or atomic output |

The numeric code narrows the stage.

The structured error supplies the actual cause.

Do not diagnose from the last filename you noticed or from the emotional intensity of the log message. Software remains unmoved by confidence.

---

## 2. Capture a clean diagnostic run

Use a dedicated stderr file and preserve the process status:

```bash
clear

set +e

./bin/fish-audio-cli \
  --config /absolute/path/to/config.json \
  --format opus \
  --output /absolute/path/to/output.opus \
  --text 'Diagnostic text' \
  2> /tmp/fish-audio-cli.stderr

status=$?

set -e

printf 'exit_status=%d\n' "$status"
printf '%s\n' '--- stderr ---'
cat /tmp/fish-audio-cli.stderr
```

This captures early failures that never reach the persistent log file.

### 2.1 Do not expose the API key

Do not add:

```text
cat secret-file
set -x
env
```

to a diagnostic transcript.

The API key is not needed to identify most failures.

### 2.2 Preserve the exact error

Do not paraphrase:

```text
it says something about a file
```

Preserve:

- exit status;
- complete final event;
- wrapped error text;
- request ID;
- output path state;
- relevant file metadata.

---

## 3. Find the request ID

Normal structured records contain:

```text
request_id=<32 lowercase hexadecimal characters>
```

Use it to correlate one invocation across stderr and the persistent log.

Example text-log search:

```bash
clear

request_id='0123456789abcdef0123456789abcdef'

grep -F "$request_id" \
  /path/to/fish-audio-cli.log
```

Example JSON-log search:

```bash
clear

request_id='0123456789abcdef0123456789abcdef'

grep -F "\"request_id\":\"$request_id\"" \
  /path/to/fish-audio-cli.log
```

A request ID is unavailable when its generation itself failed.

---

## 4. Identify the last lifecycle message

The last meaningful message usually identifies the stage.

| Last message | Stage |
|---|---|
| `option parsing failed` | CLI arguments |
| `path initialization failed` | configuration path |
| `config loading failed` | config bytes or JSON |
| `config validation failed` | config semantics |
| `logger initialization failed` | persistent logging |
| `module initialization failed` | registry or module config/build |
| `module logging initialization failed` | logging decorator |
| `application initialization failed` | final pipeline construction |
| `input failed` | `--text` or stdin |
| `text processing failed` | module execution or cancellation |
| `Fish request creation failed` | completed request validation |
| `empty secret file created` | missing key file bootstrap |
| `Fish API key loading failed` | secret filesystem or content |
| `Fish client initialization failed` | endpoint, key, model, timeout, retry |
| `synthesis failed` | Fish request, stream, or output |
| `synthesis completed` | output operation succeeded |
| `log file closing failed` | deferred logging cleanup |

---

## 5. Determine whether persistent logging was active

Persistent logging opens after:

- option parsing;
- project-path initialization;
- config loading;
- config validation.

Failures before that point are stderr-only.

### 5.1 Stderr-only failures

These do not appear in the configured file:

```text
logging error
request ID generation failed
option parsing failed
path initialization failed
config loading failed
config validation failed
logger initialization failed
```

### 5.2 Dual-destination failures

After configured logger initialization, records target:

```text
stderr
persistent log file
```

Seeing the same record twice after collecting both streams is expected duplication, not two invocations.

---

## 6. Minimal preflight checklist

Before deeper diagnosis, verify:

```bash
clear

pwd

test -x ./bin/fish-audio-cli
printf 'binary_status=%d\n' "$?"

test -r /absolute/path/to/config.json
printf 'config_read_status=%d\n' "$?"

test -d /absolute/path/to/output-parent
printf 'output_parent_status=%d\n' "$?"

df -h /absolute/path/to/output-parent
df -i /absolute/path/to/output-parent
```

Also verify the full repository state when developing:

```bash
clear

git status --short
go version
```

The module currently declares Go `1.26.5`.

---

## 7. Command not found

### Symptom

```text
fish-audio-cli: command not found
```

This is a shell failure.

The application did not start and has no exit-code contract for it.

### Checks

```bash
clear

command -v fish-audio-cli || true
ls -l ./bin/fish-audio-cli
file ./bin/fish-audio-cli
```

### Correction

Invoke the explicit path:

```bash
clear

./bin/fish-audio-cli --help
```

Or install the binary into a directory already present in `PATH`.

---

## 8. Permission denied when launching

### Symptom

```text
Permission denied
```

before any structured log appears.

### Checks

```bash
clear

ls -l ./bin/fish-audio-cli
mount | grep -F ' /path/to/filesystem ' || true
```

Possible causes:

- executable bit missing;
- filesystem mounted `noexec`;
- incompatible security policy;
- wrong binary ownership or ACL.

### Correction

For a normal local build:

```bash
clear

chmod 0755 ./bin/fish-audio-cli
```

Do not apply recursive executable permissions to the repository.

---

## 9. Exec format error

### Symptom

```text
Exec format error
```

The binary was built for another operating system or architecture.

### Checks

```bash
clear

file ./bin/fish-audio-cli
uname -s
uname -m
go env GOOS GOARCH
```

### Correction

Build for the target:

```bash
clear

go build \
  -trimpath \
  -o ./bin/fish-audio-cli \
  ./cmd/fish-audio-cli
```

---

## 10. `--help` returns nonzero

Normal help returns `0` and writes usage to stdout.

If it returns `1`, startup failed before option handling.

Possible final messages:

```text
logging error
request ID generation failed
```

Persistent logging is not active.

Capture stderr directly.

---

## 11. Help creates no persistent log

This is expected.

The command handles help before:

- config loading;
- configured logging;
- modules;
- input;
- synthesis.

No log file or output file should be created by a normal help invocation.

---

## 12. `option parsing failed`

Exit:

```text
2
```

Persistent log:

```text
not active
```

Common causes:

- unknown flag;
- missing flag value;
- positional argument;
- missing `--output`;
- unsupported format.

---

## 13. Missing `--output`

### Error

```text
--output is required
```

### Correction

```bash
clear

./bin/fish-audio-cli \
  --config /absolute/path/to/config.json \
  --format opus \
  --output /absolute/path/to/speech.opus \
  --text 'Hello'
```

There is no automatic output filename.

Audio is not written to stdout.

---

## 14. Unsupported format

Accepted CLI values:

```text
wav
mp3
opus
ogg
```

Input is lowercased.

`ogg` is normalized internally to:

```text
opus
```

### Error example

```text
unsupported format "flac": expected wav, mp3, opus or ogg
```

The filename extension does not select the format.

---

## 15. Empty format

An omitted or empty `--format` is unsupported.

Specify one of the four accepted names.

Do not assume the extension supplies the missing value.

---

## 16. Unexpected positional arguments

### Error shape

```text
unexpected positional arguments: [...]
```

Text is not accepted as an unnamed positional argument.

Use:

```bash
clear

./bin/fish-audio-cli \
  --format opus \
  --output speech.opus \
  --text 'Text to synthesize'
```

Or pipe text through stdin.

---

## 17. A value beginning with `-`

A text or path value beginning with a dash can be interpreted as another flag.

Use an explicit flag assignment where appropriate:

```bash
clear

./bin/fish-audio-cli \
  --format=opus \
  --output=./speech.opus \
  --text='-leading text'
```

---

## 18. `path initialization failed`

Exit:

```text
2
```

The configuration path was unusable before the file was opened.

Typical cause:

```text
configuration path is empty
```

Use an explicit absolute path in services.

---

## 19. Relative config path resolves from the wrong directory

A relative `--config` is resolved from the process working directory.

It is not resolved from:

- binary directory;
- repository discovered automatically;
- user home;
- output directory.

### Diagnose

```bash
clear

pwd
ls -l config/config.json
```

### Correction

Use:

```bash
clear

./bin/fish-audio-cli \
  --config /absolute/path/to/config.json \
  --format opus \
  --output /absolute/path/to/output.opus \
  --text 'Hello'
```

---

## 20. `~` in JSON does not expand

Incorrect:

```json
{
  "secrets": {
    "fishApiKeyFile": "~/.secrets/fish-key"
  }
}
```

The path contains a literal tilde.

Use a full absolute path.

Shell tilde expansion can occur in an unquoted command argument, but JSON is not interpreted by a shell.

---

## 21. Environment variables in JSON do not expand

Incorrect:

```json
{
  "logging": {
    "file": "$HOME/logs/fish-audio-cli.log"
  }
}
```

The application does not expand:

```text
$HOME
${HOME}
%USERPROFILE%
```

Use the final path explicitly.

---

## 22. Config symlink points elsewhere but relative paths stay local

The configuration loader can follow a symlinked config file.

Project-root derivation uses the lexical path supplied to `--config`, not the symlink target.

Example:

```text
--config /opt/voice/config/config.json
symlink target /etc/voice/config.json
```

Relative secret and log paths are based on:

```text
/opt/voice
```

not:

```text
/etc/voice
```

This is intentional.

Use an absolute configured secret or log path when the distinction matters.

---

## 23. Project directory seems one level too high

When the immediate config parent directory is named exactly:

```text
config
```

the project directory is its parent.

Example:

```text
/opt/fish-audio-cli/config/config.json
```

Project directory:

```text
/opt/fish-audio-cli
```

For:

```text
/etc/fish-audio-cli/settings.json
```

project directory:

```text
/etc/fish-audio-cli
```

The comparison is case-sensitive.

---

## 24. `config loading failed`

Exit:

```text
2
```

Persistent log:

```text
not active
```

Possible causes:

- file missing;
- permission denied;
- read failure;
- file above 1 MiB;
- invalid UTF-8;
- malformed JSON;
- duplicate key;
- unknown field;
- more than one JSON value;
- prohibited explicit `null`;
- secret path resolution failure.

---

## 25. Config file not found

### Error contains

```text
open configuration file
no such file or directory
```

### Checks

```bash
clear

ls -l /absolute/path/to/config.json
namei -l /absolute/path/to/config.json
```

`namei` may not be installed on every system.

The process needs search permission on every parent directory and read permission on the file.

---

## 26. Config permission denied

### Checks

```bash
clear

stat -c \
  'type=%F mode=%a owner=%U group=%G path=%n' \
  /absolute/path/to/config.json

namei -l /absolute/path/to/config.json
```

A typical non-secret config mode is:

```text
0640
```

The application does not enforce config ownership or mode.

---

## 27. Config is too large

Default maximum:

```text
1 MiB
```

Check size:

```bash
clear

stat -c 'bytes=%s path=%n' \
  /absolute/path/to/config.json
```

Remove accidental data, embedded credentials, copied logs, or repeated JSON.

Do not increase the limit by editing unrelated code unless the configuration genuinely requires it.

---

## 28. Malformed JSON

A syntax-only check:

```bash
clear

python3 -m json.tool \
  /absolute/path/to/config.json \
  > /dev/null
```

This does not detect every application rule.

In particular, generic parsers may accept duplicate keys by keeping only the last value.

The CLI remains the authoritative validator.

---

## 29. Duplicate JSON key

The strict loader rejects duplicate keys at every object level.

Examples:

```json
{
  "fish": {},
  "fish": {}
}
```

and:

```json
{
  "fish": {
    "model": "first",
    "model": "second"
  }
}
```

Escaped equivalent key names are also duplicates.

Remove one definition.

---

## 30. Unknown configuration field

The loader rejects typos instead of ignoring them.

Example:

```text
unknown JSON object key
```

Check exact field names in [`configuration.md`](configuration.md).

Do not add a misspelled field hoping the application will infer your intent. It has enough work already.

---

## 31. Explicit `null` rejected

Many configuration fields distinguish:

```text
omitted
null
zero value
```

Omit a field to receive its default.

Use `null` only where explicitly supported, such as a nullable sample-rate field.

---

## 32. Multiple JSON values

A file containing:

```text
{} {}
```

is rejected.

The configuration file must contain exactly one top-level JSON value and that value must match the config object contract.

---

## 33. Invalid config UTF-8

The file must be valid UTF-8.

Check without printing it:

```bash
clear

python3 - <<'PY'
from pathlib import Path

path = Path("/absolute/path/to/config.json")
data = path.read_bytes()

try:
    data.decode("utf-8")
except UnicodeDecodeError as exc:
    raise SystemExit(f"invalid UTF-8: {exc}")

print("UTF-8 valid")
PY
```

Save the file as UTF-8 without an incompatible legacy encoding.

---

## 34. `config validation failed`

Exit:

```text
2
```

Persistent log:

```text
not active
```

The JSON structure loaded, but one or more values violate semantic rules.

The error path usually names the field.

---

## 35. Invalid read limit

Fields such as:

```text
input.maxBytes
secrets.maxBytes
fish.maxErrorBodyBytes
```

must be positive and within implementation bounds.

Do not use zero to mean unlimited.

No unbounded mode exists.

---

## 36. Invalid logging level

Accepted config values:

```text
debug
info
warn
error
```

Use exact lowercase spelling in the public configuration contract.

---

## 37. Invalid logging format

Accepted values:

```text
text
json
```

A value such as:

```text
pretty
```

is unsupported.

---

## 38. Invalid pipeline error policy

Accepted values:

```text
use_previous
use_original
skip
abort
```

The top-level policy supplies the default.

A module can override it with its own `onError`.

---

## 39. Duplicate module name

Module instance names must be unique.

Two instances can share one type:

```text
passthrough
```

but their `name` values must differ.

Names identify steps in:

- logs;
- reports;
- errors.

---

## 40. Module array is `null`

A `null` module array is not the same as an empty array.

Valid empty pipeline:

```json
{
  "pipeline": {
    "modules": []
  }
}
```

An empty pipeline returns the input text unchanged.

---

## 41. Invalid Fish base URL

The base URL must:

- use `http` or `https`;
- include a host;
- contain no user info;
- contain no query;
- contain no fragment.

The client appends:

```text
/v1/tts
```

to the base path.

Valid example:

```text
https://api.fish.audio
```

---

## 42. Base URL already includes `/v1/tts`

Do not configure the final endpoint as the base unless the resulting joined path is intended.

The client appends its synthesis path.

A base ending in `/v1/tts` would produce another joined `v1/tts` component.

Configure the provider base or proxy base path instead.

---

## 43. Invalid retry configuration

Check:

```text
maxAttempts > 0
initialDelayMilliseconds > 0
maxDelayMilliseconds > 0
maxDelayMilliseconds >= initialDelayMilliseconds
all values within implementation maxima
```

The maximum attempt count includes the first request.

---

## 44. `logger initialization failed`

Exit:

```text
2
```

Persistent log:

```text
not available
```

The bootstrap stderr logger reports this failure.

Possible causes:

- log directory creation failure;
- log file open failure;
- chmod failure;
- invalid internal level/format;
- path resolution failure.

---

## 45. Default log path is unexpected

When `logging.file` is empty, the path is:

```text
logs/fish-audio-cli.log
```

relative to the derived project directory.

It is not necessarily relative to the current working directory.

---

## 46. Log directory missing

The application creates missing log directories automatically with requested mode:

```text
0750
```

subject to umask.

If creation fails, inspect the nearest existing parent:

```bash
clear

namei -l /absolute/path/to/logs/fish-audio-cli.log
```

---

## 47. Log path permission denied

### Checks

```bash
clear

log_path='/absolute/path/to/fish-audio-cli.log'
log_dir="$(dirname "$log_path")"

stat -c \
  'type=%F mode=%a owner=%U group=%G path=%n' \
  "$log_dir"

test -w "$log_dir"
printf 'directory_writable=%d\n' "$?"
```

The service identity must be able to:

- create or open the file;
- append;
- chmod it to `0640`.

---

## 48. Read-only log mount

The logger forces mode:

```text
0640
```

An existing read-only file can fail even when readable.

Persistent logging cannot currently be disabled.

Use a writable regular file path.

---

## 49. `/dev/null` does not disable file logging

Do not configure:

```text
/dev/null
```

The logger opens the path and calls `Chmod(0640)`.

For an unprivileged process, this normally fails logger initialization.

With excessive privilege, changing device permissions would be dangerous.

There is no supported stderr-only configuration.

---

## 50. Log file is truncated unexpectedly

The application opens logs with append mode.

It should not truncate existing content.

Check whether an external process performed:

- log rotation;
- truncation;
- replacement;
- cleanup.

The included logrotate template uses `nocreate`.

---

## 51. Log file mode changes

The logger forces the file to:

```text
0640
```

This is expected.

Existing ownership and group are not changed.

---

## 52. No logs at debug level

Ensure public config contains:

```json
{
  "logging": {
    "level": "debug"
  }
}
```

Then confirm the correct config file loaded by locating:

```text
config loaded path=...
```

The current application has limited explicit DEBUG events; lowering the threshold does not manufacture events that code never emits.

---

## 53. JSON expected, text received

Early startup logs always use bootstrap text format.

Configured JSON formatting begins only after:

- config load;
- config validation;
- logger initialization.

A failure before that remains text.

One invocation can therefore produce early text and later JSON only when both phases emit records.

---

## 54. Same log appears twice

The configured logger writes every record to:

```text
stderr
log file
```

If a collector ingests both, duplicate records are expected.

Deduplicate by:

- request ID;
- timestamp;
- message;
- fields.

Or collect one configured destination intentionally.

---

## 55. `module initialization failed`

Exit:

```text
2
```

Persistent logging:

```text
active
```

Possible causes:

- unsupported type;
- strict module config failure;
- preparation failure;
- nil builder;
- builder failure;
- typed-nil processor;
- invalid error policy.

---

## 56. Unsupported module type

The currently registered built-in type is:

```text
passthrough
```

An error resembles:

```text
module "name" at index N has unsupported type "type"
```

Check spelling and ensure the binary was built from code containing the expected registration.

Configuration alone cannot load an uncompiled plugin.

---

## 57. New module code exists but type is unsupported

Verify the module preparer was added to the registry.

Then rebuild the binary:

```bash
clear

go build \
  -trimpath \
  -o ./bin/fish-audio-cli \
  ./cmd/fish-audio-cli
```

Check that you are invoking the rebuilt file rather than an older copy elsewhere in `PATH`.

---

## 58. Module config unknown field

Each module strictly owns and decodes its `config` object.

For `passthrough`, the valid config is:

```json
{}
```

A field such as:

```json
{
  "inventMeaning": true
}
```

is rejected.

---

## 59. Preparation failure stops every build

The registry prepares all module configurations before instantiating any processor.

When preparation fails:

- no processor builder should have run;
- input has not been read;
- secret has not been loaded;
- Fish has not been contacted;
- output has not been created.

Fix the reported module config or path.

---

## 60. Builder failure after preparation

Once every module prepared successfully, builders run in configured order.

A builder failure stops later builders.

There is no general module cleanup lifecycle.

A module that acquires resources must manage its own failure safety.

---

## 61. `module logging initialization failed`

This is a defensive setup failure around the per-step logging decorator.

It usually indicates an invalid or nil step that should have been rejected earlier.

Treat it as:

- an internal invariant failure;
- a module integration defect;
- not an end-user retry condition.

Run the full test suite.

---

## 62. `application initialization failed`

Final pipeline construction rejected the prepared steps.

Possible causes include:

- duplicate identity;
- blank metadata;
- nil processor;
- unsupported policy.

This is normally a defect in module build integration or a missed earlier validation.

---

## 63. `input failed`

Exit:

```text
2
```

Persistent logging:

```text
active
```

Possible causes:

- stdin nil in package use;
- read error;
- input above limit;
- invalid UTF-8;
- empty or whitespace-only text.

---

## 64. `--text ""` waits on stdin

An exact empty text argument is treated as omitted.

The command then reads stdin.

This can look like a hang.

### Diagnose

Use a non-empty test:

```bash
clear

./bin/fish-audio-cli \
  --config /absolute/path/to/config.json \
  --format opus \
  --output /absolute/path/to/output.opus \
  --text 'test'
```

Or close stdin explicitly:

```bash
clear

./bin/fish-audio-cli \
  --config /absolute/path/to/config.json \
  --format opus \
  --output /absolute/path/to/output.opus \
  < /dev/null
```

The latter then fails because input is empty.

---

## 65. Piped command never finishes

The CLI reads stdin until EOF.

A producer that keeps the pipe open causes the CLI to wait.

Example safe producer:

```bash
clear

printf '%s' 'Hello' |
  ./bin/fish-audio-cli \
    --config /absolute/path/to/config.json \
    --format opus \
    --output /absolute/path/to/output.opus
```

Ensure the upstream process terminates or closes stdout.

---

## 66. Empty input

Error includes:

```text
input text is empty
```

Whitespace-only text is also empty under the shared text contract.

Provide at least one non-whitespace Unicode rune.

---

## 67. Invalid input UTF-8

Input must be valid UTF-8.

Binary input or text in another encoding is rejected.

Check a file:

```bash
clear

python3 - <<'PY'
from pathlib import Path

path = Path("/absolute/path/to/input.txt")
data = path.read_bytes()

try:
    data.decode("utf-8")
except UnicodeDecodeError as exc:
    raise SystemExit(f"invalid UTF-8: {exc}")

print("UTF-8 valid")
PY
```

---

## 68. Input exceeds limit

Default:

```text
1 MiB
```

The limit counts bytes, not Unicode characters.

Check:

```bash
clear

wc -c /absolute/path/to/input.txt
```

Reduce input size or deliberately adjust `input.maxBytes` within supported bounds.

---

## 69. Signal while waiting for stdin

The signal-aware context is created after input reading.

A signal received while the process is blocked reading stdin can follow default operating-system termination behavior rather than returning application code `3` or `4`.

Close the producer pipe or provide `--text` for predictable stage handling.

---

## 70. `text processing failed`

Exit:

```text
3
```

The pipeline returned an error.

Fields include:

```text
pipeline_outcome
steps_total
steps_executed
pipeline_duration_ms
error
```

Possible causes:

- module error under `abort`;
- context cancellation;
- deadline;
- invalid module output under `abort`;
- defensive policy failure.

Fish synthesis has not started.

---

## 71. Module error but final exit is `0`

This can be correct.

The step logger emits:

```text
module processing failed
```

before pipeline policy is applied.

Policies can recover:

```text
use_previous
use_original
skip
```

Inspect the final pipeline outcome:

```text
recovered
stopped
```

and the final command status.

Do not treat any ERROR record as automatic process failure.

---

## 72. `use_previous` behavior surprises you

On module failure:

- failing changes are rolled back;
- text from immediately before the module is restored;
- later modules continue.

This can produce valid output and exit `0`.

---

## 73. `use_original` behavior surprises you

On module failure:

- failing changes are rolled back;
- the original pipeline input is restored;
- later modules continue.

Earlier successful transformations are discarded at that point.

---

## 74. `skip` stops later modules but still synthesizes

`skip` means:

- restore pre-failure text;
- stop remaining modules;
- return pipeline success;
- continue to Fish.

It does not mean:

```text
skip the entire command
```

---

## 75. `abort` returns code `3`

`abort` means:

- restore pre-failure text;
- stop pipeline;
- return an error;
- do not build/send the Fish request.

---

## 76. Module returned blank output

A processor can return `nil` but produce invalid text.

The pipeline converts blank or invalid UTF-8 output into:

```text
invalid text output
```

Then the configured error policy applies.

Fix the module output contract.

---

## 77. Pipeline interrupted despite recovery policy

Cancellation and deadline errors are not converted into successful fallback results.

Even with:

```text
use_previous
use_original
skip
```

an interruption stops the pipeline and returns code `3`.

---

## 78. `Fish request creation failed`

Exit:

```text
3
```

The final processed text and selected CLI format did not form a valid request.

This stage occurs after pipeline completion but before secret loading.

Check:

- final text validity;
- format;
- reference ID;
- format-specific request fields;
- request parameter compatibility.

---

## 79. `empty secret file created`

Exit:

```text
3
```

Severity:

```text
WARN
```

This is expected on the first run when the configured secret file does not exist.

The loader:

- creates parent directories when needed;
- creates an empty file;
- sets file mode `0600`;
- stops the command.

### Correction

Write exactly one key line.

Interactive example:

```bash
clear

secret_path='/absolute/path/to/fish-api-key'

read -r -s -p 'Fish API key: ' fish_key
printf '\n'

printf '%s\n' "$fish_key" > "$secret_path"
unset fish_key

chmod 0600 "$secret_path"
```

---

## 80. Secret file exists but is empty

Error category:

```text
secret is empty
```

Check metadata without printing the key:

```bash
clear

secret_path='/absolute/path/to/fish-api-key'

stat -c \
  'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
  "$secret_path"
```

Populate it with one nonblank line.

---

## 81. Secret has two lines

Error:

```text
secret must contain exactly one line
```

One optional final LF or CRLF is accepted.

Two final newlines are not.

Inspect counts without printing content:

```bash
clear

python3 - <<'PY'
from pathlib import Path

path = Path("/absolute/path/to/fish-api-key")
data = path.read_bytes()

print(f"bytes={len(data)}")
print(f"lf_count={data.count(bytes([10]))}")
print(f"cr_count={data.count(bytes([13]))}")
PY
```

---

## 82. Secret has surrounding whitespace

Error:

```text
secret must not have surrounding whitespace
```

The loader does not silently trim the key.

Common causes:

- copied leading space;
- trailing space before newline;
- tab;
- quoted value with spaces.

Rewrite the exact key.

---

## 83. Secret was stored with quotes

Incorrect file content:

```text
"actual-key"
```

The quotes become part of the key.

The generic loader can accept them, but Fish authentication normally fails.

Write:

```text
actual-key
```

without JSON quoting.

---

## 84. Secret has a UTF-8 BOM

A UTF-8 BOM is valid UTF-8 and is not removed.

It becomes part of the API key.

Rewrite the file without BOM.

---

## 85. Secret invalid UTF-8

Error:

```text
secret is not valid UTF-8
```

The key file is plain UTF-8 text.

Do not use:

- UTF-16;
- binary key containers;
- locale-specific encoding.

---

## 86. Secret exceeds byte limit

Default maximum:

```text
16 KiB
```

A Fish API key should be far smaller.

A large file usually means the configured path points to the wrong object.

Check path and size before increasing the limit.

---

## 87. Secret directory writable by group or others

Error:

```text
secret directory "..." is writable by group or others
```

Rejected write bits:

```text
0020
0002
```

### Correction

For an owner-only directory:

```bash
clear

chmod 0700 /absolute/path/to/secrets
```

Then inspect every deployment user and ownership assumption.

Do not use `0777`.

---

## 88. Secret directory mode `0755`

The loader’s specific security check allows `0755` because group and others cannot write.

Owner-only `0700` remains the recommended deployment mode.

The loader does not automatically tighten an existing directory.

---

## 89. Secret path is a symlink

Error:

```text
secret path "..." is not a regular file
```

The secret leaf symlink is intentionally rejected.

Use a regular file at the configured path.

An absolute path to the actual regular file is acceptable when deployment policy permits it.

---

## 90. Secret path is a directory, FIFO, socket, or device

The leaf must be a regular file.

Replace the path with an ordinary file.

Do not use process substitution or a named pipe as the secret source.

---

## 91. Secret file changed while opening

Error:

```text
secret file "..." changed while it was being opened
```

The loader detected a race or replacement.

Possible causes:

- secret rotation at the same moment;
- deployment tool replacing the file;
- untrusted directory mutation.

Retry only after the replacement operation has completed and the directory is trusted.

---

## 92. Secret chmod fails

The loader forces mode:

```text
0600
```

A read-only mount or immutable file can fail even when readable.

Use a writable secure copy or redesign the secret-loading contract deliberately.

---

## 93. Container secret mount fails

Common cause:

```text
read-only secret projection
```

The current loader attempts `Chmod(0600)`.

Copy the secret at container startup into a secure writable directory owned by the runtime user, then configure that path.

Do not copy it into a world-readable location.

---

## 94. Wrong secret path created

The in-memory secret path is resolved relative to the project directory during config loading.

For:

```text
/project/config/config.json
```

default secret:

```text
/project/secrets/fish-api-key
```

For:

```text
/etc/fish/settings.json
```

default secret:

```text
/etc/fish/secrets/fish-api-key
```

Use the `path` field in the warning/error rather than guessing.

---

## 95. `Fish client initialization failed`

Exit:

```text
3
```

The secret loaded, but the HTTP client boundary rejected an option.

Possible causes:

- bad base URL;
- empty or padded key;
- ASCII control in key;
- blank or padded model;
- ASCII control in model;
- timeout out of range;
- error-body limit out of range;
- retry options invalid.

No Fish request was sent.

---

## 96. API key contains a control character

The client rejects ASCII controls including:

- NUL;
- tab;
- carriage return;
- line feed;
- DEL.

The secret loader removes one final newline but does not accept internal line breaks.

Rewrite the key with ordinary printable characters only.

---

## 97. Model contains surrounding whitespace

Error:

```text
model must not have surrounding whitespace
```

Correct the configuration.

Do not rely on implicit trimming for header values.

---

## 98. Timeout appears too short

Default:

```text
120 seconds
```

This is the HTTP client timeout for one request attempt, including response reading.

Total command duration can exceed one timeout because:

- more than one attempt can occur;
- retry delays occur between attempts;
- local pipeline time occurs before synthesis;
- file sync occurs afterward.

---

## 99. `synthesis failed`

Exit:

```text
4
```

This one final event covers two broad families:

```text
Fish HTTP / stream
output filesystem / publication
```

Read the nested error prefix.

---

## 100. Distinguish Fish and output errors

Fish-side prefixes include:

```text
send synthesis request
Fish API returned
read Fish API error response
stream synthesis response
Fish API returned an empty audio response
wait before Fish API retry
```

Output-side prefixes include:

```text
create temporary output file
write temporary output file
sync temporary output file
close temporary output file
replace output file
persist output replacement
remove temporary output file
```

A `write temporary output file` error can wrap a Fish error because synthesis streams through the output callback.

Read the full chain.

---

## 101. DNS failure

Error may include:

```text
send synthesis request
no such host
```

### Checks

```bash
clear

getent hosts api.fish.audio || true
```

For a custom base URL, test its hostname.

Check:

- DNS configuration;
- container resolver;
- VPN/proxy routing;
- `/etc/resolv.conf`;
- firewall.

Transport errors are not retried by the current Fish retry policy.

---

## 102. Connection refused

Error may include:

```text
connect: connection refused
```

Possible causes:

- wrong host or port;
- local proxy not running;
- service unavailable;
- firewall rejection.

Verify the configured base URL.

The client appends `/v1/tts`.

---

## 103. TLS certificate failure

Error may mention:

```text
x509
certificate
unknown authority
hostname
```

Check:

- system clock;
- CA bundle;
- TLS interception;
- proxy certificate;
- hostname;
- custom internal CA installation.

Do not disable TLS verification in production. The current client does not expose an insecure-skip option.

---

## 104. HTTP proxy problems

Go’s default transport can use standard proxy environment variables.

Inspect without printing unrelated secrets:

```bash
clear

env |
  grep -E '^(HTTP_PROXY|HTTPS_PROXY|NO_PROXY|http_proxy|https_proxy|no_proxy)=' \
  || true
```

A proxy can cause:

- DNS differences;
- TLS interception;
- connection refusal;
- unexpected HTTP status;
- timeout.

---

## 105. Request timeout

Error may include:

```text
Client.Timeout exceeded
context deadline exceeded
```

Check:

- `fish.timeoutSeconds`;
- network latency;
- proxy;
- provider response time;
- output disk speed during response streaming.

Because streaming writes to disk, a blocked or failing filesystem can contribute to request completion time.

---

## 106. Fish `400`

Category:

```text
request validation failed
```

Exit:

```text
4
```

The local request passed client validation, but Fish rejected it.

Check:

- model;
- format-specific fields;
- reference ID;
- request parameter combinations;
- provider-side API changes.

Blind retry is usually not useful.

---

## 107. Fish `401`

Category:

```text
authentication failed
```

Check:

- exact key;
- no quotes;
- no BOM;
- correct account;
- revoked key;
- wrong environment or provider endpoint.

Do not repeatedly retry a bad credential.

---

## 108. Fish `402`

Category:

```text
payment required
```

Check:

- credits;
- plan;
- selected model;
- billing status;
- provider account.

The CLI cannot repair provider billing.

---

## 109. Fish `403`

Category:

```text
permission denied
```

Possible causes:

- key lacks model access;
- voice/reference access denied;
- account restriction;
- endpoint permission.

Authentication can be valid while authorization fails.

---

## 110. Fish `404`

Category:

```text
resource not found
```

Check:

- base URL path;
- model;
- reference ID;
- proxy route;
- provider resource deletion.

Remember the client appends `/v1/tts`.

---

## 111. Fish `422`

Category:

```text
request validation failed
```

This commonly indicates a semantically invalid request accepted by local validation but rejected by the remote service.

Inspect the bounded provider message.

---

## 112. Fish `429`

Category:

```text
rate limit exceeded
```

The client treats `429` as retryable.

Default retry settings:

```text
max attempts: 3
initial delay: 500 ms
maximum delay: 5 s
```

The count includes the first request.

---

## 113. `429` returns immediately instead of waiting

Possible reasons:

- maximum attempts already reached;
- `Retry-After` requests a delay above configured `maxDelay`;
- context canceled;
- invalid retry configuration would have failed earlier;
- final attempt returned `429`.

A valid `Retry-After` above the maximum is not clamped.

Retry stops and returns the API error.

---

## 114. `429` appears to hang

The client can be waiting before another attempt.

Without `Retry-After`, backoff starts at:

```text
initialDelay
```

and doubles up to:

```text
maxDelay
```

With `Retry-After`, the server delay is used when valid and not above the maximum.

Use `SIGINT` or `SIGTERM` to cancel.

---

## 115. Fish `5xx`

Category:

```text
server error
```

By default:

```text
retryServerErrors = false
```

Therefore a `5xx` normally returns after the first response.

Enable server-error retry only when duplicate synthesis risk and provider semantics are acceptable.

---

## 116. Transport error is not retried

Current retry classification covers:

```text
429
optionally 5xx
```

It does not retry:

- DNS failure;
- connection refusal;
- TLS failure;
- request timeout;
- other transport errors.

An external supervisor can implement a higher-level policy with appropriate duplicate-request caution.

---

## 117. Retry may consume additional provider quota

There is no idempotency key.

A request can reach the provider even when the local caller later sees a failure.

Do not assume every code `4` is free to repeat.

---

## 118. Error response body too large

The client reads non-2xx response bodies through a bounded reader.

Default maximum:

```text
64 KiB
```

If exceeded, the returned error can contain both:

- typed API status;
- bounded-read failure.

The limit protects logs and memory.

---

## 119. Fish returned malformed JSON error

The client falls back to the trimmed bounded body as plain text.

A malformed provider error body does not hide the HTTP status category.

---

## 120. Fish returned empty error body

The error still contains the HTTP status.

Provider detail may simply be absent.

Use the status category and request context.

---

## 121. Empty successful audio response

Error:

```text
Fish API returned an empty audio response
```

The server returned 2xx but zero audio bytes.

The temp output is not published.

Check provider behavior, proxy response handling, and selected model.

---

## 122. Response streaming failed

Error prefix:

```text
stream synthesis response
```

Possible causes:

- remote connection closed mid-body;
- proxy interruption;
- local output write failure;
- disk full;
- context cancellation.

No retry occurs after successful response streaming begins.

Retrying into a non-rewindable writer could duplicate or concatenate bytes.

---

## 123. Partial audio temp file

Partial bytes can exist in the hidden temp file while streaming.

On a normal returned error, cleanup attempts to remove it.

An abrupt process death or cleanup failure can leave it behind.

The final destination is not published before successful stream, sync, close, and rename.

---

## 124. Output parent does not exist

Error prefix:

```text
create temporary output file
```

The output package does not create parent directories.

### Correction

```bash
clear

install -d -m 0750 \
  /absolute/path/to/output-parent
```

Then rerun.

---

## 125. Output path is relative to the wrong directory

Relative output follows process cwd.

It is not relative to config or project root.

### Diagnose

```bash
clear

pwd
```

### Correction

Use an absolute output path.

---

## 126. Output path is whitespace

The CLI rejects only an exact empty output path.

It does not trim whitespace.

A blank-looking filename can be created.

Use a normal explicit path and remove accidental whitespace from the caller.

---

## 127. Output extension does not match bytes

The format is selected by `--format`, not extension.

Example:

```text
--format mp3 --output speech.wav
```

creates MP3 bytes under a `.wav` name.

Correct the caller’s filename convention.

---

## 128. Output parent permission denied

### Checks

```bash
clear

output='/absolute/path/to/output.opus'
parent="$(dirname "$output")"

stat -c \
  'type=%F mode=%a owner=%U group=%G path=%n' \
  "$parent"

test -w "$parent"
printf 'parent_writable=%d\n' "$?"
```

The process needs create, rename, open-directory, and sync permissions.

---

## 129. Disk full

Check both blocks and inodes:

```bash
clear

output='/absolute/path/to/output.opus'
parent="$(dirname "$output")"

df -h "$parent"
df -i "$parent"
```

Failure can occur during:

- temp creation;
- stream write;
- temp sync;
- directory sync.

Before rename, the old destination remains.

After rename, the new destination may already exist despite code `4`.

---

## 130. Quota exceeded

Filesystem free space can be available while user, group, or project quota is exhausted.

Check platform-specific quota tools.

The application does not preflight or reserve space.

---

## 131. Existing output was preserved after failure

This is expected for failures before successful rename.

The application writes a separate temp file.

It does not truncate the old destination first.

---

## 132. Existing output mode changed to `0600`

A successful replacement publishes the newly created temp file.

Final mode is:

```text
0600
```

The old file’s mode, ACL, xattrs, owner-specific metadata, and hard-link topology are not copied.

Apply policy at the directory or post-publication layer when different metadata is required.

---

## 133. Output symlink became a regular file

This is expected.

The final rename replaces the symlink leaf rather than following its target.

The target remains unchanged.

Parent directory components can still involve symlinks through ordinary path resolution.

---

## 134. Destination is a directory

Rename normally fails with prefix:

```text
replace output file
```

Use a file path beneath the directory, not the directory path itself.

---

## 135. Exit `4` but output exists

Read the nested error.

If it begins:

```text
persist output replacement
```

then:

- temp sync succeeded;
- temp close succeeded;
- rename succeeded;
- new output was published;
- directory sync or close failed.

The application intentionally keeps the new file.

Do not delete it automatically and assume the old file can be restored.

---

## 136. Verify output after ambiguous failure

```bash
clear

output='/absolute/path/to/output.opus'

if [ -e "$output" ]; then
  stat -c \
    'type=%F mode=%a bytes=%s mtime=%y path=%n' \
    "$output"
else
  printf 'output absent\n'
fi
```

Existence alone cannot prove which concurrent invocation created the file.

Use unique output paths.

---

## 137. `persist output replacement` on network filesystem

Some network or userspace filesystems differ in:

- directory sync support;
- rename semantics;
- error timing;
- durability.

Test on the real deployment filesystem.

A local filesystem success does not certify a remote mount.

---

## 138. Stale temp files

Pattern:

```text
.<destination-base>.*.tmp
```

Inspect:

```bash
clear

find /absolute/path/to/output-parent \
  -maxdepth 1 \
  -type f \
  -name '.*.*.tmp' \
  -printf '%TY-%Tm-%Td %TH:%TM:%TS %m %s %p\n'
```

Possible causes:

- `SIGKILL`;
- power loss;
- runtime crash;
- cleanup remove failure.

Confirm no active process owns the file before deleting it.

---

## 139. Temp cleanup error

The returned error can join:

- primary synthesis/output error;
- close-temp error;
- remove-temp error.

Preserve the complete multi-line error.

The old destination normally remains because publication did not succeed.

---

## 140. Two writers target the same output

There is no destination lock.

Each invocation creates an independent temp.

The last successful rename wins.

Use unique names:

```text
output/<job-id>.opus
```

Do not infer safe coordination merely because temp names are unique.

---

## 141. Special output paths

Do not use:

```text
-
/dev/stdout
/dev/null
```

as a streaming protocol.

The output package treats them as filesystem paths and performs temp creation plus rename.

Use a regular file.

---

## 142. Output file is absent after code `3`

This is expected.

Code `3` occurs before `WriteAtomic` starts.

Possible artifacts still include:

- processed module side effects;
- newly created empty secret file.

Fish synthesis was not sent by the core Fish client.

---

## 143. Output file is absent after code `4`

Possible causes include:

- transport or API failure;
- response stream failure;
- temp create/write/sync/close failure;
- rename failure.

Inspect the nested prefix.

An old destination may still exist.

---

## 144. Output file is corrupt despite code `0`

First verify the format/extension pairing.

Then verify size:

```bash
clear

stat -c 'bytes=%s path=%n' \
  /absolute/path/to/output
```

Code `0` means the received byte stream was non-empty and publication succeeded.

It does not decode or validate the audio container.

Possible causes:

- provider returned invalid media;
- proxy transformed body;
- wrong extension;
- downstream decoder mismatch.

---

## 145. Command returns `0` but log close fails

Possible final sequence:

```text
synthesis completed
log file closing failed
```

The deferred close failure is logged to bootstrap stderr.

It does not change the selected process status.

The output operation already succeeded.

---

## 146. Runtime log write failure is not reflected in exit status

Ordinary `slog` calls do not return handler write errors to command code.

A full log disk can cause diagnostic loss without changing business status.

Monitor:

- log filesystem usage;
- stderr collector health;
- file permissions.

The exit code remains the primary machine result.

---

## 147. No `synthesis completed` but exit is `0`

Possible explanation:

- final log write failed;
- collector dropped stderr/file record;
- logging threshold or collection issue;
- wrapper reported a different process status.

Verify the actual shell status and output file.

A normal successful code path attempts the event.

---

## 148. `log file closing failed`

Check:

- filesystem errors;
- descriptor state;
- storage health;
- network filesystem;
- process limits.

The error is diagnostic-only for the selected exit status.

It appears only on stderr.

---

## 149. Text unexpectedly appears in logs

Check:

```json
{
  "logging": {
    "logText": true
  }
}
```

When true, top-level input and processed output text are logged.

Set false for sensitive data.

Module errors can still contain module-generated details, so module authors must avoid embedding full text casually.

---

## 150. Text does not appear despite `logText=true`

Confirm:

- the intended config loaded;
- logging level includes INFO;
- you are inspecting the correct request ID;
- the command reached text processing;
- JSON/text collector is not truncating fields.

The setting controls top-level text fields, not arbitrary module intermediate text.

---

## 151. Character count seems wrong

Log fields count UTF-8 runes.

They do not count:

- bytes;
- grapheme clusters;
- visible glyphs.

Emoji sequences and combining marks can produce a count different from human-perceived characters.

---

## 152. Command seems stuck before any processing log

Possible stages:

- waiting for stdin;
- blocked config filesystem;
- blocked log filesystem;
- slow module builder;
- external filesystem issue.

Use process inspection:

```bash
clear

ps -o \
  pid,ppid,stat,etime,wchan:32,cmd \
  -C fish-audio-cli
```

`strace` can reveal syscall blocking on Linux, but its output can expose paths and data. Use it carefully.

---

## 153. Command seems stuck after `synthesis started`

Possible causes:

- HTTP request in progress;
- provider generating audio;
- retry delay;
- response streaming;
- slow/full output filesystem;
- directory sync.

Default per-attempt timeout is 120 seconds.

Retry can extend total duration.

---

## 154. Cancel a running synthesis

Send:

```text
SIGINT
SIGTERM
```

Interactive:

```bash
clear

kill -TERM <pid>
```

During pipeline, normal result is code `3`.

During synthesis/retry/output callback, normal result is code `4`.

`SIGKILL` bypasses cleanup and can leave temp files.

---

## 155. Shell reports `130`, `143`, or `137`

Common shell conventions:

```text
130 = SIGINT
143 = SIGTERM
137 = SIGKILL
```

These are not explicit `run()` return values.

They can occur when:

- signal arrives before handler installation;
- process receives an unhandled signal;
- `SIGKILL` is used;
- a supervisor translates termination.

---

## 156. Service restarts endlessly on code `2`

Code `2` usually requires local correction:

- arguments;
- config;
- logging path;
- module config;
- input.

Automatic immediate restart rarely helps.

Configure the supervisor to avoid tight restart loops for permanent setup failures.

---

## 157. Service restarts endlessly on missing secret

Code `3` plus:

```text
empty secret file created
```

requires provisioning.

Repeated restart only reopens the empty file and fails again.

Populate it before restarting.

---

## 158. Blind retry after code `4`

Code `4` can mean:

- request never connected;
- provider rejected it;
- provider processed it;
- audio streamed partially;
- output published but directory sync failed.

Do not apply one universal immediate retry policy.

Inspect the error class and output state.

---

## 159. Safe retry guidance

Generally safe after correction:

```text
code 1
code 2
code 3
```

with the caveat that modules may own external side effects.

Code `4` requires more care because the Fish request may have reached the provider.

Use unique output names and provider-aware retry policy.

---

## 160. Collect a safe diagnostic bundle

```bash
clear

bundle='/tmp/fish-audio-cli-diagnostic.txt'
config='/absolute/path/to/config.json'
output='/absolute/path/to/output.opus'
secret='/absolute/path/to/fish-api-key'

{
  printf '%s\n' '== time =='
  date --iso-8601=seconds

  printf '%s\n' '== system =='
  uname -a

  printf '%s\n' '== go =='
  go version 2>&1 || true

  printf '%s\n' '== binary =='
  file ./bin/fish-audio-cli 2>&1 || true
  sha256sum ./bin/fish-audio-cli 2>&1 || true

  printf '%s\n' '== config metadata =='
  stat -c \
    'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
    "$config" 2>&1 || true

  printf '%s\n' '== secret metadata only =='
  stat -c \
    'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
    "$secret" 2>&1 || true

  printf '%s\n' '== output metadata =='
  stat -c \
    'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
    "$output" 2>&1 || true

  printf '%s\n' '== filesystem =='
  df -h "$(dirname "$output")" 2>&1 || true
  df -i "$(dirname "$output")" 2>&1 || true

  printf '%s\n' '== git =='
  git status --short 2>&1 || true
  git log -1 --oneline --decorate 2>&1 || true
} > "$bundle"

printf 'wrote %s\n' "$bundle"
```

This intentionally does not include:

- config contents;
- secret contents;
- input text;
- output bytes;
- environment variables.

Review every bundle before sharing it.

---

## 161. Add stderr to the bundle

After a controlled reproduction:

```bash
clear

cat /tmp/fish-audio-cli.stderr \
  >> /tmp/fish-audio-cli-diagnostic.txt
```

Review for:

- text content;
- provider messages;
- paths;
- sensitive identifiers.

---

## 162. Check secret structure without content

```bash
clear

python3 - <<'PY'
from pathlib import Path

path = Path("/absolute/path/to/fish-api-key")
data = path.read_bytes()

print(f"bytes={len(data)}")

try:
    text = data.decode("utf-8")
except UnicodeDecodeError as exc:
    print(f"valid_utf8=false error={exc}")
    raise SystemExit(0)

print("valid_utf8=true")
print(f"lf_count={text.count(chr(10))}")
print(f"cr_count={text.count(chr(13))}")
print(f"starts_with_whitespace={bool(text[:1] and text[:1].isspace())}")
print(f"ends_with_lf={text.endswith(chr(10))}")
print(f"ends_with_crlf={text.endswith(chr(13) + chr(10))}")
PY
```

This reveals shape metadata, not the key.

The odd-looking diagnostic restraint is deliberate. Credentials have a habit of appearing in bug reports and then developing independent careers.

---

## 163. Check local config JSON without printing it

```bash
clear

python3 - <<'PY'
from pathlib import Path
import json

path = Path("/absolute/path/to/config.json")
data = path.read_bytes()

print(f"bytes={len(data)}")

try:
    text = data.decode("utf-8")
except UnicodeDecodeError as exc:
    raise SystemExit(f"invalid UTF-8: {exc}")

try:
    json.loads(text)
except json.JSONDecodeError as exc:
    raise SystemExit(f"JSON syntax error: {exc}")

print("generic JSON syntax valid")
print("note: this does not detect all strict application rules")
PY
```

The CLI remains authoritative for:

- duplicate keys;
- exact fields;
- null rules;
- semantic validation.

---

## 164. Confirm the binary matches the repository

```bash
clear

git log -1 --oneline --decorate
go build \
  -trimpath \
  -o /tmp/fish-audio-cli-check \
  ./cmd/fish-audio-cli

sha256sum \
  ./bin/fish-audio-cli \
  /tmp/fish-audio-cli-check
```

Different hashes do not always prove different source because build metadata and environment can vary.

A clearly old timestamp or missing rebuilt feature remains useful evidence.

---

## 165. Run the local quality gates

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

Any failing command should be diagnosed before blaming deployment.

---

## 166. `gofmt -l .` prints files

Those files are not formatted.

Correct:

```bash
clear

gofmt -w .
```

Review the diff afterward.

---

## 167. `go vet ./...` fails

Read the exact analyzer output.

Common categories include:

- formatting directive mismatch;
- bad struct tag;
- suspicious standard-library usage.

Do not suppress vet globally to land one warning.

Fix the code or justify a narrowly scoped change.

---

## 168. `go test ./...` passes but CI test fails

CI uses:

```text
-count=1
```

Your plain run may have reused cached success.

Reproduce:

```bash
clear

go test -count=1 ./...
```

Also ensure your Go version matches `go.mod`.

---

## 169. Race test fails

Reproduce the focused package:

```bash
clear

go test \
  -race \
  -count=20 \
  ./path/to/package
```

Inspect both conflicting goroutine stacks.

The race can be in:

- production code;
- test handler state;
- package globals;
- `os.Args` mutation;
- shared buffers;
- cleanup.

---

## 170. Main-package tests race around `os.Args`

Tests that modify `os.Args` must not run in parallel.

Save and restore it with `t.Cleanup`.

Do not add `t.Parallel()` to such tests.

---

## 171. Tests fail only under `-race`

Race instrumentation changes timing and makes data races visible.

Do not dismiss the failure as “the race detector being slow.”

Check synchronization and timeout assumptions.

---

## 172. Test hangs

Run the package verbosely with a shorter timeout:

```bash
clear

go test \
  -count=1 \
  -v \
  -timeout=30s \
  ./path/to/package
```

Inspect the goroutine dump.

Common causes:

- channel never closed;
- HTTP body not closed;
- server shutdown waiting;
- context not canceled;
- real retry sleep;
- global deadlock.

---

## 173. HTTP tests try to reach the Internet

Normal tests should use:

```text
httptest.Server
custom RoundTripper
```

No real Fish credential or live API is required.

A test using the configured production URL is a test defect unless it belongs to an explicit opt-in live suite.

---

## 174. Filesystem tests fail under another user

Check assumptions about:

- chmod permission;
- directory ownership;
- umask;
- symlink creation;
- read-only mounts;
- network filesystem;
- mandatory access control.

Current CI runs on Linux `ubuntu-latest`.

Cross-platform equivalence is not currently a CI guarantee.

---

## 175. Test leaves files in repository

Tests should use:

```go
t.TempDir()
```

They should not create runtime artifacts under:

```text
config/
secrets/
logs/
bin/
```

A dirty working tree after tests indicates a test or manual workflow defect.

---

## 176. Config default seems different from docs

Inspect the authoritative default code and example configuration together.

Then run tests.

Default values currently include:

```text
input max: 1 MiB
secret max: 16 KiB
Fish error body max: 64 KiB
Fish attempts: 3
initial retry delay: 500 ms
maximum retry delay: 5 s
Fish timeout: 120 s
model: s2.1-pro-free
pipeline policy: use_previous
module: passthrough
logging level: info
logging format: text
logText: false
```

A documentation mismatch should be corrected before release.

---

## 177. Existing README contradicts behavior

During the documentation rewrite, older top-level text can lag behind normative docs and code.

Use the specialized document for the relevant subsystem.

The final documentation pass rewrites `README.md` after the subsystem references and troubleshooting guide are complete.

Do not silently preserve a known false operational workaround.

---

## 178. Wrong model or voice

Confirm the `config loaded` event includes the expected:

```text
fish_model
```

The reference ID is not included in that startup event.

Inspect the intended config securely.

A model can be syntactically valid but unavailable to the account, producing remote `402`, `403`, `404`, or validation failure.

---

## 179. Model availability changed

Model availability is provider-controlled.

The CLI validates local shape, not current commercial availability.

A remote error is authoritative.

Change the configured model without rebuilding the binary.

---

## 180. Wrong request reaches a proxy path

The client joins:

```text
base URL path
+
v1/tts
```

For a proxy base:

```text
https://example.test/fish
```

the endpoint becomes conceptually:

```text
https://example.test/fish/v1/tts
```

Configure the base path accordingly.

---

## 181. Authorization header missing at proxy

The client sends:

```text
Authorization: Bearer <key>
model: <model>
Content-Type: application/json
```

A reverse proxy can remove or rewrite headers.

Inspect proxy configuration without logging the full authorization value.

---

## 182. Proxy returns HTML error page

The client treats a non-2xx body as bounded diagnostic text when Fish JSON parsing does not succeed.

An HTML gateway page can therefore appear in the error.

The HTTP status still determines the stable category where mapped.

---

## 183. Provider request succeeded but local disk failed

Possible sequence:

```text
Fish returns audio
stream writes temp
temp sync/rename or directory sync fails
exit 4
```

Provider work and quota may already be consumed.

Fix the filesystem before retrying.

---

## 184. Local disk succeeded but provider connection ended late

A streaming response can write partial data before a remote error.

The partial data remains only in temp unless every later publication stage succeeds.

The output layer cleans it on normal failure.

---

## 185. Final output exists from an older run

Before interpreting an exit `4`, record destination metadata before the run or use a unique new path.

Otherwise you cannot distinguish:

- old file preserved;
- new file published;
- concurrent writer output.

Unique job paths simplify recovery.

---

## 186. Use unique output names

Example:

```bash
clear

job_id="$(date +%s)-$$"
output="/var/lib/fish-audio-cli/output/${job_id}.opus"

./bin/fish-audio-cli \
  --config /absolute/path/to/config.json \
  --format opus \
  --output "$output" \
  --text 'Hello'
```

For stronger uniqueness, use a generated random identifier.

---

## 187. Existing readers hear old audio

Atomic replacement changes the directory entry.

A process that opened the old file before rename can continue reading the old inode under normal Unix semantics.

Open the path after successful command completion to consume the new output.

---

## 188. Watcher reacts to temp files

Directory watchers can see hidden temp creation and writes.

Treat only the final path as published.

Do not ingest:

```text
.<baseName>.*.tmp
```

as completed audio.

---

## 189. Log watcher sees partial records or rotation issues

The logger appends one structured record through `slog`, but filesystem and collector behavior can still split reads.

Use line-oriented collection.

For rotation, use the provided template and correct absolute path.

The CLI does not perform internal rotation.

---

## 190. Log file grows without limit

The application has no internal:

- rotation;
- retention;
- size cap.

Install external rotation.

The repository template uses:

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

---

## 191. Rotated log is not recreated immediately

With `nocreate`, the CLI creates the next file on its next invocation.

If no invocation occurs, no new active file appears.

This is expected.

---

## 192. Log owner is wrong after service-user change

The application changes mode, not owner or group.

Remove or chown the existing file and directory deliberately.

Ensure the new service user can append and chmod.

---

## 193. Secret owner is wrong after service-user change

The loader changes mode to `0600`, not owner.

The process must be able to read and chmod the file.

Correct ownership outside the application.

---

## 194. Config file contains secrets

The Fish key is not an inline configuration field.

Unknown inline fields are rejected.

Store the key in the separate configured file.

Protect config anyway because it controls:

- endpoint;
- model;
- secret path;
- log path;
- modules.

---

## 195. Custom secret path was committed

The default `/secrets/` directory is ignored.

A custom repository-local path may not be.

Immediately:

- remove it from version control;
- rotate the exposed key;
- add an ignore rule;
- audit remote history and logs.

Deleting one working-tree file does not revoke a leaked credential.

---

## 196. Output contains sensitive speech

Final output mode is `0600`.

Still protect:

- parent directory;
- backups;
- downstream copies;
- temporary filesystem;
- process privileges.

Audio content is not encrypted by the application.

---

## 197. Error logs disclose provider details

Remote error messages are included within a configured byte limit.

They can reveal:

- account status;
- model names;
- resource IDs;
- proxy messages.

Protect logs and sanitize before sharing.

---

## 198. `request_id` differs between attempts

Fish retry attempts occur inside one command invocation.

They should share the same application request ID because logging correlation is per invocation.

The Fish client does not currently emit per-attempt structured events, so attempt count can require server-side or test instrumentation.

---

## 199. No explicit retry log events

Current command logs:

```text
synthesis started
synthesis failed or completed
```

It does not emit one normal log record per retry attempt.

A delay after `synthesis started` can be request processing or retry wait.

Use final error and configured policy to infer likely behavior.

---

## 200. Need exact retry observability

Adding attempt-level logs is a code and logging-contract change.

It should include:

- attempt number;
- maximum attempts;
- category;
- selected delay;
- no API key;
- bounded safe fields.

Do not infer this data by enabling `logText`.

---

## 201. `Retry-After` in the past

An HTTP-date earlier than current time becomes a zero delay.

The client can retry immediately unless context is canceled.

---

## 202. Invalid `Retry-After`

An invalid header is ignored for delay selection.

The client falls back to exponential backoff.

---

## 203. `Retry-After` above max delay

The client does not clamp.

It stops retrying and returns the current API error.

Increase `maxDelay` only when waiting that long is operationally acceptable.

---

## 204. Retry server errors disabled

Default:

```text
false
```

A `500` or `503` returns immediately.

`429` remains retryable regardless of this flag.

---

## 205. Maximum attempts set to `1`

No retry occurs.

The first result is final.

This can simplify duplicate-risk handling at the cost of resilience to `429`.

---

## 206. Cancellation during retry wait

The returned error joins:

- the previous Fish API error;
- `wait before Fish API retry`;
- context cancellation.

Exit remains:

```text
4
```

---

## 207. Error contains multiple lines

The project uses `errors.Join`.

A single failure can preserve more than one cause.

Examples:

- primary plus close;
- API status plus error-body read;
- output primary plus cleanup;
- directory sync plus close.

Do not keep only the first line in a bug report.

---

## 208. Error string matching is brittle

Shell automation can use exit codes for broad stages.

Fine categories are available reliably to Go package callers through:

```text
errors.Is
errors.As
```

Structured JSON logs do not currently serialize a stable `error_class`.

Avoid permanent automation based on English substrings where a package integration is possible.

---

## 209. A warning accompanies failure

`empty secret file created` uses WARN but exits `3`.

Severity does not equal process success.

Always inspect the status.

---

## 210. An error accompanies success

A module can log `module processing failed`, recover through policy, and the command can return `0`.

A deferred log close can also log ERROR after successful synthesis without changing status.

Use lifecycle and status together.

---

## 211. Output was never requested from Fish

Codes `1`, `2`, and core code `3` occur before Fish synthesis.

However, custom modules can perform their own external work during pipeline processing.

Distinguish the core Fish client from module-owned side effects.

---

## 212. Fish may have received the request

Code `4` can occur after:

- request send;
- provider processing;
- response start;
- full response;
- local publication.

Treat retries as potentially duplicate remote work.

---

## 213. Diagnose by artifact state

| Artifact | Meaning |
|---|---|
| no persistent log | early failure or log failure |
| empty secret newly exists | missing-secret bootstrap |
| old output unchanged | failure before rename, absent concurrency |
| hidden temp remains | abnormal death or cleanup failure |
| new output exists with code `4` | possible post-rename persistence failure |
| log contains module error and success | policy recovery |
| output mode `0600` | normal publication mode |

---

## 214. Diagnose by last successful message

### `config loaded`

Config and logger succeeded.

Next suspects:

- modules;
- input.

### `text processing started`

Input succeeded.

Next suspects:

- module execution;
- cancellation.

### `text processing completed`

Pipeline succeeded or recovered.

Next suspects:

- request;
- secret;
- client.

### `synthesis started`

All local setup succeeded.

Next suspects:

- Fish HTTP;
- streaming;
- output.

---

## 215. No `config loaded`

Possible stages:

- help;
- parse failure;
- path failure;
- config load;
- config validation;
- logger open.

Persistent logging may be absent.

Inspect stderr.

---

## 216. `config loaded` but no module event

Possible causes:

- module initialization failed before execution;
- input failed;
- empty pipeline proceeded directly;
- logging threshold hides events only if their level is below threshold.

Module execution starts only after valid input.

---

## 217. Module start without completion

Possible causes:

- processor blocked;
- processor panicked;
- cancellation;
- process termination;
- logging destination failure.

Use module-specific diagnostics and context cancellation.

A panic is outside the normal exit-code contract.

---

## 218. Panic stack trace

An uncaught panic indicates a defect or invalid package integration.

Preserve:

- full stack trace;
- commit SHA;
- Go version;
- input shape without sensitive content;
- config structure with secrets removed;
- reproduction command.

Run:

```bash
clear

go test -count=1 ./...
go test -race -count=1 ./...
```

Do not convert a panic into a guessed operational configuration issue without evidence.

---

## 219. Process killed by OOM

The application normally streams audio rather than buffering the complete response.

Possible memory pressure sources:

- very large input within configured limit;
- module implementations;
- large configs or secrets within raised limits;
- concurrent processes;
- system-wide pressure;
- Go runtime overhead.

Inspect kernel or container OOM logs.

---

## 220. Many concurrent processes exhaust descriptors

Each invocation can open:

- stderr;
- config;
- log;
- secret directory;
- secret file;
- HTTP connection;
- temp output;
- output directory.

Check limits:

```bash
clear

ulimit -n
```

Use controlled concurrency and ensure resource closure.

---

## 221. Too many concurrent requests

Fish can respond `429`.

The CLI’s internal retry can multiply request count during a burst.

Coordinate concurrency at the caller.

Use backpressure and unique outputs.

---

## 222. Proxy or VPN causes intermittent transport failures

Because transport errors are not retried internally, intermittent routing can surface immediately as code `4`.

Diagnose:

- route;
- DNS;
- proxy;
- VPN stability;
- MTU;
- TLS interception.

An external retry should account for possible request delivery.

---

## 223. IPv6/IPv4 mismatch

Go’s transport follows system resolution and routing.

A hostname can resolve to addresses that are unreachable in the deployment network.

Check:

```bash
clear

getent ahosts api.fish.audio || true
```

Fix system networking rather than hard-coding provider IPs, which can change and break TLS.

---

## 224. System clock is wrong

Clock skew can affect:

- TLS certificate validation;
- HTTP-date `Retry-After`;
- log timestamps;
- file timestamps.

Check:

```bash
clear

date --iso-8601=seconds
timedatectl status 2>/dev/null || true
```

Correct time synchronization.

---

## 225. Output sync is slow

`Sync` requests persistence and can block on slow storage.

Possible contributors:

- network filesystem;
- overloaded disk;
- failing disk;
- synchronous mount options;
- virtualization storage;
- full journal.

Measure storage separately.

Do not remove sync calls merely to hide latency; that changes durability guarantees.

---

## 226. Log sync expectations

The logger does not explicitly call file `Sync` before close.

A successful log write is not the same durability contract as atomic output.

Do not infer output durability behavior from logging behavior.

---

## 227. Secret file final newline

Accepted:

```text
key
key + LF
key + CRLF
```

The one final line ending is removed.

A bare final CR is not accepted as the standard line ending and remains a control/newline violation.

---

## 228. Windows-created secret file

A normal CRLF final line is accepted.

Multiple CRLF lines are not.

Ensure the file is UTF-8 rather than UTF-16.

---

## 229. Secret copied with clipboard newline

One final LF is acceptable.

Two or more lines are not.

Use the structure-only diagnostic before sharing anything.

---

## 230. Config key case differs

JSON field names are exact.

A case variant can be rejected as unknown.

Use the documented lowercase/camelCase spelling exactly.

---

## 231. Module type case differs

Registry type lookup is exact.

Use:

```text
passthrough
```

not:

```text
Passthrough
PASSThrough
```

---

## 232. Logging config seems normalized but validation rejects it

Low-level logging helpers trim/lower values.

Public config validation expects the documented exact values.

Use clean lowercase config values rather than depending on lower-level normalization.

---

## 233. Output filename has parent traversal

The output path is caller-controlled and not sandboxed.

A privileged wrapper must not pass arbitrary untrusted paths.

Constrain outputs at the caller layer.

---

## 234. Configured path escapes project root

The project resolver performs lexical joining.

It does not confine `..` beneath the project directory.

Example:

```text
../shared/key
```

can resolve outside the project.

Protect configuration integrity.

---

## 235. Log path points to a symlink

Logging does not apply secret-style leaf-symlink hardening.

Use a trusted log directory and regular file path.

Do not place logs in an untrusted writable directory.

---

## 236. Config path itself is a symlink

Allowed by current config loading.

Protect both:

- symlink entry;
- target.

Remember project-relative paths use the lexical supplied location.

---

## 237. Output parent is untrusted

Leaf symlink replacement does not make parent path races safe.

Use a dedicated directory not writable by untrusted users.

The output package uses ordinary path-based operations.

---

## 238. Secret parent contains symlink components

The loader anchors leaf operations in the final opened directory.

It does not reject every symlink in ancestor components or verify every ancestor owner/mode.

Use a trusted absolute directory hierarchy.

---

## 239. Need stderr-only logging

Not currently supported.

Operational alternatives:

- configure a writable dedicated file;
- collect stderr and rotate the file;
- add an explicit supported configuration feature in code.

Do not use a special device workaround.

---

## 240. Need output on stdout

Not currently supported.

The output contract requires a final file path and atomic publication.

Adding stdout streaming would need separate semantics for:

- partial data;
- retries;
- exit status;
- binary stdout;
- logging separation.

---

## 241. Need inline API key

Not supported.

There is no:

```text
--api-key
environment fallback
inline JSON key
stdin secret mode
```

Use the protected file.

---

## 242. Need environment substitution

Not supported in configuration.

Perform substitution in a trusted deployment templating step that writes the final JSON.

Then validate the generated file.

Do not hand the application unresolved placeholders.

---

## 243. Need live configuration reload

The command is one-shot.

Each invocation reloads config and secret.

There is no long-running daemon or reload signal.

Change files, then start a new invocation.

---

## 244. Key rotation during a request

The invocation reads the key before client creation.

Replacing the file afterward does not change the key already retained by that running client.

The next invocation reads the new value.

---

## 245. Log rotation during an invocation

The open log descriptor continues following its opened file object according to filesystem semantics.

A rotation strategy must account for short-lived one-shot invocations.

The provided `nocreate` template lets the next invocation create the next active file.

---

## 246. Output consumer starts too early

Wait for process exit `0`.

Watching for destination existence alone is insufficient because:

- an old destination may preexist;
- a post-rename error can leave a file with exit `4`;
- another process can write the same path.

Use unique paths and the command status.

---

## 247. Need checksum verification

The application does not compute or log an output checksum.

A caller can compute one after exit `0`:

```bash
clear

sha256sum /absolute/path/to/output.opus
```

This verifies local bytes, not audio semantic validity.

---

## 248. Need audio validation

Use a format-aware external tool after successful publication.

The CLI itself verifies only:

- non-empty Fish 2xx stream;
- successful filesystem publication.

It does not parse WAV, MP3, or Opus containers.

---

## 249. Need detailed provider request tracing

Do not log the authorization header.

A safe trace can include:

- endpoint host/path;
- attempt count;
- status;
- duration;
- model;
- format;
- request ID.

Adding it requires code changes.

Avoid general-purpose HTTP dumps in production.

---

## 250. Need to prove no real Fish request occurred

For codes `1` and `2`, the core Fish client was not reached.

For code `3`, core Fish synthesis was not sent.

For stronger integration proof, use a local `httptest.Server` or controlled proxy and assert request count.

Custom modules can have independent remote behavior.

---

## 251. Need to reproduce Fish errors safely

Use a local HTTP test server in Go tests.

Do not intentionally send malformed or repeated paid requests to production.

The existing test strategy covers:

- typed errors;
- retries;
- oversized error bodies;
- streaming failures;
- cancellation.

---

## 252. Need to reproduce filesystem failures

Prefer package-level tests with narrow injected seams for:

- sync failure;
- close failure;
- cleanup failure;
- directory persistence failure.

Changing real system permissions during production diagnosis can create new incidents.

---

## 253. CI build passes but deployed binary behaves differently

Possible causes:

- different binary copied;
- stale service path;
- different config;
- different cwd;
- different user;
- different filesystem;
- proxy environment;
- missing CA;
- custom module build mismatch.

Record binary checksum and absolute invocation path.

---

## 254. Service uses different working directory

Relative:

```text
--config
--output
```

can resolve differently under systemd, cron, containers, or shell.

Use absolute paths.

Project-relative config fields still use the config-derived project directory.

---

## 255. Cron has no input

Without `--text`, the command reads stdin.

Cron usually supplies empty stdin.

Result:

```text
input failed
exit 2
```

Provide explicit text or pipe a producer.

---

## 256. Systemd service waits for stdin

A service with omitted `--text` can block or immediately see EOF depending on stdin setup.

Pass the text explicitly from the calling integration.

The CLI is intended as a one-shot executable, not an idle stdin service.

---

## 257. Container output disappears

The application can successfully write inside the container filesystem while the file is lost when the container exits.

Mount a persistent output directory and use its absolute path.

This is outside the atomic-file logic.

---

## 258. Container log path is unwritable

Configured logger initialization happens before modules and input.

Mount a writable log directory or bake one with correct ownership.

Persistent logging cannot be skipped.

---

## 259. Container secret path created in image layer

Use an explicit mounted or writable runtime secret location.

Check the resolved path from the warning.

Do not assume it follows the binary location.

---

## 260. SELinux or AppArmor denial

Unix mode bits can look correct while mandatory access control denies:

- config read;
- log append;
- secret chmod;
- output create/rename;
- network access.

Inspect platform audit logs.

Do not weaken all filesystem permissions to compensate for a policy denial.

---

## 261. Read-only root filesystem

A read-only root can still work when writable mounts are provided for:

- log;
- secret when chmod is required;
- output.

Config can be read-only.

Use absolute paths to writable volumes.

---

## 262. Binary runs as root but not as service user

Root can hide:

- ownership errors;
- directory traversal denial;
- chmod limitations;
- proxy/environment differences.

Test under the actual service identity.

Do not solve routine path access by running the whole TTS client as root.

---

## 263. Binary runs manually but not in service

Compare:

```text
cwd
user/group
environment
proxy variables
HOME
CA paths
config path
output path
mounts
resource limits
```

Use absolute paths and explicit service settings.

---

## 264. Binary runs in service but not shell

The service may provide:

- different proxy;
- different permissions;
- mounted secret;
- writable directories;
- different config.

Reproduce under the same identity and environment carefully.

---

## 265. Provider rejects free model

The default model name is locally valid.

Provider access and availability can change.

A remote payment, permission, not-found, or validation response is authoritative.

Configure another model appropriate to the account.

---

## 266. Wrong reference ID

A bad or inaccessible reference can produce remote validation, permission, or not-found errors.

The CLI does not query a voice catalog.

Copy the intended identifier from a trusted provider workflow.

---

## 267. Format-specific bitrate error

Check configuration values for:

- MP3 bitrate;
- Opus bitrate;
- sample rate;
- latency;
- other request parameters.

Local validation catches known ranges.

The provider can enforce additional model-specific constraints.

---

## 268. Output format alias confusion

CLI input:

```text
ogg
```

becomes request format:

```text
opus
```

The output filename remains whatever the caller supplied.

This is expected.

---

## 269. `config loaded` lists no modules

An explicit empty pipeline is valid.

The startup fields should show:

```text
pipeline_module_count=0
```

Input proceeds unchanged.

---

## 270. Repeated passthrough modules

Multiple distinct names with type `passthrough` are valid.

They do not change text.

This can be used to verify ordering and logging but provides no transformation.

---

## 271. Module names appear in wrong order

The configuration array order is authoritative.

Core preserves order.

If logs differ, verify:

- correct config loaded;
- module builder returned expected steps;
- no stale binary;
- no log mixing across request IDs.

---

## 272. Pipeline counts differ from bytes

Pipeline and log character fields count runes.

Input limits count bytes.

A Cyrillic or emoji input can be within one count and appear larger under another.

This is not data loss.

---

## 273. Negative or zero duration in logs

Durations are derived from the monotonic component of Go time values where available and reported in whole milliseconds.

A very fast stage can report:

```text
0 ms
```

A negative value would be unexpected and should be treated as a defect.

---

## 274. No source file/line in logs

The logger does not enable source attribution.

Records contain structured fields but no source location.

Adding source is a logging-contract and performance decision.

---

## 275. Need more detailed log level

Changing from INFO to DEBUG cannot expose events that are not emitted.

For deeper diagnosis, add targeted safe events in code and tests.

Do not enable text logging merely to diagnose network timing.

---

## 276. JSON collector rejects multiline error

Joined errors can serialize as a string containing newline characters.

A correct JSON handler escapes them inside the JSON string.

A downstream line parser must parse JSON rather than split unescaped conceptual content.

---

## 277. Text log parser breaks on spaces

Text-format slog output uses key/value formatting that can quote values.

Do not parse it with a fixed `cut -d ' '` schema.

Use JSON logging for machine consumption.

---

## 278. Need one log destination

Current configured logger always uses two.

A one-destination mode requires a new explicit configuration contract.

Collector-side selection is the operational workaround.

---

## 279. Log file path resolves outside expected project

Check config lexical location and the special immediate-parent `config` rule.

An absolute log path is never rebased.

A relative log path can contain `..` and escape the project.

Protect config.

---

## 280. Secret and log resolve differently from output

This is intentional.

| Path | Relative base |
|---|---|
| config argument | cwd |
| secret config | project directory |
| log config | project directory |
| module path | module contract |
| CLI output | cwd |

Use absolute paths when integrating services.

---

## 281. Diagnostic command changes behavior

Commands such as:

```text
readlink -f
realpath
```

canonicalize symlinks.

The project resolver does not.

Do not use a canonicalized diagnostic path as proof of the application’s lexical project root.

---

## 282. Shell quoting changes text

Use single quotes for literal text where possible.

Shell expansion can modify:

- `$`;
- backslashes;
- command substitutions;
- newlines.

For complex text, pipe a UTF-8 file:

```bash
clear

cat /absolute/path/to/input.txt |
  ./bin/fish-audio-cli \
    --config /absolute/path/to/config.json \
    --format opus \
    --output /absolute/path/to/output.opus
```

---

## 283. Shell adds a newline

`printf '%s'` does not add one.

`echo` often does.

The text contract permits nonblank input including newlines, so the semantic speech can differ.

Use `printf` for exact test input.

---

## 284. Input file contains NUL

NUL is valid UTF-8 as a byte value in a Go string but may be undesirable text for the provider or modules.

The generic text contract rejects invalid UTF-8 and blank text, not every control character.

Provider validation or synthesis may fail.

Sanitize at an explicit text-processing module if required.

---

## 285. Fish API key path has surrounding whitespace

Normal configuration path resolution trims the configured path before the lower-level loader receives it.

Do not rely on padding.

Use a clean path.

Direct package calls to `secrets.Load` reject surrounding whitespace.

---

## 286. Output path has surrounding whitespace

Unlike config-owned paths, the CLI output path is not trimmed.

Correct the caller.

This asymmetry is documented and intentional in the current implementation.

---

## 287. Config path has surrounding whitespace

The project resolver trims it.

A real filename intentionally containing leading or trailing whitespace is therefore not addressable through the normal contract.

Rename the file.

---

## 288. Build fails because Go version is old

Check:

```bash
clear

go version
cat go.mod
```

Install the declared Go version or a compatible newer toolchain according to project policy.

Do not edit `go.mod` downward merely to make an old machine happy.

---

## 289. No third-party modules downloaded

The current `go.mod` contains no external requirements.

A normal build still uses the Go standard library and GitHub Actions.

A future module may introduce dependencies and change this observation.

---

## 290. CI action or runner failure

Distinguish:

- repository test failure;
- checkout/setup action failure;
- hosted-runner outage;
- cache issue.

Reproduce repository commands locally.

Do not change production code to fix a transient CI infrastructure incident.

---

## 291. `git diff --check` fails

It detects whitespace errors.

Run:

```bash
clear

git diff --check
git diff --cached --check
```

Fix the exact reported lines.

For Go files, also run `gofmt`.

---

## 292. Documentation checksum mismatch

The installer workflow verifies the downloaded generated file.

A mismatch means:

- wrong file;
- browser renamed/replaced content;
- partial download;
- manual edit;
- stale artifact.

Do not bypass the checksum.

Download the intended artifact again.

---

## 293. Documentation local link validation fails

The target document may not yet exist or the link name may be wrong.

Check the documentation sequence.

Do not create a dummy file merely to satisfy validation.

---

## 294. Test uses a live key accidentally

Stop the test.

Rotate the key if it entered:

- source;
- logs;
- CI output;
- shell history;
- test fixture.

Replace it with a synthetic value and local server.

---

## 295. API key appears in error output

The project does not intentionally log it.

Treat any appearance as a security bug.

Preserve a redacted reproduction and rotate the key.

Do not paste the original secret into an issue.

---

## 296. Input text appears in an error unexpectedly

Check module-generated errors and provider messages.

`logText=false` controls specific normal top-level fields, not every arbitrary error string created by external code.

Modules must follow a separate privacy discipline.

---

## 297. Provider error body contains input

The client logs the returned bounded message through the error chain.

The provider can echo submitted content.

Protect logs and avoid sharing them unreviewed.

A future redaction policy would require explicit design.

---

## 298. Need a minimal reproducible config

Start from the tracked example, then reduce:

- one passthrough module;
- default request parameters;
- explicit base URL;
- explicit absolute secret path;
- explicit writable log path;
- text logging false.

Do not remove required objects or replace arrays with null.

---

## 299. Minimal command reproduction

```bash
clear

./bin/fish-audio-cli \
  --config /absolute/path/to/minimal-config.json \
  --format opus \
  --output /absolute/path/to/diagnostic.opus \
  --text 'Test'
```

Use a new output path.

Capture stderr and status.

---

## 300. Reduce modules

Use:

```json
{
  "pipeline": {
    "onError": "abort",
    "modules": [
      {
        "name": "passthrough",
        "type": "passthrough",
        "config": {}
      }
    ]
  }
}
```

This isolates custom module behavior while retaining pipeline wiring.

An empty array isolates the pipeline completely.

---

## 301. Compare empty and passthrough pipelines

If both fail before synthesis, the problem is likely outside transformation behavior.

If empty succeeds and custom pipeline fails, focus on module setup or processing.

If passthrough succeeds and custom module fails, compare module config and remote dependencies.

---

## 302. Isolate output filesystem

Use a known writable local directory:

```bash
clear

directory="$(mktemp -d)"
chmod 0700 "$directory"

printf 'diagnostic_directory=%s\n' "$directory"
```

Point output there while retaining the same Fish config.

Do not put the real API key or config into an insecure temp directory.

---

## 303. Isolate log filesystem

Configure a dedicated writable regular path in the project or secure temp test environment.

Remember logger initialization occurs before input and synthesis.

A bad log path can mask every later stage.

---

## 304. Isolate network without Fish

Use package tests and `httptest.Server`.

The production CLI does not expose a dry-run HTTP mode.

Do not send a fake API key to the real endpoint merely to test output handling.

---

## 305. Isolate secret loading

A command invocation reaches secret loading only after pipeline and request construction.

For package-level diagnosis, run secret tests.

For CLI diagnosis, use a valid minimal pipeline and inspect the specific secret event.

---

## 306. Isolate pipeline

Use an empty module array.

If the command proceeds, custom module behavior is implicated.

If it still fails at `text processing failed`, inspect context and application invariants.

---

## 307. Isolate CLI input

Use a literal nonblank `--text`.

This avoids:

- producer hangs;
- stdin encoding;
- pipe closure;
- input file errors.

---

## 308. Isolate format

Use:

```text
opus
```

with a matching `.opus` output name.

Then test other formats separately.

---

## 309. Isolate model access

Use a model known to be available to the account.

Local default validity does not prove current provider access.

Remote billing and permission errors identify account/model issues.

---

## 310. Isolate reference voice

Temporarily remove the reference ID only when the model/API supports that request shape and the test objective permits it.

A reference-specific failure can otherwise look like a general model failure.

Follow the provider contract documented for the selected model.

---

## 311. Error index: exit `1`

| Message | Likely action |
|---|---|
| `logging error` | inspect stderr/runtime writer |
| `request ID generation failed` | inspect entropy/runtime environment |

No persistent log exists.

---

## 312. Error index: exit `2`

| Message | Likely action |
|---|---|
| `option parsing failed` | fix flags |
| `path initialization failed` | fix `--config` |
| `config loading failed` | fix file/JSON/path |
| `config validation failed` | fix semantic value |
| `logger initialization failed` | fix log path/permissions |
| `module initialization failed` | fix registry/module config |
| `module logging initialization failed` | internal step defect |
| `application initialization failed` | pipeline invariant defect |
| `input failed` | fix text source/encoding/size |

---

## 313. Error index: exit `3`

| Message | Likely action |
|---|---|
| `text processing failed` | module/cancellation/policy |
| `Fish request creation failed` | request fields/final text |
| `empty secret file created` | provision key |
| `Fish API key loading failed` | secret path/type/content/mode |
| `Fish client initialization failed` | endpoint/key/model/timeout/retry |

---

## 314. Error index: exit `4`

| Nested error | Likely action |
|---|---|
| `send synthesis request` | network/TLS/timeout |
| `Fish API returned ...` | provider status |
| `read Fish API error response` | provider body/read limit |
| `wait before Fish API retry` | cancellation |
| `stream synthesis response` | remote stream/local writer |
| `empty audio response` | provider/proxy |
| `create temporary output file` | parent/path/permission |
| `write temporary output file` | Fish or local write |
| `sync temporary output file` | storage |
| `close temporary output file` | storage/descriptor |
| `replace output file` | rename/destination |
| `persist output replacement` | directory sync; file published |
| `remove temporary output file` | cleanup |

---

## 315. Do not use `chmod -R 777`

It can:

- expose secrets;
- make secret directories invalid;
- enable path replacement;
- weaken logs and output;
- hide ownership mistakes;
- create a larger incident.

Set the narrow intended modes:

```text
secret directory 0700
secret file 0600
log file 0640
output file 0600
```

Directory modes for log/output depend on deployment ownership and sharing needs.

---

## 316. Do not run as root to test permissions

Root can bypass or alter the behavior being diagnosed.

Use the actual service identity.

Correct ownership and directory policy.

---

## 317. Do not print the secret

Metadata, line counts, and encoding checks are sufficient for most diagnosis.

If authentication still fails, rotate and reprovision rather than publishing the key for inspection.

---

## 318. Do not disable TLS verification

A certificate failure is a network trust problem.

Disabling verification converts a clear error into credential and text exposure risk.

Install the correct CA or fix interception.

---

## 319. Do not retry every code `4` immediately

The provider may already have generated audio.

The new output may already exist.

Inspect cause and state.

---

## 320. Do not delete all hidden temp files blindly

Shared directories can contain unrelated application files.

Match the intended destination pattern and confirm no active process.

Use a dedicated output directory.

---

## 321. Do not edit generated docs in transit

The installer verifies checksum and structure.

Apply intentional source changes in the repository after installation and review, not to the downloaded artifact before the checksum gate.

---

## 322. Do not use README as the only operational authority during rewrite

Subsystem documents are being completed before the final README rewrite.

Prefer the specialized normative document and current code.

The final index and README will link them.

---

## 323. When to file a bug

File a project bug when:

- panic occurs;
- documented valid config is rejected;
- exit code does not match stage;
- secret value is logged;
- pre-rename failure destroys old output;
- post-rename failure removes new output;
- temp cleanup leaks under normal returned failure;
- typed error identity is lost;
- race detector reports production code;
- docs and tested implementation disagree.

---

## 324. Bug report contents

Include:

- commit SHA;
- Go version;
- OS and architecture;
- exact exit status;
- complete stderr with secrets/text reviewed;
- request ID;
- relevant config fragment with credentials removed;
- file metadata;
- output state;
- reproducible command;
- expected behavior;
- actual behavior;
- whether retry could have sent multiple requests.

---

## 325. Bug report exclusions

Do not include:

- Fish API key;
- private input text unless necessary and authorized;
- full environment dump;
- private output audio;
- unrelated server logs;
- access tokens;
- home-directory archives.

Use synthetic reproduction data when possible.

---

## 326. Maintainer triage sequence

1. Confirm commit.
2. Confirm exact exit code.
3. Identify last lifecycle message.
4. Determine whether configured logging was active.
5. Inspect wrapped or joined error.
6. Check stage artifacts.
7. Reproduce with minimal config and input.
8. Run focused tests.
9. Run full uncached and race suites.
10. add a regression test before fixing code.

---

## 327. Troubleshooting invariants

The following rules are useful anchors when diagnosis becomes noisy.

1. Help returns `0`.
2. Help does not open configured logging.
3. Early setup failures are stderr-only.
4. Configured logging always targets stderr and a file.
5. Persistent file logging cannot be disabled.
6. `/dev/null` is not a supported log destination.
7. Relative config uses cwd.
8. Project paths use lexical config location.
9. Immediate parent `config` moves project root up one level.
10. JSON does not expand `~` or environment variables.
11. Output relative path uses cwd.
12. Output parent is not created.
13. Input exact empty argument selects stdin.
14. Input is bounded in bytes.
15. Input must be valid UTF-8 and nonblank.
16. Modules prepare before any processor builds.
17. The current built-in type is `passthrough`.
18. Module errors can recover and still exit `0`.
19. Cancellation bypasses pipeline recovery.
20. Missing secret creates an empty `0600` file and exits `3`.
21. Secret directory must not be group/other writable.
22. Secret leaf must be regular and not a symlink.
23. Existing secret file is forced to `0600`.
24. One final LF or CRLF is accepted.
25. Secret must contain one unpadded line.
26. Fish client adds `/v1/tts`.
27. Transport errors are not internally retried.
28. `429` is retryable.
29. `5xx` retry is optional and disabled by default.
30. `Retry-After` above max delay stops retry.
31. Successful response streaming is not retried after partial output.
32. Fish 2xx zero-byte output is rejected.
33. Temp output is created beside destination.
34. Final output mode is `0600`.
35. Before rename, old destination is preserved.
36. After rename, directory-sync failure keeps new output.
37. Exit `4` can coexist with an output file.
38. Concurrent same-path writers are unlocked.
39. Runtime log write errors do not change exit status.
40. Log close failure does not change exit status.
41. Shell signal statuses can fall outside `0` through `4`.
42. CI runs uncached tests.
43. CI runs the race detector.
44. Normal tests do not require live Fish.
45. Error severity and exit status are independent.

---

## 328. Quick decision tree

```text
No structured output?
    ├─ shell launch error → binary/path/permissions
    └─ raw logging error → exit 1

Structured error, exit 2?
    ├─ option/path/config before logger → stderr only
    └─ module/input after logger → stderr + file

Exit 3?
    ├─ text processing failed → pipeline/module/cancel
    ├─ request creation failed → final request
    ├─ empty secret created → populate file
    ├─ secret loading failed → filesystem/content
    └─ client init failed → endpoint/key/model/retry

Exit 4?
    ├─ Fish/API/stream prefix → network/provider
    └─ output prefix → filesystem/publication
         └─ persist output replacement
              → output already published
```

---

## 329. Fast safe commands

Show help:

```bash
clear

./bin/fish-audio-cli --help
```

Check repository:

```bash
clear

git log -1 --oneline --decorate
git status --short
```

Run tests:

```bash
clear

go test -count=1 ./...
go test -race -count=1 ./...
```

Check config metadata:

```bash
clear

stat -c \
  'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
  /absolute/path/to/config.json
```

Check secret metadata:

```bash
clear

stat -c \
  'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
  /absolute/path/to/fish-api-key
```

Check output filesystem:

```bash
clear

df -h /absolute/path/to/output-parent
df -i /absolute/path/to/output-parent
```

---

## 330. Summary

Troubleshoot from the stage boundary, not from guesses:

```text
capture exit code and stderr
    ↓
find request ID
    ↓
identify last lifecycle message
    ↓
inspect complete wrapped/joined error
    ↓
check artifacts created by that stage
    ↓
reduce config, modules, input, network, or filesystem
    ↓
reproduce under actual service identity
    ↓
run focused and full tests
```

The most important operational rules are:

- always capture stderr;
- use absolute config and output paths in services;
- never print the API key;
- treat missing-secret creation as a provisioning action;
- distinguish module ERROR records from final process failure;
- inspect the full nested cause beneath `synthesis failed`;
- remember transport errors are not internally retried;
- remember `429` can retry and `5xx` does not by default;
- create output parents before invocation;
- check both disk blocks and inodes;
- inspect output after code `4`;
- preserve multi-line joined errors;
- do not weaken filesystem security to make an error disappear;
- reproduce from a clean build with uncached and race tests;
- report the exact commit, status, request ID, and artifact state.
