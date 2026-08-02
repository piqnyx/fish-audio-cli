# Output files and atomic publication

> **Document status:** normative description of the current pre-release output-file behavior.
>
> **Audience:** CLI users, service integrators, operators choosing output directories, and maintainers changing atomic publication or Fish response streaming.
>
> **Scope:** this document describes output-path interpretation, temporary-file creation, permissions, streaming, synchronization, rename publication, directory synchronization, cleanup, existing destinations, symlinks, failure states, cancellation, concurrency, durability, security boundaries, and compatibility constraints. Command syntax is documented in [`cli.md`](cli.md); path domains in [`secrets-and-paths.md`](secrets-and-paths.md); Fish response behavior in [`fish-audio.md`](fish-audio.md); logging in [`logging.md`](logging.md); architecture ownership in [`architecture.md`](architecture.md).

---

## 1. Purpose

`fish-audio-cli` publishes synthesized audio through a temporary file rather than writing directly to the final destination.

The normal flow is:

```text
selected output path
    ↓
create unique temporary file beside destination
    ↓
stream Fish response into temporary file
    ↓
sync temporary file
    ↓
close temporary file
    ↓
rename temporary file to destination
    ↓
sync and close containing directory
    ↓
report success
```

This design protects the final destination from partially written synthesis data.

It separates:

1. **temporary writing**, where partial bytes are allowed;
2. **publication**, where one rename changes the destination entry;
3. **persistence confirmation**, where the containing directory is synchronized.

These are related but distinct guarantees.

---

## 2. Ownership boundary

The output package owns:

- creating a temporary file;
- choosing the temporary filename pattern;
- streaming through a supplied writer callback;
- synchronizing and closing the temporary file;
- renaming it to the destination;
- synchronizing and closing the containing directory;
- cleaning up unpublished temporary files;
- preserving cleanup errors.

The output package does not own:

- CLI format selection;
- audio encoding;
- Fish HTTP behavior;
- retry policy;
- destination extension validation;
- output directory creation;
- path sandboxing;
- file locking;
- concurrent-job coordination;
- backup creation;
- stale-temp scavenging.

The application composes the Fish client inside the output writer callback.

Conceptually:

```go
output.WriteAtomic(outputPath, func(writer io.Writer) error {
    return fishClient.Synthesize(ctx, request, writer)
})
```

---

## 3. CLI output option

The destination is supplied through:

```text
--output PATH
```

The CLI requires a non-empty option value.

Example:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output speech.opus
```

### 3.1 No default destination

The application does not invent:

- a filename;
- a timestamped output;
- a format-derived extension;
- a temporary final destination.

The caller must choose the path.

### 3.2 Audio is not stdout

The application does not stream audio to standard output.

The final audio is written only through `--output`.

### 3.3 Extension does not choose format

This path:

```text
speech.wav
```

does not cause WAV synthesis.

The selected format comes only from:

```text
--format
```

A caller can deliberately or accidentally create MP3 bytes in a `.wav` file.

The output package does not inspect audio content or extension.

---

## 4. Output path interpretation

The output path does not use the project path resolver.

A relative path is interpreted relative to the process working directory.

Example:

```text
working directory:
    /tmp/job-17

--config:
    /opt/fish-audio-cli/config/config.json

--output:
    speech.opus
```

The final destination is:

```text
/tmp/job-17/speech.opus
```

It is not:

```text
/opt/fish-audio-cli/speech.opus
```

### 4.1 No absolute conversion

The output package does not convert the destination to an absolute path before use.

### 4.2 No project rebasing

The output path is independent from:

- configuration location;
- binary location;
- log path;
- secret path.

### 4.3 Working-directory stability

The current command does not intentionally change working directory during execution.

A wrapper should still set an explicit working directory or pass an absolute output path.

---

## 5. Empty and whitespace paths

The output function rejects only an exact empty string:

```text
""
```

It does not trim the path.

A whitespace-only path is not rejected by the initial argument check.

Example:

```text
"   "
```

may be treated as a literal filename if the filesystem accepts it.

Do not use:

- blank-looking names;
- leading whitespace;
- trailing whitespace;
- control characters;
- ambiguous Unicode filenames.

The supported operational contract is a normal explicit path.

---

## 6. Parent directory

The output package does not create the parent directory.

Example:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output missing/directory/speech.opus
```

fails when:

```text
missing/directory
```

does not exist.

### 6.1 Create parents explicitly

Example:

```bash
mkdir -p /var/lib/voice-worker/output

fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output /var/lib/voice-worker/output/speech.opus
```

### 6.2 Parent requirements

The process needs sufficient permission to:

- traverse the directory;
- create the temporary file;
- write and sync the temporary file;
- rename within the directory;
- open the directory for synchronization;
- synchronize and close the directory.

A directory that permits file creation but cannot be opened or synchronized can produce a post-publication failure.

---

## 7. Destination decomposition

The output function derives:

