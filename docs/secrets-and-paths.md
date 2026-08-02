# Secrets and path resolution

> **Document status:** normative description of the current pre-release filesystem-path and secret-file behavior.
>
> **Audience:** operators installing `fish-audio-cli`, service integrators choosing filesystem locations, module authors resolving module-owned paths, and maintainers reviewing file-security boundaries.
>
> **Scope:** this document describes configuration-path normalization, project-directory derivation, core path resolution, Fish API key location and loading, secret file creation and validation, permissions, symlink and race handling, logging and output path differences, repository hygiene, and compatibility constraints. Configuration fields are documented in [`configuration.md`](configuration.md); command invocation in [`cli.md`](cli.md); architecture ownership in [`architecture.md`](architecture.md); logging-specific file behavior in [`logging.md`](logging.md); module path ownership in [`module-author-guide.md`](module-author-guide.md).

---

## 1. Purpose

`fish-audio-cli` uses several path domains with deliberately different semantics.

The main domains are:

```text
configuration file path
project-relative configured paths
Fish API key file
persistent log file
module-owned paths
CLI output path
```

They do not all resolve from the same directory.

The stable mental model is:

```text
--config
    ↓
absolute lexical config path
    ↓
derived project directory
    ↓
core/project-relative paths
        ├─ Fish API key
        ├─ persistent log file
        └─ module-owned paths when modules choose to resolve them

--output
    ↓
ordinary process-working-directory path
```

The project path resolver is intentionally small.

It provides:

- one absolute configuration path;
- one derived project directory;
- lexical resolution of absolute or project-relative paths.

It does not provide:

- filesystem canonicalization;
- symlink evaluation;
- file existence checks;
- permission checks;
- ownership checks;
- environment-variable expansion;
- home-directory expansion;
- sandboxing.

Those policies belong to the subsystem that consumes the resolved path.

---

## 2. Path ownership

| Path | Supplied by | Resolved by | Relative base |
|---|---|---|---|
| configuration file | `--config` | `projectpath.New` | process working directory |
| Fish API key | `secrets.fishApiKeyFile` | `config.Load` through resolver | derived project directory |
| log file | `logging.file` | logging package through resolver | derived project directory |
| module-owned file | module `config` | module implementation | module decides, normally resolver |
| output audio | `--output` | operating system and output package | process working directory |
| Fish base URL | `fish.baseUrl` | URL resolver, not filesystem resolver | not applicable |

A field that merely contains a string resembling a path is not automatically resolved.

Core resolves only paths it owns.

Modules resolve their own path fields.

---

## 3. Configuration path input

The command-line option is:

```text
--config PATH
```

Default:

```text
config/config.json
```

The path resolver receives the option value before configuration loading.

### 3.1 Surrounding whitespace

The configuration path is trimmed.

These values resolve identically:

```text
config/config.json
<leading-space>config/config.json
config/config.json<trailing-space>
```

A blank or whitespace-only value is rejected.

Do not intentionally use configuration filenames whose names begin or end with whitespace.

### 3.2 Relative path

A relative configuration path is converted to an absolute path using the process working directory at resolver construction time.

Example:

```text
working directory:
    /opt/fish-audio-cli

--config:
    config/config.json

absolute config path:
    /opt/fish-audio-cli/config/config.json
```

Changing the working directory changes the meaning of a relative `--config`.

### 3.3 Absolute path

An absolute configuration path remains absolute and is cleaned lexically.

Example:

```text
/opt/fish-audio-cli/config/../config/config.json
```

becomes:

```text
/opt/fish-audio-cli/config/config.json
```

### 3.4 No existence check in resolver

Creating the resolver does not prove that the configuration file:

- exists;
- is readable;
- is regular;
- is safe;
- contains valid JSON.

The configuration loader performs the later open and read.

### 3.5 No tilde expansion

The resolver does not expand:

```text
~
~/config.json
```

A shell may expand an unquoted tilde before launching the process, but JSON configuration values are not shell-expanded.

### 3.6 No environment expansion

The resolver does not expand:

```text
$HOME
${HOME}
%USERPROFILE%
```

Such text remains literal path content.

---

## 4. Lexical cleaning

Path normalization uses the operating system’s `filepath` rules.

Lexical cleaning can:

- remove redundant separators;
- remove `.` components;
- collapse ordinary `..` components;
- normalize path separators according to the host platform.

It does not inspect the filesystem.

### 4.1 Example

```text
/project/config/./sub/../config.json
```

becomes lexically equivalent to:

```text
/project/config/config.json
```

### 4.2 Symlinks are not evaluated

Lexical cleaning does not call a filesystem canonicalization operation such as:

```text
EvalSymlinks
realpath
```

The resulting absolute path may still contain symlinked components.

### 4.3 Case is not normalized

The resolver does not lowercase or otherwise normalize filename case.

The special directory name:

```text
config
```

is matched by exact string equality.

A directory named:

```text
Config
CONFIG
config<trailing-space>
```

does not trigger the special project-root rule.

---

## 5. Project-directory derivation

The resolver derives one project directory from the absolute cleaned configuration path.

Start with:

```text
directory containing config file
```

Then apply one special rule:

```text
if that directory's basename is exactly "config",
use its parent as the project directory
```

### 5.1 Standard repository layout

```text
/project/config/config.json
```

produces:

```text
config path:
    /project/config/config.json

project directory:
    /project
```

### 5.2 Configuration outside `config`

```text
/etc/fish-audio-cli/settings.json
```

produces:

```text
config path:
    /etc/fish-audio-cli/settings.json

project directory:
    /etc/fish-audio-cli
```

### 5.3 Nested `config` directory

```text
/srv/apps/voice/config/settings.json
```

produces:

```text
project directory:
    /srv/apps/voice
```

The filename does not need to be `config.json`.

Only the immediate parent directory name matters.

### 5.4 Directory named `Config`

```text
/srv/apps/voice/Config/settings.json
```

produces:

```text
project directory:
    /srv/apps/voice/Config
```

The comparison is exact.

