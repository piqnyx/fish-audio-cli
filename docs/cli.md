# Command-line interface

> **Document status:** normative reference for the current pre-release command-line interface.
>
> **Audience:** users invoking `fish-audio-cli` directly, shell-script authors, service integrators, and maintainers reviewing CLI compatibility.
>
> **Scope:** this document describes command syntax, options, text input selection, format normalization, destination handling, output streams, signals, program-controlled exit statuses, and automation guidance. JSON settings are documented in [`configuration.md`](configuration.md); runtime architecture in [`architecture.md`](architecture.md); text-processing behavior in [`pipeline.md`](pipeline.md).

---

## 1. Purpose

`fish-audio-cli` is a single-invocation command-line program.

One successful invocation:

1. parses command-line options;
2. loads and validates configuration;
3. prepares and builds the configured text modules;
4. reads one text input;
5. processes that text through the ordered pipeline;
6. creates one Fish Audio request;
7. loads the Fish API key;
8. synthesizes one audio response;
9. publishes one destination file atomically;
10. exits.

The CLI is designed for:

- direct shell use;
- subprocess integration;
- local automation;
- wrappers that provide text through an option or standard input;
- callers that require an explicit output file.

It is not an interactive prompt.

It does not:

- read multiple independent requests from one process;
- stream audio to standard output;
- infer the audio format from the destination extension;
- accept positional text;
- create an output directory tree;
- run as a persistent server.

---

## 2. Synopsis

```text
fish-audio-cli [options]
```

Current usage text:

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

A normal invocation requires:

- a supported `--format`;
- a non-empty `--output`;
- valid input from `--text` or standard input;
- a readable valid configuration file;
- a valid Fish API key file by the time synthesis begins.

---

## 3. Option summary

| Option | Value | Default | Required | Meaning |
|---|---|---|---:|---|
| `--config` | path | `config/config.json` | no | JSON configuration file |
| `--text` | string | empty | no | text argument; empty selects stdin |
| `--format` | string | empty | yes in practice | `wav`, `mp3`, `opus`, or `ogg` |
| `--output` | path | empty | yes | final audio destination |
| `--help` | none | false | no | print usage to stdout and exit successfully |

`--format` has no explicit “missing” branch.

An omitted format remains empty and fails the supported-format check.

---

## 4. Parsing order

The CLI parses options before it:

- initializes the project path resolver;
- opens the configuration file;
- initializes persistent logging;
- builds modules;
- reads text input;
- loads the API key;
- contacts Fish Audio;
- touches the destination file.

This means malformed options fail early.

At a high level:

```text
arguments
    ↓
parse options
    ↓
resolve config path
    ↓
load and validate config
    ↓
initialize logging and modules
    ↓
read text
    ↓
process and synthesize
```

---

## 5. `--help`

Use:

```bash
fish-audio-cli --help
```

Behavior:

- usage is written to standard output;
- exit status is `0`;
- `--output` is not required;
- `--format` is not required;
- configuration is not loaded;
- modules are not initialized;
- text input is not read;
- no API key is loaded;
- no output file is created.

This makes help safe to call in packaging and installation checks.

### 5.1 Help with other arguments

The underlying flag parser handles `--help` as a help request.

Do not depend on unrelated invalid arguments being diagnosed when help is present.

Use a dedicated help invocation when checking the interface:

```bash
fish-audio-cli --help
```

### 5.2 Output stream

Help goes to:

```text
stdout
```

Normal diagnostics go to:

```text
stderr
```

This distinction allows:

```bash
fish-audio-cli --help > usage.txt
```

without mixing structured application logs into the usage text.

---

## 6. `--config`

Syntax:

```text
--config PATH
```

Default:

```text
config/config.json
```

Example:

```bash
fish-audio-cli \
  --config /etc/fish-audio-cli/config.json \
  --text "Привет" \
  --format opus \
  --output /tmp/speech.opus
```

### 6.1 Path resolution

The configuration path is converted to an absolute cleaned path.

The project directory used for configured relative paths is derived from that absolute path.

Special project-root behavior:

- when the configuration file’s parent directory is literally named `config`, the project directory is the parent of that `config` directory;
- otherwise, the project directory is the configuration file’s own parent directory.