```text
directory = filepath.Dir(path)
baseName = filepath.Base(path)
```

For:

```text
speech.opus
```

the directory is normally:

```text
.
```

and the base name is:

```text
speech.opus
```

For:

```text
/var/lib/audio/speech.opus
```

the directory is:

```text
/var/lib/audio
```

and the base name is:

```text
speech.opus
```

These values are used to create the temporary file beside the final destination.

---

## 8. Temporary filename

The temporary file pattern is:

```text
.<baseName>.*.tmp
```

For destination:

```text
speech.opus
```

a possible temporary name is:

```text
.speech.opus.184739201.tmp
```

The random portion is selected by the operating-system temporary-file helper.

### 8.1 Same directory

The temporary file is created in the destination directory.

This is essential because rename-based publication should remain within one filesystem and directory namespace.

### 8.2 Hidden-name convention

On Unix-like systems, the leading dot normally makes the temporary file hidden from ordinary directory listings.

It is still a normal file.

### 8.3 Name disclosure

The temporary filename contains the destination base name.

A directory observer can infer the intended output name while synthesis is in progress.

### 8.4 Unique creation

The temporary-file helper creates a new unique file rather than reusing a predictable fixed name.

This avoids ordinary collisions among concurrent invocations.

---

## 9. Temporary file permissions

The temporary file is created with restrictive permissions:

```text
0600
```

The current tests require the published output to have:

```text
0600
```

because the final destination is the renamed temporary file.

### 9.1 Resulting access

Mode `0600` grants:

- owner read;
- owner write;
- no group permissions;
- no permissions for others.

### 9.2 Existing destination mode is not preserved

If an existing destination has mode:

```text
0644
```

a successful replacement becomes:

```text
0600
```

The output package does not copy the old mode.

### 9.3 No explicit post-rename chmod

The mode comes from the newly created temporary file.

The package does not reopen and chmod the final path after rename.

### 9.4 Ownership and group

The replacement is a newly created file.

Its owner and group follow operating-system and filesystem creation rules for the process and directory.

The package does not preserve the previous destination’s:

- owner;
- group;
- ACL;
- extended attributes;
- timestamps;
- security labels;
- hard-link relationships.

Deployment policy must account for the new-file model.

---

## 10. Writer callback

`WriteAtomic` receives a function with the conceptual type:

```go
func(io.Writer) error
```

The callback writes the complete output payload.

### 10.1 Public contract

The callback should:

- write bytes;
- return any write or generation error;
- stop when its upstream operation fails;
- leave close, sync, rename, and cleanup to the output package.

### 10.2 Do not close the writer

The callback must not close the supplied writer.

The output package owns its lifecycle.

Closing it inside the callback can cause:

- later sync failure;
- later close failure;
- cleanup close errors;
- incomplete publication.

### 10.3 Do not rename or remove the temporary file

The callback receives only an `io.Writer` contract.

It must not depend on the concrete writer being an `*os.File`.

It must not:

- inspect the temporary filename;
- rename the temporary file;
- remove it;
- chmod it;
- seek unless explicitly redesigned;
- publish it independently.

### 10.4 Error responsibility

A callback that encounters an error must return it.

If it silently ignores a short or failed write and returns `nil`, the output package cannot reconstruct the lost failure.

### 10.5 Empty success

The generic output package does not count bytes.

A callback that writes zero bytes and returns `nil` can publish an empty file.

The CLI’s Fish client separately rejects a successful HTTP response that yields zero audio bytes, so ordinary CLI synthesis does not publish an empty Fish response.

---

## 11. Fish response streaming

The Fish client copies a successful response body directly into the temporary file writer.

The complete audio is not accumulated in application memory first.

Flow:

```text
Fish 2xx response body
    ↓
io.Copy
    ↓
temporary output file
```

### 11.1 Partial temporary data

A response or disk error can leave partial audio in the temporary file.

That is expected during the unpublished phase.

### 11.2 Final-path protection

When streaming fails before rename:

- the callback returns an error;
- the output package does not rename the temporary file;
- cleanup removes the temporary file;
- an existing final destination remains unchanged.

### 11.3 No Fish retry after stream failure

The Fish client does not retry after successful-response streaming begins.

A retry could mix or duplicate bytes in a non-rewindable writer.

The output layer therefore receives the stream error and cleans up.

---

## 12. Publication stages

The output operation has six relevant stages:

```text
1. create temp
2. write temp
3. sync temp
4. close temp
5. rename temp to destination
6. sync and close destination directory
```

The rename boundary divides unpublished and published states.

---

## 13. Stage 1: create temporary file

Creation uses the destination directory and temporary-name pattern.

Failure examples:

- parent directory missing;
- permission denied;
- read-only filesystem;
- invalid path;
- quota preventing file creation;
- too many open files;
- unsupported filename.

Error context begins with:

```text
create temporary output file
```

No final destination change has occurred.

---

## 14. Stage 2: write temporary file