### 5.5 Root-level edge case

For a configuration located directly in a root-level directory named `config`, lexical parent rules determine the result.

Operators should use an explicit conventional installation layout rather than depend on unusual filesystem-root edge cases.

---

## 6. Symlinked configuration path

The resolver derives the project directory from the supplied lexical configuration path, not from the configuration file’s ultimate symlink target.

Example:

```text
supplied path:
    /opt/voice/config/config.json

actual symlink target:
    /etc/voice/config.json
```

The project directory remains derived from:

```text
/opt/voice/config/config.json
```

and is therefore:

```text
/opt/voice
```

It is not automatically changed to:

```text
/etc/voice
```

### 6.1 Configuration loader behavior

The configuration loader uses ordinary file opening for the config path.

It does not reject a symlinked configuration file.

### 6.2 Operational consequence

Relative configured paths follow the lexical location selected by `--config`, even when the configuration bytes come from a symlink target elsewhere.

This can be useful for deployment indirection, but it must be intentional.

### 6.3 Trust consequence

A configuration symlink can redirect:

- configuration content;
- Fish endpoint selection;
- model selection;
- relative secret path values;
- log path values;
- module configuration.

The resolver does not authenticate the symlink target.

Protect the configuration path and its parent directories.

---

## 7. Resolver API contract

The resolver exposes two relevant operations.

### Configuration path

Conceptually:

```go
ConfigPath() string
```

Returns:

- absolute;
- lexically cleaned;
- possibly symlink-containing;
- configuration path string.

### Generic path resolution

Conceptually:

```go
Resolve(path string) (string, error)
```

Behavior:

```text
trim surrounding whitespace
    ↓
blank?
    ├─ yes → error
    └─ no
         ↓
absolute?
    ├─ yes → lexical clean and return
    └─ no
         ↓
resolver initialized?
    ├─ no → error
    └─ yes → join with project directory
```

---

## 8. Resolving absolute paths

An absolute input path:

- is trimmed;
- is lexically cleaned;
- does not require an initialized resolver;
- is not rebased to the project directory.

Example:

```text
/var/lib/fish-audio-cli/../fish-audio-cli/secrets/key
```

becomes:

```text
/var/lib/fish-audio-cli/secrets/key
```

### 8.1 Uninitialized resolver

A zero-value resolver can still clean an absolute path.

This package behavior is useful in tests and internal callers.

The CLI normally initializes the resolver before any configured path is used.

---

## 9. Resolving relative paths

A relative path:

- is trimmed;
- must be nonblank;
- requires an initialized resolver;
- is joined to the derived project directory.

Example:

```text
project directory:
    /opt/fish-audio-cli

configured path:
    secrets/fish-api-key

resolved path:
    /opt/fish-audio-cli/secrets/fish-api-key
```

`filepath.Join` performs lexical cleaning during the join.

### 9.1 Parent traversal

A relative path may contain:

```text
..
```

Example:

```text
project directory:
    /opt/fish-audio-cli

configured path:
    ../shared/key
```

resolves lexically to:

```text
/opt/shared/key
```

The resolver does not confine paths beneath the project directory.

It is a convenience resolver, not a sandbox.

### 9.2 Absolute escape

An absolute configured path can point anywhere permitted by process filesystem permissions.

### 9.3 Security implication

Anyone who can modify trusted configuration can redirect:

- the Fish API key path;
- the log path;
- module-owned paths that use the resolver.

Configuration integrity is therefore security-sensitive.

---

## 10. Configuration loading order

The main startup sequence is:

```text
parse CLI
    ↓
create resolver from --config
    ↓
open absolute config path
    ↓
read at most 1 MiB
    ↓
strictly decode over defaults
    ↓
validate explicit-null rules
    ↓
resolve Fish API key path
    ↓
return config
    ↓
validate semantic values
```

### 10.1 Resolver initialization failure

A blank configuration path fails before the config file is opened.

### 10.2 Configuration open failure

A missing or unreadable configuration file fails before:

- persistent logging;
- module construction;
- input reading;
- secret loading.

### 10.3 Read and close failures

The configuration loader preserves both:

- read failure;
- close failure

when both occur.

### 10.4 No config permission policy

The configuration loader does not enforce:

- regular-file type;
- owner;
- mode;
- symlink prohibition;
- directory writability;
- maximum link count.

It enforces content size and JSON correctness, not configuration-file filesystem security.

Operators must protect the configuration file separately.

---

## 11. Fish API key path resolution

The default configuration contains:

```json
{
  "secrets": {
    "fishApiKeyFile": "secrets/fish-api-key"
  }
}
```

During `config.Load`, this path is immediately resolved through the project resolver.

The returned in-memory configuration therefore contains an absolute path.

### 11.1 Standard layout

For:

```text
config:
    /opt/fish-audio-cli/config/config.json
```

the default secret path becomes:

```text
/opt/fish-audio-cli/secrets/fish-api-key
```

### 11.2 Configuration elsewhere

For:

```text
config:
    /etc/fish-audio-cli/settings.json
```

the default secret path becomes:

```text
/etc/fish-audio-cli/secrets/fish-api-key
```

### 11.3 Absolute configured key path

Configuration:

```json
{
  "secrets": {
    "fishApiKeyFile": "/run/secrets/fish-api-key"
  }
}
```

remains absolute after lexical cleaning.

### 11.4 Surrounding whitespace nuance

The project resolver trims the configured key path before storing the absolute result.

Therefore a JSON value padded with spaces is normalized during config loading.

The lower-level secret loader itself rejects surrounding whitespace when called directly, but the normal CLI path has already been normalized.

Do not rely on padding.

Use a clean path value.

### 11.5 Resolution before semantic validation

The key path is resolved before the main `Config.Validate` call.

A blank key path therefore fails during path resolution before ordinary semantic validation.

---

## 12. Core does not resolve every path-like field

Only the core-owned Fish API key path is resolved inside `config.Load`.

The loader does not recursively inspect JSON strings looking for paths.

It does not resolve:

- arbitrary module configuration strings;
- output path;
- Fish base URL;
- future fields merely because their names end in `Path` or `File`.

This prevents core from guessing module semantics.

---

## 13. Module-owned paths

Each module instance owns its nested:

```text
pipeline.modules[].config
```

A module that defines a path field should resolve it during `Prepare`.

The module receives the narrow project path resolver rather than the complete core configuration.

### 13.1 Recommended module rule

For a module-owned path:

```text
absolute path
    → use cleaned absolute path

relative path
    → resolve from project directory
```

This matches core-owned configured paths.

### 13.2 Module responsibility

The module must decide and document:

- whether blank is allowed;
- whether surrounding whitespace is rejected or trimmed;
- whether the path may escape project directory;
- whether symlinks are allowed;
- whether a file must exist;
- whether directories are created;
- required permissions;
- read/write behavior;
- size limits;
- lifecycle and cleanup.

### 13.3 No automatic inheritance

A module config path is not automatically replaced by an absolute path before module preparation.

The module must call the resolver deliberately.

### 13.4 Prepare phase

Pure lexical path resolution is suitable for `Prepare`.

Resource creation and stateful file opening belong to the module builder unless the module’s contract explicitly requires another boundary.

See [`module-author-guide.md`](module-author-guide.md).

---

## 14. Persistent log path

`logging.file` is resolved later by the logging package, after full configuration validation.

Default configured value:

```json
"file": ""
```

Empty or whitespace-only selects:

```text
logs/fish-audio-cli.log
```

That default is then resolved from the project directory.

### 14.1 Standard layout

```text
config:
    /opt/fish-audio-cli/config/config.json

log:
    /opt/fish-audio-cli/logs/fish-audio-cli.log
```

### 14.2 Absolute log path

```json
{
  "logging": {
    "file": "/var/log/fish-audio-cli/application.log"
  }
}
```

is used as an absolute cleaned path.

### 14.3 Logging path security differs from secret security

The logging package:

- creates missing directories;
- opens in append mode;
- chmods the file to `0640`.

It does not apply the secret loader’s leaf-symlink and same-file checks.

See [`logging.md`](logging.md).

---

## 15. CLI output path

`--output` does not use the project resolver.

A relative output path follows the process working directory.

Example:

```text
working directory:
    /tmp/job-42

config:
    /opt/fish-audio-cli/config/config.json

--output:
    result.opus
```

Destination:

```text
/tmp/job-42/result.opus
```

not:

```text
/opt/fish-audio-cli/result.opus
```

### 15.1 No trimming by CLI parser

The output option is checked only for exact empty string.

It is not trimmed or project-resolved.

Avoid leading or trailing whitespace in output paths.

### 15.2 Parent creation

The output subsystem does not create missing parent directories.

### 15.3 Final symlink behavior

Atomic publication replaces the final destination directory entry.

A symlink at the destination path is replaced rather than followed as the final output file.

Detailed publication semantics belong to the output documentation.

---

## 16. Path-domain comparison

| Behavior | Config path | Fish secret path | Log path | Module path | Output path |
|---|---:|---:|---:|---:|---:|
| relative to cwd | yes initially | no | no | module decides | yes |
| relative to project dir | derives it | yes | yes | recommended | no |
| surrounding whitespace trimmed | yes | yes in normal config flow | yes | module decides | no |
| lexical cleaning | yes | yes | yes | resolver if used | OS operations |
| symlinks canonicalized | no | no | no | module decides | no |
| leaf symlink rejected | no | yes | no | module decides | final entry replaced |
| parent dirs auto-created | no | final secret dir path | yes | module decides | no |
| file mode enforced | no | `0600` | `0640` | module decides | temp `0600` |
| confinement beneath project | no | no | no | no by default | not applicable |

---

## 17. Secret loading timing

The Fish API key is not loaded during configuration parsing.

The command loads it after:

- CLI validation;
- config loading and validation;
- configured logger initialization;
- module construction;
- text input reading;
- pipeline processing;
- Fish request construction.

Sequence:

```text
process text locally
    ↓
build Fish request
    ↓
load Fish API key
    ↓
construct Fish client
    ↓
synthesize
```

### 17.1 Operational consequence

An invalid or missing API key file may be discovered after local text modules have already run.

A module may have performed remote or stateful work before the Fish secret failure appears.

### 17.2 Logging consequence

Secret failures occur after configured logging is active.

They are written to:

- stderr;
- persistent log file.

### 17.3 No early secret preflight

The current CLI does not provide a secret-only preflight command.

---

## 18. Secret path contract

The lower-level secret loader requires its path argument to:

- be nonblank;
- have no surrounding whitespace;
- name a leaf file rather than `.` or filesystem root;
- have a directory path that can be created or opened.

The normal CLI supplies an already absolute, cleaned path.

### 18.1 No shell expansion

Secret paths from JSON do not expand:

```text
~
$HOME
${XDG_CONFIG_HOME}
```

### 18.2 Filename

The loader operates on:

```text
directory = filepath.Dir(cleanPath)
name = filepath.Base(cleanPath)
```

The leaf filename is opened relative to an anchored directory handle.

### 18.3 Path escape has already happened

The loader does not attempt to confine the absolute path to the project directory.

Any `..` or absolute-path decision was resolved earlier.

---

## 19. Secret directory creation

The loader runs conceptually:

```go
os.MkdirAll(directory, 0o700)
```

### 19.1 Missing directory

Missing directory components are requested with mode:

```text
0700
```

The operating-system umask may remove bits and make them more restrictive.

### 19.2 Existing directory

Existing directory permissions are not automatically changed.

Instead, the final secret directory is inspected.

### 19.3 Final directory requirement

The opened final directory must:

- actually be a directory;
- not be writable by group;
- not be writable by others.

Rejected mode bits:

```text
0020
0002
```

Combined check:

```text
mode & 0022 != 0
```

### 19.4 Allowed examples

Subject to ownership and access:

```text
0700
0750
0755
0710
```

