# fish-audio-cli

[![CI](https://github.com/piqnyx/fish-audio-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/piqnyx/fish-audio-cli/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A small, single-run command-line client for generating speech through the Fish Audio API.

`fish-audio-cli` accepts one text input, passes it through an ordered local processing pipeline, sends the resulting text to Fish Audio, and atomically publishes the returned audio at a caller-selected path.

```text
text argument or standard input
    ↓
strict JSON configuration
    ↓
ordered text-processing modules
    ↓
Fish Audio API
    ↓
atomic WAV, MP3, or Opus output
```

The executable is intended for scripts, bots, scheduled jobs, automation systems, and applications that need a predictable TTS command without running a persistent local service.

> [!WARNING]
> The project is in alpha development. The core synthesis path is functional and tested, but public interfaces can still change before the first stable release.

## Features

- Text input through `--text` or standard input
- Strict UTF-8 and bounded-input validation
- Strict JSON configuration with duplicate-key and unknown-field rejection
- Ordered, configurable text-processing pipeline
- Per-pipeline and per-module failure policies
- Fish Audio request configuration and bounded API errors
- Rate-limit retry with bounded backoff
- WAV, MP3, and Opus output
- `ogg` accepted as a CLI alias for Opus
- API key loaded from a separate protected file
- Structured text or JSON logs with request IDs
- Atomic same-directory output replacement
- Temporary-file cleanup on pre-publication failure
- Race-tested Go implementation
- No daemon or local HTTP service required

The currently registered built-in module type is:

```text
passthrough
```

It validates the module lifecycle without changing the input text.

## Requirements

- Go `1.26.5`
- A Fish Audio API key
- A Fish model and optional reference voice available to the account
- A writable log location
- An existing writable parent directory for the output file

Fish Audio controls model availability, account permissions, pricing, quotas, and rate limits. The configured default model name is a project default, not a guarantee of current provider access.

## Build

Clone the repository and build the command:

```bash
clear

git clone https://github.com/piqnyx/fish-audio-cli.git
cd fish-audio-cli

mkdir -p bin

go build \
  -trimpath \
  -o bin/fish-audio-cli \
  ./cmd/fish-audio-cli
```

Show command help:

```bash
clear

./bin/fish-audio-cli --help
```

## Quick start

### 1. Create the local configuration

Copy the tracked example:

```bash
clear

cp \
  config/config.example.json \
  config/config.json
```

The default command-line configuration path is:

```text
config/config.json
```

The local `config/config.json` file is ignored by Git.

Review at least:

- `fish.model`
- `fish.referenceId`
- `secrets.fishApiKeyFile`
- `logging.file`

The complete field reference is in [`docs/configuration.md`](docs/configuration.md).

### 2. Store the Fish API key

The default secret path is:

```text
secrets/fish-api-key
```

Create it without placing the key directly in shell history:

```bash
clear

install -d -m 0700 secrets

read -r -s -p 'Fish API key: ' fish_api_key
printf '\n'

printf '%s\n' "$fish_api_key" \
  > secrets/fish-api-key

unset fish_api_key

chmod 0600 secrets/fish-api-key
```

The file must contain exactly one nonblank UTF-8 key line.

One final LF or CRLF is accepted.

When the configured secret file is missing, the command creates an empty `0600` file, logs where it was created, and exits with status `3`. Populate that file and run the command again.

Secret paths, directory checks, permissions, symlink handling, and container-mount limitations are documented in [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md).

### 3. Generate audio from `--text`

```bash
clear

./bin/fish-audio-cli \
  --config config/config.json \
  --format opus \
  --output speech.opus \
  --text 'Hello from fish-audio-cli.'
```

A successful invocation returns status `0` and publishes:

```text
speech.opus
```

### 4. Generate audio from standard input

```bash
clear

printf '%s' 'Text received through standard input.' |
  ./bin/fish-audio-cli \
    --config config/config.json \
    --format opus \
    --output speech.opus
```

Standard input is used when `--text` is omitted or its value is exactly empty.

The command reads stdin until EOF.

## Command-line interface

```text
Usage:
    fish-audio-cli [options]

Options:
    --config PATH   JSON configuration file
                    default: config/config.json

    --text TEXT     text to synthesize
                    standard input is used when omitted

    --format FORMAT output format: wav, mp3, opus or ogg

    --output PATH   destination audio file

    --help          show this help
```

### Options

| Option | Required | Meaning |
|---|---:|---|
| `--config PATH` | no | JSON configuration file; default `config/config.json` |
| `--text TEXT` | no | Text to synthesize; stdin is used when omitted or exactly empty |
| `--format FORMAT` | yes | `wav`, `mp3`, `opus`, or `ogg` |
| `--output PATH` | yes | Final destination audio file |
| `--help` | no | Print help to stdout and exit `0` |

Positional arguments are rejected.

The complete CLI contract is in [`docs/cli.md`](docs/cli.md).

## Output formats

Accepted CLI values:

| CLI value | Fish request format | Typical extension |
|---|---|---|
| `wav` | `wav` | `.wav` |
| `mp3` | `mp3` | `.mp3` |
| `opus` | `opus` | `.opus` |
| `ogg` | `opus` | `.ogg` |

The filename extension does not select or validate the format.

For example:

```text
--format mp3 --output speech.wav
```

still writes MP3 bytes.

## Configuration

`fish-audio-cli` loads one JSON configuration file for each invocation.

The tracked example is:

[`config/config.example.json`](config/config.example.json)

The configuration controls:

- maximum input size;
- pipeline modules and failure policy;
- Fish base URL;
- Fish model;
- optional reference voice;
- HTTP timeout;
- retry behavior;
- synthesis request parameters;
- secret file path and maximum size;
- log level, format, text visibility, and file path.

Configuration decoding is strict:

- malformed JSON is rejected;
- invalid UTF-8 is rejected;
- duplicate keys are rejected;
- unknown fields are rejected;
- prohibited explicit `null` values are rejected;
- semantic ranges and cross-field relations are validated.

Omitted fields receive built-in defaults.

Do not place the API key inside the JSON file.

See [`docs/configuration.md`](docs/configuration.md) for every field, default, allowed range, and null rule.

## Text-processing pipeline

`pipeline.modules` is an ordered array of configured module instances.

Each instance has:

```json
{
  "name": "passthrough",
  "type": "passthrough",
  "config": {}
}
```

The instance name must be unique.

A module type can appear more than once under different names.

An empty module array is valid and leaves the input unchanged:

```json
{
  "pipeline": {
    "modules": [],
    "onError": "use_previous"
  }
}
```

Supported failure policies are:

| Policy | Behavior after a module failure |
|---|---|
| `use_previous` | Restore text from before the failing module and continue |
| `use_original` | Restore the original pipeline input and continue |
| `skip` | Restore pre-step text and stop the remaining modules successfully |
| `abort` | Restore pre-step text and return an error |

A failed or interrupted module never commits its partial text mutation.

Cancellation is not converted into a successful fallback result.

See:

- [`docs/pipeline.md`](docs/pipeline.md)
- [`docs/modules.md`](docs/modules.md)
- [`docs/module-author-guide.md`](docs/module-author-guide.md)

## Fish Audio behavior

The configured Fish base URL is validated and joined with:

```text
/v1/tts
```

The client sends:

```text
Authorization: Bearer <API key>
Content-Type: application/json
model: <configured model>
```

The API key and model are rejected before the request when they contain surrounding whitespace, invalid UTF-8, or ASCII control characters.

The current client uses JSON and single-speaker requests. `fish.referenceId` maps to either one string or an omitted `reference_id`. It does not currently encode MessagePack, inline `references`, array-valued `reference_id`, or multi-speaker dialogue requests, even when the live provider API supports those forms.

### Retry behavior

The client can retry:

- HTTP `429`;
- HTTP `5xx` when `fish.retry.retryServerErrors` is enabled.

The client does not internally retry:

- DNS failures;
- connection refusal;
- TLS failures;
- transport timeouts;
- response streaming failures after a successful HTTP status.

A valid `Retry-After` value above the configured maximum delay stops retry instead of being silently clamped.

There is no provider idempotency key. Retrying a failed invocation can create another remote synthesis request.

See [`docs/fish-audio.md`](docs/fish-audio.md) for request fields, validation, retries, typed API errors, cancellation, and response streaming.

## Secrets and paths

Relative paths do not all use the same base.

| Path | Relative base |
|---|---|
| `--config` | Process working directory |
| `secrets.fishApiKeyFile` | Project directory derived from the config path |
| `logging.file` | Project directory derived from the config path |
| Module-owned path | Defined by that module, normally through the project resolver |
| `--output` | Process working directory |

Use absolute config and output paths in services, containers, cron jobs, and other managed environments.

The project resolver is lexical:

- it does not expand `~`;
- it does not expand environment variables;
- it does not resolve symlink targets;
- it does not confine `..` beneath the project directory.

See [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md).

## Output files

The caller must supply `--output`.

The output parent directory must already exist.

The command does not stream audio to stdout.

Audio publication follows:

```text
create unique temp beside destination
    ↓
stream Fish response into temp
    ↓
sync temp
    ↓
close temp
    ↓
rename temp to destination
    ↓
sync and close containing directory
```

The temporary filename pattern is:

```text
.<destination-base>.*.tmp
```

The final file is created with mode:

```text
0600
```

Before a successful rename, a failure preserves an existing destination and attempts to remove the temporary file.

After a successful rename, a directory-sync failure returns exit `4` but keeps the newly published output. Therefore, a nonzero synthesis/output status does not always mean the destination is absent.

Concurrent invocations should use unique output paths. There is no destination lock.

See [`docs/output-files.md`](docs/output-files.md).

## Logging

The command has two logging phases.

### Bootstrap logging

Before the configuration is validated, diagnostics use a text logger on stderr.

Early failures do not reach the persistent log file.

### Configured logging

After configuration validation, every structured record is sent to:

```text
stderr
+
persistent log file
```

When `logging.file` is empty, the default file is:

```text
logs/fish-audio-cli.log
```

relative to the derived project directory.

The command creates missing log directories, opens the file in append mode, and forces final file mode:

```text
0640
```

Persistent file logging cannot currently be disabled.

Do not use `/dev/null` as a disabling mechanism.

Supported formats:

```text
text
json
```

Supported thresholds:

```text
debug
info
warn
error
```

Input and processed text are omitted by default.

Set `logging.logText` to `true` only when recording text is operationally acceptable.

A logrotate template is included at:

[`deploy/logrotate/fish-audio-cli`](deploy/logrotate/fish-audio-cli)

See [`docs/logging.md`](docs/logging.md).

## Exit codes

| Code | Meaning |
|---:|---|
| `0` | Help displayed or invocation completed successfully |
| `1` | Bootstrap logger or request-ID generation failed |
| `2` | Arguments, paths, config, configured logging, modules, app setup, or input failed |
| `3` | Pipeline, Fish request creation, secret loading, or Fish client setup failed |
| `4` | Fish synthesis, response streaming, or atomic output failed |

Important consequences:

- A recoverable module error can be logged at ERROR while the command later exits `0`.
- A missing secret file is logged at WARN while the command exits `3`.
- Exit `4` can occur after the new output has already been renamed into place.
- Shell signal statuses can fall outside `0` through `4`.

See:

- [`docs/errors-and-exit-codes.md`](docs/errors-and-exit-codes.md)
- [`docs/troubleshooting.md`](docs/troubleshooting.md)

## Development

Run the required repository checks:

```bash
clear

unformatted="$(gofmt -l .)"

if [ -n "$unformatted" ]; then
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
```

The GitHub Actions workflow runs the same gates on:

- pushes to `main`;
- pull requests targeting `main`;
- manual dispatch.

Normal tests do not contact the live Fish API and do not require a real key.

See [`docs/testing.md`](docs/testing.md).

## Documentation

The complete documentation map is:

[`docs/index.md`](docs/index.md)

Common entry points:

| Topic | Document |
|---|---|
| Command syntax | [`docs/cli.md`](docs/cli.md) |
| Complete configuration reference | [`docs/configuration.md`](docs/configuration.md) |
| Runtime architecture | [`docs/architecture.md`](docs/architecture.md) |
| Pipeline execution and fallback | [`docs/pipeline.md`](docs/pipeline.md) |
| Module system | [`docs/modules.md`](docs/modules.md) |
| Adding a module | [`docs/module-author-guide.md`](docs/module-author-guide.md) |
| Fish request and retry behavior | [`docs/fish-audio.md`](docs/fish-audio.md) |
| Paths and secret security | [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md) |
| Atomic output publication | [`docs/output-files.md`](docs/output-files.md) |
| Logs and request correlation | [`docs/logging.md`](docs/logging.md) |
| Errors and exit codes | [`docs/errors-and-exit-codes.md`](docs/errors-and-exit-codes.md) |
| Tests and CI | [`docs/testing.md`](docs/testing.md) |
| Failure diagnosis | [`docs/troubleshooting.md`](docs/troubleshooting.md) |
| Contribution workflow | [`CONTRIBUTING.md`](CONTRIBUTING.md) |
| Security reporting and trust boundaries | [`SECURITY.md`](SECURITY.md) |
| Unreleased changes and release history | [`CHANGELOG.md`](CHANGELOG.md) |

The specialized document for a subsystem is authoritative for its exact behavior.

Repository governance and release material live at the repository root rather than under `docs/`.

## Security notes

Report suspected vulnerabilities privately as described in [`SECURITY.md`](SECURITY.md). Do not publish credentials or exploit details in a public issue.

- Keep API keys outside the repository.
- Rotate a key immediately if it appears in Git history, logs, CI output, or a bug report.
- Use a trusted secret directory that is not writable by group or others.
- Do not replace secret permission errors with `chmod -R 777`.
- Treat the configuration as trusted input because it controls paths, endpoints, modules, and logging.
- Input text is sent to the configured external Fish endpoint.
- Provider error bodies can appear in logs within a configured byte limit.
- Keep `logging.logText` disabled for sensitive content.
- Use a trusted output directory.
- Review stderr and log files before sharing diagnostic data.
- The application does not encrypt output audio at rest.

Detailed filesystem and logging boundaries are documented in:

- [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md)
- [`docs/logging.md`](docs/logging.md)
- [`docs/output-files.md`](docs/output-files.md)

## Project scope

The current project is a one-shot CLI.

It does not currently provide:

- a daemon;
- a local HTTP server;
- dynamic runtime plugins;
- stdout audio streaming;
- inline API keys;
- environment-variable interpolation in JSON;
- automatic output naming;
- automatic output-parent creation;
- destination locking;
- live configuration reload;
- internal log rotation;
- provider model discovery;
- provider idempotency.

New capabilities require explicit behavior, security, failure, logging, and testing contracts.

## License

`fish-audio-cli` is licensed under the MIT License.

See [`LICENSE`](LICENSE).