The callback writes into the temporary file.

Failure examples:

- Fish API error;
- context cancellation;
- response read failure;
- disk full;
- quota exceeded;
- local filesystem write failure;
- callback-defined error.

The primary error is wrapped with:

```text
write temporary output file
```

The final destination is not changed.

Cleanup then attempts to close and remove the temporary file.

---

## 15. Stage 3: synchronize temporary file

After the callback returns success, the package calls:

```text
tempFile.Sync()
```

This requests flushing file content and relevant file metadata before publication.

Failure context:

```text
sync temporary output file
```

### 15.1 Before rename

A sync failure occurs while the temporary file is still unpublished.

The existing destination remains unchanged.

### 15.2 Cleanup

The deferred cleanup closes and removes the temporary file.

### 15.3 Durability meaning

A successful file sync improves crash durability of the temporary file’s contents.

It does not yet persist the destination-name replacement because rename has not happened.

---

## 16. Stage 4: close temporary file

After successful sync, the package closes the temporary file.

Failure context:

```text
close temporary output file
```

### 16.1 Close failure aborts publication

A close error prevents rename.

The destination remains unchanged.

### 16.2 Removal still attempted

The cleanup path attempts to remove the temporary pathname.

### 16.3 No second normal close

After the explicit close call, the output package marks the temporary file as closed even when `Close` reports an error.

The cleanup path does not issue another normal close attempt.

It still attempts removal.

---

## 17. Stage 5: rename publication

After successful temp sync and close, the package calls conceptually:

```go
os.Rename(tempPath, destinationPath)
```

This is the publication boundary.

### 17.1 Same-directory rename

The temporary file and destination are in the same directory.

This avoids an ordinary cross-filesystem rename.

### 17.2 Atomic visibility

On filesystems and operating systems that provide atomic same-directory rename semantics, observers see either:

- the old destination entry;
- the new destination entry.

They do not see a partially overwritten final file.

### 17.3 Platform qualification

Atomic replacement details are governed by the host operating system and filesystem.

The project tests the expected replacement behavior on supported development platforms, but unusual network filesystems or platform-specific rename rules can differ.

### 17.4 Rename failure

Failure context:

```text
replace output file
```

Before a successful rename:

- `published` remains false;
- existing destination remains;
- temporary-file cleanup runs.

---

## 18. Stage 6: directory synchronization

After rename succeeds, the output is considered published.

The package then:

1. opens the containing directory;
2. calls directory `Sync`;
3. closes the directory;
4. preserves both sync and close failures.

Failure is wrapped with:

```text
persist output replacement
```

### 18.1 Why sync the directory

The rename changes directory metadata.

Synchronizing only the file does not necessarily persist the directory entry update across a crash.

### 18.2 Open failure

If the directory cannot be opened after rename, publication has already happened.

The function returns an error, but the new output remains at the destination.

### 18.3 Sync failure

If directory sync fails:

- the function returns an error;
- the output remains published;
- durability of the namespace update is not confirmed.

### 18.4 Close failure

A directory close failure is also returned.

The output remains published.

### 18.5 Combined failure

If both directory sync and close fail, both errors are joined.

---

## 19. Success definition

`WriteAtomic` returns `nil` only after:

- temporary creation succeeded;
- callback succeeded;
- temporary sync succeeded;
- temporary close succeeded;
- rename succeeded;
- directory open succeeded;
- directory sync succeeded;
- directory close succeeded.

A CLI synthesis returns status `0` only after the complete output operation succeeds.

---

## 20. Published-but-error state

The most important non-obvious state is:

```text
rename succeeded
directory persistence step failed
```

Result:

- the new output file exists;
- the old destination has already been replaced;
- the function returns an error;
- the CLI logs `synthesis failed`;
- the CLI exits with status `4`;
- the output package does not remove the new file.

### 20.1 Why the file is retained

After rename, deleting the new output would not restore the old destination.

Removal would instead leave no destination.

The package therefore preserves the published file and reports the persistence failure honestly.

### 20.2 Automation rule

On nonzero exit, do not assume the output path is absent.

Inspect the error and filesystem state when recovery policy needs to distinguish:

- no publication;
- publication with unconfirmed directory durability.

### 20.3 Success rule

On exit `0`, the complete publication and directory-sync sequence succeeded.

---

## 21. Existing destination before rename

Before rename succeeds, an existing destination is not opened or truncated.

The new bytes go into a separate file.

If any earlier stage fails:

- old content remains;
- old mode remains;
- old metadata remains;
- temporary file is removed when cleanup succeeds.

This is safer than direct truncating writes.

---

## 22. Existing regular file on success

A successful rename replaces the destination directory entry with the newly created temporary file.

Consequences:

- content becomes new audio;
- mode becomes `0600`;
- previous inode metadata is not preserved;
- previous hard links continue referring to the old inode;
- processes with an already open old descriptor may continue observing the old object according to operating-system semantics.

