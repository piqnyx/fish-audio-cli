# Configuration reference

> **Document status:** normative reference for the current pre-release configuration format.
>
> **Audience:** users operating `fish-audio-cli`, maintainers reviewing configuration changes, and module authors documenting new instance-owned settings.
>
> **Scope:** this document describes the JSON file, defaults, validation, path resolution, pipeline instances, Fish Audio settings, retry behavior, secrets, and logging. Runtime architecture is documented in [`architecture.md`](architecture.md); pipeline behavior in [`pipeline.md`](pipeline.md); the module model in [`modules.md`](modules.md).

---

## 1. Configuration file

`fish-audio-cli` reads one JSON configuration file per invocation.

The default path is:

```text
config/config.json
```

A different path can be selected with:

```text
--config /path/to/config.json
```

The repository provides a complete example at:

```text
config/config.example.json
```

Create a local configuration with:

```bash
cp config/config.example.json config/config.json
```

`config/config.json` is ignored by Git and is intended to remain machine-local.

### 1.1 File-size limit

The configuration file may contain at most:

```text
1048576 bytes
1 MiB
```

The limit applies to the complete file, including whitespace.

### 1.2 Load sequence

Configuration loading performs these steps:

```text
resolve absolute config path
    ↓
read at most 1 MiB
    ↓
start from built-in defaults
    ↓
strictly decode supplied JSON over defaults
    ↓
reject unsupported explicit null values
    ↓
resolve the Fish API key path
    ↓
validate semantic values
    ↓
prepare and build configured modules
```

A failure stops startup before text input is read.

---

## 2. Strict JSON rules

The configuration parser rejects:

- invalid UTF-8;
- malformed JSON;
- an empty file;
- more than one top-level JSON value;
- duplicate object keys at any depth;
- unknown core configuration fields;
- incorrect capitalization of core field names;
- non-object pipeline module entries;
- unsupported explicit `null` values.

Field names are exact.

This is valid:

```json
{
  "fish": {
    "model": "s2.1-pro-free"
  }
}
```

This is rejected:

```json
{
  "Fish": {
    "Model": "s2.1-pro-free"
  }
}
```

Duplicate keys are rejected even when one spelling uses a Unicode escape:

```json
{
  "fish": {
    "model": "first",
    "mo\u0064el": "second"
  }
}
```

### 2.1 Unknown fields

Unknown fields are errors rather than ignored hints.

This is rejected:

```json
{
  "fish": {
    "timeotSeconds": 120
  }
}
```

The same strictness applies to a module instance envelope.

Module-owned fields inside `pipeline.modules[].config` are decoded later by the selected module.

### 2.2 Explicit `null`

The only core configuration path that accepts explicit `null` is:

```text
fish.request.sampleRate
```

Example:

```json
{
  "fish": {
    "request": {
      "sampleRate": null
    }
  }
}
```

For core fields, omission means “keep the default”; `null` does not mean omission.

These are rejected:

```json
{
  "pipeline": {
    "modules": null
  }
}
```

```json
{
  "logging": {
    "logText": null
  }
}
```

A module instance’s `config` object itself cannot be `null`.

Values nested inside that object are owned by the module and may accept or reject `null` according to that module’s schema.

---

## 3. Defaults and partial configuration

The loader starts with a complete built-in configuration and applies supplied fields over it.

A minimal file is valid:

```json
{}
```

It uses every built-in default.

A partial nested object is also valid:

```json
{
  "fish": {
    "model": "s2.1-pro"
  }
}
```

Only `fish.model` changes. Other Fish settings remain at their defaults.

### 3.1 Objects are overlaid

Omitted fields inside ordinary objects keep their current default values.

Example:

```json
{
  "fish": {
    "retry": {
      "maxAttempts": 5
    }
  }
}
```

The retry delays and `retryServerErrors` keep their defaults.

### 3.2 Arrays are replaced

When an array field is present, its complete value replaces the default array.

Example:

```json
{
  "pipeline": {
    "modules": []
  }
}
```

This removes the default `passthrough` instance and creates an empty pipeline.

### 3.3 Module entries do not inherit defaults

Each element of `pipeline.modules` is decoded into a fresh module instance.

This is invalid:

```json
{
  "pipeline": {
    "modules": [
      {}
    ]
  }
}
```

The element does not inherit the default module’s `name`, `type`, or `config`.

Every supplied module instance must be self-contained.

---

## 4. Complete default configuration

The built-in defaults are equivalent to:

```json
{
  "input": {
    "maxBytes": 1048576
  },
  "pipeline": {
    "modules": [
      {
        "name": "passthrough",
        "type": "passthrough",
        "config": {}
      }
    ],
    "onError": "use_previous"
  },
  "fish": {
    "baseUrl": "https://api.fish.audio",
    "model": "s2.1-pro-free",
    "referenceId": "",
    "timeoutSeconds": 120,
    "maxErrorBodyBytes": 65536,
    "retry": {
      "maxAttempts": 3,
      "initialDelayMilliseconds": 500,
      "maxDelayMilliseconds": 5000,
      "retryServerErrors": false
    },
    "request": {
      "temperature": 0.7,
      "topP": 0.7,
      "prosody": {
        "speed": 1.0,
        "volume": 0.0,
        "normalizeLoudness": true
      },
      "chunkLength": 300,
      "normalize": true,
      "sampleRate": null,
      "mp3Bitrate": 192,
      "opusBitrate": 64000,
      "latency": "normal",
      "maxNewTokens": 1024,
      "repetitionPenalty": 1.2,
      "minChunkLength": 50,
      "conditionOnPreviousChunks": true,
      "earlyStopThreshold": 1.0,
      "features": []
    }
  },
  "secrets": {
    "fishApiKeyFile": "secrets/fish-api-key",
    "maxBytes": 16384
  },
  "logging": {
    "level": "info",
    "format": "text",
    "logText": false,
    "file": ""
  }
}
```

---

## 5. Field summary

| JSON path | Type | Default | Accepted value |
|---|---|---:|---|
| `input.maxBytes` | integer | `1048576` | `1` through `16777216` bytes |
| `pipeline.modules` | array | one `passthrough` instance | empty or ordered module-instance array |
| `pipeline.onError` | string | `use_previous` | `use_previous`, `use_original`, `skip`, `abort` |
| `fish.baseUrl` | string | `https://api.fish.audio` | absolute HTTP or HTTPS base URL |
| `fish.model` | string | `s2.1-pro-free` | nonblank header-safe model identifier |
| `fish.referenceId` | string | `""` | empty or valid UTF-8 reference identifier |
| `fish.timeoutSeconds` | integer | `120` | `1` through `900` seconds |
| `fish.maxErrorBodyBytes` | integer | `65536` | `1` through `1048576` bytes |
| `fish.retry.maxAttempts` | integer | `3` | `1` through `10` total attempts |
| `fish.retry.initialDelayMilliseconds` | integer | `500` | `1` through `300000` milliseconds |
| `fish.retry.maxDelayMilliseconds` | integer | `5000` | initial delay through `300000` milliseconds |
| `fish.retry.retryServerErrors` | boolean | `false` | `true` or `false` |
| `fish.request.temperature` | number | `0.7` | finite `0.0` through `1.0` |
| `fish.request.topP` | number | `0.7` | finite `0.0` through `1.0` |
| `fish.request.prosody.speed` | number | `1.0` | finite `0.5` through `2.0` |
| `fish.request.prosody.volume` | number | `0.0` | finite `-20.0` through `20.0` |
| `fish.request.prosody.normalizeLoudness` | boolean | `true` | `true` or `false` |
| `fish.request.chunkLength` | integer | `300` | `100` through `300` |
| `fish.request.normalize` | boolean | `true` | `true` or `false` |
| `fish.request.sampleRate` | integer or null | `null` | format-compatible supported rate |
| `fish.request.mp3Bitrate` | integer | `192` | `64`, `128`, `192` kbps |
| `fish.request.opusBitrate` | integer | `64000` | `-1000`, `24000`, `32000`, `48000`, `64000` bps |
| `fish.request.latency` | string | `normal` | `normal`, `balanced`, `low` |
| `fish.request.maxNewTokens` | integer | `1024` | greater than zero |
| `fish.request.repetitionPenalty` | number | `1.2` | any finite number |
| `fish.request.minChunkLength` | integer | `50` | `0` through `100` |
| `fish.request.conditionOnPreviousChunks` | boolean | `true` | `true` or `false` |
| `fish.request.earlyStopThreshold` | number | `1.0` | finite `0.0` through `1.0` |
| `fish.request.features` | array of strings | `[]` | valid UTF-8 strings |
| `secrets.fishApiKeyFile` | string | `secrets/fish-api-key` | nonblank relative or absolute path |
| `secrets.maxBytes` | integer | `16384` | `1` through `65536` bytes |
| `logging.level` | string | `info` | `debug`, `info`, `warn`, `error` |
| `logging.format` | string | `text` | `text`, `json` |
| `logging.logText` | boolean | `false` | `true` or `false` |
| `logging.file` | string | `""` | empty, relative, or absolute path |

