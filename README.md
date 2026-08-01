# fish-audio-cli

A standalone command-line client for generating speech with the Fish Audio API.

`fish-audio-cli` accepts text from a command-line argument or standard input, passes it through a configurable processing pipeline, sends the result to Fish Audio, and atomically writes the generated audio to a requested output path.

It is designed for scripts, automation systems, bots, and applications that need a small and predictable TTS executable without running a persistent service.

## Status

The project is currently in alpha development.

The core synthesis path is functional:

```text
text input
    ↓
configurable processor pipeline
    ↓
Fish Audio API
    ↓
atomic audio output
```

The initial processor is `passthrough`, which sends the input text to Fish Audio without modifying it.

## Features

- Text input through `--text` or standard input
- Configurable text-processing pipeline
- Fish Audio settings stored in JSON
- WAV, MP3, and Opus output
- `ogg` accepted as an alias for Opus
- API keys loaded from separate protected files
- Automatic creation of missing secret files
- Atomic output file replacement
- Unique request IDs in structured logs
- Safe concurrent execution as independent processes
- No shared temporary directory
- No daemon or persistent service required

## Requirements

- Go 1.26 or newer
- A Fish Audio API key

## Build

Clone the repository:

```bash
git clone https://github.com/piqnyx/fish-audio-cli.git
cd fish-audio-cli
```

Build the executable:

```bash
mkdir -p bin
go build -o bin/fish-audio-cli ./cmd/fish-audio-cli
```

Run the test suite:

```bash
go test ./...
go vet ./...
```

## Configuration

See the [configuration reference](docs/configuration.md) for a complete description of every available option.

Copy the example configuration:

```bash
cp config/config.example.json config/config.json
```

The default configuration path is:

```text
config/config.json
```

A different configuration file can be selected with:

```text
--config /path/to/config.json
```

The local `config/config.json` file is ignored by Git and may contain machine-specific paths and voice identifiers.

## API keys

By default, the Fish Audio API key is read from:

```text
secrets/fish-api-key
```

Write the key into that file:

```bash
mkdir -p secrets
chmod 700 secrets

printf '%s\n' 'YOUR_FISH_AUDIO_API_KEY' > secrets/fish-api-key
chmod 600 secrets/fish-api-key
```

When a configured secret file does not exist, the CLI creates an empty file with restricted permissions and reports where the key must be written.

API keys are not stored in the repository.

## Usage

Generate Opus audio from a command-line argument:

```bash
./bin/fish-audio-cli \
  --config config/config.json \
  --format opus \
  --output ./speech.opus \
  --text 'Hello from fish-audio-cli.'
```

Read text from standard input:

```bash
printf '%s' 'Text received through standard input.' |
  ./bin/fish-audio-cli \
    --config config/config.json \
    --format opus \
    --output ./speech.opus
```

## Command-line options

```text
--config   Path to the JSON configuration file
--text     Text to synthesize; standard input is used when omitted
--format   Output audio format
--output   Final output file path
```

Supported output formats:

```text
wav
mp3
opus
ogg
```

`ogg` is normalized to `opus`.

## Output files

The caller provides the final output path through `--output`.

Audio is first written to a unique temporary file beside the destination. After synthesis completes successfully, the temporary file is synchronized, atomically renamed to the requested path, and the containing directory is synchronized before the CLI reports success.

Temporary files are removed after failures. The completed output file belongs to the caller and is not deleted automatically.

Concurrent invocations are safe when each invocation receives its own output path.

## Processing pipeline

Processors are configured in `pipeline.modules` and executed in the listed order.

Example:

```json
{
  "pipeline": {
    "modules": [
      "passthrough"
    ],
    "onError": "use_previous"
  }
}
```

Processor failures are controlled by `pipeline.onError`:

- `use_previous` restores the last valid text and continues.
- `use_original` restores the original input and continues.
- `skip` stops the remaining processors without failing synthesis.
- `abort` stops the request with an error.

The default policy is `use_previous`.

The current processor list contains:

```text
passthrough
```

Future processors may add text normalization, pronunciation handling, emoji processing, or other text transformations without changing the synthesis and output layers.

## Fish Audio models

The Fish Audio model is selected in the configuration file:

```json
{
  "fish": {
    "model": "s2.1-pro-free"
  }
}
```

Model availability, pricing, rate limits, and free access are controlled by Fish Audio and may change.

Use a paid model for deployments that require predictable availability. Changing models does not require rebuilding the CLI.

A specific Fish Audio voice may be selected through `fish.referenceId`.

## Planned improvements

- Allow configuration values to be overridden by command-line arguments, using the precedence `CLI argument > configuration file > built-in default`.
- Add a machine-checked configuration schema or consistency test so new settings cannot be added without updating defaults, validation, the example configuration, and documentation.

## Security

- Keep API keys outside the repository.
- Secret files should use mode `0600`.
- Secret directories should use mode `0700`.
- Input text is sent to the configured external API services.
- Avoid enabling text logging when processing sensitive content.

## Passthrough module

The built-in `passthrough` module returns its input text unchanged.

It is intentionally kept as a minimal reference implementation for developing
custom text-processing modules. It can also be used to verify pipeline wiring
without modifying the synthesized text.

A custom module should follow the same processing and error-handling contract,
then be registered alongside the built-in modules.

## Logging

The CLI always writes structured logs to standard error and to a persistent
log file.

When `logging.file` is empty, the default file is:

```text
logs/fish-audio-cli.log
```

Relative log paths are resolved from the project directory determined from the
configuration file path. Absolute paths are cleaned and used without rebasing.

On Linux and other Unix-like systems, set `logging.file` to `/dev/null` to
disable persistent file logging while keeping standard error output.

A logrotate template is included at:

```text
deploy/logrotate/fish-audio-cli
```

Replace the placeholder path in the template with the absolute path to the
project log file before installing it. The template uses `nocreate` because
the CLI creates the next log file with the correct owner and permissions on
its next invocation.

## License

This project is licensed under the MIT License. See `LICENSE` for details.
