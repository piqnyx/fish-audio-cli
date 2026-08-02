# Documentation index

> **Document status:** navigation entry point and authority map for the current pre-release documentation set.
>
> **Audience:** users, operators, service integrators, shell-script authors, module authors, contributors, maintainers, and reviewers.
>
> **Scope:** this document explains where to start, which document owns each subject, recommended reading paths, documentation authority, runtime-stage navigation, and maintenance expectations. It summarizes and links; it does not replace the specialized references.

---

## 1. Project at a glance

`fish-audio-cli` is a single-run command-line client for Fish Audio text-to-speech.

One normal invocation performs:

```text
command-line parsing
    ↓
configuration loading and validation
    ↓
ordered local text-processing pipeline
    ↓
Fish API request
    ↓
streamed audio response
    ↓
atomic output-file publication
```

The program is designed for:

- scripts;
- bots;
- automation systems;
- scheduled jobs;
- applications that need a small TTS executable;
- integrations that do not want a persistent local service.

The current project status is pre-release alpha development.

The currently registered built-in module type is:

```text
passthrough
```

It leaves text unchanged.

---

## 2. Start here

Choose the path that matches the task.

| Task | Start with |
|---|---|
| Understand what the program does | [`../README.md`](../README.md) |
| Run the command correctly | [`cli.md`](cli.md) |
| Configure every JSON field | [`configuration.md`](configuration.md) |
| Diagnose a failed invocation | [`troubleshooting.md`](troubleshooting.md) |
| Interpret exit status or error stage | [`errors-and-exit-codes.md`](errors-and-exit-codes.md) |
| Understand the full runtime design | [`architecture.md`](architecture.md) |
| Configure module order and fallback | [`pipeline.md`](pipeline.md) |
| Understand module boundaries | [`modules.md`](modules.md) |
| Implement a new module | [`module-author-guide.md`](module-author-guide.md) |
| Configure Fish requests and retries | [`fish-audio.md`](fish-audio.md) |
| Deploy paths, secrets, and permissions | [`secrets-and-paths.md`](secrets-and-paths.md) |
| Collect and interpret logs | [`logging.md`](logging.md) |
| Understand output atomicity | [`output-files.md`](output-files.md) |
| Run tests or reproduce CI | [`testing.md`](testing.md) |

For an operational failure, begin with [`troubleshooting.md`](troubleshooting.md), not with architecture archaeology.

---

## 3. Documentation authority

The documentation set uses specialized ownership.

A document is authoritative for the subject named in its status and scope.

When two documents discuss the same behavior:

1. use the more specialized subsystem document;
2. use the broader document for context;
3. treat a disagreement with tested implementation as a defect;
4. do not invent a compromise between contradictory descriptions.

### 3.1 Specialized reference wins

Examples:

| Question | Primary authority |
|---|---|
| Exact CLI flag behavior | [`cli.md`](cli.md) |
| Exact JSON field or default | [`configuration.md`](configuration.md) |
| Pipeline rollback | [`pipeline.md`](pipeline.md) |
| Module preparation contract | [`modules.md`](modules.md) |
| Module implementation procedure | [`module-author-guide.md`](module-author-guide.md) |
| Fish retry classification | [`fish-audio.md`](fish-audio.md) |
| Log destination behavior | [`logging.md`](logging.md) |
| Secret file hardening | [`secrets-and-paths.md`](secrets-and-paths.md) |
| Rename and directory sync | [`output-files.md`](output-files.md) |
| Exit-code mapping | [`errors-and-exit-codes.md`](errors-and-exit-codes.md) |
| Required CI commands | [`testing.md`](testing.md) |
| Symptom-driven recovery | [`troubleshooting.md`](troubleshooting.md) |

### 3.2 Architecture supplies ownership

[`architecture.md`](architecture.md) explains:

- which package owns each stage;
- where validation occurs;
- how components are composed;
- which dependencies are intentionally narrow;
- which lifecycle phases exist;
- which responsibilities do not belong in core.

Use it to understand boundaries.

Use the specialized reference for exact values and subsystem rules.

### 3.3 Troubleshooting is practical, not overriding

[`troubleshooting.md`](troubleshooting.md) translates current behavior into checks and corrective actions.

It does not redefine:

- CLI syntax;
- configuration semantics;
- pipeline policy;
- secret security;
- Fish protocol behavior;
- output durability;
- test requirements.

When troubleshooting guidance and a normative subsystem reference appear inconsistent, investigate the mismatch.

### 3.4 README is the public overview

[`../README.md`](../README.md) is the repository front page and quick introduction.

It is intentionally shorter than the subsystem references.

During development, exact behavior belongs in the specialized documents listed here.

---

## 4. Recommended reading paths

The complete set is large because the program has explicit behavior at every boundary. Nobody is required to read every page before generating one audio file. Civilization can survive a staged curriculum.

---

## 5. First-time user path

Read in this order:

1. [`../README.md`](../README.md)
2. [`cli.md`](cli.md)
3. [`configuration.md`](configuration.md)
4. [`secrets-and-paths.md`](secrets-and-paths.md)
5. [`troubleshooting.md`](troubleshooting.md)

This path answers:

- how to build;
- how to provide text;
- how to choose a format;
- how to select output;
- where configuration lives;
- where the Fish key lives;
- what to do after a failure.

For initial use, the pipeline can remain the default `passthrough` module.

---

## 6. Shell automation path

Read:

1. [`cli.md`](cli.md)
2. [`errors-and-exit-codes.md`](errors-and-exit-codes.md)
3. [`output-files.md`](output-files.md)
4. [`logging.md`](logging.md)
5. [`troubleshooting.md`](troubleshooting.md)

Important automation facts include:

- `--output` is required;
- relative output follows process working directory;
- audio is not written to stdout;
- exit codes identify broad stages;
- exit `4` can coexist with a newly published output;
- configured logs go to stderr and one persistent file;
- module errors can be recovered and still end with exit `0`;
- warnings can accompany a nonzero exit;
- blind retry after synthesis failure can duplicate remote work.

---

## 7. Operator and deployment path

Read:

1. [`secrets-and-paths.md`](secrets-and-paths.md)
2. [`logging.md`](logging.md)
3. [`output-files.md`](output-files.md)
4. [`configuration.md`](configuration.md)
5. [`errors-and-exit-codes.md`](errors-and-exit-codes.md)
6. [`troubleshooting.md`](troubleshooting.md)

This path covers:

- absolute and relative paths;
- project-directory derivation;
- secret file creation;
- permissions;
- symlink behavior;
- log-file ownership and rotation;
- output directory requirements;
- atomic replacement;
- filesystem durability;
- service identity;
- container mounts;
- retry-safe operations.

Use absolute config and output paths in service definitions.

---

## 8. Pipeline configuration path

Read:

1. [`pipeline.md`](pipeline.md)
2. [`modules.md`](modules.md)
3. [`configuration.md`](configuration.md)
4. [`logging.md`](logging.md)
5. [`errors-and-exit-codes.md`](errors-and-exit-codes.md)

This path explains:

- strict array order;
- repeated module types;
- unique instance names;
- default and per-instance error policy;
- rollback;
- `use_previous`;
- `use_original`;
- `skip`;
- `abort`;
- cancellation;
- step reports;
- module lifecycle logging.

---

## 9. Module author path

Required reading:

1. [`architecture.md`](architecture.md)
2. [`pipeline.md`](pipeline.md)
3. [`modules.md`](modules.md)
4. [`module-author-guide.md`](module-author-guide.md)
5. [`configuration.md`](configuration.md)
6. [`testing.md`](testing.md)
7. [`logging.md`](logging.md)
8. [`errors-and-exit-codes.md`](errors-and-exit-codes.md)

Read [`secrets-and-paths.md`](secrets-and-paths.md) when the module owns files.

Read [`fish-audio.md`](fish-audio.md) when the module communicates over HTTP and needs comparable retry, error, or privacy design.

A production module must preserve:

- strict instance-owned configuration;
- prepare-before-build ordering;
- independent processor instances;
- valid nonblank UTF-8 text;
- cancellation;
- wrapped error identity;
- pipeline rollback assumptions;
- safe logging;
- deterministic tests;
- registry integration.

---

## 10. Maintainer and reviewer path

Read the full set in this order:

1. [`architecture.md`](architecture.md)
2. [`cli.md`](cli.md)
3. [`configuration.md`](configuration.md)
4. [`pipeline.md`](pipeline.md)
5. [`modules.md`](modules.md)
6. [`module-author-guide.md`](module-author-guide.md)
7. [`fish-audio.md`](fish-audio.md)
8. [`secrets-and-paths.md`](secrets-and-paths.md)
9. [`output-files.md`](output-files.md)
10. [`logging.md`](logging.md)
11. [`errors-and-exit-codes.md`](errors-and-exit-codes.md)
12. [`testing.md`](testing.md)
13. [`troubleshooting.md`](troubleshooting.md)

This order moves from ownership to public interfaces, execution, external boundaries, diagnostics, verification, and operations.

---

## 11. Failure-triage path

Read:

1. [`troubleshooting.md`](troubleshooting.md)
2. [`errors-and-exit-codes.md`](errors-and-exit-codes.md)
3. the subsystem document named by the final error stage.

Stage map:

| Final event | Next document |
|---|---|
| `option parsing failed` | [`cli.md`](cli.md) |
| `path initialization failed` | [`secrets-and-paths.md`](secrets-and-paths.md) |
| `config loading failed` | [`configuration.md`](configuration.md) |
| `config validation failed` | [`configuration.md`](configuration.md) |
| `logger initialization failed` | [`logging.md`](logging.md) |
| `module initialization failed` | [`modules.md`](modules.md) |
| `input failed` | [`cli.md`](cli.md) |
| `text processing failed` | [`pipeline.md`](pipeline.md) |
| `Fish request creation failed` | [`fish-audio.md`](fish-audio.md) |
| `empty secret file created` | [`secrets-and-paths.md`](secrets-and-paths.md) |
| `Fish API key loading failed` | [`secrets-and-paths.md`](secrets-and-paths.md) |
| `Fish client initialization failed` | [`fish-audio.md`](fish-audio.md) |
| `synthesis failed` with HTTP cause | [`fish-audio.md`](fish-audio.md) |
| `synthesis failed` with file cause | [`output-files.md`](output-files.md) |
| `log file closing failed` | [`logging.md`](logging.md) |

---

## 12. Catalog: architecture

### [`architecture.md`](architecture.md)

**Title:** Architecture

**Primary audience:**

- maintainers;
- reviewers;
- module authors;
- operators needing the complete model.

**Authoritative for:**

- process lifecycle;
- package ownership;
- dependency direction;
- composition root;
- validation boundaries;
- core versus module responsibilities;
- cancellation propagation;
- request flow;
- runtime guarantees;
- deliberate non-goals.

**Read it when:**

- deciding where new behavior belongs;
- reviewing a cross-package change;
- tracing startup order;
- understanding why a package receives a narrow interface;
- evaluating whether a feature belongs in core or a module.

**Do not use it as the sole source for:**

- exact JSON ranges;
- exact CLI syntax;
- complete retry rules;
- complete file-permission behavior.

Use the relevant specialized reference.

---

## 13. Catalog: command-line interface

### [`cli.md`](cli.md)

**Title:** Command-line interface

**Primary audience:**

- direct users;
- shell-script authors;
- service integrators;
- maintainers preserving CLI compatibility.

**Authoritative for:**

- command syntax;
- `--config`;
- `--text`;
- stdin selection;
- `--format`;
- `ogg` normalization;
- `--output`;
- positional-argument rejection;
- stdout and stderr use;
- signal handling at command level;
- CLI-visible status behavior.

**Read it when:**

- writing a command invocation;
- integrating with a shell;
- deciding whether text comes from argument or stdin;
- choosing an output format;
- changing a flag.

**Related documents:**

- JSON fields: [`configuration.md`](configuration.md)
- broad exit stages: [`errors-and-exit-codes.md`](errors-and-exit-codes.md)
- output publication: [`output-files.md`](output-files.md)

---

## 14. Catalog: configuration

### [`configuration.md`](configuration.md)

**Title:** Configuration reference

**Primary audience:**

- users;
- operators;
- maintainers;
- module authors documenting settings.

**Authoritative for:**

- top-level JSON structure;
- exact field names;
- defaults;
- strict JSON behavior;
- unknown fields;
- duplicate keys;
- explicit `null`;
- ranges;
- cross-field validation;
- pipeline instance schema;
- Fish settings;
- retry settings;
- secret settings;
- logging settings.

**Read it when:**

- editing `config.json`;
- reviewing `config.example.json`;
- adding a config field;
- interpreting a validation error;
- checking a default or allowed range.

**Related documents:**

- path meaning: [`secrets-and-paths.md`](secrets-and-paths.md)
- pipeline runtime: [`pipeline.md`](pipeline.md)
- Fish HTTP meaning: [`fish-audio.md`](fish-audio.md)

---

## 15. Catalog: text-processing pipeline

### [`pipeline.md`](pipeline.md)

**Title:** Text-processing pipeline

**Primary audience:**

- users configuring module chains;
- operators interpreting outcomes;
- module authors;
- maintainers.

**Authoritative for:**

- ordered execution;
- document state;
- original and current text;
- rollback;
- policy parsing;
- `use_previous`;
- `use_original`;
- `skip`;
- `abort`;
- cancellation;
- invalid module output;
- reports;
- step counts;
- pipeline outcomes;
- logging integration.

**Read it when:**

- a module failure was recovered;
- later modules did not run;
- text unexpectedly reverted;
- adding a new error policy;
- changing report behavior.

---

## 16. Catalog: module system

### [`modules.md`](modules.md)

**Title:** Module system

**Primary audience:**

- users configuring instances;
- maintainers;
- developers deciding extension boundaries.

**Authoritative for:**

- module type versus instance;
- module-owned configuration;
- registry lookup;
- preparation;
- processor builders;
- prepare-all-before-build;
- independent instances;
- lifecycle limitations;
- current built-in modules;
- security and ownership boundaries.

**Read it when:**

- a type is unsupported;
- repeated module types are configured;
- deciding whether behavior belongs in a module;
- reviewing a module’s resource lifecycle;
- understanding registry errors.

---

## 17. Catalog: module author guide

### [`module-author-guide.md`](module-author-guide.md)

**Title:** Module author guide

**Primary audience:**

- Go developers adding a built-in text-processing module.

**Authoritative for:**

