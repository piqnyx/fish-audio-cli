# Fish Audio integration

> **Document status:** normative description of the current pre-release Fish Audio client boundary.
>
> **Audience:** operators configuring Fish synthesis, maintainers reviewing HTTP behavior, and developers changing request validation, retry logic, or API error handling.
>
> **Scope:** this document describes endpoint construction, authentication, model selection, request creation, JSON mapping, local validation, HTTP behavior, retries, API errors, response streaming, cancellation, security boundaries, and compatibility constraints. Configuration values and ranges are documented in [`configuration.md`](configuration.md); command invocation in [`cli.md`](cli.md); overall ownership in [`architecture.md`](architecture.md).

---

## 1. Purpose

The Fish Audio layer converts final processed text into one audio response.

Its boundary is:

```text
valid processed text
    +
validated Fish configuration
    +
selected output format
    +
Fish API key
    ↓
validated POST /v1/tts request
    ↓
streamed response bytes
```

The Fish layer owns:

- synthesis endpoint construction;
- client configuration;
- request JSON;
- authentication and model headers;
- HTTP request execution;
- API status classification;
- bounded API error bodies;
- response retry policy;
- response-body streaming;
- empty-response detection.

It does not own:

- command-line parsing;
- text input selection;
- local text-module execution;
- pipeline rollback;
- Fish API key file security;
- final destination publication;
- logging destinations;
- process exit-code selection.

The application composes those boundaries around the Fish client.

---

## 2. Runtime sequence

After the text pipeline succeeds, the application performs:

```text
processed text
    ↓
build and validate SynthesisRequest
    ↓
load Fish API key
    ↓
create Fish client
    ↓
marshal request JSON once
    ↓
send one or more permitted attempts
    ↓
stream successful response to atomic temp file
    ↓
publish final destination outside Fish client
```

Important ordering:

1. text processing completes before the API key is loaded;
2. request parameters are validated before network I/O;
3. the request body is encoded once before the retry loop;
4. retries occur only for selected non-success HTTP responses;
5. successful response bytes stream directly to the supplied writer;
6. the outer output layer decides whether temporary bytes become the final file.

---

## 3. Endpoint construction

The configured base URL is:

```text
fish.baseUrl
```

Default:

```text
https://api.fish.audio
```

The client appends:

```text
v1/tts
```

using URL path joining.

Default resolved endpoint:

```text
https://api.fish.audio/v1/tts
```

### 3.1 Accepted base URL

The base URL must:

- be nonblank after trimming;
- parse as a URL;
- use `http` or `https`;
- contain a hostname;
- contain no user information;
- contain no query string;
- contain no fragment.

A base path is allowed.

Example:

```text
https://proxy.example.test/fish
```

resolves to:

```text
https://proxy.example.test/fish/v1/tts
```

### 3.2 Scheme normalization

The scheme is lowercased.

Example:

```text
HTTPS://api.fish.audio
```

is normalized to an HTTPS URL.

### 3.3 Surrounding whitespace

The resolver trims the complete configured base URL before parsing.

A padded value such as:

```text
" https://api.fish.audio "
```

therefore resolves successfully.

This differs from model and API-key handling, where surrounding whitespace is rejected.

### 3.4 Rejected examples

User information:

```text
https://user:password@api.fish.audio
```

Query:

```text
https://api.fish.audio?region=test
```

Fragment:

```text
https://api.fish.audio#tts
```

Unsupported scheme:

```text
ftp://api.fish.audio
```

Missing hostname:

```text
https:///v1
```

### 3.5 HTTP versus HTTPS

The client accepts both:

```text
http
https
```

The default is HTTPS.

Plain HTTP may be useful for a trusted local test server or explicit proxy boundary, but it does not provide transport encryption.

Production operators should normally use HTTPS.

---

## 4. HTTP method and headers

Each synthesis attempt sends:

```text
POST <resolved endpoint>
```

The request includes:

```text
Authorization: Bearer <Fish API key>
Content-Type: application/json
model: <configured model>
```

### 4.1 Authorization

The API key is placed only in the `Authorization` header:

```text
Bearer <key>
```

It is not placed in:

- the request URL;
- query parameters;
- request JSON;
- command-line arguments;
- ordinary log fields.

### 4.2 Model header

The configured model is sent as an HTTP header named:

```text
model
```

It is not part of the request JSON body.

Changing:

```text
fish.model
```

changes this header.

### 4.3 Content type

The request body is JSON:

```text
Content-Type: application/json
```

The client does not currently set an `Accept` header.

### 4.4 No custom user agent

The client does not define a project-specific `User-Agent`.

The Go HTTP stack may provide its ordinary default behavior.

### 4.5 No idempotency header

The client does not currently send an idempotency key.

Retry safety is therefore limited to the explicitly supported response cases and the current API semantics.

Adding an idempotency mechanism would be a deliberate protocol change.

---

## 5. API key validation

The Fish client receives the already loaded secret string.

Before constructing the HTTP client, it requires:

- nonblank content;
- no leading or trailing whitespace;
- valid UTF-8;
- no ASCII control characters.

ASCII control characters include bytes:

```text
0x00 through 0x1f
0x7f
```

This rejects embedded:

- newlines;
- carriage returns;
- tabs;
- NUL;
- other header-breaking control bytes.

### 5.1 Secret file boundary

File creation, permissions, symlink checks, size limits, and one-line normalization are handled before the value reaches the Fish client.

The Fish client validates the final string again because it becomes an HTTP header value.

