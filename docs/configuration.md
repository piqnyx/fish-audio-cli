# Configuration reference

`fish-audio-cli` uses a JSON configuration file.

The default path is:

```text
config/config.json
```

Create a local configuration from the provided example:

```bash
cp config/config.example.json config/config.json
```

A different path can be selected with:

```text
--config /path/to/config.json
```

Unknown configuration fields are rejected. This helps detect spelling mistakes instead of silently ignoring them.

## Pipeline

### `pipeline.modules`

Ordered array of configured text-processing module instances.

Modules run from first to last. The output of one module becomes the input of
the next module.

The array may be empty. An empty pipeline returns the original input text
unchanged.

Each array item is an object with the following fields.

#### `name`

Required unique name of this particular module instance.

The name is used in logs and errors, so two configured instances must not have
the same name.

_Validation:_ Must be a non-empty string without leading or trailing
whitespace.

#### `type`

Required registered module implementation type.

Several module instances may use the same type. Their `name`, `config`, and
optional `onError` values may differ.

Whether a type is supported is checked by the module registry during
initialization.

_Validation:_ Must be a non-empty string without leading or trailing
whitespace.

#### `config`

Required JSON object containing settings owned by the selected module type.

Each module strictly decodes and validates its own configuration. Unknown
module-specific fields are rejected instead of being silently ignored.

Even a module with no configurable options must provide an empty object:

```json
"config": {}
```

Values such as `null`, arrays, strings, numbers, or a missing `config` field
are rejected.

#### `onError`

Optional failure-policy override for this particular module instance.

When omitted, the module inherits `pipeline.onError`.

Supported values are the same as for the global policy:

```text
use_previous
use_original
skip
abort
```

Example pipeline:

```json
{
  "pipeline": {
    "onError": "use_previous",
    "modules": [
      {
        "name": "first-pass",
        "type": "passthrough",
        "config": {}
      },
      {
        "name": "second-pass",
        "type": "passthrough",
        "onError": "abort",
        "config": {}
      }
    ]
  }
}
```

This example deliberately uses the same module type twice under different
instance names.

Empty pipeline example:

```json
{
  "pipeline": {
    "onError": "use_previous",
    "modules": []
  }
}
```

Currently registered module types:

```text
passthrough
```

### `pipeline.onError`

Default policy used when a module fails and does not provide its own
`onError` override.

Supported values:

```text
use_previous
use_original
skip
abort
```

#### `use_previous`

Restores the text produced before the failed module and continues with the
next module.

This is the default policy and is suitable for optional text-enhancement
modules.

#### `use_original`

Restores the original pipeline input and continues with the next module.

#### `skip`

Restores the text produced before the failed module, stops the remaining
pipeline, and continues synthesis without returning a module error.

#### `abort`

Restores the text produced before the failed module, stops processing, and
returns an error. Audio synthesis is not started.

Default:

```text
use_previous
```

Changes made by a module that returns an error are never kept.

Context cancellation and deadline expiration always stop the pipeline and
return an interruption error, regardless of the configured failure policy.
Changes made by the interrupted module are rolled back first.

## Fish Audio

### `fish.baseUrl`

Base URL of the Fish Audio API.

Default:

```text
https://api.fish.audio
```

The CLI appends `/v1/tts` when sending synthesis requests.

_Validation:_ Must be an absolute HTTP or HTTPS URL containing a host. User
information, query parameters, and fragments are rejected. A base path is
allowed and `/v1/tts` is appended to it.

### `fish.model`

Fish Audio model identifier.

Example:

```text
s2.1-pro-free
```

Model availability, pricing, and free access are controlled by Fish Audio and may change.

_Validation:_ Must be a non-empty string.

### `fish.referenceId`

Optional Fish Audio voice or reference identifier.

An empty value uses the model's default voice:

```json
"referenceId": ""
```

A specific voice may be selected with:

```json
"referenceId": "YOUR_REFERENCE_ID"
```

_Validation:_ May be an empty string or a reference identifier accepted by Fish Audio. No
additional local format validation is applied.

### `fish.timeoutSeconds`

Maximum duration of one Fish Audio HTTP request.

Default:

```text
120
```

_Validation:_ Must be an integer greater than zero.

## Fish synthesis request

The `fish.request` object controls synthesis parameters sent to Fish Audio.

### `temperature`

Sampling temperature.

Default:

```text
0.7
```

Higher values may produce more variation. Lower values are generally more deterministic.

Valid configuration range:

```text
0.0 to 1.0
```

### `topP`

Nucleus sampling probability.

Default:

```text
0.7
```

Valid configuration range:

```text
0.0 to 1.0
```

### `prosody.speed`

Speech speed multiplier.

Default:

```text
1.0
```

Accepted range:

```text
0.5 to 2.0
```

### `prosody.volume`

Output volume adjustment.

Default:

```text
0.0
```

Accepted range:

```text
-20.0 to 20.0
```