Examples:

```text
/project/config/config.json
    → project directory: /project
```

```text
/etc/fish-audio-cli/settings.json
    → project directory: /etc/fish-audio-cli
```

Relative configured paths such as secret and log paths are resolved from that project directory.

### 6.2 Blank path

The path resolver trims surrounding whitespace.

An empty or whitespace-only configuration path is rejected.

Do not use filenames that intentionally depend on leading or trailing spaces.

### 6.3 Missing or invalid file

Failures include:

- missing file;
- unreadable file;
- file larger than 1 MiB;
- malformed JSON;
- invalid UTF-8;
- duplicate fields;
- unknown fields;
- invalid configuration values.

These are startup failures.

Text input is not read after configuration loading or validation fails.

### 6.4 Partial configuration

The JSON file may be partial because supplied fields overlay built-in defaults.

That behavior belongs to the configuration format, not to the CLI parser.

See [`configuration.md`](configuration.md).

---

## 7. `--text`

Syntax:

```text
--text TEXT
```

Example:

```bash
fish-audio-cli \
  --text "Привет, мир!" \
  --format opus \
  --output speech.opus
```

A non-empty option value takes precedence over standard input.

Example:

```bash
printf '%s' 'stdin is ignored' |
  fish-audio-cli \
    --text 'argument wins' \
    --format opus \
    --output speech.opus
```

The synthesized input is:

```text
argument wins
```

### 7.1 Exact empty-string rule

Source selection uses an exact string comparison.

If:

```text
--text value != ""
```

the text argument is used.

If:

```text
--text value == ""
```

standard input is used.

Therefore this explicit option:

```bash
fish-audio-cli \
  --text "" \
  --format opus \
  --output speech.opus
```

does not request an empty text argument.

It selects standard input.

### 7.2 Whitespace-only argument

Whitespace-only text is non-empty, so it takes precedence over stdin.

Then text validation rejects it.

Example:

```bash
printf '%s' 'valid stdin text' |
  fish-audio-cli \
    --text '   ' \
    --format opus \
    --output speech.opus
```

This fails.

The program does not fall back to stdin after rejecting the selected argument.

### 7.3 Shell quoting

Quote text containing spaces or shell metacharacters:

```bash
fish-audio-cli \
  --text 'Он сказал: "Привет".' \
  --format mp3 \
  --output speech.mp3
```

Without quoting:

```bash
--text Привет мир
```

the shell passes `мир` as a positional argument, which the CLI rejects.

### 7.4 Newlines in an argument

Shell quoting may include newlines.

Example:

```bash
fish-audio-cli \
  --text $'Первая строка.\nВторая строка.' \
  --format wav \
  --output speech.wav
```

The CLI does not trim input text.

The newline remains part of the text sent through the pipeline.

### 7.5 Argument-size limit

The configured `input.maxBytes` limit applies to `--text`.

The count is UTF-8 bytes, not characters.

An argument exactly at the configured byte limit is accepted.

An argument one byte above it is rejected.

---

## 8. Standard input

Standard input is used when the `--text` value is empty.

Typical invocation:

```bash
printf '%s' 'Текст через stdin' |
  fish-audio-cli \
    --format opus \
    --output speech.opus
```

### 8.1 Omitted `--text`

This is the normal stdin form:

```bash
printf '%s' 'Текст через stdin' |
  fish-audio-cli \
    --format opus \
    --output speech.opus
```

### 8.2 Explicit empty `--text`

This behaves the same way:

```bash
printf '%s' 'Текст через stdin' |
  fish-audio-cli \
    --text "" \
    --format opus \
    --output speech.opus
```

### 8.3 Complete read

The CLI reads standard input to completion before text processing starts.

It is not a streaming text protocol.

The writing side must:

- close the pipe;
- finish the redirected file;
- terminate the heredoc.

Otherwise the CLI continues waiting for EOF.

### 8.4 No newline removal

The CLI does not remove a trailing newline from stdin.

Compare:

```bash
printf '%s' 'Привет'
```

with:

```bash
echo 'Привет'
```

`echo` normally appends a newline.

That newline remains part of the input.

Use `printf '%s'` when an exact text value is required.

### 8.5 Multiline stdin

A heredoc is supported:

