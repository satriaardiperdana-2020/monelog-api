# PLAN-003: Autentikasi, state akun, dan otorisasi role

Status: Disempurnakan — menunggu keputusan keamanan yang belum ditetapkan sebelum implementasi.
Diperbarui: 16 September 2026 (v0.4)
Issue: [ISSUE-003](../issues/ISSUE-003-authentication.md)
Repositori: monelog-api
Prasyarat: 002
Persyaratan: FR-01, FR-15, FR-16, FR-17

## Temuan repositori

- ISSUE-001 hanya menyediakan Echo, health route, konfigurasi dasar, middleware logging aman, dan `cmd/api/main.go` sebagai composition root.
- ISSUE-002 menyediakan `users`, `categories`, dan `refresh_sessions`, query sqlc, `pgxpool`, serta `Queries.WithTx`. Skema sesi sudah memuat `family_id`, `token_hash`, `revoked_at`, dan `replaced_by`; tidak ada migrasi baru yang diperlukan untuk ISSUE-003.
- `db/queries/users.sql` tidak memasukkan role pada registrasi, sehingga default database `user` sudah melindungi registrasi publik. Query tambahan untuk bootstrap operator harus terpisah dan tidak boleh digunakan route publik.
- `api/openapi.yaml` adalah kontrak kanonis untuk health, authentication, dan profil yang diimplementasikan saat ini. `oapi-codegen` menghasilkan binding Echo v5 dan model dari kontrak tersebut; route domain berikutnya ditambahkan bersama implementasinya.

## Batas scope

ISSUE-003 mengirim registrasi, login, refresh, logout, `/me`, middleware actor/role, bootstrap admin operator, dan keamanan transport token. ISSUE ini tidak mengirim kategori atau transaksi CRUD, route `/admin/users`, perubahan role lewat HTTP, atau audit endpoint; area itu tetap milik issue berikutnya.

## Keputusan keamanan

### Password

- Terima password UTF-8 apa adanya tanpa trim atau normalisasi. Panjang wajib 12–128 byte UTF-8; tidak ada aturan komposisi supaya passphrase tetap didukung.
- Hash baru memakai Argon2id dengan salt acak 16 byte dari `crypto/rand`, memory 65,536 KiB (64 MiB), iterations 3, parallelism 1, dan output 32 byte.
- Simpan hanya string PHC `$argon2id$v=19$m=65536,t=3,p=1$<salt>$<hash>` pada `users.password_hash`. Verifikasi membatasi parsing PHC pada parameter tepat di atas dan membandingkan output dengan waktu konstan.
- Password mentah hanya hidup selama request/command berlangsung. Tidak boleh muncul di log, error, response, audit metadata, SQL, atau token.

### Access token JWT

- Gunakan JWT HMAC-SHA-256 saja. `AUTH_JWT_HMAC_KEY` adalah secret base64 yang didekode menjadi minimal 32 byte; tidak ada fallback/dev key dan nilai tidak pernah dicatat.
- Issue ini menambahkan `AUTH_JWT_ISSUER` dan `AUTH_JWT_AUDIENCE` sebagai konfigurasi wajib. Access token berlaku maksimum 24 jam.
- Token bertipe `access` berisi `iss`, `aud`, `sub`, `exp`, `iat`, dan `jti`. `sub` dan `jti` adalah UUID v4; role tidak menjadi sumber otorisasi JWT.
- Validator menerima hanya header `alg=HS256` dan menolak `none`, algoritma lain, key kosong, signature tidak sah, issuer/audience tidak tepat, subject atau token ID bukan UUID, `exp`/`iat` hilang, `iat` di masa depan, `exp <= iat`, atau lifetime lebih dari 24 jam ditambah clock skew 30 detik.
- Middleware mengambil user aktif saat ini dari database untuk setiap Bearer request dan menaruh `{userID, role}` dalam context. `is_delete=true` menghasilkan 401 meskipun JWT masih berlaku. Guard admin memakai role database saat ini, sehingga demotion berikutnya tidak dapat memakai klaim lama.

### Refresh token dan replay