Group or other read/execute is not rejected by this specific check.

### 19.5 Rejected examples

```text
0720
0702
0770
0777
```

Any group-write or other-write bit causes failure.

### 19.6 Existing directory not tightened

A final directory at `0755` is accepted and remains `0755`.

The loader does not force it to `0700`.

Operators wanting owner-only directory visibility should set `0700` themselves.

---

## 20. Directory security boundary

The loader opens the final secret directory through an anchored root handle.

Leaf operations then occur relative to that root.

This reduces path-replacement and traversal risk for the secret filename.

### 20.1 What is checked

The loader checks the opened final directory’s:

- directory type;
- group/other write bits.

### 20.2 What is not checked

It does not explicitly verify:

- directory owner;
- every ancestor directory’s mode;
- every ancestor directory’s owner;
- absence of symlink components in the directory path;
- mount properties;
- filesystem encryption;
- network-filesystem trust.

### 20.3 Deployment rule

Use a trusted absolute directory owned and controlled by the service account or administrator.

The loader is hardened, not omniscient. Filesystems remain a group project humanity started without a product manager.

---

## 21. Missing secret file

When the leaf file does not exist, the loader attempts an exclusive create:

```text
read-only descriptor
create
exclusive
mode 0600
```

The exclusive create prevents silently opening a file that appeared between absence detection and creation.

### 21.1 File mode

After creation, the loader explicitly applies:

```text
0600
```

### 21.2 Empty file

The new file is empty.

The loader does not write a placeholder, newline, or template key.

### 21.3 Reported result

Creation returns an error wrapping:

```text
ErrFileCreated
```

The value result is empty.

The CLI logs:

```text
empty secret file created
```

and exits with status:

```text
3
```

### 21.4 Required operator action

Write exactly one Fish API key line into the created file, then run the command again.

---

## 22. Existing secret file

For an existing leaf path, the loader:

1. performs `Lstat`;
2. requires a regular file;
3. opens it relative to the directory root;
4. obtains metadata from the opened descriptor;
5. performs another `Lstat`;
6. requires the current leaf still to be regular;
7. requires both metadata objects to identify the same file;
8. applies mode `0600`;
9. reads through the already opened descriptor.

### 22.1 Purpose

The sequence detects replacement between inspection and open.

### 22.2 Descriptor stability

After a successful open, reading uses the opened file descriptor.

A later directory-entry replacement does not redirect that descriptor to another file.

### 22.3 Permission tightening

An existing regular file such as:

```text
0644
0660
0666
```

is changed to:

```text
0600
```

when permitted.

### 22.4 Ownership

The loader does not change file owner or group.

The process must have permission to open and chmod the file.

---

## 23. Rejected secret leaf types

The existing secret leaf must be a regular file.

Rejected leaf types include:

- symbolic link;
- directory;
- FIFO;
- socket;
- block device;
- character device;
- other non-regular filesystem objects.

### 23.1 Symbolic link

A symlink at the secret leaf is rejected.

The loader does not follow it to read the target.

### 23.2 Hard links

A hard-linked regular file still appears as a regular file.

The loader does not inspect or reject link count.

Protect the directory and underlying file ownership accordingly.

### 23.3 Configuration-file contrast

A symlinked configuration file is accepted by the ordinary configuration loader.

A symlinked secret leaf is rejected by the secret loader.

---

## 24. Secret file race detection

The existing-file path is inspected before and after open.

The loader compares:

```text
opened descriptor metadata
current directory-entry metadata
```

using same-file identity.

Failure produces an error conceptually stating:

```text
secret file changed while it was being opened
```

### 24.1 Detected case

An attacker or concurrent process replaces the leaf between initial inspection and completed open/reinspection.

### 24.2 Remaining trust requirements

Race checking does not remove the need for:

- non-writable secret directory;
- trusted ownership;
- secure deployment;
- least-privilege process identity.

---

## 25. Secret file size limit

Configuration:

```text
secrets.maxBytes
```

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

### 25.1 Byte count

The limit counts file bytes before newline removal.

Example:

```text
key bytes + trailing LF
```

must together fit within the limit.

### 25.2 Exact boundary

A file exactly at the configured maximum can be read.

A file above it is rejected with a bounded-read error.

### 25.3 Why bounded

The limit protects startup from:

- accidental large files;
- wrong-path selection;
- pipes or generated content exposed as files;
- unbounded memory use;
- oversized diagnostic data.

The regular-file check already rejects FIFOs and devices, while the byte bound limits ordinary files.

---

## 26. Secret text encoding

Secret bytes must be valid UTF-8.

Invalid byte sequences are rejected.

### 26.1 No alternate encoding

The loader does not convert:

- UTF-16;
- Latin-1;
- locale-specific encodings;
- binary token formats.

### 26.2 Byte-order mark

A UTF-8 byte-order mark is valid UTF-8 and is not specially removed.

For a Fish API key, it becomes part of the value and will normally cause authentication failure or later header-value concerns.

Do not include a BOM.

---

## 27. Accepted line endings

The loader accepts one optional final line ending.

Accepted forms:

```text
secret
secret\n
secret\r\n
```

The final LF or CRLF is removed before returning the value.

### 27.1 Editor compatibility

A normal text editor that saves one line with a final Unix or Windows newline is supported.

### 27.2 No repeated final lines

This is rejected:

```text
secret\n\n
```

Only one final line ending is removed.

The remaining newline makes the value multiline.

### 27.3 Bare carriage return

A final bare:

```text
\r
```

is not treated as the accepted line ending.

It remains in the value and triggers the single-line check.

---

## 28. Single-line requirement

After removing one optional final LF or CRLF, the value must contain no:

```text
\r
\n
```

Examples rejected:

```text
first\nsecond
first\rsecond
first\r\nsecond
```

The error reports that the secret must contain exactly one line.

### 28.1 Empty second line

A file containing:

```text
secret\n\n
```

is still multiple-line input and is rejected.

---

## 29. Empty secret

The secret is empty when, after removing one optional final line ending:

```text
strings.TrimSpace(value) == ""
```

Rejected examples include:

```text
empty file
\n
\r\n
spaces only
tabs only
Unicode whitespace only
```

The loader exposes the stable category:

```text
ErrEmpty
```

---

## 30. Surrounding whitespace

After optional final newline removal, surrounding Unicode whitespace is rejected.

Rejected:

```text
<leading-space>secret
secret<trailing-space>
<tab>secret
secret<tab>
```

The loader does not trim the secret and continue.

It fails so that accidental formatting cannot silently change authentication behavior.

### 30.1 Interior whitespace

The generic secret loader does not reject all interior whitespace.

For example, an interior ordinary space is not inherently invalid:

```text
part one
```

Whether such a value is meaningful belongs to the consumer.

### 30.2 Fish API key tightening

The Fish client additionally rejects ASCII control characters anywhere in the API key.

Therefore an interior tab, NUL, or other ASCII control that might pass generic one-line normalization is rejected before HTTP client construction.

---

## 31. Secret normalization order

The normalization sequence is:

```text
validate UTF-8
    ↓
convert bytes to string
    ↓
remove one final CRLF or LF
    ↓
reject blank-after-trim
    ↓
reject remaining CR or LF
    ↓
reject surrounding whitespace
    ↓
return exact remaining string
```

The returned value is not otherwise transformed.

It is not:

- lowercased;
- uppercased;
- Unicode-normalized;
- decoded from base64;
- JSON-decoded;
- quote-stripped.

---

## 32. Quotation marks

Do not wrap the key in quotes inside the file.

Incorrect:

```text
"actual-api-key"
```

The quotation marks become part of the secret.

Correct:

```text
actual-api-key
```

The secret file is plain text containing the value, not JSON.

---

## 33. Secret memory handling

The loader reads into a byte slice.

After normalization, it schedules clearing of that byte slice.

### 33.1 What is cleared

The original mutable byte buffer is overwritten through Go’s `clear`.

### 33.2 What remains

The returned Go string remains in memory.

Additional copies may exist due to:

- string conversion;
- client construction;
- runtime implementation;
- logging of errors that do not include the secret;
- compiler and garbage collector behavior.

### 33.3 No secure-memory guarantee

The project does not provide:

- locked memory;
- guaranteed zeroization of every copy;
- swap prevention;
- core-dump filtering;
- hardware-backed secret storage.

Clearing the byte buffer is a hygiene measure, not a cryptographic erasure guarantee.

---

## 34. Secret file and directory close errors

The loader closes:

- secret file;
- anchored secret directory handle.

Close failures are joined with any primary error.

### 34.1 Successful read can become failure

A value can be read successfully, but a later close error can still make `Load` return an error.

### 34.2 Missing-file case

A created-file result can be joined with a close failure while still preserving `ErrFileCreated` for `errors.Is`.

### 34.3 Error inspection

Callers should use error categories rather than parse the formatted string.

---

## 35. Stable secret error categories

The package exposes:

```text
ErrFileCreated
ErrEmpty
```

### `ErrFileCreated`

Means:

- the configured path was missing;
- an empty file was created securely;
- the caller must populate it;
- synthesis must not continue.

### `ErrEmpty`

Means:

- the file exists;
- no usable non-whitespace secret value is present.

Other errors are descriptive wrapped failures for:

- paths;
- permissions;
- file type;
- races;
- reads;
- encoding;
- line structure;
- close operations.

---

## 36. CLI handling of secret errors

### Missing file created

Log severity:

```text
WARN
```

Message:

```text
empty secret file created
```

Exit status:

```text
3
```

### Other load failure

Log severity:

```text
ERROR
```

Message:

```text
Fish API key loading failed
```

Exit status:

```text
3
```

### Logged metadata

The command may log:

- path;
- error;
- remediation action.

It does not intentionally log the secret value.

---

## 37. Fish client validation after load

After a successful secret load, Fish client construction requires the API key to be:

- nonblank;
- without surrounding whitespace;
- valid UTF-8;
- free of ASCII control characters.

The secret loader and Fish client therefore form two layers:

```text
filesystem and generic one-line secret validation
    ↓
HTTP-header-specific API key validation
```

### 37.1 Why both exist

A generic secret value and an HTTP header have different safety requirements.

The loader should not assume every future secret is a header.

The Fish client must protect its own protocol boundary.

---

## 38. Default local layout

A conventional repository or installation layout is:

```text
fish-audio-cli/
├── bin/
│   └── fish-audio-cli
├── config/
│   ├── config.example.json
│   └── config.json
├── logs/
│   └── fish-audio-cli.log
└── secrets/
    └── fish-api-key
```

With:

```text
--config config/config.json
```

the project directory is the repository or installation root.

Defaults resolve to:

```text
secret:
    <project>/secrets/fish-api-key

log:
    <project>/logs/fish-audio-cli.log
```

A relative output path still resolves from the process working directory.

---

## 39. Recommended Unix permissions

Typical owner-only secret layout:

```bash
install -d -m 0700 /opt/fish-audio-cli/secrets
install -m 0600 /dev/null /opt/fish-audio-cli/secrets/fish-api-key
```

Then write the key without adding extra lines.

Example:

```bash
printf '%s\n' "$FISH_API_KEY" \
  > /opt/fish-audio-cli/secrets/fish-api-key

chmod 0600 \
  /opt/fish-audio-cli/secrets/fish-api-key
```

### 39.1 Shell history caution

Placing the secret directly in a shell command can expose it through:

- shell history;
- terminal scrollback;
- process environment;
- audit logs.

Use an appropriate secret-provisioning mechanism for production.

### 39.2 Service identity

The service account must be able to:

- traverse the secret directory;
- read the file;
- chmod the file to `0600`.

A read-only mounted secret that cannot be chmodded may fail under the current loader.

---

## 40. Read-only secret mounts

Many container and orchestration secret mounts are read-only.

The current loader always attempts to apply:

```text
0600
```

to an existing secret file.

