# PLAN-003: Authentication and profile

Status: Draft — refine against actual repository and review before coding.
Issue: ../issues/ISSUE-003-authentication.md
Repository: monelog-api; prerequisites: 002.

## Before implementation
Read linked issue and requirements/architecture/database/API. Verify prerequisite issues are Done. Inspect actual tree and existing changes; proposed paths are not verified existing paths. Resolve any product/security choices relevant to this issue.

## Implementation sequence
1. Finalize password policy, hash choice, session TTL and JWT validation settings.
2. Implement registration with default category seeding in one DB transaction.
3. Implement access JWT plus hashed rotating refresh sessions.
4. Implement browser cookie/CSRF/origin handling and native token transport.
5. Add profile timezone validation and login rate limiting.

## Affected areas
Backend: cmd as needed, internal handlers/service/repository, db queries/migrations, API contract and integration tests.

Narrow these areas to exact file paths during repository inspection; do not edit all listed areas automatically.

## Validation
Wrong password; duplicate email; expired JWT; invalid issuer/audience; refresh replay/race; logout; CSRF failure; user isolation.
Run go test ./... and go vet ./... plus configured PostgreSQL integration suite; add race tests where concurrency is involved.

Record actual results; unavailable infrastructure is a stated blocker, not a pass.

## Risks and recovery
No Google login or provider backup credentials. Preserve user changes. Keep PR focused. Use disposable database fixtures; if schema changes, test forward migration and document recovery before rollout. Roll back application through a reviewed prior build, never by erasing shared data. Incompatible/data-changing migrations require a separate recovery plan.

## Review checkpoint
Present refined sequence, exact file scope, unresolved decisions and test commands. Wait for implementation approval. After coding, compare each acceptance criterion with concrete test evidence.

