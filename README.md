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

The currently registered module type is `passthrough`, which sends the input text to Fish Audio without modifying it.

## Features

- Text input through `--text` or standard input
- Configurable text-processing pipeline
- Fish Audio settings stored in JSON
- WAV, MP3, and Opus output
- `ogg` accepted as an alias for Opus
- API keys loaded from separate protected files
- Automatic creation and permission hardening of secret files
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

When a configured secret file does not exist, the CLI creates an empty file with restricted permissions and reports where the key must be written. If the file already exists, the CLI requires it to be a regular file and sets its mode to `0600`.

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
--help     Show command-line help
```

Supported output formats:

```text
wav
mp3
opus
ogg
```

`ogg` is normalized to `opus`.

## Exit codes

- `0`: synthesis completed successfully or command-line help was displayed.
- `1`: startup logging or request ID initialization failed.
- `2`: arguments, configuration, logging, secrets, modules, or input initialization failed.
- `3`: text processing, API key loading, request construction, or Fish client initialization failed.
- `4`: synthesis or atomic output persistence failed.

## Output files

The caller provides the final output path through `--output`.

Audio is first written to a unique temporary file beside the destination. After synthesis completes successfully, the temporary file is synchronized, atomically renamed to the requested path, and the containing directory is synchronized before the CLI reports success.

Temporary files are removed after failures. The completed output file belongs to the caller and is not deleted automatically.

Concurrent invocations are safe when each invocation receives its own output path.

## Processing pipeline

`pipeline.modules` is an ordered array of configured module instances. Modules
run from first to last, and the output of one module becomes the input of the
next module.

Each module instance contains:

- `name`: a unique instance name used in logs and errors;
- `type`: the registered module implementation;
- `config`: a required JSON object owned and validated by that module;
- `onError`: an optional failure-policy override for that instance.

Module types may be repeated under different names. This allows several
instances of the same implementation to use different configurations.

Example:

```json
{
  "pipeline": {
    "onError": "use_previous",
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

An empty module array is valid:

```json
{
  "pipeline": {
    "onError": "use_previous",
    "modules": []
  }
}
```

An empty pipeline returns the input text unchanged.

`pipeline.onError` provides the default failure policy. A module may override
it with its own `onError` value.

Supported policies:

- `use_previous` restores the text from before the failed module and continues;
- `use_original` restores the original input and continues;
- `skip` restores the previous text and stops the remaining modules;
- `abort` restores the previous text and returns an error.

Changes made by a failed or interrupted module are always rolled back.
Cancellation and deadline errors always stop processing and are not converted
into successful fallback results.

The currently registered module type is:

```text
passthrough
```

The `passthrough` module accepts an empty configuration object and leaves its
input unchanged.

Future modules may add normalization, pronunciation handling, emoji
processing, language-model transformations, or other text processing without
changing the synthesis and output layers.

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
- Extend machine-checked configuration consistency beyond defaults and the example file to validation rules and documentation.

## Security

- Keep API keys outside the repository.
- Secret files should use mode `0600`.
- Secret directories should use mode `0700`.
- Input text is sent to the configured external API services.
- Avoid enabling text logging when processing sensitive content.

## Passthrough module

The built-in `passthrough` module returns its input text unchanged.

A configured instance looks like this:

```json
{
  "name": "passthrough",
  "type": "passthrough",
  "config": {}
}
```

The instance name may be changed, and the `passthrough` type may be used more
than once. Its configuration object has no fields, so unknown fields are
rejected.

The module is intentionally kept as a minimal reference implementation and as
a way to verify configuration, registry, pipeline, logging, and synthesis
wiring without transforming text.

Each custom module owns strict decoding and validation of its configuration,
then implements the common processing contract before being registered.

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