The package performs replacement, not in-place mutation.

---

## 23. Destination symlink

The final destination path is not opened for writing.

A successful rename replaces the symlink directory entry itself.

Example:

```text
speech.opus -> protected-target.opus
```

After successful output:

```text
speech.opus
```

is a regular `0600` file containing new audio.

The original:

```text
protected-target.opus
```

remains unchanged.

### 23.1 Leaf behavior

This protects against following a symlink at the destination leaf.

### 23.2 Parent components

Symlinked parent-directory components are still resolved by normal operating-system path handling.

The output package does not use the anchored anti-race directory handling used by the secret loader.

### 23.3 Trust boundary

Use a trusted output directory that untrusted users cannot replace or redirect.

---

## 24. Destination directory

If the destination path already names a directory, rename normally fails rather than replacing that directory with the file.

The existing directory remains.

The unpublished temporary file is cleaned up.

The error appears under:

```text
replace output file
```

---

## 25. Other destination leaf types

Because publication uses rename rather than opening the final leaf, an existing non-directory entry may be replaced according to host rename semantics.

Possible leaf types include:

- regular file;
- symlink;
- FIFO;
- socket;
- device node.

The package does not prevalidate the destination leaf type.

Do not point output at special files.

Use an ordinary path inside a dedicated output directory.

---

## 26. Temporary cleanup

A deferred cleanup runs whenever publication has not succeeded.

It can perform:

1. temporary-file close when not already closed;
2. temporary-path removal;
3. joining cleanup errors with the primary failure.

### 26.1 Missing temporary path

Removal treats:

```text
file does not exist
```

as already cleaned.

### 26.2 Close during cleanup

A cleanup close failure receives context:

```text
close temporary output file during cleanup
```

### 26.3 Remove failure

A removal failure receives context:

```text
remove temporary output file
```

### 26.4 Primary error preservation

Cleanup failure does not replace the original synthesis or output error.

Errors are joined so callers can inspect both.

### 26.5 No cleanup after publication

Once rename succeeds, the temporary pathname no longer represents an unpublished file.

The cleanup defer returns without removing anything.

---

## 27. Cleanup failure consequences

When temporary removal fails before publication:

- the final destination remains unchanged;
- a hidden temp file may remain;
- the returned error contains both primary and cleanup information.

Possible causes:

- directory permission changed;
- filesystem error;
- antivirus or external process interference;
- callback closed or manipulated the file unexpectedly;
- operating-system-specific file-locking behavior.

Operators should inspect the output directory after cleanup-related errors.

---

## 28. Abnormal process termination

Deferred cleanup runs only while Go control flow unwinds normally through the function.

It does not run after:

- `SIGKILL`;
- abrupt process termination;
- machine power loss;
- kernel panic;
- runtime termination that bypasses defers.

A stale file matching:

```text
.<baseName>.*.tmp
```

may remain.

### 28.1 Managed cancellation

The CLI handles `SIGINT` and `SIGTERM` through context cancellation during synthesis.

That normally causes the callback to return an error, allowing deferred cleanup to run.

### 28.2 No built-in scavenger

The application does not scan for or delete stale temp files on startup.

### 28.3 Manual cleanup

Remove stale temp files only after confirming no active process is using them.

Example pattern:

```text
.speech.opus.*.tmp
```

A broad command deleting every hidden `.tmp` file in a shared directory is not an acceptable maintenance policy, despite humanity’s affection for wildcard-based regret.

---

## 29. Cancellation

During Fish synthesis, the signal-aware context can be canceled by:

- `SIGINT`;
- `SIGTERM`;
- caller-provided cancellation in package use.

The Fish client returns an error.

The output package then:

- treats the callback as failed;
- avoids rename;
- closes and removes the temporary file;
- preserves an existing destination.

The CLI exits with status:

```text
4
```

Cancellation before output creation fails earlier in another command stage.

---

## 30. Disk-full behavior

Disk exhaustion can appear during:

- temporary-file creation;
- response streaming;
- temporary-file sync;
- directory metadata persistence.

### 30.1 During write

Partial temp bytes are removed when cleanup succeeds.

The old destination remains.

### 30.2 During file sync

The old destination remains because rename has not happened.

### 30.3 During directory sync

The new destination is already published.

The command reports failure and retains it.

### 30.4 Space requirement during replacement

Until rename, the filesystem may need space for both:

- existing destination;
- complete new temporary output.

Large replacements therefore require enough free space for the new audio in addition to the old file.

---

## 31. Read-only filesystem

A read-only destination filesystem can fail at temporary creation or another write stage.

No direct write to the final destination is attempted first.

An existing destination remains unchanged unless the filesystem becomes read-only only after a successful rename.

---

## 32. Quotas

User, group, project, or filesystem quotas can affect:

- temp creation;
- streaming;
- sync;
- directory metadata update.

Quota behavior is surfaced as wrapped filesystem errors.