- implementation sequence;
- package structure;
- strict config decoding;
- semantic validation;
- path resolution;
- preparation;
- processor construction;
- `Process`;
- cancellation;
- testing;
- registry registration;
- documentation and review checklist.

**Prerequisites:**

- [`architecture.md`](architecture.md)
- [`pipeline.md`](pipeline.md)
- [`modules.md`](modules.md)

**Read it when:**

- creating a new module package;
- converting a prototype into production structure;
- deciding what belongs in `Prepare`;
- writing module tests;
- registering a module type.

---

## 18. Catalog: Fish Audio integration

### [`fish-audio.md`](fish-audio.md)

**Title:** Fish Audio integration

**Primary audience:**

- operators configuring synthesis;
- maintainers of HTTP behavior;
- developers changing request validation or retry logic.

**Authoritative for:**

- endpoint construction;
- `/v1/tts` joining;
- authorization header;
- model header;
- request mapping;
- request validation;
- HTTP timeout;
- retries;
- `429`;
- optional `5xx` retry;
- `Retry-After`;
- typed API errors;
- bounded error bodies;
- response streaming;
- empty response rejection;
- cancellation;
- protocol security.

**Read it when:**

- Fish returns `400`, `401`, `402`, `403`, `404`, `422`, `429`, or `5xx`;
- synthesis appears to pause between attempts;
- changing a request field;
- adding retry behavior;
- diagnosing a stream failure.

---

## 19. Catalog: secrets and paths

### [`secrets-and-paths.md`](secrets-and-paths.md)

**Title:** Secrets and path resolution

**Primary audience:**

- operators;
- service integrators;
- module authors owning paths;
- security reviewers.

**Authoritative for:**

- config path normalization;
- project-directory derivation;
- lexical path resolution;
- absolute and relative paths;
- config symlink behavior;
- Fish API key path;
- secret directory creation;
- directory write-bit checks;
- secret file creation;
- mode `0600`;
- symlink rejection;
- race checking;
- one-line UTF-8 secret format;
- read-only mount limitations;
- repository ignore behavior;
- differences among config, secret, log, module, and output paths.

**Read it when:**

- a secret file is created;
- a secret path appears in an unexpected directory;
- deployment uses containers or services;
- a chmod or symlink error occurs;
- designing module-owned file settings.

---

## 20. Catalog: output files

### [`output-files.md`](output-files.md)

**Title:** Output files and atomic publication

**Primary audience:**

- CLI users;
- automation integrators;
- operators;
- maintainers of output and streaming.

**Authoritative for:**

- output path interpretation;
- parent-directory responsibility;
- temporary filename;
- mode `0600`;
- write callback;
- stream-to-temp behavior;
- temp sync;
- temp close;
- rename;
- directory sync;
- cleanup;
- existing destination preservation;
- destination symlink replacement;
- stale temp files;
- concurrency;
- crash windows;
- published-but-error state.

**Read it when:**

- output is missing;
- an old output survived failure;
- exit `4` occurred but a new file exists;
- a hidden temp file remains;
- two jobs target one path;
- changing atomic publication.

---

## 21. Catalog: logging

### [`logging.md`](logging.md)

**Title:** Logging

**Primary audience:**

- operators;
- log collectors;
- service integrators;
- maintainers;
- privacy reviewers.

**Authoritative for:**

- bootstrap logger;
- configured logger;
- stderr destination;
- persistent file destination;
- request ID;
- levels;
- text and JSON formats;
- lifecycle messages;
- module logging;
- optional input/output text fields;
- log file path;
- directory creation;
- mode `0640`;
- append behavior;
- runtime writer failures;
- close failure;
- rotation expectations;
- unsupported disable behavior.

**Read it when:**

- logs are duplicated;
- early records are text while later records are JSON;
- no persistent log exists;
- text appears or does not appear;
- adding an event or field;
- configuring rotation.

---

## 22. Catalog: errors and exit codes

### [`errors-and-exit-codes.md`](errors-and-exit-codes.md)

**Title:** Errors and exit codes

**Primary audience:**

- CLI users;
- shell-script authors;
- service integrators;
- operators;
- module authors;
- maintainers.

**Authoritative for:**

- application exit codes `0` through `4`;
- stage boundaries;
- final lifecycle messages;
- stderr versus persistent logging;
- recovered module errors;
- warning with nonzero exit;
- Fish error categories;
- cancellation mapping;
- wrapped errors;
- joined errors;
- retry safety;
- output state after failure;
- signal-derived external statuses.

**Read it when:**

- writing status-based automation;
- interpreting a nonzero exit;
- deciding whether retry is safe;
- preserving typed error identity;
- changing a return-code boundary.

---

## 23. Catalog: testing

### [`testing.md`](testing.md)

**Title:** Testing

**Primary audience:**

- contributors;
- module authors;
- maintainers;
- operators reproducing failures.

**Authoritative for:**