A read-only filesystem or immutable mount may therefore fail even when the file already has acceptable apparent permissions.

### 40.1 Current contract

The loader prioritizes permission enforcement over read-only-mount compatibility.

### 40.2 Deployment options

Use a writable secure copy owned by the service account, or deliberately change the loader contract after design review.

Do not document a read-only mount as supported without testing its chmod behavior on the target platform.

---

## 41. Container layout

A container deployment can use:

```text
/app/config/config.json
/app/logs/fish-audio-cli.log
/run/fish-audio-cli-secrets/fish-api-key
/output/result.opus
```

Configuration:

```json
{
  "secrets": {
    "fishApiKeyFile": "/run/fish-audio-cli-secrets/fish-api-key"
  },
  "logging": {
    "file": "/app/logs/fish-audio-cli.log"
  }
}
```

Invocation:

```bash
/app/bin/fish-audio-cli \
  --config /app/config/config.json \
  --format opus \
  --output /output/result.opus
```

Ensure:

- secret directory is not group/other writable;
- secret file can be chmodded to `0600`;
- log directory is writable;
- output parent exists.

---

## 42. System installation layout

Example:

```text
/opt/fish-audio-cli/
    bin/fish-audio-cli

/etc/fish-audio-cli/
    config.json
    secrets/fish-api-key

/var/log/fish-audio-cli/
    application.log

/var/lib/fish-audio-cli/output/
```

Configuration:

```json
{
  "secrets": {
    "fishApiKeyFile": "secrets/fish-api-key"
  },
  "logging": {
    "file": "/var/log/fish-audio-cli/application.log"
  }
}
```

Because the config is:

```text
/etc/fish-audio-cli/config.json
```

the relative secret becomes:

```text
/etc/fish-audio-cli/secrets/fish-api-key
```

The binary location does not determine the project directory.

---

## 43. Binary location is irrelevant

The resolver does not derive paths from:

- executable path;
- source repository path;
- installation prefix;
- `PATH`;
- Go module root.

Only the configuration path determines the project directory.

This allows one binary to serve multiple isolated configurations.

Example:

```text
same binary:
    /usr/local/bin/fish-audio-cli

config A:
    /srv/tenant-a/config/config.json

config B:
    /srv/tenant-b/config/config.json
```

Each config receives its own project-relative secret and log locations.

---

## 44. Working-directory changes

A relative `--config` is resolved when the resolver is created.

After that, the stored config path and project directory are absolute.

A later working-directory change does not alter them.

The current command does not intentionally change working directory during execution.

Relative `--output` is interpreted when the output subsystem opens its destination and therefore follows the process working directory in effect at that time.

---

## 45. Repository ignore rules

The repository currently ignores:

```text
/bin/
/secrets/
/config/config.json
/logs/
/docs/maintainers/final-audit-anchor.md
```

### 45.1 Default secret protection

The default repository-local path:

```text
secrets/fish-api-key
```

falls under ignored `/secrets/`.

### 45.2 Custom path warning

A custom secret path elsewhere in the repository may not be ignored.

Example:

```json
{
  "secrets": {
    "fishApiKeyFile": "local/fish-key"
  }
}
```

The repository does not automatically ignore:

```text
/local/fish-key
```

Add an appropriate ignore rule before creating or copying the secret.

### 45.3 Absolute path

An absolute secret path outside the repository is not governed by repository `.gitignore`.

### 45.4 Config file

`config/config.json` is ignored because it may contain machine-local operational settings.

The example file remains tracked and must not contain real secrets.

---

## 46. Secret must not be stored in JSON

The current configuration accepts only a secret file path.

There is no field for an inline Fish API key.

Do not add the key as an unknown JSON property.

Strict configuration decoding rejects unknown fields.

### 46.1 No environment fallback

The current application does not read a Fish API key from an environment variable.

### 46.2 No CLI key option

There is no:

```text
--api-key
```

This avoids routine exposure in process listings and command history.

### 46.3 No stdin key protocol

Standard input is reserved for text when `--text` is empty.

It is not a secret-input channel.

---

## 47. Config trust boundary

The configuration file is not secret storage, but it is security-sensitive.

It controls:

- Fish endpoint;
- model;
- reference ID;
- secret location;
- log location;
- module types;
- module-owned settings;
- output-related request parameters.

A malicious config can redirect text and credentials to another endpoint.

### 47.1 Protect config writes

Only trusted administrators or the service owner should modify configuration.

### 47.2 Symlinked config

If using a symlinked config intentionally, protect both:

- symlink path;
- target path.

### 47.3 Mode recommendation

A typical non-secret config can use:

```text
0640
```

with appropriate owner and group.

The application does not enforce this mode.

---

## 48. Log path trust boundary

The configured log path receives operational metadata and possibly text when `logging.logText` is enabled.

The logging subsystem does not reject a symlink leaf.

Use a trusted log directory.

Do not point logging at:

- secret files;
- device nodes;
- FIFOs;
- unrelated privileged files;
- paths writable by untrusted users.

Persistent logging cannot currently be disabled through a special path.

---

## 49. Output path trust boundary

The output path is caller-controlled and may replace an existing destination entry atomically.

The caller must ensure the destination directory is appropriate.

Do not allow untrusted users to choose arbitrary output paths when the process has broader filesystem privileges.

Use:

- a dedicated output directory;
- per-job filenames;
- least-privilege service identity.

---

## 50. Common path mistakes

### Running from the wrong directory

Command:

```bash
fish-audio-cli \
  --config config/config.json \
  --format opus \
  --output speech.opus
```

fails to find config when the working directory is not the installation root.

Use an absolute config path in services.

### Assuming binary-relative paths

Incorrect assumption:

```text
relative paths are beside the executable
```

They are not.

### Assuming config-target-relative symlink behavior

Incorrect assumption:

```text
relative paths follow the symlink target directory
```

They follow the supplied lexical config path.

### Using `~` in JSON

Incorrect:

```json
"fishApiKeyFile": "~/.secrets/fish-key"
```

No home expansion occurs.