### 5.2 Empty key file

When the configured secret file is missing, the secrets layer creates it securely and reports that action.

The application stops before creating the Fish client.

After the operator writes exactly one API key line, the invocation can be repeated.

### 5.3 Key lifetime

The application clears its temporary local string variable after the client has retained the key.

Go strings are immutable and memory erasure is not guaranteed for every copy.

This reduces an unnecessary reference but is not a formal secure-memory guarantee.

---

## 6. Model validation

The configured model must:

- be nonblank;
- have no leading or trailing whitespace;
- be valid UTF-8;
- contain no ASCII control characters.

Examples rejected locally:

```text
""
" model-name"
"model-name "
"model\nname"
```

### 6.1 Model existence

The local client does not know the live Fish Audio model catalog.

A syntactically valid model can still be rejected by the remote API because of:

- availability;
- account access;
- billing;
- deprecation;
- provider-side validation.

The local contract validates header safety, not remote entitlement.

### 6.2 Default model

The built-in configuration currently uses:

```text
s2.1-pro-free
```

This is a project default, not a promise that the remote provider will keep that model available indefinitely.

---

## 7. Request construction

The application builds the Fish request from:

- `fish.request`;
- processed pipeline text;
- `fish.referenceId`;
- normalized CLI format.

Conceptually:

```go
request := cfg.Request.SynthesisRequest()
request.Text = processedText
request.ReferenceID = cfg.ReferenceID
request.Format = selectedFormat
request.Validate()
```

The request is validated before it reaches the HTTP client.

The Fish client validates it again before marshaling.

This creates two defensive boundaries:

```text
application request creation
    ↓
Fish client send boundary
```

---

## 8. JSON body

The request JSON uses Fish-style snake-case field names.

Conceptual complete body:

```json
{
  "text": "Привет",
  "reference_id": "voice-reference",
  "temperature": 0.7,
  "top_p": 0.7,
  "prosody": {
    "speed": 1.0,
    "volume": 0.0,
    "normalize_loudness": true
  },
  "chunk_length": 300,
  "normalize": true,
  "format": "opus",
  "sample_rate": null,
  "mp3_bitrate": 192,
  "opus_bitrate": 64000,
  "latency": "normal",
  "max_new_tokens": 1024,
  "repetition_penalty": 1.2,
  "min_chunk_length": 50,
  "condition_on_previous_chunks": true,
  "early_stop_threshold": 1.0,
  "features": []
}
```

Actual omission rules modify this shape.

---

## 9. Request-field mapping

| Config/runtime source | JSON field | Behavior |
|---|---|---|
| processed text | `text` | always included |
| `fish.referenceId` | `reference_id` | omitted when empty |
| `fish.request.temperature` | `temperature` | always included |
| `fish.request.topP` | `top_p` | always included |
| prosody object | `prosody` | always included |
| prosody speed | `prosody.speed` | always included |
| prosody volume | `prosody.volume` | always included |
| loudness flag | `prosody.normalize_loudness` | always included |
| `fish.request.chunkLength` | `chunk_length` | always included |
| `fish.request.normalize` | `normalize` | always included |
| normalized CLI format | `format` | always included |
| `fish.request.sampleRate` | `sample_rate` | included; `null` when unset |
| `fish.request.mp3Bitrate` | `mp3_bitrate` | always included |
| `fish.request.opusBitrate` | `opus_bitrate` | always included |
| `fish.request.latency` | `latency` | always included |
| `fish.request.maxNewTokens` | `max_new_tokens` | always included |
| repetition penalty | `repetition_penalty` | always included |
| minimum chunk length | `min_chunk_length` | always included |
| previous-chunk flag | `condition_on_previous_chunks` | always included |
| early-stop threshold | `early_stop_threshold` | always included |
| `fish.request.features` | `features` | omitted when empty |

### 9.1 Empty reference ID

When:

```json
"referenceId": ""
```

the JSON omits:

```text
reference_id
```

The remote API then applies its own behavior for a request without an explicit reference.

### 9.2 Empty feature array

When:

```json
"features": []
```

the JSON omits:

```text
features
```

An empty array is not sent.

### 9.3 Null sample rate

When:

```json
"sampleRate": null
```

the JSON includes:

```json
"sample_rate": null
```

`sample_rate` does not use an omission rule.

### 9.4 Mutable config values

The config-to-request conversion copies:

- `sampleRate`;
- `features`.

The resulting request does not intentionally share those mutable values with the configuration object.

---

## 10. Text validation

Request text must:

- be valid UTF-8;
- contain at least one non-whitespace Unicode code point.

The Fish request boundary does not trim text.

It preserves:

- leading spaces;
- trailing spaces;
- newlines;
- module-produced markup;
- punctuation.

A pipeline module that returns invalid or blank text should already fail earlier, but the request performs the shared check again.

### 10.1 Text size

The original CLI input is bounded by:

```text
input.maxBytes
```

The current Fish request layer does not impose a separate post-pipeline byte limit.

A module can increase text size after the input limit has been applied.

Modules capable of substantial expansion should define and document their own bound until a shared post-pipeline limit is deliberately added.

---

## 11. Reference ID validation

`fish.referenceId` becomes:

```text
reference_id
```

Local validation requires only valid UTF-8.

The value may be empty.

The request layer does not currently:

- trim it;
- reject surrounding whitespace;
- reject control characters;
- enforce a UUID shape;
- enforce a length;
- verify remote existence.

This is the current code contract, not a recommendation to use arbitrary values.

Remote validation remains authoritative.