- declared Go toolchain;
- CI workflow;
- `gofmt`;
- `go vet`;
- uncached tests;
- race detector;
- build gate;
- test placement;
- same-package tests;
- `t.Parallel`;
- process-global state;
- `t.TempDir`;
- `httptest.Server`;
- handwritten fakes;
- typed-nil tests;
- error assertions;
- filesystem side effects;
- module test expectations;
- optional coverage, shuffle, fuzzing, and benchmarks.

**Read it when:**

- preparing a commit;
- reproducing CI;
- adding tests;
- diagnosing a race or hang;
- adding a module;
- reviewing test quality.

---

## 24. Catalog: troubleshooting

### [`troubleshooting.md`](troubleshooting.md)

**Title:** Troubleshooting

**Primary audience:**

- users;
- operators;
- service integrators;
- module authors;
- maintainers diagnosing actual failures.

**Authoritative for:**

- symptom-to-stage navigation;
- safe diagnostic collection;
- common config mistakes;
- path surprises;
- secret remediation;
- Fish status diagnosis;
- retry observations;
- output ambiguity;
- stale temp inspection;
- service and container pitfalls;
- CI reproduction;
- bug-report contents.

**Read it when:**

- the command failed;
- the command appears stuck;
- output exists unexpectedly;
- logs are missing or duplicated;
- tests pass locally but fail in CI;
- preparing a safe bug report.

It is the practical entry point for failures.

---

## 25. Runtime-stage map

The program’s main stages map to documents as follows.

```text
CLI arguments
    → cli.md
    → errors-and-exit-codes.md

configuration file
    → configuration.md
    → secrets-and-paths.md

configured logging
    → logging.md

module preparation and building
    → modules.md
    → module-author-guide.md

input and pipeline execution
    → cli.md
    → pipeline.md

Fish request and HTTP
    → fish-audio.md

secret loading
    → secrets-and-paths.md

output publication
    → output-files.md

final status and recovery
    → errors-and-exit-codes.md
    → troubleshooting.md

repository verification
    → testing.md
```

---

## 26. Ownership map

| Concern | Owner document |
|---|---|
| Overall component boundaries | [`architecture.md`](architecture.md) |
| CLI flags and streams | [`cli.md`](cli.md) |
| JSON schema and defaults | [`configuration.md`](configuration.md) |
| Text execution and fallback | [`pipeline.md`](pipeline.md) |
| Extension model | [`modules.md`](modules.md) |
| New module implementation | [`module-author-guide.md`](module-author-guide.md) |
| Fish request and response | [`fish-audio.md`](fish-audio.md) |
| Path and secret filesystem rules | [`secrets-and-paths.md`](secrets-and-paths.md) |
| Atomic audio publication | [`output-files.md`](output-files.md) |
| Observability | [`logging.md`](logging.md) |
| Error stages and process status | [`errors-and-exit-codes.md`](errors-and-exit-codes.md) |
| Repository verification | [`testing.md`](testing.md) |
| Practical diagnosis | [`troubleshooting.md`](troubleshooting.md) |

---

## 27. Path quick map

Relative paths do not all share one base.

| Path | Relative base |
|---|---|
| `--config` | process working directory |
| `secrets.fishApiKeyFile` | derived project directory |
| `logging.file` | derived project directory |
| module-owned configured path | module contract, normally project resolver |
| `--output` | process working directory |

Primary reference:

[`secrets-and-paths.md`](secrets-and-paths.md)

Output-specific behavior:

[`output-files.md`](output-files.md)

---

## 28. Error-code quick map

| Exit | Stage |
|---:|---|
| `0` | help or successful completion |
| `1` | bootstrap diagnostics or request ID |
| `2` | invocation, config, initialization, input |
| `3` | pipeline, request, secret, Fish client setup |
| `4` | Fish HTTP, streaming, output publication |

Primary reference:

[`errors-and-exit-codes.md`](errors-and-exit-codes.md)

Practical diagnosis:

[`troubleshooting.md`](troubleshooting.md)

---

## 29. Pipeline-policy quick map

| Policy | On module failure |
|---|---|
| `use_previous` | restore text from before the failing step and continue |
| `use_original` | restore original pipeline input and continue |
| `skip` | restore pre-step text and stop later modules successfully |
| `abort` | restore pre-step text and return an error |

Primary reference:

[`pipeline.md`](pipeline.md)

Configuration schema:

[`configuration.md`](configuration.md)

---

## 30. Fish retry quick map

| Condition | Internal retry |
|---|---|
| `429` | yes, within configured attempt and delay limits |
| `5xx` | only when enabled |
| DNS failure | no |
| connection refusal | no |
| TLS failure | no |
| transport timeout | no |
| response stream failure after 2xx | no |

Primary reference:

[`fish-audio.md`](fish-audio.md)

Operational guidance:

[`troubleshooting.md`](troubleshooting.md)

---

## 31. File-mode quick map