### Using environment syntax in JSON

Incorrect:

```json
"fishApiKeyFile": "$HOME/.secrets/fish-key"
```

No environment expansion occurs.

### Making secret directory group-writable

Mode such as:

```text
0770
```

is rejected.

### Supplying a symlink secret leaf

Rejected even when the target is a regular `0600` file.

### Mounting secret read-only

May fail because loader enforces `0600` with chmod.

### Adding a second newline

A file with two lines, including an empty second line, is rejected.

### Quoting the key

Quotes become part of the secret.

### Expecting output directory creation

The output subsystem does not create parents.

---

## 51. Diagnosing resolved paths

The configured logger’s `config loaded` event includes:

```text
path
```

for the absolute config path.

Secret errors include the resolved absolute secret path.

Synthesis start includes the caller-supplied output path.

### 51.1 No project-directory log field

The project directory is not logged as a separate field.

It can be derived from the documented config-path rule.

### 51.2 No secret-path success log

A successful secret load does not log the secret path as a dedicated success event.

### 51.3 Avoid exposing paths unnecessarily

Paths can reveal:

- usernames;
- tenant names;
- directory structure;
- deployment topology.

Protect logs accordingly.

---

## 52. Missing-secret bootstrap workflow

For the default layout:

1. create or select configuration;
2. run the CLI once;
3. loader creates:

```text
secrets/fish-api-key
```

with mode `0600`;
4. command logs remediation;
5. command exits `3`;
6. operator writes exactly one key line;
7. run again.

### 52.1 Directory mode

The missing directory is requested as `0700`.

### 52.2 Existing unsafe directory

If `secrets/` already exists and is group/other writable, creation is refused.

### 52.3 No placeholder content

The created file remains zero bytes until the operator populates it.

---

## 53. Safe provisioning example

From the installation root:

```bash
clear

install -d -m 0700 secrets
install -m 0600 /dev/null secrets/fish-api-key

read -r -s -p 'Fish API key: ' fish_key
printf '\n'

printf '%s\n' "$fish_key" > secrets/fish-api-key
unset fish_key

chmod 0600 secrets/fish-api-key
```

This example:

- avoids placing the key in command history;
- writes one line;
- adds one accepted final LF;
- restores mode `0600`.

Terminal and process security still depend on the local environment.

---

## 54. Validation example

Check type and permissions without printing contents:

```bash
clear

stat -c 'type=%F mode=%a owner=%U group=%G path=%n' \
  secrets \
  secrets/fish-api-key
```

Expected general shape:

```text
type=directory mode=700 ...
type=regular file mode=600 ...
```

The application accepts some non-group-writable directory modes beyond `0700`, but `0700` is the simplest owner-only recommendation.

---

## 55. Avoid content-revealing diagnostics

Do not use commands that print the key into shared logs or chat.

Avoid:

```bash
cat secrets/fish-api-key
set -x
env
```

when their output may be captured.

To inspect line count without content:

```bash
clear

python3 - <<'PY'
from pathlib import Path

path = Path("secrets/fish-api-key")
data = path.read_bytes()

print(f"bytes={len(data)}")
print(f"lf_count={data.count(bytes([10]))}")
print(f"cr_count={data.count(bytes([13]))}")
PY
```

This still reveals length metadata, which may matter in unusually sensitive environments.

---

## 56. Backup policy

Treat the Fish API key file as a credential.

A backup system may copy it unless excluded.

Decide explicitly whether the secret should be:

- backed up encrypted;
- restored from a secret manager;
- excluded and reprovisioned;
- rotated after disaster recovery.

The application does not manage backups or rotation.

---

## 57. Key rotation

To rotate the key:

1. obtain a new key;
2. replace the file contents with one line;
3. preserve mode `0600`;
4. run a new CLI invocation.

Each invocation loads the key fresh.

There is no long-running process cache across CLI invocations.

### 57.1 Atomic secret replacement

The project does not provide a dedicated secret-write command.

An operator can use an appropriately secure temporary file and rename within the same directory.

Ensure the replacement remains a regular file and can be chmodded to `0600`.

### 57.2 Concurrent invocations

An invocation already holding the old key string continues with that value.

A later invocation loads the replacement.

---

## 58. Multiple configurations

One binary can use multiple independent config roots.

Example:

```text
/srv/a/config/config.json
/srv/a/secrets/fish-api-key
/srv/a/logs/fish-audio-cli.log

/srv/b/config/config.json
/srv/b/secrets/fish-api-key
/srv/b/logs/fish-audio-cli.log
```

Invocations:

```bash
fish-audio-cli \
  --config /srv/a/config/config.json \
  --format opus \
  --output /srv/a/output/message.opus
```

```bash
fish-audio-cli \
  --config /srv/b/config/config.json \
  --format opus \
  --output /srv/b/output/message.opus
```

The request IDs and files remain independent.

---

## 59. Testing expectations for project paths

Tests should cover:

- blank config path;
- relative config path to absolute conversion;
- lexical cleaning;
- exact `config` parent rule;
- non-`config` parent rule;
- absolute path resolution;
- relative path resolution;
- blank generic path;
- uninitialized resolver with relative path;
- uninitialized resolver with absolute path;
- `..` lexical escape behavior;
- symlinked config lexical-root behavior where supported;
- no tilde or environment expansion.

---

## 60. Testing expectations for secrets

Tests should cover:

### Creation

- missing directory;
- missing file;
- `ErrFileCreated`;
- file mode `0600`;
- final directory not group/other writable;
- exclusive-create race.

### Existing file

- secure regular file;
- insecure regular file tightened to `0600`;
- symlink rejection;
- directory rejection;
- FIFO and device rejection where practical;
- changed-between-inspection-and-open detection;
- chmod failure;
- open failure;
- file close failure;
- directory close failure.

### Content

- no final newline;
- one final LF;
- one final CRLF;
- empty file;
- whitespace-only;
- leading whitespace;
- trailing whitespace;
- internal newline;
- repeated final newline;
- bare carriage return;
- invalid UTF-8;
- exact byte limit;
- one byte above limit;
- UTF-8 BOM behavior;
- quote preservation.