### 11.1 Header versus JSON distinction

Unlike model and API key, reference ID is JSON data rather than an HTTP header.

It therefore follows JSON-string handling rather than header control-character validation.

---

## 12. Feature validation

Each `fish.request.features` element must be valid UTF-8.

The request layer does not currently:

- trim feature strings;
- reject blank feature strings;
- reject duplicate features;
- reject control characters;
- enforce a local allowlist;
- enforce a maximum element count;
- enforce a maximum string length.

An empty overall slice is omitted from JSON.

Remote API behavior remains authoritative for feature names and combinations.

### 12.1 Configuration strictness

The `features` field itself must be a JSON array of strings.

Unknown sibling configuration fields are rejected by strict config decoding.

---

## 13. Numeric request validation

The Fish request validates all configured numeric parameters locally.

Floating-point values requiring ranges must also be finite.

JSON configuration cannot directly encode `NaN` or infinity, but programmatic callers are still protected.

### 13.1 Temperature

```text
0.0 through 1.0
```

inclusive and finite.

### 13.2 Top P

```text
0.0 through 1.0
```

inclusive and finite.

### 13.3 Prosody speed

```text
0.5 through 2.0
```

inclusive and finite.

### 13.4 Prosody volume

```text
-20.0 through 20.0
```

inclusive and finite.

### 13.5 Chunk length

```text
100 through 300
```

inclusive.

### 13.6 MP3 bitrate

Allowed values:

```text
64
128
192
```

The unit is kilobits per second according to the configuration contract.

### 13.7 Opus bitrate

Allowed values:

```text
-1000
24000
32000
48000
64000
```

The unit is bits per second according to the configuration contract.

The value `-1000` is passed through as the configured Fish option.

The local client does not reinterpret it.

### 13.8 Maximum new tokens

Must be:

```text
greater than zero
```

There is currently no local upper bound.

### 13.9 Repetition penalty

Must be finite.

There is currently no local minimum or maximum.

### 13.10 Minimum chunk length

```text
0 through 100
```

inclusive.

### 13.11 Early-stop threshold

```text
0.0 through 1.0
```

inclusive and finite.

---

## 14. Latency validation

Allowed values:

```text
normal
balanced
low
```

The value is exact and case-sensitive.

These are rejected:

```text
Normal
LOW
<leading-space>low
low<trailing-space>
```

No normalization is applied.

---

## 15. Format validation

The Fish request supports these internal format values:

```text
wav
pcm
mp3
opus
```

The public CLI accepts:

```text
wav
mp3
opus
ogg
```

The CLI normalizes:

```text
ogg → opus
```

Therefore ordinary CLI calls never pass `ogg` into the Fish request.

### 15.1 Internal PCM support

The request validator understands `pcm`.

The command-line parser does not expose it.

`pcm` is an internal capability, not a documented public CLI format.

### 15.2 Exact format values

The request validator does not lowercase or trim format.

Public CLI normalization happens earlier.

Programmatic callers must pass an exact supported value.

---

## 16. Sample-rate validation

A non-null sample rate first must belong to the global supported set:

```text
8000
16000
24000
32000
44100
48000
```

It must then be compatible with the selected format.

### 16.1 WAV and PCM

Allowed:

```text
8000
16000
24000
32000
44100
```

### 16.2 MP3

Allowed:

```text
32000
44100
```

### 16.3 Opus

Allowed:

```text
48000
```

### 16.4 Null value

A null sample rate leaves the choice to Fish Audio.

The JSON still contains:

```json
"sample_rate": null
```

### 16.5 Validation timing

General parameter validation occurs when configuration is validated.

Format-specific compatibility occurs after the CLI format is known and the complete request is built.

A configuration may therefore load successfully but fail one invocation because its sample rate conflicts with that invocation’s `--format`.

---

## 17. Bitrate validation nuance

Both bitrate fields are validated regardless of the selected format.

For example, a WAV invocation still requires:

```text
fish.request.mp3Bitrate
```

to contain one supported MP3 value and:

```text
fish.request.opusBitrate
```

to contain one supported Opus value.

This is because parameter validation covers the complete request configuration before format relevance is considered.

Do not place an arbitrary placeholder in an “unused” bitrate field.

Use a valid supported value.

---

## 18. Normalize options

Two different normalization concepts appear in the request.

### 18.1 Text normalization

```json
"normalize": true
```

maps to:

```text
normalize
```

in the Fish request.

This is a remote synthesis option.

It is separate from local text-processing modules.

### 18.2 Loudness normalization

```json
"normalizeLoudness": true
```

maps to:

```text
prosody.normalize_loudness
```

This controls the remote prosody option.

The two flags are independent.

---

## 19. Request encoding

The complete validated request is encoded with Go’s JSON encoder before the retry loop.

Encoding happens once:

```text
validate request
    ↓
marshal JSON
    ↓
attempt loop reuses same bytes
```

Consequences:

- every retry sends the same request body;
- module output does not change between attempts;
- request parameters do not change between attempts;
- encoding failure prevents all network attempts.

The body is recreated as a fresh reader for each HTTP request.

---

## 20. Client timeout

`fish.timeoutSeconds` configures the Go HTTP client timeout.

Default:

```text
120 seconds
```

Allowed maximum:

```text
900 seconds
15 minutes
```

### 20.1 Per-attempt behavior

The HTTP client timeout applies to each HTTP request execution.

With retries, total invocation time can exceed one timeout because each attempt can consume time and retry waits occur between attempts.