- Refresh secret adalah 32 byte acak dari `crypto/rand`, dienkode base64url tanpa padding. Masa berlaku setiap refresh session tepat 30 hari dari penerbitan.
- Sebelum disimpan atau dicari, hash `SHA-256` dibuat atas representasi secret yang diterima dan disimpan sebagai string hex. Secret mentah tidak disimpan atau dicatat.
- Login membuat session ID dan family ID UUID baru. Refresh sukses berjalan dalam satu transaksi: lock session menurut hash, lock user, validasi session aktif/belum kedaluwarsa dan user aktif, buat session pengganti dengan family ID sama, lalu revoke session lama dengan `replaced_by` sebelum commit.
- Secret yang cocok dengan session ter-revoke adalah replay. Dalam transaksi yang sama revoke seluruh family, lalu kembalikan kegagalan refresh generik. Secret yang hanya expired tetap gagal generik dan family aktifnya juga dicabut sebagai tindakan pertahanan. Race dua refresh pada secret sama harus membuat paling banyak satu rotasi sukses; request lain mendeteksi replay dan mencabut family.
- Logout menemukan session dari hash lalu mencabut seluruh family dalam transaksi. Secret tidak dikenal tetap menghasilkan 204 agar keberadaan session tidak bocor. Penghapusan akun dan perubahan role di masa depan memanggil `RevokeAllUserRefreshSessions` dalam transaksi perubahan akun yang sama.

### Browser, CSRF, Origin, dan native

- Browser menerima access token hanya pada JSON response untuk disimpan di memori. Refresh secret ada hanya pada cookie host-only `__Host-monelog-refresh` dengan `HttpOnly`, `Secure`, `Path=/`, `SameSite=Strict`, tanpa `Domain`, dan masa berlaku paling lama 30 hari.
- Browser juga menerima cookie host-only `__Host-monelog-csrf`, `Secure`, `Path=/`, `SameSite=Strict`, non-HttpOnly, berisi secret acak yang berbeda. Refresh dan logout browser wajib mengirim header `X-CSRF-Token` yang cocok dengan cookie menggunakan perbandingan waktu konstan.
- Login, refresh, dan logout browser wajib memiliki `Origin` yang persis cocok dengan daftar `AUTH_ALLOWED_ORIGINS`; Origin yang hilang atau berbeda ditolak. Respons auth dan semua respons terautentikasi memakai `Cache-Control: no-store`.
- Login browser menerbitkan ulang kedua cookie. Refresh dan logout menghapus cookie dengan atribut yang sama. Route browser tidak menerima refresh secret dari JSON.
- Native Android/iOS wajib memakai HTTPS. Login dan refresh native menerima/mengembalikan refresh secret hanya dalam JSON request/response; aplikasi menyimpannya pada storage aman OS dan menyimpan access token hanya di memori. Native tidak memakai cookie atau localStorage.
- Rate limiter in-memory berjalan sebelum verifikasi password pada `POST /api/v1/auth/login`: maksimum lima percobaan per 15 menit untuk gabungan hash email-normalized+IP dan maksimum 20 per 15 menit untuk IP. Bucket kedaluwarsa dibersihkan, key tidak dicatat, dan penolakan memberi 429 generik dengan `Retry-After`. Password salah dan email tak ada selalu memberi respons 401 yang sama.

## Alur implementasi