| Artifact | Current mode behavior |
|---|---|
| secret file | forced to `0600` |
| output file | created/published as `0600` |
| log file | forced to `0640` |
| missing secret directory | requested as `0700`, subject to umask |
| missing log directory | requested as `0750`, subject to umask |

Primary references:

- secrets: [`secrets-and-paths.md`](secrets-and-paths.md)
- output: [`output-files.md`](output-files.md)
- logs: [`logging.md`](logging.md)

---

## 32. Current public capabilities

The documented current capabilities include:

- one text input per invocation;
- `--text` or stdin;
- strict JSON configuration;
- ordered compiled-in text modules;
- `passthrough` built-in module;
- configurable pipeline error policies;
- Fish Audio synthesis;
- WAV;
- MP3;
- Opus;
- `ogg` CLI alias for Opus;
- separate protected API key file;
- structured text or JSON logs;
- request correlation;
- bounded reads;
- retry for rate limiting;
- optional retry for server errors;
- atomic output replacement;
- race-tested Go implementation.

Exact details belong in the specialized references.

---

## 33. Current deliberate limitations

The current documentation describes a one-shot CLI, not:

- a daemon;
- a local HTTP service;
- a plugin loader;
- a dynamic shared-library system;
- stdout audio streaming;
- inline API keys;
- environment-variable interpolation in JSON;
- automatic output naming;
- output directory creation;
- live configuration reload;
- internal log rotation;
- destination locking;
- provider idempotency;
- current live model catalog discovery.

Do not infer these features from adjacent behavior.

---

## 34. Document status terminology

### Normative description

A normative document states the intended current contract.

A disagreement with implementation or tests is a defect to resolve.

### Normative reference

A normative reference defines exact public values, fields, or commands.

### Normative implementation guide

A normative guide defines the required implementation process for contributors.

### Practical guide

A practical guide converts normative behavior into operational steps.

It does not override the underlying contract.

---

## 35. Reading exact values

Use these documents for exact values.

| Value type | Document |
|---|---|
| CLI flag and accepted format | [`cli.md`](cli.md) |
| JSON field and default | [`configuration.md`](configuration.md) |
| Pipeline policy | [`pipeline.md`](pipeline.md) |
| Fish validation and retry | [`fish-audio.md`](fish-audio.md) |
| File permission | subsystem file document |
| Exit code | [`errors-and-exit-codes.md`](errors-and-exit-codes.md) |
| CI command | [`testing.md`](testing.md) |

Do not rely on a passing mention in a broad overview when a specialized table exists.

---

## 36. Reading examples

Examples demonstrate one supported use.

They do not silently expand the formal contract.

For example:

- a sample output filename does not define format inference;
- a sample relative path does not make every path cwd-relative;
- a sample module config does not permit unknown fields;
- a sample retry value does not replace the documented range.

Read the surrounding rule and validation section.

---

## 37. Reading errors

Error strings contain contextual layers.

A typical chain can be:

```text
write temporary output file
    ↓
send synthesis request
    ↓
context canceled
```

Use:

- the final lifecycle message for broad stage;
- the full wrapped or joined error for cause;
- the subsystem reference for semantics;
- the exit status for automation.

See:

- [`errors-and-exit-codes.md`](errors-and-exit-codes.md)
- [`troubleshooting.md`](troubleshooting.md)

---

## 38. Reading logs

A configured invocation uses:

```text
stderr
+
persistent log file
```

Early failures can be stderr-only.

Module ERROR records can be recovered by pipeline policy.

A WARN record can accompany exit `3`.

A deferred log close ERROR does not alter the selected status.

See:

- [`logging.md`](logging.md)
- [`errors-and-exit-codes.md`](errors-and-exit-codes.md)

---

## 39. Reading filesystem state

A failed invocation can still change the filesystem.

Examples:

- missing secret creates an empty `0600` file;
- cleanup failure can leave a hidden temp;
- directory-sync failure can return exit `4` after publishing output;
- successful output replaces existing metadata with a new `0600` file.

See:

- [`secrets-and-paths.md`](secrets-and-paths.md)
- [`output-files.md`](output-files.md)
- [`troubleshooting.md`](troubleshooting.md)

---

## 40. Source of truth during development

The intended consistency set is:

```text
implementation
tests
configuration example
normative subsystem documentation
README and navigation
```

These should agree.

A mismatch is not resolved by choosing whichever text is convenient.

During review:

1. identify the actual intended contract;
2. update code or tests when behavior is wrong;
3. update every affected document;
4. add a regression test for behavioral defects;
5. rerun the required verification suite.

---

## 41. Documentation update matrix

When changing a subsystem, review the listed documents.