### Directory

- secure existing directory;
- group-writable rejection;
- other-writable rejection;
- existing read/execute permissions preserved;
- owner mismatch behavior documented rather than assumed.

---

## 61. Review checklist

### Resolver

- Is config path still trimmed and absolutized?
- Is cleaning lexical rather than canonical?
- Is project root still based on exact parent basename `config`?
- Are relative paths still allowed to escape with `..`?
- Are absolute paths still accepted?
- Are tilde and environment expansion still absent?

### Core paths

- Is Fish API key still resolved during config load?
- Is the in-memory key path still absolute?
- Is log path still resolved by logging rather than config loader?
- Is output still independent from project root?
- Are module-owned paths still module-owned?

### Secret directory

- Are missing directories requested as `0700`?
- Is the final directory still checked for group/other write?
- Are existing permissions intentionally not tightened?
- Are owner and ancestor checks unchanged?

### Secret file

- Is exclusive missing-file creation preserved?
- Is mode `0600` enforced?
- Are non-regular leaves rejected?
- Are symlink leaves rejected?
- Is same-file race detection preserved?
- Are close errors joined?

### Secret content

- Is the byte limit applied before normalization?
- Is UTF-8 required?
- Is exactly one optional final LF or CRLF removed?
- Are multiple lines rejected?
- Is surrounding whitespace rejected?
- Is blank content categorized as `ErrEmpty`?
- Is the mutable byte buffer cleared?

### Operations

- Are missing-secret exit and log semantics unchanged?
- Are repository ignore rules accurate?
- Are read-only mount limitations documented?
- Are no unsupported secret sources promised?

---

## 62. Path and secret invariants

The following rules are normative for the current implementation.

1. `--config` defaults to `config/config.json`.
2. Configuration path is trimmed.
3. Blank configuration path is rejected.
4. Relative configuration path is made absolute from current working directory.
5. Paths are cleaned lexically.
6. Symlinks are not canonicalized by the resolver.
7. Project directory starts as the config file’s lexical parent.
8. An exact parent basename `config` moves project directory up one level.
9. Generic resolved paths are trimmed.
10. Blank generic paths are rejected.
11. Absolute generic paths do not require an initialized resolver.
12. Relative generic paths require an initialized resolver.
13. Relative paths are not confined beneath project directory.
14. Tilde expansion is absent.
15. Environment-variable expansion is absent.
16. Configuration file is opened without secret-style symlink hardening.
17. Fish API key path is resolved during config loading.
18. The in-memory Fish key path is absolute.
19. Logging resolves its own path later through the same resolver.
20. Module-owned paths are not resolved recursively by core.
21. Output path does not use project resolver.
22. Secret loading occurs after text processing and request construction.
23. Missing final secret directory components are requested as `0700`.
24. Final secret directory must not be group/other writable.
25. Existing secret directory mode is not tightened.
26. Secret leaf must name a file rather than `.` or root.
27. Missing secret file is created exclusively.
28. Missing secret file is created empty.
29. Secret file mode is enforced as `0600`.
30. Missing-file creation returns `ErrFileCreated`.
31. Existing secret leaf must be regular.
32. Secret leaf symlink is rejected.
33. Existing file is inspected before and after open.
34. Opened and current leaf must identify the same file.
35. Hard-link count is not checked.
36. File owner is not changed or explicitly validated.
37. Secret size is bounded in bytes before normalization.
38. Secret bytes must be valid UTF-8.
39. One final LF or CRLF is accepted and removed.
40. Remaining CR or LF is rejected.
41. Blank-after-trim content returns `ErrEmpty`.
42. Surrounding Unicode whitespace is rejected.
43. The exact remaining string is returned.
44. Original mutable read bytes are cleared.
45. Returned string has no secure-erasure guarantee.
46. File and directory close failures are joined.
47. Fish client performs additional HTTP-header validation.
48. Secret value is not intentionally logged.
49. Default repository secret and log directories are ignored by Git.
50. Custom secret paths require their own repository-hygiene review.

Changing one of these rules is a filesystem or secret-handling compatibility change.

---

## 63. Non-goals

The current path and secret system does not provide:

- filesystem sandboxing;
- chroot;
- path allowlists;
- symlink canonicalization for config;
- universal anti-symlink policy;
- ownership enforcement;
- ancestor-directory permission validation;
- hard-link rejection;
- encrypted secret storage;
- operating-system keyring integration;
- environment-variable secret loading;
- CLI secret input;
- remote secret-manager integration;
- automatic secret rotation;
- secret expiry checks;
- secret format discovery;
- read-only mount compatibility guarantee;
- memory locking;
- complete zeroization;
- project-root discovery from executable location;
- environment or tilde expansion;
- automatic output parent creation.

These require explicit design rather than accidental path magic.

---

## 64. Summary

The project path model is:

```text
config argument
    ↓
trim
    ↓
absolute lexical path
    ↓
derive project directory
    ↓
resolve core and module-owned configured paths deliberately
```

The secret model is:

```text
absolute resolved key path
    ↓
create or open trusted final directory
    ↓
reject group/other-writable directory
    ↓
create exclusive 0600 file or securely open regular leaf
    ↓
detect leaf replacement
    ↓
enforce 0600
    ↓
bounded UTF-8 read
    ↓
remove one optional final LF or CRLF
    ↓
require one nonblank unpadded line
    ↓
Fish header validation
```

The most important operational rules are:

- use absolute `--config` paths in services;
- remember project root follows the lexical config location, not executable location or symlink target;
- keep secret directories non-writable by group and others;
- use a regular, writable-to-owner `0600` secret file;
- do not use a symlink as the secret leaf;
- store exactly one unquoted key line;
- expect read-only mounts to fail when chmod is unavailable;
- protect configuration because it controls credential and endpoint paths;
- treat log, module, and output paths as separate policy domains;
- add ignore rules for any custom repository-local secret path.
