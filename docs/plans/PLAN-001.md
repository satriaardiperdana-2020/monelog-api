# PLAN-001 — Setup Proyek dan Fondasi Layanan

Issue terkait: [ISSUE-001](../issues/ISSUE-001-project-setup.md)

Status: **Refined — siap diimplementasikan setelah review**

## 1. Tujuan

Membuat fondasi executable API Go yang dapat dikompilasi, dikonfigurasi dari environment variable, terhubung ke PostgreSQL melalui pgxpool, menyediakan health check, mencatat request secara aman, dan berhenti dengan graceful shutdown.

Plan ini hanya mencakup ISSUE-001. Domain keuangan, autentikasi, migrasi database, OpenAPI generation, dan sqlc generation tetap dikerjakan dalam issue masing-masing.

## 2. Kondisi Repository Saat Plan Diperbarui

- Repository masih berisi dokumentasi; belum ada go.mod, source code Go, workflow CI, migrasi, atau spesifikasi OpenAPI.
- Branch aktif adalah feature/ISSUE-001-project-setup.
- Go SDK belum tersedia di PATH pada environment saat ini: perintah go version menghasilkan go: command not found.
- Tidak ada AGENTS.md atau instruksi repository tambahan yang ditemukan.
- Ada perubahan dokumentasi yang belum di-commit dan direktori .idea yang tidak dilacak. Implementasi ISSUE-001 harus mempertahankan perubahan tersebut dan tidak menyentuh .idea.
- Tidak ada file .env atau secret yang dibaca dalam penyusunan plan ini.

## 3. Keputusan Teknis dan Versi

| Komponen | Pilihan | Versi/pin | Alasan dan kompatibilitas |
|---|---|---|---|
| Go SDK | Go | 1.27.1 | Patch stabil terbaru saat plan diperbarui. go.mod memakai go 1.27.0 dan toolchain go1.27.1. |
| Module path | GitHub module | github.com/satriaardiperdana-2020/monelog-api | Sesuai lokasi repository proyek. |
| HTTP framework | Echo v5 | github.com/labstack/echo/v5 v5.0.2 | Versi v5 stabil; menyediakan routing, recovery middleware, dan integrasi dengan slog. |
| PostgreSQL driver/pool | pgx v5 | github.com/jackc/pgx/v5 v5.11.0 | Mendukung Go 1.25+, sehingga kompatibel dengan Go 1.27.1. |
| Logging | log/slog | Go standard library | Structured logging tanpa dependency tambahan. |
| Config | os, time, net/url | Go standard library | Environment variable sederhana dan validasi eksplisit; tidak memerlukan Viper atau dotenv. |
| Tests | testing, httptest | Go standard library | Cukup untuk unit test dan lifecycle test dengan fake dependency. |
| CI | GitHub Actions | Action resmi dipin ke full commit SHA saat implementasi | Menjalankan format check, test, race detector, vet, build, dan module verification. |

Dependency yang sengaja belum ditambahkan pada ISSUE-001:

- sqlc v1.31.1 menjadi baseline kompatibilitas untuk ISSUE-002.
- oapi-codegen v2.8.0 menjadi baseline kompatibilitas untuk ISSUE-004 dan mendukung Echo v5.
- Library JWT, password hashing, authorization, dan rate limiting ditentukan dalam ISSUE-003.

Referensi versi yang diverifikasi pada 15 September 2026:

- [Go release history](https://go.dev/doc/devel/release)
- [Echo repository dan release](https://github.com/labstack/echo)
- [pgx changelog](https://github.com/jackc/pgx/blob/master/CHANGELOG.md)
- [oapi-codegen releases](https://github.com/oapi-codegen/oapi-codegen/releases)
- [sqlc releases](https://github.com/sqlc-dev/sqlc/releases)

## 4. Struktur dan File

### 4.1 File yang dibuat

~~~text
monelog-api/
├── .github/
│   └── workflows/
│       └── ci.yml
├── cmd/
│   └── api/
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── handlers/
│   │   ├── health.go
│   │   ├── health_test.go
│   │   └── routes.go
│   ├── middleware/
│   │   ├── request_logger.go
│   │   └── request_logger_test.go
│   ├── repository/
│   │   └── postgresql/
│   │       ├── connect.go
│   │       └── connect_test.go
│   └── service/
│       ├── health.go
│       └── health_test.go
├── .env.example
├── .gitignore
├── .go-version
├── go.mod
└── go.sum
~~~

Catatan struktur:

- internal/repository/postgresql berisi kode koneksi pgxpool yang ditulis manual. Output sqlc akan ditempatkan di internal/repository/sqlc oleh ISSUE-002.
- internal/api belum dibuat karena generated OpenAPI code adalah scope ISSUE-004.
- internal/service pada ISSUE-001 hanya mengorkestrasi status liveness/readiness. Business service untuk autentikasi dan keuangan tetap dibuat oleh issue domain terkait.
- cmd/admin dibuat pada ISSUE-003. cmd/worker dibuat saat background jobs mulai diperlukan.
- db/migrations dan db/queries dibuat pada ISSUE-002.

### 4.2 File yang dimodifikasi

| File | Perubahan |
|---|---|
| README.md | Tambahkan prerequisite, environment variable, command run/test, dan health endpoint dalam Bahasa Indonesia. Pertahankan pemilih bahasa. |
| README.en.md | Tambahkan informasi yang sama dalam bahasa Inggris. |
| docs/architecture.md | Sinkronkan struktur final hanya jika implementasi menghasilkan perbedaan terhadap struktur yang sudah didokumentasikan. |
| docs-en/architecture.md | Sinkronkan perubahan struktur yang sama dalam bahasa Inggris. |
| docs/issues/ISSUE-001-project-setup.md | Tandai acceptance criteria setelah seluruh verifikasi lulus. |
| docs/issues.md | Ubah status ISSUE-001 hanya setelah implementasi dan verifikasi selesai. |

Perubahan dokumentasi lokal yang sudah ada harus digabungkan dengan hati-hati; file tidak boleh ditimpa secara keseluruhan tanpa memeriksa diff.

## 5. Konfigurasi Runtime

Konfigurasi dibaca dari environment variable. Aplikasi tidak melakukan automatic dotenv loading. File .env.example hanya berisi placeholder aman.

| Variable | Wajib | Default | Validasi |
|---|---|---|---|
| APP_ENV | Tidak | development | Nilai non-kosong. |
| HTTP_ADDR | Tidak | 127.0.0.1:8080 | Format host:port yang valid. |
| HTTP_READ_HEADER_TIMEOUT | Tidak | 5s | Duration lebih dari nol. |
| HTTP_READ_TIMEOUT | Tidak | 10s | Duration lebih dari nol. |
| HTTP_WRITE_TIMEOUT | Tidak | 15s | Duration lebih dari nol. |
| HTTP_IDLE_TIMEOUT | Tidak | 60s | Duration lebih dari nol. |
| SHUTDOWN_TIMEOUT | Tidak | 10s | Duration lebih dari nol. |
| DATABASE_URL | Ya | — | URL valid dengan scheme postgres atau postgresql. |

Nilai DATABASE_URL, header Authorization, cookie, query string, dan request body tidak boleh ditulis ke log atau response error.

## 6. Perilaku Fondasi Layanan

### 6.1 Startup dan PostgreSQL

1. Muat dan validasi konfigurasi.
2. Buat logger slog.
3. Parse konfigurasi pgxpool dan buat pool.
4. Daftarkan routes dan middleware.
5. Jalankan http.Server dengan timeout eksplisit.
6. Terima SIGINT atau SIGTERM melalui signal.NotifyContext.
7. Jalankan Shutdown dengan batas SHUTDOWN_TIMEOUT, lalu tutup pool tepat satu kali.

DATABASE_URL yang tidak valid menggagalkan startup. PostgreSQL yang sementara tidak tersedia tidak harus mematikan proses setelah konfigurasi valid; endpoint readiness melaporkan kondisi tersebut. Ini memungkinkan orchestrator membedakan proses hidup dari service yang siap menerima traffic.

### 6.2 Health endpoints

- GET /health/live mengembalikan HTTP 200 jika proses HTTP hidup dan tidak bergantung pada database.
- GET /health/ready melakukan Ping melalui interface kecil yang dapat diganti fake pada test.
- Readiness mengembalikan HTTP 200 saat PostgreSQL tersedia dan HTTP 503 saat koneksi gagal atau timeout.
- Response menggunakan payload JSON kecil dan stabil, misalnya status: ok atau status: unavailable, tanpa detail konfigurasi atau error internal.
- Health check database mempunyai timeout sendiri agar request tidak menggantung.

### 6.3 Middleware

- Recovery menangkap panic dan mengembalikan HTTP 500 yang tersanitasi.
- Request logger mencatat request ID, HTTP method, route template, status, duration, dan remote address yang sudah dinormalisasi.
- Logger tidak mencatat Authorization, Cookie, query string, body, DATABASE_URL, password, atau token.
- Request ID yang valid dari client dapat dipertahankan; jika tidak ada, middleware membuat ID baru.

## 7. Urutan Implementasi

1. **Siapkan toolchain** — pasang Go 1.27.1, pastikan go version dan go env sesuai, lalu buat .go-version dan module metadata.
2. **Tambahkan dependency minimum** — pin Echo v5.0.2 dan pgx v5.11.0, lalu hasilkan go.sum dengan go mod tidy.
3. **Implementasikan config** — parsing environment variable, default, dan seluruh validasi sebelum komponen runtime dibuat.
4. **Implementasikan adapter PostgreSQL** — pgxpool configuration, Ping abstraction untuk readiness, timeout, dan Close.
5. **Implementasikan health service** — orkestrasi liveness/readiness dan pemetaan kegagalan dependency tanpa detail sensitif.
6. **Implementasikan health handlers dan routes** — terjemahkan hasil service menjadi response HTTP tersanitasi.
7. **Implementasikan middleware** — request ID, safe structured logging, dan recovery.
8. **Implementasikan entry point** — dependency wiring, http.Server timeout, signal handling, graceful shutdown, dan pool cleanup.
9. **Tambahkan tests** — unit test lebih dahulu pada boundary config/health/logging, lalu lifecycle test server.
10. **Tambahkan CI** — format check, go test, race detector, vet, build, dan go mod verify.
11. **Perbarui dokumentasi** — quick start bilingual dan struktur arsitektur jika diperlukan.
12. **Verifikasi acceptance criteria** — jalankan seluruh command pada checkout bersih sebelum mengubah status issue.

## 8. Command Implementasi dan Verifikasi

### 8.1 Pemeriksaan SDK

~~~bash
go version
go env GOVERSION GOOS GOARCH
~~~

Expected SDK: go1.27.1. Implementasi tidak dimulai sampai kedua command tersedia.

### 8.2 Inisialisasi module dan dependency

Command ini dijalankan satu kali saat file belum ada:

~~~bash
go mod init github.com/satriaardiperdana-2020/monelog-api
go get github.com/labstack/echo/v5@v5.0.2
go get github.com/jackc/pgx/v5@v5.11.0
go mod tidy
~~~

### 8.3 Quality checks

~~~bash
test -z "$(gofmt -l cmd internal)"
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/api
go mod verify
~~~

Jika coverage dibutuhkan untuk review:

~~~bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
~~~

Coverage adalah bukti tambahan, bukan target persentase yang mendorong test tanpa nilai perilaku.

### 8.4 Run lokal dan smoke test

Gunakan DSN lokal milik developer tanpa menyimpannya di Git atau output CI:

~~~bash
APP_ENV=development \
HTTP_ADDR=127.0.0.1:8080 \
DATABASE_URL='<local-postgresql-dsn>' \
go run ./cmd/api

curl --fail http://127.0.0.1:8080/health/live
curl --fail http://127.0.0.1:8080/health/ready
~~~

Uji juga kondisi database tidak tersedia. /health/live harus tetap 200 dan /health/ready harus 503.

## 9. Tests yang Bermakna

### 9.1 Config tests

- Konfigurasi valid menghasilkan seluruh nilai dan default yang diharapkan.
- DATABASE_URL kosong atau scheme tidak valid menghasilkan startup error yang jelas tanpa membocorkan DSN.
- HTTP_ADDR tidak valid ditolak.
- Duration tidak valid, nol, atau negatif ditolak.
- Tidak ada test yang membaca .env lokal; test memakai t.Setenv.

### 9.2 Service, handler, dan repository boundary tests

- Liveness mengembalikan 200 tanpa memanggil database.
- Readiness mengembalikan 200 ketika fake checker berhasil.
- Readiness mengembalikan 503 ketika checker gagal atau context timeout.
- Error readiness tidak memuat host database, credential, DSN, atau pesan driver.
- PostgreSQL config menolak DSN malformed sebelum server berjalan.
- Health handler hanya menerjemahkan hasil service menjadi status HTTP dan payload publik; error dependency tetap berada di service/log internal.

### 9.3 Middleware tests

- Panic pada handler dipulihkan menjadi response 500 yang tersanitasi.
- Log sukses dan gagal memuat request ID, method, route template, status, dan duration.
- Log tidak memuat Authorization, Cookie, query string, request body, atau nilai secret uji.
- Request ID diteruskan ke response dan dibuat ketika header tidak tersedia atau tidak valid.

### 9.4 Lifecycle tests

- http.Server memakai semua timeout dari config.
- Context yang dibatalkan menyebabkan server berhenti dalam SHUTDOWN_TIMEOUT.
- Cleanup menutup pool tepat satu kali, termasuk ketika startup setelah pool gagal pada komponen berikutnya.
- In-flight request diberi kesempatan selesai selama grace period; test menggunakan handler yang dikendalikan channel, bukan sleep panjang.
- Route domain seperti /api/v1/transactions tetap 404 untuk membuktikan ISSUE-001 tidak membuat placeholder business endpoint.

Test unit menggunakan fake readiness checker dan fake closer. PostgreSQL container bukan syarat ISSUE-001 karena belum ada migrasi atau query domain; smoke test lokal cukup untuk memverifikasi koneksi nyata.

## 10. CI

Workflow .github/workflows/ci.yml berjalan pada pull request dan push ke main dengan langkah berikut:

1. Checkout source.
2. Install Go dari .go-version dengan module cache.
3. Pastikan gofmt tidak menghasilkan perubahan.
4. Jalankan go mod verify.
5. Jalankan go test ./... dan go test -race ./....
6. Jalankan go vet ./....
7. Jalankan go build ./cmd/api.

CI tidak membutuhkan DATABASE_URL karena test menggunakan fake. Action GitHub resmi dipin ke full commit SHA yang sesuai dengan major release aktif saat workflow dibuat.

## 11. Risiko dan Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Go SDK tidak terpasang pada environment saat ini | Implementasi dan validasi tidak dapat dijalankan | Pasang Go 1.27.1 dan verifikasi SDK sebelum membuat module. |
| Echo v5 atau pgx baru membawa perubahan API | Compile atau integrasi middleware gagal | Pin versi exact, gunakan API resmi versi tersebut, dan validasi dengan build serta tests. |
| PostgreSQL tidak tersedia saat startup | Service bisa crash-loop atau salah dianggap siap | Pisahkan liveness dan readiness; validasi DSN saat startup dan laporkan availability melalui readiness. |
| Secret masuk log atau error | Kebocoran credential/token | Allowlist field log dan test dengan nilai sentinel yang wajib tidak muncul. |
| Shutdown tidak selesai atau resource ditutup dua kali | Request terputus dan pool error | Context timeout, cleanup idempotent, serta lifecycle tests terkontrol. |
| Perubahan dokumentasi lokal tertimpa | Pekerjaan yang sudah ada hilang | Periksa git diff dan edit bagian terkait saja; jangan reset atau checkout file. |
| Struktur ISSUE-001 tumpang tindih dengan issue berikutnya | Scope membesar dan generated code dini | Buat hanya package yang mempunyai implementasi nyata pada fondasi layanan. |

## 12. Scope Exclusions

Plan ini tidak mencakup:

- Schema, migrations, seed data, query SQL, dan sqlc generation — ISSUE-002.
- Login, JWT, password hashing, role/ownership authorization, dan rate limiting — ISSUE-003.
- OpenAPI specification, validation middleware dari kontrak, dan oapi-codegen — ISSUE-004.
- CRUD category/transaction, summary, report, admin operation, audit write, session persistence, dan business rules keuangan.
- Worker, job queue, export, backup, template system, frontend, atau mobile client.
- Docker/Kubernetes, deployment cloud, TLS termination, dan provisioning PostgreSQL production.
- Membaca file .env lokal, menyimpan secret, atau membuat credential contoh yang dapat digunakan.

## 13. Pemetaan Acceptance Criteria ISSUE-001

Beberapa acceptance criteria ISSUE-001 berasal dari kontrak lintas issue. Penerapannya pada setup ini dibatasi sebagai berikut:

| Kriteria issue | Bukti pada ISSUE-001 |
|---|---|
| Fondasi fungsional | Config, health, middleware, startup, dan shutdown memiliki test perilaku serta lolos build. |
| Scope pengguna/admin dan owner/actor | Belum ada endpoint atau input domain. Route /api/v1 yang tidak dikenal harus 404; tidak ada cara untuk mengirim role atau owner. Implementasi otorisasi tetap milik ISSUE-003 dan issue domain. |
| isDelete, version, trash/restore, dan resource lintas owner | Tidak ada model, query, atau endpoint data pada ISSUE-001. Perilaku ini tidak boleh dibuat sebagai placeholder dan diuji pada issue pemiliknya. |
| Audit dan generated code | Tidak ada audit write, sqlc output, atau OpenAPI output pada ISSUE-001. Konsistensi dibuktikan dengan tidak menambahkan artefak tersebut sebelum ISSUE-002/004. |
| Dokumentasi | README bilingual dan architecture disinkronkan hanya untuk fondasi yang benar-benar dibuat. |

ISSUE-001 baru dapat ditandai selesai jika bukti yang relevan di atas lulus. Kriteria lintas domain yang belum applicable tetap diteruskan ke issue pemiliknya dan tidak boleh dinyatakan sudah teruji.

## 14. Definition of Done ISSUE-001

- Go 1.27.1 tersedia dan versi dicatat melalui .go-version serta go.mod.
- go build ./cmd/api berhasil.
- Semua config, health, middleware, dan lifecycle tests lulus.
- go test -race ./..., go vet ./..., dan go mod verify lulus.
- Liveness dan readiness menunjukkan perilaku berbeda saat PostgreSQL tidak tersedia.
- SIGINT/SIGTERM menghentikan HTTP server dengan graceful shutdown dan menutup pool.
- Log dan response tidak membocorkan secret atau detail koneksi.
- README Indonesia dan Inggris berisi command setup/run/test yang akurat.
- CI menjalankan seluruh quality checks pada checkout bersih.
- Tidak ada endpoint domain, migration, generated sqlc/OpenAPI code, atau auth implementation yang ikut masuk.

## 15. Recovery

Semua file ISSUE-001 harus masuk dalam satu commit implementasi yang terfokus setelah tests lulus. Jika rollback diperlukan, revert commit tersebut. Jangan memakai reset atau menimpa perubahan dokumentasi yang sudah ada.

## 16. Checkpoint Plan

Plan ini hanya memperjelas pekerjaan. Tidak ada source code, GitHub state, commit, branch, atau Git history yang diubah sebagai bagian dari refinement ini.