| Change | Review |
|---|---|
| CLI flag or argument behavior | `cli`, `errors`, `troubleshooting`, README |
| JSON field/default/range | `configuration`, subsystem doc, troubleshooting, README |
| Pipeline policy or report | `pipeline`, `errors`, `logging`, `testing` |
| Module lifecycle | `modules`, author guide, architecture, testing |
| Fish request/retry/error | `fish-audio`, configuration, errors, troubleshooting, testing |
| Secret or path behavior | `secrets-and-paths`, configuration, troubleshooting, testing |
| Output publication | `output-files`, errors, troubleshooting, testing |
| Log event or field | `logging`, errors, troubleshooting, testing |
| Exit code | `errors`, CLI, troubleshooting, README |
| CI command | `testing`, contributor material, README |
| New document | this index and relevant cross-links |

---

## 42. Adding a new module document trail

A new built-in module normally requires updates to:

- module package tests;
- registry;
- configuration example;
- [`modules.md`](modules.md);
- [`module-author-guide.md`](module-author-guide.md) when the general pattern changes;
- [`configuration.md`](configuration.md);
- [`testing.md`](testing.md);
- [`troubleshooting.md`](troubleshooting.md) for module-specific operational failures when justified;
- this index when a dedicated document is added;
- README feature summary.

Do not document a compiled-in module as dynamically loadable.

---

## 43. Adding a new configuration field

Before adding the field, decide:

- owner package;
- default;
- omitted behavior;
- explicit `null` behavior;
- type;
- allowed range;
- whitespace rules;
- path semantics;
- security boundary;
- logging exposure;
- tests;
- compatibility impact.

Then update:

- Go config types;
- defaults;
- strict decoding tests;
- validation tests;
- example config;
- [`configuration.md`](configuration.md);
- owning subsystem document;
- troubleshooting where operators need remediation.

---

## 44. Adding a new exit code

A new application-defined exit code is a CLI compatibility change.

Review:

- [`cli.md`](cli.md);
- [`errors-and-exit-codes.md`](errors-and-exit-codes.md);
- [`troubleshooting.md`](troubleshooting.md);
- command tests;
- service integration guidance;
- README.

Do not add a numeric code merely to avoid carrying a typed error internally.

---

## 45. Adding a new log event

Define:

- message;
- level;
- stage;
- request correlation;
- fields;
- privacy;
- duplication expectations;
- failure behavior;
- compatibility for collectors.

Update:

- [`logging.md`](logging.md);
- [`errors-and-exit-codes.md`](errors-and-exit-codes.md) when it marks a failure stage;
- [`troubleshooting.md`](troubleshooting.md);
- tests.

---

## 46. Adding a new output mode

A new output mode must define:

- destination;
- partial-data semantics;
- retry compatibility;
- atomicity;
- durability;
- cleanup;
- stdout/stderr separation;
- exit behavior;
- logging;
- concurrency.

File output currently owns a specific atomic contract.

Stdout streaming would be a separate contract rather than a filename spelling trick.

---

## 47. Testing documentation claims

Behavioral claims should be traceable to tests where practical.

High-priority claims include:

- strict JSON;
- exact limits;
- error policies;
- typed-nil rejection;
- secret hardening;
- Fish retry;
- response-body closure;
- atomic output;
- exit codes;
- log fields;
- CI gates.

See [`testing.md`](testing.md).

---

## 48. Local documentation links

Documents in `docs/` use relative links.

Examples:

```text
configuration.md
pipeline.md
../README.md
```

A document move or rename requires updating every inbound link.

The documentation installation workflow validates that local link targets exist.

---

## 49. Documentation language and terminology

Use project terms consistently:

```text
Fish Audio
Fish API key
module type
module instance
processor
preparer
processor builder
project directory
pipeline outcome
request ID
temporary output file
published output
```

Avoid introducing synonyms that imply a different architecture, such as:

```text
runtime plugin
daemon worker
streaming stdout mode
environment secret fallback
```

unless those features are actually designed and implemented.

---

## 50. Key distinctions

Several distinctions recur across the documentation.

### Type versus instance

One module type can appear in multiple configured instances.

### Prepare versus build

All instances prepare before any processor is built.

### Failure versus process failure

A module can fail and be recovered by policy.

### Atomicity versus durability

Rename visibility and directory persistence are separate.

### Relative path versus project-relative path

Output and config arguments differ from configured secret and log paths.

### Log severity versus exit status

ERROR can coexist with exit `0`; WARN can coexist with exit `3`.

### Local validation versus remote authority

A request can pass local validation and still be rejected by Fish.

---

## 51. Frequently confused questions

### Where is the API key?

Start with [`secrets-and-paths.md`](secrets-and-paths.md).

The default configured path is project-relative.

### Why did `--text ""` wait?

Read [`cli.md`](cli.md).

Exact empty text selects stdin.

### Why did a module error still produce audio?

Read [`pipeline.md`](pipeline.md).

The configured policy recovered.

### Why did exit `4` leave a file?

Read [`output-files.md`](output-files.md).

Rename may have succeeded before directory persistence failed.

### Why are logs duplicated?

Read [`logging.md`](logging.md).

Configured records go to stderr and file.

### Why was `5xx` not retried?

Read [`fish-audio.md`](fish-audio.md).

Server retries are optional and disabled by default.