### 20.2 Context interaction

The request also carries the application context.

The effective end can therefore be caused by:

- HTTP client timeout;
- `SIGINT`;
- `SIGTERM`;
- caller cancellation in package use;
- deadline on a caller-provided context.

### 20.3 No separate connect/read timeout configuration

The current client exposes one overall `http.Client.Timeout`.

It does not separately configure:

- DNS timeout;
- TCP connect timeout;
- TLS handshake timeout;
- response-header timeout;
- idle read timeout.

Adding granular transport controls would require new client and configuration design.

---

## 21. Attempt count

`fish.retry.maxAttempts` is the total number of allowed attempts.

It includes the initial request.

Examples:

```text
maxAttempts = 1
    → one request, no retry
```

```text
maxAttempts = 3
    → initial request plus at most two retries
```

Allowed range:

```text
1 through 10
```

The default is:

```text
3
```

---

## 22. Retryable responses

The client retries:

- HTTP `429` always, subject to attempt and delay rules;
- HTTP `5xx` only when `retryServerErrors` is `true`.

Default:

```json
"retryServerErrors": false
```

Therefore the default retry category is:

```text
429 only
```

### 22.1 Not retried

The client does not retry ordinary:

```text
400
401
402
403
404
422
```

It also does not currently retry:

```text
408
409
3xx
other unclassified non-2xx statuses
```

### 22.2 Transport errors

Network and transport failures returned by `http.Client.Do` are not retried.

Examples may include:

- DNS failure;
- connection refusal;
- TLS failure;
- timeout;
- connection reset before a response;
- context cancellation.

The retry predicate operates on classified Fish API response errors, not arbitrary transport errors.

### 22.3 Request creation errors

Failure to create the HTTP request is not retried.

### 22.4 Response-stream errors

Failure while streaming a successful response is not retried.

Once a `2xx` response begins writing bytes, another attempt could duplicate or mix output.

The current client returns the stream error immediately.

### 22.5 Empty successful response

A `2xx` response containing zero bytes is an error.

It is not classified as a retryable API status and is not retried.

---

## 23. Server-error retry switch

When:

```json
"retryServerErrors": false
```

HTTP `5xx` is returned immediately.

When:

```json
"retryServerErrors": true
```

HTTP statuses from:

```text
500 through 599
```

may be retried.

Attempt and delay constraints still apply.

The switch does not enable retries for transport errors or arbitrary `4xx` statuses.

---

## 24. Retry delay configuration

Two settings control fallback backoff:

```text
fish.retry.initialDelayMilliseconds
fish.retry.maxDelayMilliseconds
```

Defaults:

```text
initial: 500 ms
maximum: 5000 ms
```

Allowed range for each:

```text
1 through 300000 ms
```

The maximum delay must be greater than or equal to the initial delay.

---

## 25. `Retry-After`

For a retryable response, the client first examines:

```text
Retry-After
```

Supported forms:

- decimal seconds;
- HTTP date.

### 25.1 Decimal seconds

Example:

```text
Retry-After: 10
```

means:

```text
10 seconds
```

Negative decimal values are not accepted because only decimal digits are parsed.

### 25.2 HTTP date

Example shape:

```text
Retry-After: Wed, 21 Oct 2015 07:28:00 GMT
```

The delay is computed from the current time.

A date in the past becomes a zero delay.

### 25.3 Surrounding whitespace

The header value is trimmed before parsing.

### 25.4 Invalid header

When the header is absent or invalid, the client uses exponential backoff.

### 25.5 Delay above configured maximum

When a valid `Retry-After` delay exceeds:

```text
fish.retry.maxDelayMilliseconds
```

the client does not clamp it.

It declines the retry and returns the original API error.

This prevents the server header from forcing a wait beyond the configured ceiling.

### 25.6 Zero delay

A zero delay still checks context before proceeding.

---

## 26. Exponential backoff

When no usable `Retry-After` is available, retry delay is:

```text
initialDelay × 2^(retryNumber - 1)
```

capped at:

```text
maxDelay
```

For defaults:

```text
retry 1: 500 ms
retry 2: 1000 ms
retry 3: 2000 ms
...
cap:     5000 ms
```

The numbering refers to the wait after a failed attempt.

### 26.1 No jitter

The current backoff has no random jitter.

Concurrent clients receiving the same failure may retry at similar times.

Adding jitter would be a behavioral change requiring tests and documentation.

### 26.2 Overflow protection

Backoff growth stops at the configured maximum without overflowing duration arithmetic.

---

## 27. Retry waiting and cancellation

Retry waiting uses a timer and the same request context.

If context is canceled while waiting:

- no next attempt is sent;
- the prior API error is preserved;
- the cancellation/wait error is joined with it.

This allows error inspection to retain both:

- what response caused the retry;
- why retrying stopped.

### 27.1 Immediate retry

A zero delay performs a nonblocking context check before continuing.

### 27.2 No background retry

Retries occur synchronously inside the current `Synthesize` call.

The client does not create a background retry worker.

---

## 28. Retry decision flow

```text
attempt
    ↓
success?
    ├─ yes → return success
    └─ no
         ↓
attempt limit reached?
    ├─ yes → return error
    └─ no
         ↓
classified retryable API response?
    ├─ no → return error
    └─ yes
         ↓
usable delay within configured maximum?
    ├─ no → return error
    └─ yes
         ↓
wait with context
    ├─ canceled → return joined error
    └─ completed → next attempt
```

---

## 29. Non-success HTTP responses

Any status outside:

```text
200 through 299
```

is treated as an API failure.

The client:

1. captures `Retry-After`;
2. reads the response body with a configured bound;
3. constructs a typed `APIError`;
4. classifies selected HTTP statuses;
5. decides whether retry is allowed.

### 29.1 Any 2xx status

The client accepts every `2xx` status as a potential success.

It then streams the body.

A `204 No Content` normally results in the separate empty-audio error because zero bytes are written.

### 29.2 Redirects

The client uses Go’s ordinary HTTP client behavior and does not define a custom redirect policy.

Endpoint redirects are therefore subject to standard library behavior.

Operators should configure the final trusted endpoint directly rather than depend on redirects for authentication-sensitive requests.

---

## 30. Error-body limit

`fish.maxErrorBodyBytes` bounds the body read from a non-success response.

Default:

```text
65536 bytes
64 KiB
```

Allowed maximum:

```text
1048576 bytes
1 MiB
```

The limit protects:

- memory use;
- logs and error strings from unbounded remote bodies;
- malformed or hostile error responses.

### 30.1 Boundary behavior

A body at the configured maximum is accepted.

A body exceeding it causes a bounded-read error.

### 30.2 Classification survives read failure

When reading an oversized or otherwise failing API error body:

- the client still constructs an `APIError` from the HTTP status;
- the body message is unavailable;
- the read error is joined with the typed API error.

Therefore a `429` or optionally retryable `5xx` can still remain classifiable even when its error body cannot be read successfully.

---

## 31. Typed API errors

A non-success response becomes:

```go
*fish.APIError
```

Fields:

```text
HTTPStatusCode
HTTPStatus
APIStatus
Message
```

### 31.1 HTTP status

Examples:

```text
401
429
500
```

and status text such as:

```text
401 Unauthorized
```

### 31.2 API payload

The client attempts to decode:

```json
{
  "status": 123,
  "message": "provider message"
}
```

When decoding succeeds and `message` is nonblank:

- `APIStatus` uses `status`;
- `Message` uses the trimmed message.

### 31.3 Fallback message

When the body is not a matching JSON object with a nonblank message, the error message uses the trimmed raw body text.

Malformed JSON is therefore still available as diagnostic text, within the configured bound.

### 31.4 Empty message

When no message is available, the error reports the Fish HTTP status without appending a body message.

---

## 32. Stable API error categories

`APIError.Unwrap` exposes stable categories for `errors.Is`.

| HTTP status | Category |
|---:|---|
| `400` | `ErrValidation` |
| `401` | `ErrAuthentication` |
| `402` | `ErrPaymentRequired` |
| `403` | `ErrPermission` |
| `404` | `ErrNotFound` |
| `422` | `ErrValidation` |
| `429` | `ErrRateLimit` |
| `500–599` | `ErrServer` |
| other non-2xx | no stable category |

### 32.1 Authentication

```text
ErrAuthentication
```

indicates HTTP `401`.

It is not retried.

### 32.2 Payment required

```text
ErrPaymentRequired
```

indicates HTTP `402`.

It is not retried.

### 32.3 Permission

```text
ErrPermission
```

indicates HTTP `403`.

It is not retried.

### 32.4 Not found

```text
ErrNotFound
```

indicates HTTP `404`.

It is not retried.

### 32.5 Validation

```text
ErrValidation
```

covers HTTP:

```text
400
422
```

It is not retried.

### 32.6 Rate limit

```text
ErrRateLimit
```

indicates HTTP `429`.

It is retryable subject to configuration, remaining attempts, and delay rules.

### 32.7 Server error

```text
ErrServer
```

covers:

```text
500 through 599
```

It is retryable only when:

```text
retryServerErrors = true
```

---

## 33. API error formatting

Typical forms are:

Without body message:

```text
Fish API returned 401 Unauthorized
```

With raw or parsed message:

```text
Fish API returned 422 Unprocessable Entity: invalid request
```

With provider API status:

```text
Fish API returned 422 Unprocessable Entity (API status 1001): invalid request
```

If the HTTP status string is unavailable, the error falls back to:

```text
HTTP <code>
```

or:

```text
unknown HTTP status
```

for a structurally empty error.

---

## 34. Error-body privacy

Remote error-body text may appear in returned errors and application logs.

The body is bounded but not generally redacted.

Operators should assume a provider or proxy could return:

- request-related diagnostics;
- model or reference identifiers;
- echoed text;
- internal error details.

Do not configure an untrusted endpoint that can inject sensitive or misleading log content.

The API key itself is not part of the JSON request body and should not be present in ordinary provider errors.

---

## 35. Successful response streaming

For a `2xx` response, the client copies the response body directly into the supplied writer.

It does not first buffer the entire audio in memory.

This supports large audio responses with bounded application memory.

### 35.1 Writer ownership

The Fish client does not:

- create the destination file;
- close the supplied writer;
- rename files;
- synchronize directories.

The caller owns the writer lifecycle.

In the CLI, the atomic output layer supplies a temporary-file writer.

### 35.2 Partial bytes

If response reading or output writing fails after some data has been copied, the supplied writer may contain partial audio.

The Fish client returns an error.

The CLI’s atomic output layer prevents those partial temporary bytes from becoming the final destination.

Programmatic callers using another writer must provide their own publication or rollback semantics.

### 35.3 No retry after stream failure

A stream error is not retried.

The client cannot safely assume it can overwrite or rewind an arbitrary writer.

### 35.4 Empty response

When exactly zero bytes are copied, the client returns:

```text
Fish API returned an empty audio response
```

A successful HTTP status alone is therefore insufficient.

---

## 36. Response content type

The client does not currently validate:

- `Content-Type`;
- filename;
- audio container signature;
- codec;
- sample rate in returned bytes;
- correspondence between requested and returned format.

Any non-empty `2xx` body is streamed as the response payload.

The caller and provider contract are expected to make it audio.

Adding MIME or container validation would be a new behavior and must account for provider variations.

---

## 37. Response-body close

Every received HTTP response body is scheduled for close.

This applies to:

- success responses;
- API error responses;
- retries.

The current attempt returns after reading or streaming the body.

A body-close error is not surfaced separately because the close occurs through deferred cleanup after the main result is determined.

---

## 38. Context and cancellation

The caller context is attached to each HTTP request.

Cancellation can interrupt:

- request transmission;
- waiting for response headers;
- response reading;
- retry waiting.

### 38.1 Nil context

A nil or typed-nil context is rejected before request encoding.

### 38.2 CLI signals

The CLI context is canceled by:

- `SIGINT`;
- `SIGTERM`.

Cancellation during Fish synthesis ultimately causes CLI failure status `4`.

### 38.3 Joined retry-wait error

When cancellation occurs during a retry wait, the returned error joins:

- the retryable API error;
- the context-related wait error.

### 38.4 No hidden continuation

The client does not keep sending requests after `Synthesize` returns.

---

## 39. Output writer validation

A nil or typed-nil writer is rejected.

The client does not require:

- seek support;
- file semantics;
- buffering;
- synchronization methods.

It needs only an `io.Writer`.

Because the writer may be non-seekable, retry after partial success-stream output is intentionally not attempted.

---

## 40. HTTP transport behavior

The client creates:

```go
&http.Client{
    Timeout: configuredTimeout,
}
```

It uses Go’s default transport through that client.

The current configuration does not expose:

- proxy selection;
- custom certificate authority;
- client certificate;
- TLS minimum version;
- connection pool sizes;
- keep-alive controls;
- redirect policy;
- custom resolver;
- custom dialer.

Environment and Go standard-library defaults may affect ordinary proxy behavior.

Adding transport customization requires a deliberate configuration and security design.

---

## 41. Request attempt isolation

Each retry creates a new:

```text
http.Request
```

with:

- the same context;
- the same endpoint;
- the same encoded JSON bytes;
- the same headers;
- a fresh body reader.

No response object is reused.

No attempt mutates the request parameters for later attempts.

---

## 42. Retry limitations

Current retry behavior deliberately does not include:

- transport-error retry;
- `408` retry;
- `409` retry;
- arbitrary `4xx` retry;
- redirect-specific retry;
- content-type failure retry;
- empty-body retry;
- stream-error retry;
- jitter;
- hedged requests;
- concurrent attempts;
- background retries;
- idempotency tokens;
- per-status custom limits.

These are not accidental documentation omissions.

They are outside the present client contract.

---

## 43. Default behavior summary

With built-in defaults:

```text
endpoint:
    https://api.fish.audio/v1/tts

model:
    s2.1-pro-free

timeout per HTTP client request:
    120 seconds

maximum error body:
    65536 bytes

attempts:
    3 total

fallback retry delays:
    500 ms initial
    5000 ms maximum

retry categories:
    429
    not 5xx by default
```

Request defaults are listed completely in [`configuration.md`](configuration.md).

---

## 44. Example default request

For:

```bash
fish-audio-cli \
  --text "Привет" \
  --format opus \
  --output speech.opus
```

with default configuration and empty reference ID, the request body is conceptually:

```json
{
  "text": "Привет",
  "temperature": 0.7,
  "top_p": 0.7,
  "prosody": {
    "speed": 1.0,
    "volume": 0.0,
    "normalize_loudness": true
  },
  "chunk_length": 300,
  "normalize": true,
  "format": "opus",
  "sample_rate": null,
  "mp3_bitrate": 192,
  "opus_bitrate": 64000,
  "latency": "normal",
  "max_new_tokens": 1024,
  "repetition_penalty": 1.2,
  "min_chunk_length": 50,
  "condition_on_previous_chunks": true,
  "early_stop_threshold": 1.0
}
```

Omitted:

```text
reference_id
features
```

Headers conceptually include:

```text
Authorization: Bearer <secret>
Content-Type: application/json
model: s2.1-pro-free
```

---

## 45. Custom reference example

Configuration:

```json
{
  "fish": {
    "referenceId": "YOUR_REFERENCE_ID"
  }
}
```

Request includes:

```json
{
  "reference_id": "YOUR_REFERENCE_ID"
}
```

The identifier is not verified locally against the remote account.

A missing or inaccessible reference can produce a remote API error.

---

## 46. Custom base URL example

Configuration:

```json
{
  "fish": {
    "baseUrl": "https://tts-proxy.example.test/fish"
  }
}
```

Resolved endpoint:

```text
https://tts-proxy.example.test/fish/v1/tts
```

The proxy must understand the same:

- Bearer authorization header;
- model header;
- request JSON;
- audio response semantics;
- error behavior expected by the client.

### 46.1 Trust boundary

A custom endpoint receives:

- Fish API key;
- processed text;
- reference ID;
- synthesis parameters.

Only use an endpoint authorized to receive all of that data.

---

## 47. Rate-limit example

Assume:

```json
{
  "fish": {
    "retry": {
      "maxAttempts": 3,
      "initialDelayMilliseconds": 500,
      "maxDelayMilliseconds": 5000,
      "retryServerErrors": false
    }
  }
}
```