The output package does not:

- preflight free space;
- estimate output size;
- reserve quota;
- retry quota failures.

---

## 33. Concurrent invocations

The output package does not lock the destination.

Two invocations targeting the same path can run concurrently.

Each creates its own unique temporary file.

### 33.1 Possible outcome

Both can synthesize successfully, then rename in sequence.

The last successful rename wins.

### 33.2 No lost-update detection

The package does not compare:

- prior destination inode;
- modification time;
- request ID;
- content hash;
- expected generation.

### 33.3 Failure meaning under concurrency

A failed invocation preserves whatever destination exists at its failure time.

That destination may have been written by another concurrent invocation rather than being the original file.

### 33.4 Operational rule

Use unique output paths for parallel jobs.

Example:

```text
output/<request-id>.opus
```

---

## 34. Readers during replacement

Because the new output is written separately and renamed:

- path-based readers opening before rename see the old entry;
- path-based readers opening after rename see the new entry;
- already open file descriptors follow host filesystem semantics for the object they opened.

The output package does not coordinate readers.

Consumers should open the file after successful command completion when they require the new audio.

---

## 35. Watchers and indexing services

Directory watchers may observe:

- temporary-file creation;
- temporary-file writes;
- temporary-file removal;
- final rename;
- replacement of an existing entry.

A watcher should not treat:

```text
.<baseName>.*.tmp
```

as a completed output.

Use the final path and successful process status as completion signals.

---

## 36. Atomicity versus durability

These terms must not be conflated.

### Atomic visibility

The rename can make the destination switch from old entry to new entry as one namespace operation where the filesystem supports it.

### File-content durability

Temporary-file `Sync` requests persistence of the new file’s bytes before rename.

### Namespace durability

Directory `Sync` requests persistence of the rename itself.

### Application success

The function reports success only when all required stages return successfully.

### 36.1 What is not guaranteed

The package cannot guarantee against:

- broken storage hardware;
- lying device caches;
- filesystem implementation defects;
- unsupported directory sync;
- network filesystem semantics;
- platform-specific rename behavior;
- catastrophic system failure beyond operating-system guarantees.

---

## 37. Crash windows

### Before temporary sync

Possible state after crash:

- old destination remains;
- partial temporary file may remain.

### After temporary sync but before rename

Possible state:

- old destination remains;
- complete temporary file may remain.

### After rename but before directory sync

Possible state:

- new destination may be visible before crash;
- persistence of the rename after reboot is not confirmed.

### After successful directory sync

The application has completed its requested durability sequence.

Hardware and filesystem guarantees still define ultimate persistence.

---

## 38. Error phases

| Error context | Publication state | Existing destination |
|---|---|---|
| `output path is empty` | not started | unchanged |
| `write function is nil` | not started | unchanged |
| `create temporary output file` | not started | unchanged |
| `write temporary output file` | unpublished | unchanged |
| `sync temporary output file` | unpublished | unchanged |
| `close temporary output file` | unpublished | unchanged |
| `replace output file` | unpublished | unchanged |
| `persist output replacement` | published | replaced |
| cleanup close/remove error | unpublished | unchanged |

### 38.1 Joined errors

An error can include both:

- primary stage failure;
- cleanup failure.

Use Go error inspection rather than parsing only the final string in package integrations.

---

## 39. CLI exit status

Any error returned by the atomic output operation is handled as:

```text
synthesis failed
```

The CLI exits with:

```text
4
```

This category includes:

- Fish API errors;
- response streaming errors;
- temp creation errors;
- disk write errors;
- file sync errors;
- temp close errors;
- rename errors;
- directory sync errors;
- cleanup errors.

### 39.1 Exit `4` does not imply absence

A directory persistence error happens after publication.

The file can exist despite exit `4`.

### 39.2 Exit `0`

Exit `0` means the output function completed all stages successfully.

---

## 40. Logging

Before output starts, the application logs:

```text
synthesis started
```

with:

```text
model
format
output_path
request_id
```

On failure:

```text
synthesis failed
```

with:

```text
error
request_id
```

On success:

```text
synthesis completed
```

with:

```text
output_path
request_id
```

### 40.1 Temporary path not logged

The output package does not expose or log the random temporary path during normal operation.

Cleanup errors can include it inside the wrapped error text.

### 40.2 Output path privacy

Output paths may expose:

- usernames;
- tenant identifiers;
- message identifiers;
- directory structure.

Protect logs accordingly.

---

## 41. Format and filename examples

Recommended pairs:

| CLI format | Typical filename |
|---|---|
| `wav` | `speech.wav` |
| `mp3` | `speech.mp3` |
| `opus` | `speech.opus` |
| `ogg` | `speech.ogg` |

`ogg` is normalized to an Opus Fish request.

The output package itself does not know these formats.

It writes whatever bytes the callback supplies.

---

## 42. Safe service invocation