### Why did a secret symlink fail?

Read [`secrets-and-paths.md`](secrets-and-paths.md).

The secret leaf must be a regular file.

### Why did plain `go test ./...` pass but CI fail?

Read [`testing.md`](testing.md).

CI uses uncached tests and the race detector.

---

## 52. Quick verification commands

Run the current required local checks:

```bash
clear

gofmt -w .
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
go build \
  -trimpath \
  -o /tmp/fish-audio-cli \
  ./cmd/fish-audio-cli
```

See [`testing.md`](testing.md) for the complete contract and optional diagnostic commands.

---

## 53. Quick failure capture

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
cat /tmp/fish-audio-cli.stderr
```

Then open:

- [`troubleshooting.md`](troubleshooting.md);
- [`errors-and-exit-codes.md`](errors-and-exit-codes.md);
- the subsystem reference named by the last message.

Never include the API key in a diagnostic transcript.

---

## 54. Documentation maintenance order

When the implementation changes, update in this order:

1. code;
2. focused tests;
3. full tests;
4. owning subsystem reference;
5. related cross-cutting references;
6. troubleshooting;
7. this index when navigation changes;
8. README summary;
9. contributor, security, and release material as applicable.

This order keeps the detailed contract stable before producing summaries.

---

## 55. Index maintenance rules

Update this index when:

- a document is added;
- a document is renamed;
- a document is removed;
- ownership moves between documents;
- a new audience path is needed;
- a new subsystem gets a dedicated reference;
- the documentation authority model changes.

Do not copy entire subsystem tables into the index.

Summarize and link.

The index should remain navigable even as the specialized documents become detailed enough to frighten lesser table-of-contents generators.

---

## 56. Current document set

| Document | Role |
|---|---|
| [`architecture.md`](architecture.md) | overall runtime and ownership |
| [`cli.md`](cli.md) | command-line contract |
| [`configuration.md`](configuration.md) | exact JSON reference |
| [`pipeline.md`](pipeline.md) | ordered text execution |
| [`modules.md`](modules.md) | extension model |
| [`module-author-guide.md`](module-author-guide.md) | module implementation procedure |
| [`fish-audio.md`](fish-audio.md) | Fish HTTP boundary |
| [`secrets-and-paths.md`](secrets-and-paths.md) | filesystem paths and secret security |
| [`output-files.md`](output-files.md) | atomic audio publication |
| [`logging.md`](logging.md) | structured observability |
| [`errors-and-exit-codes.md`](errors-and-exit-codes.md) | failure stages and statuses |
| [`testing.md`](testing.md) | tests and CI |
| [`troubleshooting.md`](troubleshooting.md) | practical diagnosis |

Count:

```text
13 subsystem and operational documents
```

This index is the navigation layer above them.

---

## 57. Documentation invariants

The following rules define the current documentation structure.

1. This file is the entry point for detailed documentation.
2. README remains the public repository overview.
3. Architecture owns broad runtime boundaries.
4. CLI owns command syntax.
5. Configuration owns exact JSON fields and defaults.
6. Pipeline owns ordered execution and recovery.
7. Modules owns the extension model.
8. Module author guide owns the implementation procedure.
9. Fish Audio owns HTTP and provider behavior.
10. Secrets and paths owns project paths and key files.
11. Output files owns atomic publication.
12. Logging owns destinations and event behavior.
13. Errors and exit codes owns stage-to-status mapping.
14. Testing owns required local and CI verification.
15. Troubleshooting owns symptom-driven operational guidance.
16. Specialized references override broad summaries on detail.
17. Practical guidance does not override normative contracts.
18. A documentation and implementation disagreement is a defect.
19. Relative links must remain valid.
20. New documents must be added to the catalog.
21. Removed documents must be removed from every route and table.
22. Public interface changes require documentation updates.
23. Security-sensitive claims require tests where practical.
24. Exact values belong in the owning reference.
25. This index should summarize rather than duplicate.

---

## 58. Non-goals

This index does not provide:

- a full quick-start tutorial;
- a duplicate configuration reference;
- a duplicate API reference;
- every error string;
- complete module implementation code;
- current Fish commercial model availability;
- provider account documentation;
- installation packaging for every operating system;
- release notes;
- contribution policy;
- vulnerability reporting policy.

Those subjects belong in their own documents or external provider material.

---

## 59. Summary

Use the documentation by task:

```text
run it
    → README
    → CLI
    → configuration

operate it
    → secrets and paths
    → logging
    → output files

extend it
    → architecture
    → pipeline
    → modules
    → module author guide
    → testing

diagnose it
    → troubleshooting
    → errors and exit codes
    → owning subsystem reference
```

The shortest reliable rule is:

```text
start with the task-specific document
use architecture for context
use troubleshooting for symptoms
use testing before changing behavior
```

The documentation set is deliberately explicit because a TTS command that touches credentials, networks, modules, logs, and atomic files has more failure boundaries than its small binary would like to admit.