All range endpoints are inclusive.

---

## 6. Input

### `input.maxBytes`

Maximum byte count accepted from:

- `--text`;
- standard input when `--text` is omitted or empty.

Default:

```text
1048576 bytes
1 MiB
```

Allowed range:

```text
1 through 16777216 bytes
1 byte through 16 MiB
```

The limit counts encoded UTF-8 bytes, not Unicode characters.

Text must also:

- be valid UTF-8;
- contain at least one non-whitespace Unicode code point.

The limit applies before local pipeline processing.

It is not a post-module output limit.

---

## 7. Pipeline

### `pipeline.modules`

Ordered array of text-processing module instances.

Modules execute from first to last.

The retained output of one successful or recovered step becomes the input to the next step.

The array may be empty:

```json
{
  "pipeline": {
    "modules": []
  }
}
```

An empty pipeline returns valid input text unchanged.

The full runtime contract is documented in [`pipeline.md`](pipeline.md).

### 7.1 Module instance envelope

Each element has this shape:

```json
{
  "name": "unique-instance-name",
  "type": "registered-type",
  "onError": "optional-policy",
  "config": {}
}
```

#### `name`

Required for every supplied instance.

Rules:

- nonblank;
- no leading or trailing whitespace;
- unique across the pipeline.

The name appears in logs, initialization errors, runtime errors, and step reports.

#### `type`

Required for every supplied instance.

Rules:

- nonblank;
- no leading or trailing whitespace;
- exact registered type name.

Type matching is case-sensitive and no aliases are inferred.

Currently registered type:

```text
passthrough
```

#### `onError`

Optional per-instance override.

Omit the field to inherit `pipeline.onError`.

Do not use:

```json
"onError": null
```

Supported values:

```text
use_previous
use_original
skip
abort
```

#### `config`

Required JSON object owned by the selected module.

A module with no options still uses:

```json
"config": {}
```

These are invalid:

```json
"config": null
```

```json
"config": []
```

```json
"config": "value"
```

```json
"config": 42
```

The core validates the object boundary.

The selected module strictly decodes its own fields.

See [`modules.md`](modules.md) for configuration ownership.

### 7.2 Repeated module types

The same type may be configured more than once:

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

Every instance receives its own config, builder, processor, and effective error policy.

### `pipeline.onError`

Default policy for instances without their own override.

Default:

```text
use_previous
```

#### `use_previous`

Restore text from immediately before the failed step and continue.

#### `use_original`

Restore the original pipeline input and continue.

This discards successful transformations from earlier steps.

#### `skip`

Restore text from immediately before the failed step, stop remaining steps, and continue to Fish synthesis without a pipeline error.

#### `abort`

Restore text from immediately before the failed step, stop remaining steps, and return an error.

Fish synthesis is not started.

Cancellation and deadline expiration always interrupt the pipeline regardless of the configured policy.

---

## 8. Fish Audio connection

### `fish.baseUrl`

Base URL used to resolve the synthesis endpoint.

Default:

```text
https://api.fish.audio
```

The client appends:

```text
/v1/tts
```

A base path is preserved.

Example:

```text
https://example.invalid/proxy
```

resolves to:

```text
https://example.invalid/proxy/v1/tts
```

Validation requires:

- absolute URL;
- `http` or `https` scheme;
- nonempty host;
- no user information;
- no query;
- no fragment.

The configured value is trimmed for endpoint resolution.

Use HTTPS for real credentials and production traffic.

### `fish.model`

