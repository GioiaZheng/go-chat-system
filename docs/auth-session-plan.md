# Auth Session Plan

## Current problem

The API currently treats the bearer token as the user ID. This makes authentication easy to spoof, prevents server-side logout, and provides no secure way to expire or revoke credentials after login. Authorization checks can verify whether a user ID is allowed to access a resource, but they cannot prove the caller actually owns that identity.

## Proposed `sessions` table schema

Add a server-side sessions table and store only hashed tokens:

```sql
CREATE TABLE sessions (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    token_hash   TEXT NOT NULL UNIQUE,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at   TIMESTAMP NOT NULL,
    revoked_at   TIMESTAMP,
    last_seen_at TIMESTAMP,

    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_user_active ON sessions(user_id, expires_at, revoked_at);
```

## Login flow

1. Validate the login request and resolve or create the user according to the existing login behavior.
2. Generate a cryptographically random opaque token, for example 32 bytes from `crypto/rand` encoded with base64url.
3. Hash the token and insert a new `sessions` row with `user_id`, `token_hash`, `created_at`, and `expires_at`.
4. Return the raw token to the client once. Do not persist or log the raw token.
5. Clients continue sending `Authorization: Bearer <token>` on API requests.

## Request validation flow

1. Parse the bearer token from the `Authorization` header.
2. Reject missing or malformed tokens with `401 Unauthorized`.
3. Hash the presented token using the same token hashing strategy.
4. Look up a session by `token_hash`.
5. Reject the request with `401 Unauthorized` if the session is missing, expired, or revoked.
6. Load the session user and place the real `user_id` in `reqcontext.RequestContext`.
7. Continue using existing resource authorization checks, such as conversation membership checks, with the session-derived user ID.
8. Optionally update `last_seen_at` asynchronously or with throttling to avoid writing on every request.

## Token hashing strategy

- Use opaque random tokens rather than signed user-identifying tokens.
- Store only a keyed hash of the token, such as HMAC-SHA-256 with a server secret, encoded as hex or base64url.
- Compare token hashes using constant-time comparison when comparing in application code.
- Keep the HMAC secret outside the database in environment/configuration.
- Support secret rotation by adding a `hash_version` column if rotation is needed later.

## Expiration and revocation

- Set a finite `expires_at` for each session, for example 7 to 30 days depending on product requirements.
- Treat sessions with non-null `revoked_at` as invalid immediately.
- Add a logout endpoint that sets `revoked_at` for the current session.
- Add an optional “logout all devices” operation that revokes all active sessions for a user.
- Periodically delete old expired or revoked sessions after an audit-retention window.

## Migration plan

1. Add the `sessions` table in a database migration while keeping current token behavior unchanged.
2. Update login to create sessions and return opaque session tokens.
3. Update request context middleware to validate session tokens and set `ctx.UserID` from the session row.
4. Temporarily support legacy `token == userID` only behind a development or migration flag, if needed.
5. Update clients and tests to use login-issued session tokens.
6. Remove the legacy fallback after clients have migrated.
7. Verify authorization-sensitive endpoints still rely on `ctx.UserID`, not raw tokens.

## Tests to add

- Login creates a session row with the correct `user_id`, a future `expires_at`, and no stored raw token.
- Valid session token populates `RequestContext.UserID`.
- Missing, malformed, unknown, expired, and revoked tokens return `401 Unauthorized`.
- Logout revokes only the current session.
- Logout-all revokes all active sessions for the user.
- Expired and revoked sessions cannot access protected endpoints.
- Existing authorization tests still pass when `ctx.UserID` comes from a session instead of the raw bearer token.