Possible sequence:

```text
attempt 1 → 429, no Retry-After
wait 500 ms

attempt 2 → 429, no Retry-After
wait 1000 ms

attempt 3 → 200 with non-empty body
success
```

If attempt 3 also returns `429`, the error is returned without another wait.

---

## 48. `Retry-After` ceiling example

Configuration:

```json
{
  "fish": {
    "retry": {
      "maxAttempts": 3,
      "initialDelayMilliseconds": 500,
      "maxDelayMilliseconds": 5000,
      "retryServerErrors": false
    }
  }
}
```

Response:

```text
HTTP 429
Retry-After: 30
```

Parsed delay:

```text
30000 ms
```

Configured maximum:

```text
5000 ms
```

Result:

- no retry;
- no 5-second clamping;
- original rate-limit error returned.

---

## 49. Server-error example

Default:

```json
"retryServerErrors": false
```

Response:

```text
HTTP 503
```

Result:

```text
return immediately
```

With:

```json
"retryServerErrors": true
```

the same response can enter the retry flow.

---

## 50. Transport-error example

Failure:

```text
DNS lookup failed
```

or:

```text
connection refused
```

Even with:

```json
"maxAttempts": 10
```

the current client does not retry the transport error.

It returns the send error from the first attempt.

---

## 51. Successful stream failure example

Sequence:

```text
HTTP 200
    ↓
some audio bytes copied
    ↓
response connection breaks
```

Result:

- `Synthesize` returns a stream error;
- no retry occurs;
- the supplied writer may contain partial bytes;
- CLI atomic publication removes the temporary file and preserves any old destination.

---

## 52. Error inspection in Go

Programmatic callers can inspect the typed error.

Conceptual example:

```go
var apiErr *fish.APIError

if errors.As(err, &apiErr) {
    fmt.Println(apiErr.HTTPStatusCode)
    fmt.Println(apiErr.APIStatus)
    fmt.Println(apiErr.Message)
}

switch {
case errors.Is(err, fish.ErrRateLimit):
    // Handle rate limiting.

case errors.Is(err, fish.ErrAuthentication):
    // Handle invalid credentials.

case errors.Is(err, fish.ErrValidation):
    // Handle request rejection.
}
```

Joined errors preserve category checks where one joined component unwraps to the category.

---

## 53. Logging behavior

The application logs:

At synthesis start:

```text
model
format
output_path
```

At synthesis failure:

```text
error
```

At completion:

```text
output_path
```

The API key is not logged.

### 53.1 Request body

The application does not log the complete Fish JSON request by default.

Processed text can be logged earlier only when:

```json
"logging.logText": true
```

### 53.2 Remote error body

A bounded remote API error message may appear inside the logged error.

---

## 54. Security boundary

The Fish request discloses to the configured endpoint:

- processed text;
- reference ID when configured;
- synthesis parameters;
- model selection;
- Fish API key.

The endpoint can infer:

- requested output format;
- sampling settings;
- text length;
- voice selection;
- retry traffic.

### 54.1 Local modules first

All configured text modules run before Fish synthesis.

A module may:

- remove sensitive data;
- add markup;
- translate;
- expand text;
- send text to another provider.

The Fish layer receives only the final successful pipeline text.

### 54.2 Custom endpoint risk

A custom `fish.baseUrl` can capture the API key and text.

Configuration files that change the base URL are security-sensitive.

### 54.3 Plain HTTP risk

Using `http` exposes credentials and text to network interception unless another trusted secure tunnel provides protection.

---

## 55. Performance boundary

The Fish client:

- marshals one request body;
- streams successful audio;
- bounds error bodies;
- reuses one `http.Client` within the invocation;
- performs retries sequentially.

It does not:

- buffer the entire successful audio response;
- parallelize attempts;
- maintain a process-wide client pool across CLI invocations;
- cache synthesis responses.

### 55.1 CLI process lifetime

Each CLI invocation constructs a new Fish client.

Persistent connection reuse is limited to what occurs within that client’s short process lifetime and retry sequence.

---

## 56. Testing a custom endpoint

A test server should verify:

- method is `POST`;
- path ends with `/v1/tts`;
- authorization is `Bearer <key>`;
- model header is present;
- content type is JSON;
- request body fields use expected snake-case names;
- empty reference ID is omitted;
- empty features are omitted;
- null sample rate is included;
- response status classification;
- bounded error bodies;
- retry timing;
- cancellation;
- partial stream behavior;
- zero-byte response rejection.

Do not test only a successful `200`.

Most client defects live in failure behavior, apparently because success has insufficient imagination.

---

## 57. Unit-test expectations

### Endpoint

Test:

- default base URL;
- base path;
- uppercase scheme;
- missing scheme;
- unsupported scheme;
- missing host;
- user information;
- query;
- fragment;
- surrounding whitespace.

### Headers

Test:

- valid key and model;
- blank key;
- blank model;
- surrounding whitespace;
- invalid UTF-8;
- ASCII control characters.

### Request parameters

Test:

- every numeric boundary;
- below and above each range;
- `NaN` and infinity through programmatic construction;
- exact latency values;
- every format;
- every format/sample-rate combination;
- invalid text;
- empty reference;
- feature UTF-8;
- defensive copies.

### API errors

Test:

- every stable category;
- structured JSON payload;
- raw text body;
- malformed JSON;
- blank message;
- missing status text;
- `errors.Is`;
- `errors.As`;
- bounded-body read failure.