```bash
clear

output_dir="/var/lib/voice-worker/output"
output_path="$output_dir/message-123.opus"

install -d -m 0750 "$output_dir"

fish-audio-cli \
  --config /opt/fish-audio-cli/config/config.json \
  --text "Привет" \
  --format opus \
  --output "$output_path"

status=$?

if [ "$status" -ne 0 ]; then
  printf 'synthesis failed with status %d\n' "$status" >&2
  exit "$status"
fi

printf 'published: %s\n' "$output_path"
```

The wrapper creates the parent directory.

The CLI creates and publishes only the file.

---

## 43. Unique-path example

```bash
clear

request_id="$(python3 - <<'PY'
import secrets
print(secrets.token_hex(16))
PY
)"

output="/var/lib/voice-worker/output/${request_id}.opus"

fish-audio-cli \
  --config /opt/fish-audio-cli/config/config.json \
  --text "$TEXT" \
  --format opus \
  --output "$output"
```

Using unique destinations avoids same-path writer races.

The CLI’s internal log request ID is generated independently and is not exposed as a filename variable.

---

## 44. Post-failure inspection

When a failure may have occurred after rename:

```bash
clear

output="/var/lib/voice-worker/output/message.opus"

if [ -e "$output" ]; then
  stat -c 'mode=%a size=%s path=%n' "$output"
else
  printf 'output is absent\n'
fi
```

Do not treat existence alone as proof that the invocation returned success.

Use both:

- exit status;
- application logs;
- filesystem state when recovery requires it.

---

## 45. Stale temp inspection

Example:

```bash
clear

find /var/lib/voice-worker/output \
  -maxdepth 1 \
  -type f \
  -name '.*.*.tmp' \
  -printf '%TY-%Tm-%Td %TH:%TM:%TS %m %s %p\n'
```

Review results before deletion.

The generic pattern can include temp files from unrelated software in a shared directory.

A dedicated output directory reduces ambiguity.

---

## 46. Metadata behavior

A replacement file receives new metadata.

The output package does not preserve old:

- inode;
- owner configuration beyond creation rules;
- group beyond creation rules;
- mode;
- access-control lists;
- extended attributes;
- creation time;
- modification time;
- security labels;
- immutable flags;
- hard-link topology.

### 46.1 Deployment consequence

Do not pre-create an output file with custom mode or ACL and expect those settings to survive replacement.

Apply policy at the directory or post-publication layer when needed.

### 46.2 Mandatory access control

SELinux, AppArmor, or other policy can affect create, rename, and sync operations.

The package does not relabel files explicitly.

---

## 47. Backups

The output package does not create:

- `.bak` files;
- versioned copies;
- previous-content snapshots;
- rollback records.

Before rename, the old destination is naturally preserved.

After rename, the old destination is no longer available through that path.

Use unique versioned filenames or an external archive when history matters.

---

## 48. Retrying an invocation

A caller can retry after failure, but must understand the failure phase.

### Before publication

A retry begins with the old destination still present.

### After publication persistence failure

A retry begins with the newly published file already at the destination.

A later successful retry replaces it again.

### Fish semantic duplication

The Fish API request may already have been processed remotely before a local output error.

The current Fish integration has no idempotency key.

A retry can create another synthesis request and consume additional provider quota.

---

## 49. Security boundary

The output path is caller-controlled.

A privileged process can replace entries anywhere it has directory write permission.

The output subsystem does not confine destinations to:

- project directory;
- configured output root;
- a safe allowlist;
- the current user’s home.

### 49.1 Untrusted path input

Do not pass an untrusted arbitrary path directly to `--output` when the process has broader filesystem access.

### 49.2 Dedicated directory

Recommended:

```text
/var/lib/voice-worker/output/
```

with controlled ownership and permissions.

### 49.3 Parent symlink races

The output code uses ordinary path-based filesystem operations.

It does not anchor operations with `os.Root` or equivalent.

A hostile actor able to mutate parent path components can create race conditions.

### 49.4 Leaf symlink protection is limited

Replacing the destination symlink rather than following it is useful.

It does not make an untrusted parent directory safe.

---

## 50. Special paths

Values such as:

```text
-
/dev/stdout
/dev/null
```

have no documented special output meaning.

### `-`

Treated as a literal filename named `-`.

### `/dev/stdout`

Treated as a filesystem path.

Atomic rename semantics are incompatible with using it as a streaming stdout protocol.

### `/dev/null`

Treated as a filesystem path and subject to same-directory temp creation and rename behavior.

Do not use special-device paths.

Use a regular destination file.

---

## 51. Network filesystems

Network and userspace filesystems can differ in:

- rename atomicity;
- directory sync support;
- cache coherence;
- error timing;
- durability guarantees;
- permission behavior.

The output package does not detect filesystem type.

Test the complete failure and crash behavior on the actual deployment filesystem.

A successful local ext4 test does not certify every remote mount devised by civilization.

---

## 52. Windows considerations

