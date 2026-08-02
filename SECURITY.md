# Security policy

Security reports for `fish-audio-cli` are taken seriously.

The project handles:

- a Fish Audio API key;
- user-provided text;
- configurable network destinations;
- persistent logs;
- local filesystem paths;
- temporary and final audio files;
- compiled-in text-processing modules.

A defect at one of these boundaries can expose credentials, disclose text, overwrite files, weaken permissions, misroute requests, or leave misleading filesystem state.

This policy explains supported versions, private reporting, report contents, testing rules, security boundaries, examples of relevant vulnerabilities, and behavior that is intentionally outside the project’s security guarantees.

> [!IMPORTANT]
> Do not include a real API key, private input text, production audio, authorization header, or exploit details in a public GitHub issue.

## Table of contents

- [Supported versions](#supported-versions)
- [Reporting a vulnerability](#reporting-a-vulnerability)
- [Private reporting channels](#private-reporting-channels)
- [What to include](#what-to-include)
- [What not to include](#what-not-to-include)
- [Response and disclosure process](#response-and-disclosure-process)
- [No guaranteed response time](#no-guaranteed-response-time)
- [No bug bounty promise](#no-bug-bounty-promise)
- [Safe research guidelines](#safe-research-guidelines)
- [Scope](#scope)
- [High-impact examples](#high-impact-examples)
- [Secrets and credentials](#secrets-and-credentials)
- [Configuration trust boundary](#configuration-trust-boundary)
- [Network trust boundary](#network-trust-boundary)
- [Filesystem trust boundary](#filesystem-trust-boundary)
- [Logging and text privacy](#logging-and-text-privacy)
- [Modules](#modules)
- [Output files](#output-files)
- [Denial of service](#denial-of-service)
- [Third-party services](#third-party-services)
- [Expected behavior and non-vulnerabilities](#expected-behavior-and-non-vulnerabilities)
- [Credential exposure response](#credential-exposure-response)
- [Maintainer handling](#maintainer-handling)
- [Security changes](#security-changes)
- [Security checklist](#security-checklist)
- [Policy limitations](#policy-limitations)

---

## Supported versions

The project is currently in alpha development and does not yet maintain stable release branches.

| Version | Security support |
|---|---|
| Current `main` branch | Supported |
| Older commits | Best effort only |
| Unmaintained forks | Not supported by this project |
| Locally modified builds | Supported only when the issue reproduces in current upstream code |
| Future tagged releases | Will be documented when a release policy exists |

Before reporting, reproduce the issue against the current upstream `main` branch when practical.

Include the exact commit SHA.

```bash
clear

git rev-parse HEAD
git log -1 --oneline --decorate
```

A vulnerability that affects an older commit but no longer affects current `main` can still be useful historical information, but it may not receive a separate backport.

---

## Reporting a vulnerability

Report suspected vulnerabilities privately.

Do not begin with:

- a public issue containing exploit steps;
- a public pull request containing a working exploit;
- a discussion post containing a leaked credential;
- a log paste containing private text or authorization data;
- a social-media disclosure before maintainers can assess the issue.

A good report should let the maintainer:

1. reproduce the behavior;
2. understand the trust boundary;
3. determine impact;
4. identify affected code;
5. distinguish intended behavior from a defect;
6. create a regression test;
7. coordinate a fix and disclosure.

---

## Private reporting channels

### Preferred channel

Use GitHub private vulnerability reporting when the repository exposes:

```text
Security
→ Report a vulnerability
```

That route keeps the report inside GitHub’s private security-advisory workflow.

### When private vulnerability reporting is unavailable

Use a private contact method explicitly listed by the repository owner on his GitHub profile or repository metadata.

Do not guess an email address from:

- a username;
- a commit author string;
- a domain;
- an organization name.

### When no private contact is available

Open a minimal public issue that says only:

```text
I need a private channel to report a potential security vulnerability.
```

Do not include:

- the vulnerable component;
- exploit conditions;
- proof-of-concept code;
- paths;
- credentials;
- private text;
- impact details that reveal exploitation.

Wait for the maintainer to provide a private route.

### Do not use unrelated third parties

Do not send project vulnerability details to:

- Fish Audio support, unless the issue is actually in Fish Audio;
- package mirrors;
- random maintainers of other Go projects;
- public chat rooms;
- automated paste services.

---

## What to include

Include enough information for reliable reproduction.

### Required details

- Short vulnerability summary
- Exact upstream commit SHA
- Operating system
- Architecture
- Go version
- Relevant filesystem type when applicable
- Relevant mount behavior when applicable
- Whether the process ran as root or another user
- Minimal configuration fragment with secrets removed
- Exact command or package call
- Expected behavior
- Actual behavior
- Security impact
- Preconditions
- Reproduction steps
- Whether a real external request occurred
- Whether any output file was published
- Whether any secret, text, or provider response entered logs
- Suggested remediation when known

### Useful supporting evidence

- Minimal test case
- Synthetic proof of concept
- File metadata
- Permission bits
- Symlink layout
- Request count against a local test server
- Redacted structured log
- Wrapped error chain
- Race-detector report
- Stack trace
- Before-and-after filesystem state

### File metadata example

```bash
clear

stat -c \
  'type=%F mode=%a owner=%U group=%G bytes=%s path=%n' \
  /path/to/file
```

Do not include file contents when metadata is sufficient.

---

## What not to include

Never include:

- Fish Audio API keys;
- bearer tokens;
- authorization headers;
- production configuration containing credentials;
- private text submitted for synthesis;
- private generated audio;
- unredacted provider account identifiers;
- full environment dumps;
- home-directory archives;
- unrelated logs;
- private SSH keys;
- cloud credentials;
- browser cookies;
- personal access tokens;
- credentials belonging to another user;
- data obtained from systems you do not own or have permission to test.

Use synthetic markers such as:

```text
test-key
synthetic-input
fake-audio
```

If a real credential was exposed during research, rotate it before continuing.

---

## Response and disclosure process

When possible, the maintainer will:

1. acknowledge the report privately;
2. determine whether the issue affects current upstream code;
3. request missing reproduction details;
4. assess impact and affected boundaries;
5. create or confirm a regression test;
6. prepare a fix;
7. update affected documentation;
8. coordinate disclosure;
9. publish a GitHub security advisory when appropriate;
10. identify fixed commits or releases.

The exact process depends on:

- reproducibility;
- severity;
- maintainer availability;
- release status;
- third-party coordination;
- whether the defect has already been publicly disclosed.

### Coordinated disclosure

Please allow a reasonable opportunity to investigate and fix the issue before public disclosure.

Do not publish:

- exploit code;
- exact vulnerable paths;
- credential-exfiltration methods;
- race timing details;
- bypass sequences

while the report is being actively handled, unless the maintainer has stopped responding for an extended period and responsible disclosure requires escalation.

### Public credit

A reporter can request:

- public credit;
- anonymous credit;
- no credit.

Credit is not guaranteed when identity or authorship cannot be verified.

---

## No guaranteed response time

The project does not currently promise a fixed:

- acknowledgement SLA;
- remediation SLA;
- release deadline;
- advisory publication deadline.

This is an open-source alpha project, not a staffed commercial security response center.

A lack of immediate response does not authorize disclosure of credentials or user data.

For an urgent active compromise:

1. rotate affected credentials;
2. stop vulnerable deployments;
3. restrict filesystem and network access;
4. preserve evidence;
5. report privately;
6. apply a local mitigation if safe.

---

## No bug bounty promise

This project does not currently operate a bug bounty program.

Submitting a report does not create a promise of:

- payment;
- reward;
- merchandise;
- employment;
- public recognition;
- CVE assignment.

Do not incur costs or perform risky testing based on an assumption of compensation.

---

## Safe research guidelines

Security research must be authorized and non-destructive.

### Use your own environment

Test against:

- your own clone;
- your own machine or container;
- your own files;
- your own synthetic API key;
- a local HTTP server;
- an account you control.

### Prefer local doubles

For Fish HTTP behavior, use:

- `httptest.Server`;
- a custom `http.RoundTripper`;
- synthetic response bodies;
- fake credentials.

Normal project tests do not need the live Fish API.

### Minimize impact

Stop after demonstrating the issue.

Do not:

- exfiltrate data beyond the minimum proof;
- read unrelated files;
- alter another user’s files;
- retain credentials;
- establish persistence;
- perform lateral movement;
- evade monitoring;
- degrade third-party services;
- create paid provider usage without authorization.

### Avoid denial-of-service testing against third parties

Do not perform:

- request floods;
- retry storms;
- disk exhaustion;
- inode exhaustion;
- connection exhaustion;
- oversized provider traffic;
- resource-intensive fuzzing

against Fish Audio, GitHub, public infrastructure, or systems you do not own.

### No social engineering

Do not impersonate maintainers, provider staff, or users.

Do not request credentials as part of testing.

### No public credential validation

Do not test whether a leaked key is active by sending it to Fish Audio.

Rotate or revoke it.

### Preserve privacy

Use synthetic text.

The fact that the program handles TTS does not make a user’s text suitable for a public proof of concept. Apparently this clarification remains necessary in the age of screenshots.

---

## Scope

A security issue is generally in scope when current upstream code violates a documented trust, confidentiality, integrity, availability, or authentication boundary.

Examples include:

- Fish API key disclosure;
- authorization header disclosure;
- unexpected transmission of text or credentials;
- secret symlink protection bypass;
- secret regular-file check bypass;
- secret permission hardening bypass;
- unsafe secret file replacement race;
- output symlink behavior differing from the documented contract in a way that overwrites an unintended target;
- temporary-file collision or predictable-name exploit;
- arbitrary file overwrite outside the caller-selected path without trusted configuration control;
- path-resolution behavior that escapes a documented confinement guarantee;
- command execution caused by config, text, provider response, filename, or log data;
- log injection that changes record structure or hides attacker-controlled records;
- sensitive text logged when `logging.logText` is false and no error/provider boundary justifies it;
- API key present in errors or logs;
- unbounded parsing that bypasses configured resource limits;
- remote response handling that causes unsafe local file behavior;
- retry behavior that sends credentials or text to an unintended destination;
- HTTP redirect behavior that exposes authorization data contrary to Go and project expectations;
- TLS verification bypass;
- race condition causing secret or output operations to target another object;
- malicious module configuration bypassing strict decoding;
- typed-nil or malformed dependency causing exploitable panic in a privileged deployment;
- dependency compromise introduced by the project;
- CI workflow behavior exposing repository secrets.

A report should explain the attacker model.

---

## High-impact examples

The following examples are likely high impact when reproducible under realistic conditions.

### Credential disclosure

- API key written to stderr
- API key written to persistent log
- API key included in an error
- API key sent to a host not selected by trusted configuration
- API key exposed through redirect handling
- API key included in crash output

### Unintended file modification

- Secret leaf symlink followed despite rejection contract
- Output destination symlink target overwritten instead of replacing the leaf
- Predictable temp name permits another user to replace or read output
- Attacker-controlled input changes output path without caller/config authority
- Cleanup removes an unrelated file

### Trust-boundary bypass

- Unknown JSON fields silently accepted despite strict contract
- Duplicate JSON keys bypass validation
- Module config bypasses module-owned strict decoding
- Untrusted text changes the Fish endpoint
- Provider response executes a local command
- Log value injects a second forged structured record

### Security-control failure

- Secret file remains group-readable after successful load
- Secret directory write-bit check can be bypassed
- Invalid header control characters reach the HTTP transport
- TLS certificate verification is disabled
- Error-body limit can be bypassed to exhaust memory

---

## Secrets and credentials

The default Fish API key path is:

```text
secrets/fish-api-key
```

The configured path is project-relative unless absolute.

The secret loader currently requires:

- trusted final directory;
- no group/other write bit on the final directory;
- regular file leaf;
- no leaf symlink;
- same-file verification after open;
- final file mode `0600`;
- bounded bytes;
- valid UTF-8;
- exactly one nonblank line;
- no arbitrary surrounding whitespace.

Security reports affecting these guarantees are in scope.

### Missing secret behavior

A missing secret file is created empty with restricted permissions.

The command then exits with status `3`.

This behavior is intentional.

### Existing file chmod

The loader forces the existing secret file to `0600`.

Failure on a read-only mount is intentional under the current contract.

### Parent components

The loader does not claim to reject every symlink or insecure owner in every ancestor component.

Deployments must use a trusted absolute hierarchy.

A report based only on an attacker already controlling a trusted ancestor directory should explain why the documented trust boundary is insufficient.

### Credential rotation

The application does not rotate Fish API keys.

Rotation is controlled by the Fish account and deployment process.

---

## Configuration trust boundary

The JSON configuration is trusted administrative input.

It controls:

- Fish base URL;
- model;
- reference ID;
- retry behavior;
- secret path;
- log path;
- module list;
- module-owned settings;
- text logging;
- resource limits.

### Arbitrary base URL is intentional

`fish.baseUrl` can point to an operator-selected HTTP or HTTPS service.

The application sends:

- processed text;
- Fish API key;
- model header;
- synthesis settings

to the configured endpoint.

This is not a vulnerability when a trusted operator selected the endpoint.

It becomes a vulnerability when an attacker who is not authorized to modify trusted configuration can cause the endpoint to change.

### Relative path escape

Project-relative paths are lexically joined and can contain `..`.

The project resolver does not provide sandbox confinement.

This is documented behavior.

A report must distinguish:

- untrusted configuration modification, which already compromises the deployment;
- an unexpected escape from a separately promised confinement boundary.

### No environment interpolation

The application does not expand:

- `~`;
- `$HOME`;
- `${VAR}`;
- `%VAR%`.

Literal treatment is intentional.

### Config file permissions

The application does not enforce a specific mode or owner for the config file.

Operators must protect it because it controls sensitive behavior.

A report should focus on an actual bypass or escalation rather than the absence of a config chmod feature.

---

## Network trust boundary

The application trusts:

- the configured base URL;
- the host operating system;
- the Go HTTP transport;
- the system CA store;
- configured proxy environment;
- the remote provider to return audio or bounded errors.

### TLS

HTTPS uses normal Go certificate verification.

There is no supported insecure-skip option.

A TLS validation failure is expected when:

- CA trust is missing;
- hostname mismatches;
- a proxy performs untrusted interception;
- system time is wrong.

### HTTP base URLs

The configuration currently permits both:

```text
http
https
```

Using HTTP sends text and credentials without transport encryption.

That is an operator security decision under trusted configuration.

A proposal to forbid HTTP is a compatibility and policy change.

### Proxy environment

The default Go transport can use standard proxy environment variables.

The deployment environment is trusted.

An attacker who can change service environment variables may already control request routing.

A report should explain any privilege crossing beyond that expected control.

### Redirects

Redirect behavior follows the configured Go HTTP client and standard library.

A report that demonstrates credential forwarding to an unintended host is in scope.

Use only local test hosts and synthetic keys.

### Retry duplication

The client has no provider idempotency key.

A retry can produce duplicate remote work or charges.

This is documented operational risk, not by itself a vulnerability.

An unbounded retry loop or retry to an unintended host would be in scope.

---

## Filesystem trust boundary

The application expects trusted directories for:

- configuration;
- secrets;
- logs;
- output.

It does not provide a general filesystem sandbox.

### Process identity

The process acts with the permissions of its operating-system user.

It does not:

- drop privileges;
- enter a chroot;
- create a mount namespace;
- apply seccomp;
- isolate modules;
- restrict system calls.

Running as root increases impact and is not recommended.

### Untrusted directories

Do not place secrets, logs, or output in directories writable by untrusted users.

Security behavior at a trusted leaf cannot compensate for an attacker replacing parent components.

### Permissions

Current file modes include:

| Artifact | Mode |
|---|---:|
| Secret file | `0600` |
| Output file | `0600` |
| Persistent log file | `0640` |

Missing secret directories are requested as `0700`.

Missing log directories are requested as `0750`.

Umask can reduce requested directory permissions.

### Ownership

The application changes modes where documented.

It does not generally change owner or group.

A deployment must provision correct ownership.

---

## Logging and text privacy

Configured logging writes to:

```text
stderr
persistent log file
```

Persistent file logging cannot currently be disabled.

### Text logging

`logging.logText` defaults to:

```text
false
```

When false, documented top-level input and processed-output text fields are omitted.

When true, those text values are intentionally logged.

Enabling text logging is an operator privacy decision.

### Errors and provider messages

A module error or bounded provider error can contain text or identifiers.

`logging.logText=false` is not a universal redaction engine for arbitrary external error strings.

A report is relevant when:

- the project itself unnecessarily includes the full secret;
- the project logs text contrary to its documented fields and boundaries;
- provider bodies bypass the configured bound;
- structured log injection occurs.

### Log permissions

The persistent log file is forced to `0640`.

The directory and group remain deployment responsibilities.

### Log rotation

Rotation is external.

The included template does not create a security guarantee for every installation.

### Runtime write failure

Normal `slog` calls do not return handler write errors to the command.

A full or failing log filesystem can cause diagnostic loss without changing the business exit status.

This is documented behavior, not a confidentiality bypass by itself.

---

## Modules

Modules are compiled into the binary.

The current built-in module type is:

```text
passthrough
```

The project does not currently load arbitrary runtime plugins.

### Module security responsibilities

A module must:

- strictly decode its config;
- validate semantic values;
- respect cancellation;
- preserve valid UTF-8;
- return nonblank text on success;
- avoid leaking secrets;
- avoid logging full text by default;
- bound remote error bodies;
- close resources;
- test failure side effects.

### Custom forks

A vulnerability in a third-party module or fork is not automatically an upstream core vulnerability.

Report upstream when:

- core permits a module to bypass a promised boundary;
- registry or pipeline behavior is unsafe;
- shared module APIs encourage an unavoidable vulnerability;
- the issue reproduces with upstream built-ins or a minimal conforming test module.

### No module sandbox

Compiled modules run in the same process and with the same operating-system privileges.

A malicious compiled module can access process memory, files, environment, and network.

The module system is an extension architecture, not a security sandbox.

---

## Output files

The output package publishes audio through a temporary file beside the destination.

The intended sequence is:

```text
create temp
write response
sync temp
close temp
rename to destination
sync directory
close directory
```

### Destination selection

The caller controls `--output`.

The application does not confine output beneath a project directory.

This is intentional.

An untrusted caller allowed to choose arbitrary output paths can request writes wherever the process has permission.

A wrapper that accepts untrusted output paths must enforce its own directory policy.

### Destination symlink

A symlink at the final destination is replaced as a leaf.

The target should remain unchanged.

A bypass that causes the target to be overwritten is in scope.

### Parent symlinks

Parent path components follow ordinary filesystem resolution.

Use a trusted output hierarchy.

### Existing destination

Before successful rename, a failure preserves the existing destination.

After successful rename, a directory persistence failure returns an error but keeps the new file.

This state is intentional and documented.

### Concurrency

There is no destination lock.

Two authorized writers targeting the same path can race, and the last successful rename can win.

This is not a vulnerability unless an attacker crosses a trust boundary or causes writes outside the authorized destination.

### Temp files

Temporary names are generated by the operating-system file API and created securely.

A report demonstrating predictability, collision, disclosure, or replacement is in scope.

---

## Denial of service

Denial-of-service reports are evaluated based on realistic attacker control and documented bounds.

Potentially relevant examples:

- bypassing config byte limit;
- bypassing input byte limit;
- bypassing secret byte limit;
- bypassing Fish error-body limit;
- unbounded retry;
- infinite loop from valid input;
- deadlock;
- goroutine leak triggered remotely;
- file-descriptor leak;
- response-body leak;
- memory growth disproportionate to configured limits;
- panic from untrusted text or provider response;
- output cleanup deleting unrelated files.

Usually lower priority or out of scope:

- huge limits deliberately configured by a trusted operator;
- filling a caller-selected output filesystem with authorized synthesis;
- many concurrent authorized invocations;
- provider latency within configured timeout;
- machine exhaustion caused by running arbitrary malicious compiled modules;
- local user killing the process with `SIGKILL`.

Do not perform destructive DoS testing against shared systems.

---

## Third-party services

### Fish Audio

Report a provider-side vulnerability to Fish Audio through its own security process.

Examples primarily belonging to Fish Audio:

- unauthorized access to another account’s voice;
- provider authentication bypass;
- provider billing bypass;
- provider data exposure independent of this client;
- provider model or platform vulnerability.

Report to this project when client behavior contributes, such as:

- leaking credentials;
- sending data to the wrong host;
- unsafe redirect handling;
- unbounded response handling;
- incorrect local validation causing a security impact;
- logging sensitive provider responses contrary to contract.

### GitHub

Repository hosting, Actions infrastructure, and private advisory tooling are operated by GitHub.

A vulnerability in GitHub itself belongs to GitHub’s security process.

A workflow committed by this project that leaks secrets belongs to this project.

### Go standard library

A vulnerability solely in Go’s standard library should also be reported to the Go security team.

Report here when the project:

- uses the API unsafely;
- fails to apply a required mitigation;
- is affected in a way users need to understand;
- needs a toolchain update.

---

## Expected behavior and non-vulnerabilities

The following are generally expected under the current documented model.

### Trusted config can redirect data

A trusted operator can configure another HTTP or HTTPS base URL.

Text and the key are sent there.

### Trusted config can select filesystem paths

A trusted operator can configure secret and log paths, including absolute paths and lexical `..`.

### Caller selects output

The caller can select any output path writable by the process.

### Input text leaves the machine

Processed text is sent to the configured Fish endpoint.

The application is not offline TTS.

### Text can be logged when enabled

`logging.logText=true` intentionally logs top-level text fields.

### Logs are persistent

The configured logger always opens a file and stderr destination.

There is no supported stderr-only mode.

### HTTP is configurable

A trusted operator can choose an HTTP endpoint and accept plaintext transport risk.

### Custom compiled module is trusted code

A malicious compiled module is equivalent to malicious application code.

### Root execution has root impact

The process does not sandbox itself.

### Output is not encrypted

Mode `0600` controls Unix file access.

It does not encrypt audio at rest.

### Output content is not decoded

A non-empty successful provider stream can be published without validating the audio container.

This is integrity limitation, not necessarily a security vulnerability.

### Shell signal status can differ

External statuses such as `130`, `143`, or `137` can result from signals.

### Missing secret creates a file

The first run can create an empty protected secret file and exit `3`.

### Post-rename error can leave output

Exit `4` can coexist with a new destination after directory-sync failure.

### Same-path writers can race

There is no destination lock.

### Log close failure does not change exit status

The command can complete successfully and later report log close failure to bootstrap stderr.

A report can still be valid if one of these behaviors produces an unexpected privilege crossing, confidentiality breach, or integrity failure beyond the documented boundary.

---

## Credential exposure response

If a Fish API key is exposed:

1. revoke or rotate it through the Fish account;
2. stop affected deployments if necessary;
3. remove the key from active files;
4. inspect logs and CI output;
5. inspect Git history;
6. inspect issue and pull-request content;
7. update deployment secrets;
8. verify old credentials no longer work;
9. review how the exposure occurred;
10. add a regression test or process control.

Deleting the latest file does not remove a secret from Git history.

Changing file permissions does not revoke a copied credential.

### Do not validate leaked credentials publicly

Do not send a leaked key to the provider to see whether it still works.

Assume compromise and rotate it.

### Redaction

When preserving evidence, replace the key with a stable marker:

```text
<REDACTED_FISH_API_KEY>
```

Keep only the minimum private evidence required for investigation.

---

## Maintainer handling

A maintainer receiving a report should:

1. move discussion to a private channel;
2. avoid reproducing with the reporter’s real key;
3. create a synthetic local reproduction;
4. confirm current `main`;
5. identify the trust boundary;
6. assess confidentiality, integrity, availability, and authentication impact;
7. inspect related code and tests;
8. check whether documentation already promises the behavior;
9. create a regression test;
10. prepare the smallest safe fix;
11. run full tests and the race detector;
12. update all affected documentation;
13. review adjacent boundaries;
14. coordinate disclosure;
15. advise credential rotation where relevant.

### Evidence handling

Do not copy sensitive report content into:

- public issues;
- ordinary PR descriptions;
- CI logs;
- commit messages;
- public test fixtures.

Use synthetic regression data.

### Fix commits

A public fix commit should avoid revealing an easily weaponized exploit before coordinated disclosure when practical.

A private security-advisory fork can be used when GitHub private vulnerability reporting is active.

### Documentation review

A security fix may require updates to:

- `README.md`;
- `CONTRIBUTING.md`;
- `SECURITY.md`;
- `docs/architecture.md`;
- `docs/configuration.md`;
- `docs/fish-audio.md`;
- `docs/logging.md`;
- `docs/secrets-and-paths.md`;
- `docs/output-files.md`;
- `docs/errors-and-exit-codes.md`;
- `docs/testing.md`;
- `docs/troubleshooting.md`;
- `CHANGELOG.md`.

---

## Security changes

Security-sensitive code changes should follow [`CONTRIBUTING.md`](CONTRIBUTING.md).

At minimum:

- define the threat model;
- identify trusted and untrusted inputs;
- preserve error identity;
- add deterministic regression tests;
- test failure side effects;
- test permissions;
- test symlinks;
- test typed nil where relevant;
- test cancellation;
- use local HTTP servers;
- avoid real credentials;
- update owning documentation;
- run the race detector;
- review for adjacent bypasses.

### Required verification

```bash
clear

set -euo pipefail

unformatted="$(gofmt -l .)"

if [[ -n "$unformatted" ]]; then
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

git diff --check
```

### Security regression tests

Prefer tests that prove both:

```text
attack rejected
intended behavior preserved
```

Examples:

- symlink rejected while regular secret succeeds;
- control character rejected while printable header succeeds;
- oversized body rejected while exact limit succeeds;
- old output preserved before rename;
- new output retained after post-rename failure;
- API key absent from logs;
- duplicate JSON key rejected while separate-object keys succeed.

---

## Security checklist

### Reporting

- [ ] Report is private.
- [ ] No real credential is included.
- [ ] No private text or audio is included.
- [ ] Exact commit is provided.
- [ ] Impact and attacker model are explained.
- [ ] Reproduction uses owned systems.

### Secrets

- [ ] API key never appears in logs or errors.
- [ ] Secret leaf is regular.
- [ ] Leaf symlink is rejected.
- [ ] Directory write-bit checks remain.
- [ ] Same-file race check remains.
- [ ] Final mode is `0600`.
- [ ] Reads remain bounded.
- [ ] UTF-8 and one-line rules remain.

### Network

- [ ] Endpoint comes only from trusted config.
- [ ] Header controls are rejected.
- [ ] TLS verification remains enabled for HTTPS.
- [ ] Redirect behavior is reviewed.
- [ ] Response bodies close.
- [ ] Error bodies remain bounded.
- [ ] Retry remains bounded and cancellable.
- [ ] No retry follows partial streaming.
- [ ] API key is never printed.

### Filesystem

- [ ] Trusted directory assumptions are explicit.
- [ ] Temp files are securely created.
- [ ] Cleanup targets only the owned temp.
- [ ] Old output is preserved before rename.
- [ ] New output remains after post-rename persistence failure.
- [ ] Destination symlink target is not followed.
- [ ] Modes remain correct.
- [ ] Same-path concurrency is documented.

### Logging

- [ ] Text privacy follows `logging.logText`.
- [ ] Structured values cannot forge records.
- [ ] Provider errors are bounded.
- [ ] Log file mode is `0640`.
- [ ] No unsupported `/dev/null` guidance exists.
- [ ] Early stderr-only behavior is preserved.

### Config and modules

- [ ] Strict JSON remains.
- [ ] Duplicate keys are rejected.
- [ ] Unknown fields are rejected.
- [ ] Null semantics are explicit.
- [ ] Module config remains module-owned.
- [ ] Prepare-before-build remains.
- [ ] Custom compiled modules are not described as sandboxed.

### Tests and docs

- [ ] Regression test fails before the fix.
- [ ] Full uncached tests pass.
- [ ] Race tests pass.
- [ ] Vet passes.
- [ ] Command builds.
- [ ] Owning docs are updated.
- [ ] Troubleshooting is updated.
- [ ] Security implications are documented.
- [ ] No sensitive fixture was committed.

---

## Policy limitations

This policy does not:

- create a legal safe-harbor agreement;
- authorize access to systems you do not own;
- authorize testing Fish Audio;
- authorize testing GitHub;
- waive third-party terms of service;
- promise payment;
- promise a response deadline;
- promise a CVE;
- promise a release;
- promise backports;
- guarantee confidentiality outside the chosen reporting platform;
- provide legal advice.

Researchers remain responsible for:

- obtaining authorization;
- obeying applicable law;
- respecting third-party systems;
- minimizing data access;
- avoiding service disruption;
- protecting credentials and user data.

---

## Related documentation

- Project overview: [`README.md`](README.md)
- Contribution workflow: [`CONTRIBUTING.md`](CONTRIBUTING.md)
- Documentation map: [`docs/index.md`](docs/index.md)
- Architecture: [`docs/architecture.md`](docs/architecture.md)
- Configuration: [`docs/configuration.md`](docs/configuration.md)
- Fish HTTP boundary: [`docs/fish-audio.md`](docs/fish-audio.md)
- Logging: [`docs/logging.md`](docs/logging.md)
- Secrets and paths: [`docs/secrets-and-paths.md`](docs/secrets-and-paths.md)
- Atomic output: [`docs/output-files.md`](docs/output-files.md)
- Errors and exit codes: [`docs/errors-and-exit-codes.md`](docs/errors-and-exit-codes.md)
- Testing: [`docs/testing.md`](docs/testing.md)
- Troubleshooting: [`docs/troubleshooting.md`](docs/troubleshooting.md)

---

## Summary

Use a private channel.

Use synthetic data.

Do not disclose credentials.

Do not test systems you do not own.

A useful security report explains:

```text
attacker control
    ↓
boundary crossed
    ↓
observable impact
    ↓
minimal reproduction
    ↓
current upstream commit
```

The project’s most sensitive boundaries are:

- Fish API key confidentiality;
- trusted configuration;
- endpoint selection;
- header safety;
- secret-file integrity;
- path and symlink behavior;
- log privacy;
- temporary-file ownership;
- atomic output publication;
- bounded parsing;
- retry and response lifecycle.

Documented operator control is not automatically a vulnerability.

Unexpected privilege crossing, credential disclosure, unintended transmission, unsafe file mutation, or bypass of a promised security control is.