### Retry

Test:

- `429`;
- `5xx` disabled;
- `5xx` enabled;
- attempt count;
- decimal `Retry-After`;
- HTTP-date `Retry-After`;
- past date;
- invalid header;
- delay above maximum;
- exponential cap;
- cancellation during wait;
- transport error not retried;
- stream error not retried;
- empty body not retried.

### Streaming

Test:

- non-empty success;
- empty success;
- writer failure;
- body read failure;
- response body close;
- no retry after bytes begin.

---

## 58. Integration-test expectations

Integration tests around the application should verify:

- processed text reaches the Fish request;
- selected CLI format reaches the request;
- `ogg` reaches Fish as `opus`;
- reference ID is copied;
- API key is loaded after text processing;
- invalid format/sample-rate combination fails before network I/O;
- missing secret file prevents client construction;
- successful bytes reach atomic output;
- API failure does not replace an existing destination;
- cancellation does not publish partial output;
- program-controlled synthesis failure returns exit status `4`.

---

## 59. Review checklist

### Endpoint and headers

- Is `/v1/tts` still joined correctly?
- Are URL restrictions preserved?
- Is the API key only in authorization?
- Is model still an HTTP header?
- Are header control characters rejected?
- Is HTTPS still the production default?

### Request

- Does JSON mapping match code?
- Are omission rules unchanged?
- Is sample-rate null behavior unchanged?
- Are all numeric values finite and bounded where required?
- Is format compatibility checked?
- Are mutable values copied?

### Retry

- Does `maxAttempts` still include the initial attempt?
- Is `429` still retryable?
- Are `5xx` retries still opt-in?
- Are transport errors still non-retryable?
- Is `Retry-After` bounded by configured maximum?
- Is exponential fallback capped?
- Is context honored during waits?
- Is there still no retry after stream output begins?

### Errors

- Are error bodies bounded?
- Are status categories stable?
- Does structured payload parsing remain safe?
- Are raw fallback messages trimmed?
- Are joined errors inspectable?
- Are secrets absent?

### Streaming

- Is success streamed rather than buffered?
- Is zero-byte success rejected?
- Is content type intentionally unchecked?
- Does the caller still own writer close and publication?
- Are partial bytes contained by the CLI output layer?

### Documentation

- Are defaults synchronized with code?
- Are ranges synchronized with validation?
- Are remote capabilities not falsely promised?
- Are unsupported retry modes omitted?
- Are security disclosures accurate?

---

## 60. Fish-client invariants

The following rules are normative for the current client.

1. The endpoint is the validated base URL joined with `v1/tts`.
2. Requests use HTTP `POST`.
3. Authentication uses a Bearer header.
4. Model selection uses the `model` header.
5. Request content type is JSON.
6. API key and model must be header-safe UTF-8 strings.
7. Processed text is carried in request JSON.
8. Empty reference ID is omitted.
9. Empty features are omitted.
10. Null sample rate is included as JSON null.
11. Request parameters are validated before network I/O.
12. Format-specific sample-rate validation uses the final selected format.
13. The encoded request body is reused across attempts.
14. `maxAttempts` includes the initial request.
15. HTTP `429` is retryable.
16. HTTP `5xx` is retryable only when enabled.
17. Transport errors are not retried.
18. Successful-stream errors are not retried.
19. Empty successful responses are errors and are not retried.
20. `Retry-After` supports decimal seconds and HTTP dates.
21. A valid server delay above the configured maximum prevents retry.
22. Missing or invalid `Retry-After` uses exponential backoff.
23. Backoff has no jitter.
24. Retry waits honor context cancellation.
25. Non-success error bodies are bounded.
26. Stable categories are exposed through `errors.Is`.
27. Typed details are exposed through `errors.As`.
28. Any non-empty `2xx` body is streamed without MIME validation.
29. The Fish client may leave partial bytes in its writer on stream failure.
30. Atomic final-file safety belongs to the outer output layer.
31. The API key is not placed in request JSON.
32. The current client has no custom transport configuration.
33. Retries are synchronous and sequential.
34. The client does not run background work after return.

Changing one of these rules is a Fish integration compatibility change.

---

## 61. Non-goals

The current Fish client does not provide:

- live model discovery;
- live voice discovery;
- account or quota inspection;
- automatic model fallback;
- automatic reference fallback;
- transport-error retries;
- adaptive concurrency;
- retry jitter;
- idempotency keys;
- response MIME validation;
- audio signature validation;
- audio transcoding;
- output extension selection;
- output publication;
- response caching;
- persistent cross-process connection pooling;
- custom TLS configuration;
- provider-agnostic TTS abstraction.

These may be considered when a concrete requirement justifies the added contract.

---

## 62. Summary

The current Fish integration is intentionally narrow:

```text
validated final text
    ↓
validated Fish request
    ↓
POST /v1/tts
    ↓
selected bounded retries
    ↓
typed API error or streamed non-empty bytes
```

The most important operational rules are:

- protect `fish.baseUrl` as a credential and text disclosure boundary;
- keep model and API key free of whitespace and control characters;
- use only validated format/sample-rate combinations;
- remember that both bitrate settings must remain valid;
- treat `maxAttempts` as total attempts;
- expect retries for `429`, and for `5xx` only when enabled;
- do not expect transport or stream retries;
- keep `Retry-After` within the configured maximum;
- rely on atomic output, not the Fish client alone, to hide partial audio;
- inspect typed API errors rather than parsing error strings.