1. Tambahkan konfigurasi auth tervalidasi dan dependency crypto/JWT. Startup gagal tanpa key HMAC yang valid, issuer, audience, atau origin browser yang valid; error konfigurasi tidak mencetak secret.
2. Buat `internal/auth` untuk hash/verifikasi password, penerbitan/validasi access JWT, pembuatan/hash refresh secret, dan helper CSRF. Semua primitive memakai `crypto/rand`; dependency JWT tidak boleh memilih algoritma dari token.
3. Tambahkan query sqlc `GetRefreshSessionByTokenHashForUpdate` dan query operator `CreateInitialAdmin`. Bootstrap mengunci advisory transaction key konstan sebelum query agar hanya satu admin awal dapat dibuat. Regenerasi sqlc. Jangan mengubah file migration ISSUE-002 yang sudah diterapkan.
4. Buat `service.Auth`: registrasi, login, refresh, logout, profil, dan self-delete. Registrasi membuka transaction pgx, membuat user lewat query publik tanpa role, lalu menanam kategori default terkonfigurasi untuk user itu dan commit hanya bila seluruh insert berhasil. Kegagalan seed membatalkan user dan seluruh seed.
5. Implementasikan login dengan rate limiter lalu pembacaan user aktif dan verifikasi Argon2id. Kegagalan user tidak ada, password salah, akun terhapus, atau session tidak valid memakai body/status generik yang sama agar tidak mengungkap email.
6. Implementasikan rotasi/replay refresh dan logout sesuai transaksi/lock di atas. Profile `GET /me` memakai actor context; `PATCH /me` hanya menerima timezone dan version; `DELETE /me` memakai version, mengubah `is_delete`, dan mencabut seluruh refresh sessions dalam satu transaksi.
7. Tambahkan middleware Bearer authentication dan `RequireAdmin`. USER dapat memakai route personal yang tersedia; ADMIN juga mendapat actor context tetapi belum menerima route admin baru pada issue ini. Semua route masa depan tetap mengotorisasi role dari database, bukan JWT.
8. Validasi timezone menggunakan `time.LoadLocation`: terima `UTC` atau identifier IANA dari tzdata, tolak kosong, `Local`, abbreviation/fixed-offset yang bukan identifier IANA, dan nilai lebih dari 64 byte. Simpan `location.String()` yang tervalidasi; currency tetap `IDR` dan tidak dapat diubah profil.
9. Tambahkan handler Echo untuk `POST /api/v1/auth/register`, `/login`, `/refresh`, `/logout`, serta `GET`, `PATCH`, `DELETE /api/v1/me`. Decoder JSON menolak field tidak dikenal, terutama `role`, `isDelete`, owner, dan `client_type` yang tidak valid. Handler tidak pernah mengembalikan hash/password/session secret.
10. Tambahkan `cmd/admin` untuk bootstrap operator eksplisit. Command tidak melakukan promosi otomatis atau tersedia lewat HTTP; ia meminta password dari terminal tanpa echo, memakai `CreateInitialAdmin`, dan gagal jika active admin sudah ada. Bootstrap memakai password/timezone validator dan transaction sehingga tidak ada admin setengah jadi.

## Endpoint dan perilaku

| Route | Perilaku ISSUE-003 |
| --- | --- |
| `POST /api/v1/auth/register` | `{email,password,timezone}`; role selalu `user`; transaction membuat user dan kategori default; hasil 201 tidak berisi hash/token. |
| `POST /api/v1/auth/login` | `{email,password,client_type:web/native}`; rate-limited dan generic 401; sukses membuat family/session baru dan memberi access JWT plus cookie browser atau refresh secret native. |
| `POST /api/v1/auth/refresh` | Browser memakai cookie + Origin + CSRF; native memakai `{refresh_token}`; rotasi atomik, replay mencabut family, respons generik 401 saat gagal. |
| `POST /api/v1/auth/logout` | Browser memakai cookie + Origin + CSRF; native memakai `{refresh_token}`; revoke family idempotent dan 204. |
| `GET /api/v1/me` | Bearer actor aktif; kembalikan metadata diri tanpa password/token/hash. |
| `PATCH /api/v1/me` | Bearer actor aktif; hanya `{timezone,version}`, timezone valid IANA, optimistic-lock conflict 409. |
| `DELETE /api/v1/me` | Bearer actor aktif dan `If-Match`; transaction soft-delete akun dan cabut semua family/session, lalu 204. |

## File yang dibuat atau diubah