```bash
fish-audio-cli \
  --format mp3 \
  --output speech.mp3 <<'EOF'
Первая строка.
Вторая строка.
EOF
```

The heredoc’s final newline is part of the input.

### 8.6 Redirected file

Example:

```bash
fish-audio-cli \
  --format wav \
  --output speech.wav < input.txt
```

The entire file is subject to:

- `input.maxBytes`;
- UTF-8 validation;
- nonblank text validation.

### 8.7 Empty stdin

Empty stdin is rejected.

Whitespace-only stdin is also rejected.

Examples:

```bash
: |
  fish-audio-cli \
    --format opus \
    --output speech.opus
```

```bash
printf ' \n\t ' |
  fish-audio-cli \
    --format opus \
    --output speech.opus
```

Both fail before text processing.

---

## 9. Input validity

Selected text must:

- fit within `input.maxBytes`;
- be valid UTF-8;
- contain at least one non-whitespace Unicode code point.

The CLI does not:

- trim text;
- normalize Unicode;
- remove a byte-order mark;
- remove trailing newlines;
- repair malformed UTF-8;
- convert another character encoding to UTF-8.

Any local transformation happens later in configured modules.

### 9.1 Byte count versus character count

Example:

```text
ASCII character: often 1 UTF-8 byte
Cyrillic character: commonly 2 UTF-8 bytes
emoji: commonly 4 UTF-8 bytes
```

A one-million-character string may exceed a one-million-byte limit.

Operational logs report Unicode character counts, while the input limit is enforced in bytes.

### 9.2 Validation timing

Input is read only after:

- CLI parsing;
- path initialization;
- configuration loading;
- configuration validation;
- logger initialization;
- module preparation and construction.

An invalid module config can therefore fail before the process consumes stdin.

This is useful for callers using a pipe: configuration defects are discovered before text processing, though the producer may still need to handle a closed pipe.

---

## 10. Input-source precedence

The complete precedence rule is:

```text
non-empty --text
    ↓
use argument and ignore stdin

empty or omitted --text
    ↓
read stdin
```

There is no third source.

The CLI does not read text from:

- positional arguments;
- a JSON field;
- the output filename;
- an environment variable;
- a prompt.

### 10.1 Decision table

| `--text` | stdin | Selected source | Result |
|---|---|---|---|
| omitted | valid text | stdin | accepted |
| `""` | valid text | stdin | accepted |
| non-empty valid | anything | argument | accepted |
| whitespace-only | valid text | argument | rejected |
| non-empty oversized | valid text | argument | rejected |
| omitted | empty | stdin | rejected |
| omitted | invalid UTF-8 | stdin | rejected |

---

## 11. `--format`

Syntax:

```text
--format FORMAT
```

Supported user-facing values:

```text
wav
mp3
opus
ogg
```

The option is required in practice because its default empty string is unsupported.

### 11.1 Case normalization

The CLI converts the supplied format to lowercase.

These are accepted:

```text
WAV
Wav
MP3
OpUs
OGG
```

They normalize to:

```text
wav
wav
mp3
opus
opus
```

### 11.2 No whitespace trimming

The format value is lowercased but not trimmed.

These are rejected:

```text
" wav"
"wav "
" opus "
```

Pass the exact value without surrounding whitespace.

### 11.3 `ogg` alias

`ogg` is a CLI alias for `opus`.

When the caller passes:

```text
--format ogg
```

the stored internal format becomes:

```text
opus
```

Fish Audio receives an Opus-format request.

The output bytes are Opus data suitable for an Ogg/Opus destination.

### 11.4 Extension does not select format

The CLI does not inspect the destination filename extension.

This command requests MP3 data despite the `.wav` name:

```bash
fish-audio-cli \
  --text "Привет" \
  --format mp3 \
  --output speech.wav
```

The caller is responsible for matching extension and format.

Recommended combinations:

| Format option | Typical extension |
|---|---|
| `wav` | `.wav` |
| `mp3` | `.mp3` |
| `opus` | `.opus` |
| `ogg` | `.ogg` |

### 11.5 Public format set

The public CLI accepts only:

- `wav`;
- `mp3`;
- `opus`;
- `ogg`.

