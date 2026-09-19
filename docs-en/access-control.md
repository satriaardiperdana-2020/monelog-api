# Access control and soft deletion
Version 0.3 • 14 September 2026
Confirmed: admin can perform all application operations on any user's data. A regular user can CRUD only their own data.
This replaces the earlier restriction on admin modifications.

## Permission matrix
| Operation | Regular user | Admin |
| --- | --- | --- |
| View/create/edit income and expenses | Own only | Any selected owner |
| Delete/restore income and expenses | Own only, soft delete | Any selected owner, soft delete |
| Category CRUD/restore | Own only | Any selected owner |
| Reports and Excel/PDF exports, including job status/download | Own only | Any selected owner |
| Template CRUD/restore/apply | Own only | Any selected owner |
| Drive settings, backup jobs and restore | Own only | Any selected owner, with valid provider authorization |
| Profile update/self-account deletion | Own only | Any user |
| List/create/manage accounts and assign user/admin roles | Public registration cannot assign admin | Allowed through protected admin operations |
| View business records in Trash | Own only | Any selected owner |
| View admin audit records | Not exposed | Allowed through the admin audit endpoint |

All business operations use validation, optimistic concurrency and retained-row deletion. Passwords/tokens/provider secrets are never returned as ordinary application data; reset/disconnect flows operate on credentials without exposing them.
Admins can choose any existing account, including other admins or themselves. My Data remains a convenient self scope. Cross-user management uses explicit admin routes.

## Actor and owner
actor_user_id is the authenticated account and never changes when selecting another user.
owner_user_id is the financial data owner: actor on personal routes; explicit target on authorized admin routes.
New admin-created records retain the selected user's ownership. created_by/updated_by and admin audit events identify the acting admin.
All financial queries and writes retain owner predicates; do not make is_admin remove scoping. Record IDs, categories, form context, cursors and jobs must match the selected owner.

## Authorization
- Authenticate and load current actor account state on every protected request. Deleted actors get 401 even with an unexpired JWT.
- Every admin read/write checks the current database role. A regular user gets 403 before target lookup.
- Public registration assigns role=user. Personal profile/body parameters cannot supply role or owner overrides.
- Admin account creation and role updates are allowed on protected /admin/users operations. Initial admin bootstrap is an explicit operator action, never automatic first-user promotion.
- Role change or account soft deletion revokes refresh sessions. The next admin request after committed demotion is denied. Authorization already granted to an in-flight request may finish; mutations serialize their account/role checks with changes as described in database.md.
- Target owner and related resource mismatches return 404. Missing/malformed scope gets 400, never a global financial operation.
- Ordinary routes always use the actor's own scope, including for admins. To manage B, admin C calls B's admin route.
- Regular users cannot access another user's Trash, exports, templates or backups. Admins can perform these operations for a selected target.
- Normal writes require an active target account. Admins can inspect a deleted target's retained data and restore the account before further financial changes.

## isDelete contract
| Layer | Name | Values |
| --- | --- | --- |
| JSON response field | isDelete | boolean false / true |
| Go field | IsDelete | bool |
| PostgreSQL column | is_delete | BOOLEAN NOT NULL DEFAULT FALSE |

Applies to users, categories, transactions and templates. Sessions/connections keep their token revocation lifecycle, while job status/expiry is separate.
Create → false. Delete → true. Restore → false. No physical row deletion from these application endpoints.
Normal list/detail requires false. Trash list/detail explicitly selects isDelete=true. A boolean filter never changes authorization. Reports and report exports always use active transactions.
Delete/restore use required row version, owner predicate and expected lifecycle state. A stale version or duplicate lifecycle operation returns 409; a missing or foreign record returns 404.
Deleted categories stay available as historical labels, but new/restored/edited transactions and templates require an active compatible category.
Users can restore their own financial data while their account is active. Restoring a deleted account requires an admin because that account cannot log in.

## Admin Management UI
User search and selection start empty. Clearly display owner identity with fully enabled management actions appropriate to the feature's delivery stage.
Bind form/dialog owner at creation. Before switching, save or discard unsaved input; never redirect an open form's payload to the newly selected account.
Clear old data and cursors on switch, cancel reads and ignore old response generations. In-flight writes/jobs keep their original owner; their completion cannot mutate the new screen.
Provide Trash/Restore per selected owner and account-management/role controls for admins. Owner UI offers the same financial CRUD for self.
Use separate actor/mode/owner/filter/version state keys; discard state on logout, role denial or account deletion. All authenticated responses use Cache-Control: no-store.

## Exports, templates and backups
Personal operations set owner=actor. Admin operations set owner=the verified target; persist both requester and owner plus request_mode.
Regular users can access only jobs/data with their own owner ID. Current admins can manage any target's jobs via that target's route, irrespective of the original requester.
Workers revalidate requester/owner activity and, for admin requests, current admin role at execution. Cancel unauthorized queued jobs; retries never change owner.
Recurring schedules retain owner, authorizing actor and mode. Invalid actor/role or deleted owner pauses them; resume requires renewed authorization.
Google operations additionally need a valid OAuth connection for the target's Drive account. App admin status does not supply provider credentials or consent.
Backup formats preserve isDelete flags but exclude roles, password/session/provider secrets and audit records. Restore validates/remaps all rows to the authorized target and cannot grant roles. Role changes use the dedicated admin endpoint.

## Audit
Log actor, target, action, resource ID, outcome, request ID, UTC time and safe changed-field names/version metadata.
Successful admin writes commit the data mutation and its audit event in one database transaction. Audit failure rolls back writes; successful reads persist an event before responding.
Audit events are append-only; admins may inspect them, not rewrite history as part of business CRUD. Record role-change old/new values without secret payloads.
External jobs atomically persist authorization, job and audit before provider calls. Retry and external failure use explicit job states, not a claim of database rollback of remote work.
Do not log finance titles/amounts, passwords, tokens or full response bodies.

## Required acceptance matrix
| Test | Expected |
| --- | --- |
| A performs CRUD/restore on own data | Succeeds with correct totals/lifecycle |
| A uses B resource IDs, Trash filters, export/template/backup paths | 404 on personal resources; 403 on admin routes |
| C creates/edits/deletes/restores B records via admin route | Succeeds; owner B retained, actor C audited |
| C uses B ID inside A admin route | 404; neither owner modified |
| C submits old version | 409, no lost update |
| Delete followed by normal list/report/export | Row retained with true, excluded from active results/totals |
| Trash/restore | True rows visible only in authorized Trash; restore false and totals include once |
| Delete/edit/restore race | At most one expected-version operation wins; other request conflicts |
| Create/edit payload supplies isDelete or arbitrary owner | 400; lifecycle/ownership cannot be bypassed |
| C changes A role through protected operation | Succeeds and audited; public/profile role mutation denied |
| C demoted or actor account deleted | Next protected/admin request denied as appropriate |
| Target changes while form/write is open | No write sent to the wrong owner |
| Admin export/template/backup for B | Succeeds with target-scoped authorization; A cannot use it |
| Queued admin job after requester demotion | Canceled/paused, no provider call |
| Audit insert fails during admin write | Database mutation rolled back |
| Deleted category with active transactions | Historical amounts/labels retained; category unavailable for selection |
| Account soft delete | Sessions revoked, jobs paused, child rows retained |
| Malicious backup role/owner fields | Rejected; target ownership and role protections retained |

Delivery: 002–004 schema/auth/contract; 005–007 core CRUD and reports; 013 full admin backend; 008 UI; 009 exports; 010 mobile; 011 templates; 012 backups.