The code uses portable Go filesystem calls, but destination replacement and open-file behavior can differ by platform.

Potential differences include:

- replacing an existing destination;
- removing an open temp file;
- rename conflicts with scanners or readers;
- directory synchronization support.

Platform support must be verified with the project test suite on the target system.

Do not infer Unix semantics from function names alone.

---

## 53. Unix considerations

On ordinary Unix-like local filesystems:

- same-directory rename commonly provides atomic path replacement;
- a destination symlink entry is replaced rather than followed;
- open descriptors can continue referring to the old inode;
- directory sync is commonly used to persist rename metadata.

Actual guarantees remain filesystem-specific.

---

## 54. Package use outside the CLI

Programmatic callers can use `WriteAtomic` for non-audio content.

The function does not validate:

- MIME type;
- file extension;
- payload size;
- payload non-emptiness;
- content checksum;
- audio structure.

The caller must define those rules.

### 54.1 Callback panic

The function does not recover a panic from the callback.

Deferred cleanup runs during ordinary panic unwinding, then the panic continues.

### 54.2 `os.Exit`

A callback or caller invoking `os.Exit` bypasses defers and can leave a temp file.

Do not terminate the process from inside the callback.

---

## 55. Why same-directory temp files

Creating the temp beside the destination provides:

- same-filesystem rename under normal path conditions;
- destination-directory permission validation early;
- no separate temp-volume dependency;
- atomic replacement capability;
- predictable cleanup location.

Tradeoffs:

- destination directory must hold the complete temp;
- watchers see temp activity;
- filename base is disclosed;
- no fallback to a global temp directory.

---

## 56. Why sync before close

The package synchronizes the temporary file before closing it.

This makes a sync error an explicit pre-publication failure.

Closing alone is not treated as an equivalent durability request.

---

## 57. Why close before rename

The package closes the temporary file before publication.

Benefits include:

- surfacing close errors before rename;
- improved portability;
- avoiding publication of a file whose writer lifecycle is incomplete.

A close failure aborts publication.

---

## 58. Why sync the directory after rename

The file’s own sync does not persist the directory-entry replacement on every filesystem.

Directory sync is the final requested durability step.

The application therefore does not report success immediately after rename.

---

## 59. Why cleanup errors are joined

Ignoring cleanup errors can hide:

- leaked temporary files;
- file-descriptor problems;
- permission changes;
- filesystem damage.

Replacing the primary error with cleanup error would also be wrong.

Joining preserves both:

```text
synthesis failed
+
temporary cleanup failed
```

Package callers can use `errors.Is` and `errors.As` across joined errors.

---

## 60. Why published output is retained after sync failure

After rename:

- the old destination is already gone from that path;
- the new destination is visible;
- deleting the new destination cannot restore the old one.

The least destructive behavior is:

- retain the new output;
- return a persistence error;
- let the caller decide recovery.

---

## 61. Failure matrix

| Failure point | Temp may exist after cleanup failure | Old destination preserved | New destination may exist | CLI status |
|---|---:|---:|---:|---:|
| invalid argument | no | yes | no | `4` only when reached through app |
| temp create | no | yes | no | `4` |
| callback/write | yes | yes | no | `4` |
| temp sync | yes | yes | no | `4` |
| temp close | yes | yes | no | `4` |
| rename | yes | yes | no | `4` |
| directory open | no temp | no | yes | `4` |
| directory sync | no temp | no | yes | `4` |
| directory close | no temp | no | yes | `4` |
| complete success | no | replaced if present | yes | `0` |

“Old destination preserved” assumes no concurrent writer replaces it.

---

## 62. Testing expectations

### Arguments

Test:

- empty path;
- nil callback;
- relative path;
- absolute path;
- whitespace path behavior where supported.

### Parent directory

Test:

- existing parent;
- missing parent;
- unwritable parent;
- read-only filesystem where available.

### Temporary file

Test:

- same-directory placement;
- unique pattern;
- mode `0600`;
- callback receives writable output;
- callback failure after partial write;
- cleanup after failure.

### Replacement

Test:

- create new destination;
- replace existing regular file;
- old file preserved on pre-rename failure;
- replacement mode `0600`;
- destination symlink replaced without target modification;
- destination directory causes rename failure.

### Cleanup

Test:

- close during cleanup;
- remove failure through injectable boundary if introduced;
- missing temp treated as cleaned;
- primary and cleanup errors preserved;
- no temp remains on ordinary failure.

### Durability

Test:

- temp sync failure through injection;
- temp close failure;
- rename failure;
- directory open failure;
- directory sync and close errors combined;
- published output retained after post-rename failure.

### Integration

Test:

- Fish partial stream does not publish;
- empty Fish response does not publish;
- cancellation does not publish;
- existing output preserved on Fish failure;
- success publishes requested bytes;
- CLI returns `4` on output failure;
- CLI returns `0` only after full output success.

---

## 63. Review checklist

### Path