Internal Fish request code also understands `pcm`, but `pcm` is not a public CLI format.

Passing:

```text
--format pcm
```

is rejected during option parsing.

### 11.6 Format-specific configuration

`fish.request.sampleRate` is validated against the selected format later when the Fish request is built.

Supported non-null combinations:

| Format | Sample rates |
|---|---|
| `wav` | `8000`, `16000`, `24000`, `32000`, `44100` |
| `mp3` | `32000`, `44100` |
| `opus` / `ogg` | `48000` |

The bitrate fields remain part of configuration validation even when the current invocation selects another format.

See [`configuration.md`](configuration.md).

---

## 12. `--output`

Syntax:

```text
--output PATH
```

Example:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output /tmp/speech.opus
```

The option is mandatory.

An exact empty value is rejected:

```bash
fish-audio-cli \
  --format opus \
  --output ""
```

### 12.1 No extension validation

The destination extension is not compared with `--format`.

The CLI writes the selected audio format to the supplied path exactly.

### 12.2 Parent directory

The output subsystem does not create missing parent directories.

Create them first:

```bash
mkdir -p /var/lib/my-service/audio

fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output /var/lib/my-service/audio/speech.opus
```

### 12.3 Relative path

A relative output path is interpreted by the operating system relative to the process working directory.

It is not resolved relative to the configuration project directory.

Example:

```bash
cd /tmp/job-123

fish-audio-cli \
  --config /opt/fish-audio-cli/config/config.json \
  --text "Привет" \
  --format opus \
  --output result.opus
```

The destination is:

```text
/tmp/job-123/result.opus
```

not:

```text
/opt/fish-audio-cli/result.opus
```

### 12.4 Atomic publication

Audio is written through the atomic output subsystem.

At a high level:

1. create a temporary file beside the destination;
2. stream synthesized audio into the temporary file;
3. synchronize and close it;
4. rename it over the destination;
5. synchronize the containing directory.

Consequences:

- partial synthesis data is not published under the final destination name;
- an existing destination remains in place until replacement;
- the temporary file uses restrictive permissions;
- a symlink at the final path is replaced as a directory entry rather than followed as the output file;
- a failure before rename preserves the old destination.

### 12.5 Existing destination

A successful invocation replaces an existing destination atomically where supported by the filesystem.

Callers should not expect append behavior.

### 12.6 Audio is not written to stdout

There is no output-path value meaning “standard output”.

Values such as:

```text
-
/dev/stdout
```

are treated as ordinary filesystem paths according to operating-system behavior, not as a documented streaming interface.

For stable integration, use a real destination file.

### 12.7 Whitespace path nuance

Option parsing checks only whether the output value is exactly empty.

It does not trim the path.

Avoid whitespace-only or surrounding-whitespace paths even if the operating system permits them.

---

## 13. Positional arguments

Positional arguments are not supported.

This is rejected:

```bash
fish-audio-cli \
  --format opus \
  --output speech.opus \
  "Привет"
```

Use:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output speech.opus
```

or:

```bash
printf '%s' 'Привет' |
  fish-audio-cli \
    --format opus \
    --output speech.opus
```

All unparsed positional values produce an option-parsing failure.

This includes accidental shell-splitting caused by missing quotes.

---

## 14. Unknown and malformed options

Unknown flags are rejected.

Examples:

```text
--formats
--outfile
--voice
```

A missing flag value is rejected by the flag parser.

Example:

```bash
fish-audio-cli --output
```

Option-parsing failures:

- are logged to stderr;
- use program-controlled exit status `2`;
- do not load configuration;
- do not read stdin;
- do not contact Fish Audio.

---

## 15. Standard output

On an ordinary successful synthesis, the CLI writes no result payload to stdout.

Standard output is reserved for current help output.

Therefore:

```bash
audio_path="$(
  fish-audio-cli \
    --text "Привет" \
    --format opus \
    --output speech.opus
)"
```

does not return `speech.opus`.

The caller already supplied the destination and should retain that value itself.

### 15.1 Stable automation rule

Treat stdout as:

- usage text for `--help`;
- otherwise currently empty.

Do not use stdout as the audio transport.

---

## 16. Standard error and persistent logs

Diagnostics are written to stderr.

After configuration is loaded and the configured logger is opened, records are written to:

- stderr;
- the configured persistent log file.

Early failures before configured logger initialization appear only on stderr.

Examples include:

- bootstrap logger failure;
- request-ID generation failure;
- CLI option failure;
- project-path failure;
- config loading failure;
- config validation failure;
- configured logger initialization failure.

### 16.1 Log format

The configured logger may emit:

- text records;
- JSON records.

This affects stderr as well as the persistent log after logger initialization.

Before that point, the bootstrap logger uses its own initial text behavior.

### 16.2 Text privacy

Full input and processed text are logged only when:

```json
{
  "logging": {
    "logText": true
  }
}
```

The default is `false`.

Character counts are still logged.

### 16.3 Shell redirection

Capture diagnostics:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output speech.opus \
  2> fish-audio-cli.stderr
```

Capture help only:

```bash
fish-audio-cli --help > usage.txt
```

Capture both separately:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output speech.opus \
  > stdout.txt \
  2> stderr.txt
```

On successful synthesis, `stdout.txt` is normally empty.

---

## 17. Signals and cancellation

After input has been read and validated, the CLI creates a context canceled by:

- `SIGINT`;
- `SIGTERM`.

Typical sources:

```text
Ctrl+C → SIGINT
service stop → often SIGTERM
```

The same context is passed through:

- pipeline processing;
- module processors;
- Fish synthesis;
- retry waits.

### 17.1 During pipeline processing

When cancellation is observed during pipeline processing:

- the interrupted step’s current text is rolled back;
- fallback policies do not override cancellation;
- remaining modules do not run;
- Fish synthesis does not start;
- the CLI returns the text-processing failure stage.

The program-controlled exit status is `3`.

### 17.2 During Fish synthesis

When cancellation is observed during synthesis or a retry wait:

- the Fish request stops according to context behavior;
- atomic output does not publish a partial final destination;
- the CLI returns the synthesis/output failure stage.

The program-controlled exit status is `4`.

### 17.3 Module cooperation

A long-running module must use the supplied context.

The pipeline cannot forcibly interrupt module code that ignores it.

### 17.4 Initialization timing

Signal-aware context is created after:

- CLI parsing;
- configuration loading;
- module initialization;
- input reading.

Before that point, process behavior follows ordinary operating-system signal handling rather than the managed pipeline/synthesis cancellation path.

### 17.5 Shell conventions

Shells may report different statuses when a process is terminated by a signal before the CLI can return one of its own statuses.

The table in this document covers statuses explicitly returned by the program.

---

## 18. Program-controlled exit statuses

The current executable returns these statuses:

| Status | Meaning |
|---:|---|
| `0` | successful synthesis or help request |
| `1` | bootstrap logging or request-ID initialization failure |
| `2` | options, paths, configuration, logger, modules, application, or input failure |
| `3` | text processing, Fish request construction, API key, or Fish client initialization failure |
| `4` | Fish synthesis or atomic output failure |

### 18.1 Status `0`

Returned for:

- successful complete invocation;
- `--help`.

A successful exit means the destination was published.

A `pipeline` outcome of `recovered` or `stopped` may still lead to successful synthesis and exit `0`.

### 18.2 Status `1`

Reserved for very early internal startup failures:

- bootstrap logger creation;
- request-ID generation.

These failures happen before ordinary configured logging exists.

### 18.3 Status `2`

Covers pre-processing failures:

- invalid or unknown CLI option;
- missing `--output`;
- unsupported or omitted `--format`;
- positional arguments;
- invalid config path;
- config read or decode failure;
- config semantic validation failure;
- logger initialization failure;
- module initialization failure;
- pipeline application initialization failure;
- text input failure.

### 18.4 Status `3`

Covers failures after valid text has been selected but before synthesis output begins successfully:

- pipeline processing error or cancellation;
- selected format incompatible with request settings;
- Fish request construction failure;
- Fish API key file creation notice;
- empty or invalid API key file;
- Fish client initialization failure.

When a missing API key file is created securely, the invocation still returns `3`.

The user must populate the created file and run the command again.

### 18.5 Status `4`

Covers:

- Fish HTTP/API synthesis failure;
- exhausted retryable Fish responses;
- response streaming failure;
- zero-byte successful response rejection;
- temporary output failure;
- file synchronization or close failure;
- destination rename failure;
- final directory synchronization failure;
- cancellation during synthesis.

### 18.6 Do not parse log text for status

Automation should inspect the numeric process status.

Log wording is for diagnostics and may evolve.

---

## 19. Complete examples

### 19.1 Direct text to Opus

```bash
fish-audio-cli \
  --text "Привет, мир!" \
  --format opus \
  --output speech.opus
```

### 19.2 Direct text to Ogg/Opus

```bash
fish-audio-cli \
  --text "Привет, мир!" \
  --format ogg \
  --output speech.ogg
```

Internally, `ogg` becomes `opus`.

### 19.3 Standard input to MP3

```bash
printf '%s' 'Текст через stdin' |
  fish-audio-cli \
    --format mp3 \
    --output speech.mp3
```

### 19.4 File input to WAV

```bash
fish-audio-cli \
  --format wav \
  --output speech.wav < input.txt
```

### 19.5 Custom configuration

```bash
fish-audio-cli \
  --config /opt/fish-audio-cli/config/config.json \
  --text "Привет" \
  --format opus \
  --output /tmp/speech.opus
```

### 19.6 Multiline text

```bash
fish-audio-cli \
  --format mp3 \
  --output speech.mp3 <<'EOF'
Первый абзац.

Второй абзац.
EOF
```

### 19.7 Uppercase format

```bash
fish-audio-cli \
  --text "Привет" \
  --format OPUS \
  --output speech.opus
```

Accepted and normalized to `opus`.

### 19.8 Preserve caller-selected output path

```bash
output_path="/tmp/speech-$$.opus"

if fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output "$output_path"
then
  printf 'created: %s\n' "$output_path"
else
  status=$?
  printf 'synthesis failed with status %d\n' "$status" >&2
  exit "$status"
fi
```

### 19.9 Safe exact stdin text

```bash
text='Строка без автоматически добавленного перевода строки.'

printf '%s' "$text" |
  fish-audio-cli \
    --format opus \
    --output speech.opus
```

### 19.10 Service-style invocation

```bash
install -d -m 0750 /var/lib/voice-worker/output

printf '%s' "$MESSAGE_TEXT" |
  /opt/fish-audio-cli/bin/fish-audio-cli \
    --config /opt/fish-audio-cli/config/config.json \
    --format opus \
    --output /var/lib/voice-worker/output/message.opus
```

The parent output directory is created by the service wrapper, not by the CLI.

---

## 20. Automation guidance

### 20.1 Always pass format explicitly

Do not infer that a `.opus` destination causes an Opus request.

Use both:

```text
--format opus
--output file.opus
```

### 20.2 Always retain the output path

The CLI does not print it to stdout as a machine-readable result.

### 20.3 Inspect exit status

Example:

```bash
if ! fish-audio-cli \
  --text "$text" \
  --format opus \
  --output "$output"
then
  status=$?
  printf 'fish-audio-cli failed: %d\n' "$status" >&2
  exit "$status"
fi
```

Be careful with shell negation: in some patterns, `$?` reflects the negated condition rather than the original command.

A clearer form is:

```bash
fish-audio-cli \
  --text "$text" \
  --format opus \
  --output "$output"

status=$?

if [ "$status" -ne 0 ]; then
  printf 'fish-audio-cli failed: %d\n' "$status" >&2
  exit "$status"
fi
```

### 20.4 Use unique destination paths for parallel jobs

Two concurrent invocations targeting the same destination race at the filesystem publication boundary.

Use per-job paths:

```text
output/<request-id>.opus
```

### 20.5 Bound producer input

The CLI enforces `input.maxBytes`, but a producer should also avoid generating unbounded text.

### 20.6 Close stdin

A subprocess wrapper must close the child’s stdin after writing text.

Otherwise the CLI waits for EOF.

### 20.7 Treat stderr as diagnostics

Do not parse stderr as stable structured output unless the caller intentionally configures JSON logging and accepts the documented log schema as a separate integration concern.

### 20.8 Do not rely on current silence of stdout

Normal stdout is currently empty, but callers should not use “empty stdout” as the success test.

Use exit status and destination publication.

---