| Path | Perubahan |
| --- | --- |
| `go.mod`, `go.sum` | Tambahkan JWT, Argon2id, dan UUID dependencies yang dipilih. |
| `.env.example`, `README.md`, `README.en.md` | Dokumentasikan variabel auth tanpa nilai rahasia dan cara bootstrap/operator yang aman. |
| `internal/config/config.go`, `internal/config/config_test.go` | Tambahkan `AuthConfig`, parse/validasi HMAC key, issuer, audience, origins, dan cookie security. |
| `internal/auth/password.go`, `internal/auth/password_test.go` | Argon2id PHC dan kebijakan password. |
| `internal/auth/jwt.go`, `internal/auth/jwt_test.go` | Penerbitan dan validasi JWT strict. |
| `internal/auth/refresh.go`, `internal/auth/refresh_test.go` | Random secret, SHA-256, CSRF token, constant-time comparison. |
| `db/queries/users.sql` | Tambahkan query `CreateInitialAdmin` operator; query registrasi publik tetap tidak menerima role. |
| `db/queries/refresh_sessions.sql` | Tambahkan lookup refresh session `FOR UPDATE` untuk rotasi/logout serial. |
| `internal/repository/sqlc/users.sql.go`, `internal/repository/sqlc/refresh_sessions.sql.go`, `internal/repository/sqlc/querier.go` | Output sqlc yang diregenerasi. |
| `internal/service/auth.go`, `internal/service/auth_test.go`, `internal/service/auth_integration_test.go` | Use case auth/profile, transaction boundaries, dan fixture PostgreSQL nyata. |
| `internal/middleware/auth.go`, `internal/middleware/auth_test.go` | Bearer parsing, actor state database, USER/ADMIN guard, cache control. |
| `internal/middleware/rate_limit.go`, `internal/middleware/rate_limit_test.go` | Throttle login bounded dan aman terhadap race. |
| `internal/handlers/auth.go`, `internal/handlers/auth_test.go`, `internal/handlers/routes.go` | HTTP decoder/response/cookie dan pendaftaran route `/api/v1`. |
| `cmd/api/main.go` | Compose pool, sqlc, auth service, handler, rate limiter, dan middleware. |
| `cmd/admin/main.go`, `cmd/admin/main_test.go` | Bootstrap admin operator eksplisit. |
| `Makefile` | Tambahkan target integration auth bila suite memerlukan `TEST_DATABASE_URL`. |

Tidak ada file `db/migrations/*.sql` yang dimodifikasi atau dibuat pada issue ini: schema ISSUE-002 sudah mendukung auth dan applied migration tidak boleh ditulis ulang. `categories.sql` dipakai apa adanya untuk seed; tidak ada endpoint CRUD kategori pada ISSUE-003.

## Validasi

- Unit: batas password/PHC cacat, Argon2 verification, JWT algorithm confusion/signature/issuer/audience/subject/expiry/issued-at/JTI, random token/hash, dan config tanpa kebocoran secret.
- Service: role injection registrasi, default USER, seed atomik, timezone invalid, generic login failure, 15-minute JWT, 30-day refresh, rotate/logout, expiry, replay revoke-family, self-delete revoke-all, dan lock/race refresh.
- HTTP/security: Bearer kosong/malformed, actor terhapus, USER versus ADMIN guard, unknown JSON fields, native versus browser transport, HttpOnly/Secure/SameSite/Path cookie attributes, Origin dan CSRF mismatch, `no-store`, safe logger tanpa password/JWT/cookie/refresh secret, dan 429 rate limit.
- PostgreSQL integration: registrasi+seed rollback, concurrent refresh creates one successor then revokes family, logout/replay behavior, transaction atomicity, serta profile/self-delete version conflict pada PostgreSQL disposable.
- Quality: `make fmt-check`, `make test`, `make test-race`, `make vet`, `make build`, `make verify`, `make sqlc-vet`, `make sqlc-check`, dan suite auth integration dengan `TEST_DATABASE_URL`.

## Keputusan yang belum terselesaikan sebelum implementasi

1. Daftar kategori default (nama dan type income/expense) belum ada di requirements, database, atau API. Nilai ini diperlukan agar registrasi dapat melakukan seed deterministik.
2. Nilai production untuk `AUTH_ALLOWED_ORIGINS` belum ditentukan. Domain harus cocok persis dengan browser origin; wildcard tidak akan diterima.
3. Topologi reverse proxy belum ditentukan. Rate limiter akan memakai alamat peer TCP sampai daftar proxy tepercaya diset; jangan percaya `X-Forwarded-For` tanpa keputusan ini.
4. Cookie `Secure` mengharuskan HTTPS. Local development harus memakai HTTPS proxy atau diberi konfigurasi development yang jelas dan tidak boleh dipakai production.
5. Bootstrap operator perlu menetapkan apakah satu active ADMIN maksimum saat bootstrap dan bagaimana command memperoleh password secara non-interaktif di deployment. Rencana ini memilih stdin TTY dan menolak bootstrap kedua.

## Batasan dan recovery

Jangan menyimpan atau mencatat password, JWT, cookie, refresh secret, atau key signing. Jangan mengubah applied migration atau menghapus data bersama. Uji rollback hanya di fixture disposable; pemulihan production menggunakan migrasi maju yang direview. Sesudah keputusan terbuka disetujui dan implementasi selesai, perbarui issue/index dengan bukti test nyata.