### `prosody.normalizeLoudness`

Enables Fish Audio loudness normalization.

Default:

```text
true
```

### `chunkLength`

Requested synthesis chunk length.

Default:

```text
300
```

Accepted range:

```text
100 to 300
```

### `normalize`

Enables Fish Audio text normalization.

Default:

```text
true
```

This is an API-side synthesis option and is separate from local pipeline modules.

### `sampleRate`

Optional output sample rate.

Default:

```json
"sampleRate": null
```

A null value lets Fish Audio choose the normal sample rate for the selected format.

Accepted non-null values depend on the output format:

```text
wav:        8000, 16000, 24000, 32000, 44100
mp3:        32000, 44100
opus:       48000
```

### `mp3Bitrate`

MP3 bitrate in kilobits per second.

Default:

```text
192
```

This setting is relevant only when the requested output format is `mp3`.

Accepted values:

```text
64
128
192
```

### `opusBitrate`

Opus bitrate in bits per second.

Default:

```text
64000
```

This setting is relevant only when the requested output format is `opus` or its `ogg` alias.

Accepted values:

```text
-1000
24000
32000
48000
64000
```

### `latency`

Fish Audio latency mode.

Default:

```text
normal
```

Accepted values:

```text
normal
balanced
low
```

### `maxNewTokens`

Maximum number of generated audio tokens.

Default:

```text
1024
```

The value must be greater than zero.

### `repetitionPenalty`

Penalty used to reduce unwanted repetition.

Default:

```text
1.2
```

_Validation:_ Numeric value. No local minimum or maximum is currently enforced; unsupported
values may be rejected by Fish Audio.

### `minChunkLength`

Minimum synthesis chunk length.

Default:

```text
50
```

Accepted range:

```text
0 to 100
```

### `conditionOnPreviousChunks`

Allows later chunks to use earlier chunks as context.

Default:

```text
true
```

This may improve continuity for longer speech.

### `earlyStopThreshold`

Threshold used by Fish Audio for early stopping.

Default:

```text
1.0
```

Accepted range:

```text
0.0 to 1.0
```

### `features`

Optional Fish Audio feature flags.

Default:

```json
"features": []
```

An empty list is omitted from the HTTP request.

## Secrets

The Fish Audio API key is read from a separate file and is never stored directly in the JSON configuration.

The key file is initialized on every run. If it is missing, it is created empty with mode `0600`. If it already exists, it must be a regular file and its mode is reset to `0600`.

The containing directory should use mode `0700`. The file should contain only the API key and an optional trailing newline.

### `secrets.fishApiKeyFile`

Path to the Fish Audio API key file.

Relative paths are resolved from the project directory determined from the
configuration file path. Absolute paths are cleaned and used without rebasing.

Default:

```text
secrets/fish-api-key
```

_Validation:_ Must be a non-empty path string.

## Logging

### `logging.level`

Minimum logging level.

Accepted values:

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

### `logging.format`

Log output format.

Accepted values:

```text
text
json
```

Default:

```text
text
```

### `logging.logText`

Controls whether input or processed text may be included in logs.

Default:

```text
false
```

Keep this disabled when processing private or sensitive text.

_Validation:_ Boolean value: `true` or `false`.

### `logging.file`

Path to the persistent log file.

Default:

```json
"file": ""
```

An empty value uses the standard project log file:

```text
logs/fish-audio-cli.log
```

Relative paths are resolved from the project directory determined from the
configuration file path. Absolute paths are cleaned and used without rebasing.

Log records are always written to standard error in addition to the persistent
log file. This allows the parent process to capture diagnostic output.

On Linux and other Unix-like systems, use `/dev/null` to disable persistent
file logging while keeping standard error output:

```json
"file": "/dev/null"
```

A logrotate template is provided at:

```text
deploy/logrotate/fish-audio-cli
```

Replace the placeholder path in the template with the absolute path to the
project log file before installing it into the system logrotate configuration.

_Validation:_ May be empty, relative, or absolute. An empty value uses the standard project
log path. On Unix-like systems, `/dev/null` disables only the persistent file
output; standard error remains enabled.

## Passthrough module

The built-in `passthrough` module returns its input text unchanged.

Configuration:

```json
{
  "name": "passthrough",
  "type": "passthrough",
  "config": {}
}
```

The instance name may be changed, and several instances may use the
`passthrough` type.

The module has no configurable fields. Its `config` value must still be an
object, and unknown fields are rejected.

It is useful for:

- verifying module configuration and registry initialization;
- testing pipeline ordering and failure-policy wiring;
- comparing original and processed output;
- serving as a minimal reference when implementing another module.

The module is intentionally retained even though it performs no
transformation. Each custom module must strictly decode and validate its own
configuration before implementing the shared processing contract.

## Local configuration

The following files are intended to remain local and are ignored by Git:

```text
config/config.json
secrets/
bin/
logs/
```

Do not commit API keys, private voice identifiers, or machine-specific configuration.