## 21. Compatibility-oriented invocation

Wrappers commonly provide text in one of two ways.

### Argument style

```bash
fish-audio-cli \
  --text "$TEXT" \
  --format opus \
  --output "$OUTPUT"
```

### Stdin style

```bash
printf '%s' "$TEXT" |
  fish-audio-cli \
    --format opus \
    --output "$OUTPUT"
```

Both paths use the same:

- byte limit;
- UTF-8 validation;
- nonblank text contract;
- pipeline;
- Fish request;
- atomic output behavior.

### 21.1 Prefer stdin for large or complex text

Stdin avoids:

- shell quoting complexity;
- command-line length limits;
- accidental exposure in process listings on some systems.

The input still must fit `input.maxBytes`.

### 21.2 Prefer `--text` for short controlled values

`--text` is convenient when:

- the value is short;
- the caller already has safe subprocess argument handling;
- shell expansion is not involved.

### 21.3 No positional compatibility mode

A wrapper that currently invokes:

```text
fish-audio-cli "text"
```

must be changed to use:

```text
fish-audio-cli --text "text"
```

or stdin.

---

## 22. Common mistakes

### Missing `--format`

Incorrect:

```bash
fish-audio-cli \
  --text "Привет" \
  --output speech.opus
```

The empty default format is unsupported.

### Missing `--output`

Incorrect:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus
```

The CLI never selects a destination automatically.

### Unquoted text

Incorrect:

```bash
fish-audio-cli \
  --text Привет мир \
  --format opus \
  --output speech.opus
```

`мир` becomes an unexpected positional argument.

### Assuming extension selects format

Incorrect assumption:

```text
speech.mp3 means MP3 automatically
```

The actual format comes only from `--format`.

### Passing padded format text

Incorrect:

```bash
--format " opus "
```

Format is not trimmed.

### Expecting stdin fallback after invalid argument

Incorrect:

```bash
printf '%s' 'valid stdin' |
  fish-audio-cli \
    --text '   ' \
    --format opus \
    --output speech.opus
```

Whitespace-only `--text` wins source selection and then fails validation.

### Expecting output on stdout

Incorrect:

```bash
fish-audio-cli ... > speech.opus
```

Audio is written through `--output`, not stdout.

### Missing parent directory

Incorrect:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output missing/directory/speech.opus
```

The CLI does not create `missing/directory`.

### Assuming `echo` preserves exact text

`echo` normally appends a newline.

Use:

```bash
printf '%s' "$text"
```

for exact stdin content.

---

## 23. Security considerations

### 23.1 Command-line visibility

Text passed through `--text` may be visible to:

- local process inspection tools;
- shell history;
- service-manager diagnostics;
- audit systems.

For sensitive text, stdin is generally preferable.

### 23.2 API key

The API key is never accepted as a CLI option.

It is loaded from the configured secret file after text processing.

This avoids exposing it in command-line arguments.

### 23.3 Output path

The caller controls the output path.

Run the CLI with filesystem permissions appropriate for that destination.

### 23.4 Logs

Input and processed text remain excluded from logs by default.

Keep:

```json
"logText": false
```

for sensitive workloads.

### 23.5 Configuration path

A custom config path may redirect:

- relative secret paths;
- relative log paths;
- future module-owned relative paths.

Treat config selection as a security-relevant choice.

---

## 24. Performance considerations

### 24.1 Startup per invocation

Every invocation performs initialization again:

- config read and validation;
- logger setup;
- module preparation and construction;
- secret load;
- Fish client construction.

The CLI is optimized for correctness and process isolation, not persistent-process throughput.

### 24.2 Input buffering

The complete selected text is read into memory within the configured byte limit.

### 24.3 Ordered modules

Module processing is sequential.

Total preprocessing latency is approximately the sum of configured module latencies.

### 24.4 Audio streaming

Fish audio is streamed into a temporary output file rather than accumulated entirely in memory.

### 24.5 Retry delay

Retryable Fish responses may extend invocation duration according to configured retry delays and `Retry-After`.

The Fish HTTP timeout applies to each HTTP request attempt.

---

## 25. CLI invariants

The following rules are normative for the current interface.