Fish Audio model identifier sent as a request header.

Default:

```text
s2.1-pro-free
```

Rules:

- nonblank;
- no leading or trailing whitespace;
- valid UTF-8;
- no ASCII control characters.

Availability, pricing, and provider-side support are controlled by Fish Audio.

The CLI validates the local string contract, not current provider inventory.

### `fish.referenceId`

Optional voice or reference identifier placed in the JSON synthesis request.

Default:

```json
"referenceId": ""
```

An empty string omits `reference_id` from the encoded request.

A configured value must be valid UTF-8.

No additional local length, character-set, or provider-existence validation is currently applied.

The current configuration exposes only the single-speaker string form. It cannot represent an array-valued `reference_id`, inline `references`, or a MessagePack request body. Those live provider capabilities are outside the current client contract rather than accepted and ignored.

### `fish.timeoutSeconds`

HTTP client timeout for a Fish synthesis request.

Default:

```text
120 seconds
```

Allowed range:

```text
1 through 900 seconds
```

The timeout includes connection, request, response headers, and response-body streaming governed by the Go HTTP client.

### `fish.maxErrorBodyBytes`

Maximum non-success Fish response body captured for error reporting.

Default:

```text
65536 bytes
64 KiB
```

Allowed range:

```text
1 through 1048576 bytes
1 byte through 1 MiB
```

This limit applies to API error bodies.

It is not a limit on successful audio output.

---

## 9. Fish retry settings

### `fish.retry.maxAttempts`

Maximum total request attempts, including the initial request.

Default:

```text
3
```

Allowed range:

```text
1 through 10
```

A value of `1` disables retries while still allowing one request attempt.

### `fish.retry.initialDelayMilliseconds`

Initial exponential-backoff delay when no usable `Retry-After` value is supplied.

Default:

```text
500 milliseconds
```

Allowed range:

```text
1 through 300000 milliseconds
```

### `fish.retry.maxDelayMilliseconds`

Maximum accepted or computed retry delay.

Default:

```text
5000 milliseconds
```

Allowed range:

```text
initialDelayMilliseconds through 300000 milliseconds
```

It must be greater than or equal to `initialDelayMilliseconds`.

### `fish.retry.retryServerErrors`

Controls retries for Fish server errors.

Default:

```text
false
```

Behavior:

| Condition | Retried when false | Retried when true |
|---|---:|---:|
| HTTP `429` | yes | yes |
| HTTP `5xx` | no | yes |
| authentication or authorization errors | no | no |
| other non-retryable HTTP errors | no | no |
| transport error | no | no |

Retry behavior is deliberately conservative.

The client does not retry ambiguous transport failures.

### 9.1 Delay selection

For a retryable API response:

1. use a valid `Retry-After` header when present;
2. otherwise use exponential backoff;
3. cap exponential backoff at `maxDelayMilliseconds`.

Backoff sequence with the defaults:

```text
500 ms
1000 ms
2000 ms
4000 ms
5000 ms
5000 ms
...
```

No jitter is added.

`Retry-After` may be:

- decimal seconds;
- an HTTP date.

If the parsed delay exceeds `maxDelayMilliseconds`, the client does not perform that retry.

Cancellation or deadline expiration interrupts the wait.

### 9.2 No retry after successful streaming begins

Retries occur only for classified API error responses before successful audio streaming.

A read or write failure during a successful audio response is returned directly.

The atomic output layer prevents a partial temporary file from replacing the destination.

---

## 10. Fish synthesis request

The `fish.request` object supplies configurable JSON parameters for `POST /v1/tts`.

The application adds runtime values separately:

| Runtime value | Source |
|---|---|
| `text` | retained pipeline output |
| `format` | `--format` |
| `reference_id` | `fish.referenceId` |
| model header | `fish.model` |
| authorization header | secret file |

The CLI accepts:

```text
wav
mp3
opus
ogg
```

`ogg` is normalized to `opus`.

Internal support for `pcm` is not exposed by the current CLI.

### `fish.request.temperature`

Sampling temperature.

Default:

```text
0.7
```

Allowed range:

```text
0.0 through 1.0
```

The value must be finite.

### `fish.request.topP`

Nucleus-sampling probability.

Default:

```text
0.7
```

Allowed range:

```text
0.0 through 1.0
```

The value must be finite.

### `fish.request.prosody.speed`

Speech-speed multiplier.

Default:

```text
1.0
```

Allowed range:

```text
0.5 through 2.0
```

The value must be finite.

### `fish.request.prosody.volume`

Volume adjustment.

Default:

```text
0.0
```

Allowed range:

```text
-20.0 through 20.0
```

The value must be finite.

### `fish.request.prosody.normalizeLoudness`

Fish Audio loudness-normalization flag.

Default:

```text
true
```

This is separate from local text processing and from `fish.request.normalize`.

The [Fish Audio OpenAPI schema](https://api.fish.audio/openapi.json) reviewed on 2026-08-03 marks `normalize_loudness` as S2-Pro-only. The CLI does not compare this flag with `fish.model`; it sends the configured value and leaves model compatibility to the provider.

### `fish.request.chunkLength`

Requested synthesis chunk length.

Default:

```text
300
```

Allowed range:

```text
100 through 300
```

### `fish.request.normalize`

Fish Audio text-normalization flag.

Default:

```text
true
```

This is provider-side request behavior.

It is independent from local pipeline modules.

### `fish.request.sampleRate`

Optional output sample rate.

Default:

```json
"sampleRate": null
```

`null` delegates the normal sample rate to Fish Audio.

The value is validated twice:

1. against the global supported-rate set during config validation;
2. against the selected output format when building the request.

Globally accepted non-null values:

```text
8000
16000
24000
32000
44100
48000
```

Format compatibility:

| CLI format | Accepted sample rates |
|---|---|
| `wav` | `8000`, `16000`, `24000`, `32000`, `44100` |
| `mp3` | `32000`, `44100` |
| `opus` or `ogg` | `48000` |

A value may pass configuration validation and later fail when incompatible with the invocation’s `--format`.

### `fish.request.mp3Bitrate`

MP3 bitrate in kilobits per second.

Default:

```text
192
```

Accepted values:

```text
64
128
192
```

The field is relevant to `mp3` output.

It is still validated for every configuration, even when the current invocation uses another format.

### `fish.request.opusBitrate`

Opus bitrate in bits per second.

Default:

```text
64000
```

Accepted values:

```text
-1000
24000
32000
48000
64000
```

The field is relevant to `opus` and the `ogg` alias.

It is still validated for every configuration, even when the current invocation uses another format.

The special `-1000` value is passed to Fish Audio unchanged; provider semantics are not reinterpreted locally.

### `fish.request.latency`

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

Values are exact and lowercase.

### `fish.request.maxNewTokens`

Maximum generated audio-token count supplied to Fish Audio.

Default:

```text
1024
```

Validation:

```text
greater than zero
```

No local upper bound is currently enforced.

Provider-side limits may still apply.

### `fish.request.repetitionPenalty`

Repetition penalty supplied to Fish Audio.

Default:

```text
1.2
```

The value must be finite.

No local minimum or maximum is currently enforced.

Provider-side validation may reject values accepted locally.

### `fish.request.minChunkLength`

Minimum synthesis chunk length.

Default:

```text
50
```

Allowed range:

```text
0 through 100
```

### `fish.request.conditionOnPreviousChunks`

Allows later chunks to use previous chunks as context.

Default:

```text
true
```

### `fish.request.earlyStopThreshold`

Fish Audio early-stop threshold.

Default:

```text
1.0
```

Allowed range:

```text
0.0 through 1.0
```

The value must be finite.

### `fish.request.features`

Optional Fish Audio feature strings.

Default:

```json
"features": []
```

An empty list is omitted from the encoded request.

Every element must be valid UTF-8.

The core currently does not enforce:

- a feature-name allowlist;
- uniqueness;
- nonempty strings;
- a maximum list length;
- a maximum element length.

Unsupported feature values may be rejected by Fish Audio.

---

## 11. Secrets

The Fish API key is stored outside the JSON configuration.

### `secrets.fishApiKeyFile`

Path to the Fish API key file.

Default:

```text
secrets/fish-api-key
```

The configured path is trimmed, resolved to an absolute path, and stored in the loaded configuration.

Relative-path behavior is documented in [Path resolution](#13-path-resolution).

### `secrets.maxBytes`

Maximum Fish API key file size.

Default:

```text
16384 bytes
16 KiB
```

Allowed range:

```text
1 through 65536 bytes
1 byte through 64 KiB
```

### 11.1 Missing key file

When the key file is missing:

- the containing directory is created when possible;
- the file is created empty with mode `0600`;
- startup reports that the file was created;
- synthesis does not continue until the file is populated.

### 11.2 Directory requirements

The secret directory must:

- be a directory;
- not be writable by group or others.

Missing directories are requested with mode:

```text
0700
```

Existing directory permissions are inspected rather than silently trusted.

### 11.3 File requirements

An existing secret path must:

- be a regular file;
- remain the same file while being securely opened;
- not resolve through an accepted non-regular path;
- be set to mode `0600`.

### 11.4 Content format

The file must contain exactly one nonblank UTF-8 line.

Accepted endings:

- no line ending;
- one trailing LF;
- one trailing CRLF.

Rejected content includes:

- an empty value;
- whitespace-only value;
- leading or trailing spaces;
- additional lines;
- embedded CR or LF;
- invalid UTF-8.

Fish client initialization also rejects ASCII control characters because the key is sent in an HTTP header.

Example creation:

```bash
install -d -m 700 secrets
printf '%s\n' 'YOUR_FISH_API_KEY' > secrets/fish-api-key
chmod 600 secrets/fish-api-key
```

Do not commit the key file.

---

## 12. Logging

The application writes structured logs to:

- standard error;
- one persistent log destination.

### `logging.level`

Minimum emitted level.

Default:

```text
info
```

Accepted exact values:

```text
debug
info
warn
error
```

### `logging.format`

Handler format.

Default:

```text
text
```

Accepted exact values:

```text
text
json
```

### `logging.logText`

Controls whether the application may include input or processed text in logs.

Default:

```text
false
```

Keep this disabled for private or sensitive text.

The flag does not authorize modules to bypass logging policy with independent global loggers.

### `logging.file`

Persistent log path.

Default:

```json
"file": ""
```

An empty or whitespace-only value uses:

```text
logs/fish-audio-cli.log
```

Relative paths are resolved from the project directory.

Absolute paths are cleaned and used without rebasing.

The logger:

- creates missing parent directories with requested mode `0750`;
- does not rewrite permissions of already existing parent directories;
- opens the configured destination in append mode;
- creates a missing file with requested mode `0640`;
- applies mode `0640` to the opened destination.

Standard error remains enabled in addition to file output.

There is currently no configuration value that disables the persistent log destination. An empty value selects the default file rather than disabling file logging.

A logrotate template is provided at:

```text
deploy/logrotate/fish-audio-cli
```

---

## 13. Path resolution

The config path determines the project directory used for configured relative paths.

### 13.1 Config path

The selected `--config` value is:

- trimmed;
- converted to an absolute path;
- cleaned.

### 13.2 Project directory rule

Let the absolute config path be:

```text
/path/to/project/config/config.json
```

Its containing directory is named `config`, so the project directory becomes:

```text
/path/to/project
```

For:

```text
/path/to/project/settings.json
```

the containing directory is not named `config`, so the project directory becomes:

```text
/path/to/project
```

In general:

```text
if parent directory basename == "config":
    project directory = parent of that directory
else:
    project directory = config file directory
```

### 13.3 Relative paths

Configured relative paths passed through the resolver are joined to the project directory.

This includes:

- `secrets.fishApiKeyFile`;
- `logging.file`;
- future module-owned paths resolved by module preparers.

### 13.4 Absolute paths

Absolute configured paths are cleaned and used without project rebasing.

### 13.5 Output path

`--output` is a CLI argument, not a configuration path.

It is passed to the output layer without project-root rebasing.

---

## 14. Built-in `passthrough` module

Configuration:

```json
{
  "name": "passthrough",
  "type": "passthrough",
  "config": {}
}
```

The module has no config fields.

Unknown fields are rejected:

```json
{
  "name": "passthrough",
  "type": "passthrough",
  "config": {
    "inventMeaning": true
  }
}
```

It checks cancellation and leaves text unchanged.

The module is useful for:

- validating registry wiring;
- testing pipeline execution;
- running synthesis without local text transformation;
- serving as the minimal module implementation reference.

An empty pipeline is also valid, so `passthrough` is not required merely to make configuration load.

---

## 15. Practical partial configurations

### Change only the model

```json
{
  "fish": {
    "model": "s2.1-pro"
  }
}
```

### Disable local text modules

```json
{
  "pipeline": {
    "modules": []
  }
}
```

### Enable server-error retries

```json
{
  "fish": {
    "retry": {
      "maxAttempts": 4,
      "retryServerErrors": true
    }
  }
}
```

This retains the default retry delays.

### JSON logging without text content

```json
{
  "logging": {
    "format": "json",
    "level": "info",
    "logText": false
  }
}
```

### Explicit Opus output settings

Configuration:

```json
{
  "fish": {
    "request": {
      "sampleRate": 48000,
      "opusBitrate": 64000
    }
  }
}
```

Invocation:

```bash
fish-audio-cli \
  --format opus \
  --output speech.opus \
  --text "Hello"
```

### Custom project-relative paths

```json
{
  "secrets": {
    "fishApiKeyFile": "private/fish.key"
  },
  "logging": {
    "file": "var/log/fish-audio-cli.log"
  }
}
```

Both paths resolve from the project directory determined by the config file path.

---

## 16. Common configuration failures

### Unknown field

```text
unknown JSON object key
```

Check spelling and exact capitalization.

### Duplicate key

```text
duplicate JSON object key
```

Remove the duplicate, including duplicates expressed with escaped characters.

### Unexpected null

```text
<path> must not be null
```

Omit the field to keep its default.

Use `null` only for `fish.request.sampleRate` or where a module explicitly allows a nested null inside its own config object.

### Missing module config

```text
pipeline.modules[N].config must be present
```

Add at least:

```json
"config": {}
```

### Duplicate module name

```text
pipeline.modules contains duplicate module name
```

Give each instance a unique operational name.

### Unsupported module type

```text
unsupported type
```

Use a type compiled into the current executable.

### Invalid sample rate for format

A rate may be globally recognized but incompatible with the selected `--format`.

Consult the format table under `fish.request.sampleRate`.

### Invalid unused bitrate

Both bitrate fields are validated even when the current invocation uses another format.

Keep every configured bitrate at a supported value.

### Retry delay relationship

```text
fish.retry.maxDelayMilliseconds must be greater than or equal to fish.retry.initialDelayMilliseconds
```

Increase the maximum or lower the initial delay.

### Missing API key

The application may create the configured key file securely and then stop.

Populate the file with exactly one key line and rerun.

---

## 17. Local files and version control

These repository paths are intended to remain local:

```text
config/config.json
secrets/
logs/
bin/
```

Do not commit:

- Fish API keys;
- machine-specific paths;
- private voice identifiers when they should remain private;
- generated audio;
- local log files.

The tracked example config must contain safe placeholders and current defaults.

---

## 18. Configuration change checklist

When changing configuration code or adding a field:

1. update the Go config type;
2. choose and implement a default;
3. define exact JSON spelling;
4. define null behavior;
5. add semantic validation and bounds;
6. decide path-resolution behavior;
7. update `config/config.example.json`;
8. update the field-summary table;
9. update the detailed field section;
10. add load and validation tests;
11. test exact capitalization;
12. test duplicate-key handling where relevant;
13. test omitted versus explicit values;
14. test boundary values;
15. update module docs for module-owned fields;
16. run `go test -race ./...`;
17. run `go vet ./...`;
18. verify documentation examples against current code.

A field is not complete merely because JSON can decode it.

---

## 19. Summary

The configuration system is designed to fail early and predictably:

- complete safe defaults;
- partial object overlays;
- full array replacement;
- independent module instances;
- strict UTF-8 JSON;
- exact field names;
- duplicate-key rejection;
- narrow null support;
- bounded files and delays;
- explicit path rules;
- separate secret storage;
- exact Fish request validation.

The canonical example is:

```text
config/config.example.json
```

Local operation normally uses:

```text
config/config.json
```

Keep the example, defaults, validation, and this reference synchronized. Four competing interpretations of one JSON field are not flexibility; they are paperwork with a runtime.
