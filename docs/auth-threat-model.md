# Authentication Threat Model

## Scope

This note covers the current local-development authentication model used by the Go API and Vue client. The backend accepts `Authorization: Bearer <token>` and a legacy bare token format. In this project, the token is the user ID returned by `POST /session`.

## Assets

- User identity and profile data.
- Conversation membership.
- Message content, comments, forwards, and attachments.
- Uploaded avatars and message files under `/uploads`.

## Trust Boundaries

- Browser storage to API requests: the client reads the token from `localStorage` and sends it with each protected request.
- Public API to database: handlers must check authorization before reading or mutating conversation-scoped data.
- Static upload serving: files stored under `uploads/` are public once a URL is returned.

## Threats and Current Controls

| Threat | Current control | Remaining risk |
| --- | --- | --- |
| Missing or empty auth token | `wrap` rejects missing and empty `Authorization` headers with `401`. | Token format is still lightweight and should not be treated as production-grade. |
| Access to another conversation | Conversation and message handlers call membership checks before returning or mutating data. | Authorization coverage should continue to be tested when new endpoints are added. |
| Forged user token | The current token is the user ID, so any guessed ID can be sent. | Replace with signed, expiring session tokens before production use. |
| Token theft from browser storage | Client stores the token in `localStorage`. | Prefer HttpOnly secure cookies or short-lived access tokens with refresh rotation. |
| Upload abuse | Avatar upload checks extension, detected MIME type, and a 10 MiB limit. | Add malware scanning and stricter content storage controls for production. |
| CORS exposure | CORS is configured at the server entrypoint. | Restrict origins outside local development. |

## Hardening Backlog

1. Replace user-ID tokens with signed opaque sessions or JWTs with expiration.
2. Verify that the token maps to an active user on each authenticated request.
3. Add CSRF protection if moving auth into cookies.
4. Add rate limits to `/session`, upload endpoints, and message mutation endpoints.
5. Add structured audit logs for login, uploads, membership changes, and message deletion.