1. The command accepts options only; positional arguments are rejected.
2. `--config` defaults to `config/config.json`.
3. `--output` is mandatory.
4. `--format` must resolve to `wav`, `mp3`, or `opus`.
5. User-facing `ogg` normalizes to `opus`.
6. Format is lowercased but not trimmed.
7. Destination extension does not determine format.
8. A non-empty `--text` takes precedence over stdin.
9. An empty or omitted `--text` selects stdin.
10. Whitespace-only selected text is rejected.
11. Text is not trimmed.
12. Input must be valid UTF-8.
13. Input must fit `input.maxBytes`.
14. Input must contain non-whitespace text.
15. Audio is not emitted to stdout.
16. Help is emitted to stdout.
17. Diagnostics are emitted to stderr.
18. Persistent logging begins only after configuration and logger initialization.
19. Output parent directories are not created.
20. Output publication is atomic.
21. Existing destinations are replaced, not appended.
22. Managed `SIGINT` and `SIGTERM` cancellation begins after input reading.
23. Help returns status `0`.
24. Successful publication returns status `0`.
25. Program-controlled failure stages use statuses `1` through `4`.

A change to one of these rules is a CLI compatibility change.

It requires:

- implementation review;
- tests;
- documentation updates;
- migration consideration for wrappers.

---

## 26. Test expectations

CLI behavior should remain covered by tests for:

### Options

- default config path;
- custom config path;
- required output;
- every supported format;
- case normalization;
- `ogg` normalization;
- unsupported formats;
- omitted format;
- help;
- unknown flags;
- missing flag values;
- positional arguments.

### Input selection

- non-empty argument wins over stdin;
- empty argument uses stdin;
- omitted argument uses stdin;
- whitespace-only argument fails without stdin fallback;
- empty stdin;
- whitespace-only stdin;
- nil stdin in package tests.

### Input validation

- valid UTF-8;
- invalid UTF-8 argument;
- invalid UTF-8 stdin;
- exactly at byte limit;
- one byte above limit;
- multibyte UTF-8 byte counting;
- newline preservation.

### Main command

- help writes usage to stdout;
- option failures return `2`;
- config failures return `2`;
- input failures return `2`;
- pipeline failures return `3`;
- secret-file creation returns `3`;
- synthesis failures return `4`;
- success returns `0`;
- destination publication occurs only on success.

### Signals

- pipeline cancellation;
- synthesis cancellation;
- retry-wait cancellation;
- no partial final destination.

---

## 27. Review checklist

When changing CLI code, verify:

### Arguments

- Are existing flag names preserved?
- Are defaults unchanged or deliberately migrated?
- Are positional arguments still rejected?
- Is help still available without other required options?
- Is format normalization explicit?

### Input

- Is the argument-versus-stdin rule unchanged?
- Is text still bounded before processing?
- Is UTF-8 validation preserved?
- Is whitespace-only text rejected?
- Is input free from undocumented trimming?

### Output

- Is `--output` still explicit?
- Is format independent from extension?
- Is atomic publication preserved?
- Are parent-directory semantics unchanged?
- Is stdout still free of audio data?

### Automation

- Are exit statuses stable?
- Are logs kept on stderr?
- Are errors wrapped with useful stage context?
- Are signals propagated through context?
- Can wrappers close stdin and observe completion reliably?

### Documentation

- Does usage text match this document?
- Do examples use real supported options?
- Are public formats accurate?
- Are program-controlled statuses accurate?
- Are security implications disclosed?

---

## 28. Summary

The stable CLI model is:

```text
choose config
    ↓
provide text by --text or stdin
    ↓
choose format explicitly
    ↓
choose destination explicitly
    ↓
inspect exit status
```

Canonical argument invocation:

```bash
fish-audio-cli \
  --config config/config.json \
  --text "Привет" \
  --format opus \
  --output speech.opus
```

Canonical stdin invocation:

```bash
printf '%s' 'Привет' |
  fish-audio-cli \
    --config config/config.json \
    --format opus \
    --output speech.opus
```

The most important integration rules are:

- quote text or use stdin;
- close stdin after writing;
- pass `--format` explicitly;
- pass `--output` explicitly;
- do not infer format from extension;
- do not expect audio on stdout;
- use the numeric exit status;
- create the output parent directory before invocation.