- Is output still relative to cwd rather than project root?
- Is exact-empty validation unchanged?
- Are parent directories still caller-owned?
- Are special paths still unsupported?
- Is the destination directory trusted?

### Temporary file

- Is it still created beside destination?
- Is the pattern still unique and hidden-style?
- Is mode `0600` preserved?
- Does the callback still receive only `io.Writer`?
- Are empty payload semantics intentional?

### Pre-publication

- Does callback error prevent rename?
- Does file sync occur before close?
- Does close occur before rename?
- Are old destinations preserved?
- Are cleanup failures joined?

### Publication

- Is rename still the only final-entry replacement?
- Are destination symlinks replaced rather than followed?
- Are platform limitations documented?
- Is existing metadata intentionally not preserved?

### Post-publication

- Is the directory still opened, synced, and closed?
- Are sync and close failures both preserved?
- Is published output retained after these failures?
- Does the CLI still report status `4`?

### Operations

- Are concurrent same-path writers still unlocked?
- Is stale-temp cleanup still external?
- Are crash windows documented?
- Are network filesystem caveats accurate?
- Are output paths and errors logged without exposing audio contents?

---

## 64. Output invariants

The following rules are normative for the current implementation.

1. The caller must supply `--output`.
2. Output has no automatic filename.
3. Audio is not written to stdout.
4. Destination extension does not select audio format.
5. Output path does not use project path resolution.
6. Relative output follows process working directory.
7. Exact empty path is rejected.
8. Output path is not trimmed.
9. Parent directories are not created.
10. The writer callback must be non-nil.
11. Temporary file is created in the destination directory.
12. Temporary name uses `.<baseName>.*.tmp`.
13. Temporary file is created with restrictive `0600` permissions.
14. The final file inherits the temporary file’s mode.
15. Existing destination mode is not preserved.
16. Existing destination ownership and extended metadata are not copied.
17. Callback writes directly to the temporary file.
18. Generic output code does not reject zero-byte successful callbacks.
19. Callback failure prevents publication.
20. Temporary file is synced before publication.
21. Temporary file is closed before publication.
22. Close failure prevents publication.
23. Publication uses same-directory rename.
24. Existing destination remains until rename succeeds.
25. Existing regular destination is replaced, not modified in place.
26. A destination leaf symlink is replaced rather than followed.
27. Parent symlink components are not hardened.
28. Rename failure triggers cleanup.
29. Before publication, cleanup closes and removes the temp.
30. Missing temp during cleanup is treated as already removed.
31. Cleanup errors are joined with primary errors.
32. After rename, output is marked published.
33. Published output is never removed by cleanup.
34. Containing directory is opened after rename.
35. Directory sync is required before success.
36. Directory close is required before success.
37. Directory sync and close errors are both preserved.
38. A post-rename error returns failure while retaining the new output.
39. CLI maps output errors to status `4`.
40. Exit `4` does not guarantee output absence.
41. Exit `0` means the full output sequence succeeded.
42. There is no destination lock.
43. Concurrent writers use independent temp files.
44. Last successful rename can win.
45. There is no stale-temp startup scavenger.
46. Managed cancellation normally cleans unpublished temp files.
47. Abrupt termination can leave temp files.
48. The output package does not validate audio content.
49. The output package does not create backups.
50. Atomic and durability guarantees depend on host filesystem semantics.

Changing one of these rules is an output compatibility or durability change.

---

## 65. Non-goals

The current output system does not provide:

- stdout audio streaming;
- automatic output naming;
- extension inference;
- parent-directory creation;
- output-root confinement;
- caller path allowlists;
- destination locking;
- compare-and-swap replacement;
- version history;
- backup files;
- append mode;
- resume support;
- partial-download exposure;
- post-publication checksum verification;
- MIME validation;
- audio container validation;
- transcode support;
- free-space preflight;
- quota reservation;
- automatic stale-temp removal;
- cross-process transaction coordination;
- portable identical semantics on every filesystem;
- rollback after a successful rename;
- preservation of previous mode, ownership, ACL, or xattrs.

These require explicit product and failure-semantics design.

---

## 66. Summary

The publication contract is:

```text
write only to hidden same-directory temp
    ↓
sync temp
    ↓
close temp
    ↓
rename to final destination
    ↓
sync and close directory
```

Before rename:

```text
failure
    → keep old destination
    → clean temp
```

After rename:

```text
directory persistence failure
    → keep new destination
    → return error
```

The most important operational rules are:

- create the output parent directory first;
- use absolute paths in services;
- use unique paths for concurrent jobs;
- expect final mode `0600`;
- do not expect existing metadata to survive replacement;
- treat a destination symlink as a replaced leaf, not a followed target;
- keep the parent directory trusted;
- ignore hidden temp files as incomplete outputs;
- inspect filesystem state after a post-rename failure;
- remember exit `4` can coexist with a published file;
- rely on exit `0` for the complete sync-and-publication success path.
