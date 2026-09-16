# Arsitektur

Draf v0.3 • 14 September 2026
Khusus perencanaan; versi akan diverifikasi terhadap dokumentasi kompatibilitas resmi dan dipatok saat setup.

## Komponen dan batas

monelog-api menggunakan Go, Echo, PostgreSQL, sqlc, dan OpenAPI/oapi-codegen.
monelog-app menggunakan Vue 3 + JavaScript untuk browser serta Capacitor Android/iOS. Backend diselesaikan sebelum frontend.
Gunakan satu API dan proses worker di tahap berikutnya untuk job ekspor/backup. Redis dan microservice tidak diperlukan untuk desain MVP.
Deployment browser sebaiknya mem-proxy /api pada origin yang sama; client native memanggil API HTTPS langsung. Server tetap menjadi sumber kebenaran; paket mobile tidak berarti sinkronisasi offline.

Middleware mengautentikasi aktor dan state akun saat ini → handler memvalidasi input → service mengotorisasi tindakan/pemilik → repository menjalankan query dengan scope eksplisit.
Scope personal: actor=owner. Scope admin: actor tetap admin yang login; owner berasal dari path target yang tervalidasi. Scope admin mengizinkan create/read/update/delete/restore serta operasi ekspor/templat/backup di tahap berikutnya.
Aturan domain yang sama berlaku pada kedua scope. Pertahankan predicate owner, constraint kategori/tipe, dan versi optimistis pada setiap mutasi. Jangan mengandalkan kontrol frontend atau klaim role JWT saja.

## Struktur proyek

Struktur kanonis yang direncanakan untuk monelog-api:

```
monelog-api/
├── cmd/
│   ├── api/
│   │   └── main.go              # komposisi dan server HTTP
│   ├── worker/
│   │   └── main.go              # job ekspor/backup tahap berikutnya
│   └── admin/
│       └── main.go              # bootstrap/pemulihan admin
├── internal/
│   ├── api/                     # kode OpenAPI hasil generate
│   ├── config/                  # konfigurasi bertipe dan tervalidasi
│   ├── handlers/                # adapter HTTP personal dan admin
│   ├── middleware/              # autentikasi, role, logging aman, recovery
│   ├── service/                 # aturan bisnis, scope, otorisasi, dan job
│   └── repository/
│       ├── postgresql/          # koneksi/pool PostgreSQL yang ditulis manual
│       └── sqlc/                # query SQL hasil generate
├── db/
│   ├── migrations/              # migrasi skema berversi
│   └── queries/                 # sumber query sqlc
├── api/
│   └── openapi.yaml             # kontrak API kanonis
├── tests/
│   └── integration/             # pengujian PostgreSQL/HTTP
├── docs/                        # dokumentasi kanonis
├── go.mod
└── go.sum
```

Business logic berada di internal/service; internal/handlers hanya menerjemahkan HTTP ke service. Jika implementasi memakai script/sqlc atau nama direktori lain, sesuaikan plan dan dokumentasi setelah keputusan struktur dibuat.

## Direktori yang diusulkan

| Path | Tanggung jawab |
| --- | --- |
| cmd/api/main.go | Komposisi, startup, dan shutdown HTTP |
| cmd/worker/main.go | Job latar belakang tahap berikutnya |
| cmd/admin/main.go | Bootstrap/pemulihan admin awal untuk operator server tepercaya |
| internal/config | Konfigurasi bertipe dan tervalidasi |
| internal/handlers | Adapter HTTP personal dan admin |
| internal/middleware | Autentikasi, pemeriksaan akun/peran, logging aman, recovery, throttling |
| internal/service | Otorisasi aksi, scope, aturan keuangan, dan siklus hidup job |
| internal/repository | Batas repository yang ditulis manual |
| internal/repository/postgresql | Koneksi dan pool PostgreSQL |
| internal/repository/sqlc | Query SQL hasil generate |
| internal/api | Kode OpenAPI hasil generate |
| db/migrations, db/queries | Skema berversi dan sumber sqlc |
| api/openapi.yaml | Kontrak kanonis yang dapat dibaca mesin |
| tests/integration | Pengujian PostgreSQL/HTTP nyata |
| docs | Paket perencanaan kanonis |

Direktori frontend: src/views, components, services, stores, router, dan utils dengan pengujian terarah. Admin Management memakai ulang form transaksi tervalidasi dengan konteks owner eksplisit.

## Autentikasi dan peran

Gunakan access JWT dengan TTL maksimum 24 jam dan secret refresh acak berotasi yang di-hash dalam sesi yang dapat dicabut.
Browser: access token di memori; cookie refresh HttpOnly/Secure dengan SameSite, pemeriksaan CSRF, dan origin yang sesuai.
Native: penyimpanan refresh aman berbasis OS; access token di memori. Jangan gunakan localStorage untuk secret refresh.
Validasi algoritma JWT, issuer, audience, dan expiry; sub selalu aktor. JWT tidak membawa role; role aktor selalu dimuat dari database saat ini.
Registrasi publik default ke user dan menolak override role/owner. Pembuatan akun/perubahan role admin memakai endpoint terlindungi dan pemeriksaan role saat ini. GET /me mengembalikan role untuk UI.
Bootstrap admin pertama melalui perintah operator eksplisit; tidak ada akun hardcode atau promosi pengguna pertama otomatis.
Akun dengan is_delete=true menolak semua request terlindungi, termasuk dengan access JWT lama. Perubahan role/penghapusan akun mencabut sesi refresh; demotion yang tersimpan memblokir request admin baru.

## Transaksi, otorisasi, dan audit

Request baca memeriksa state/peran aktor sebelum mencari target. Semua mutasi bisnis membuka transaksi DB, mengunci baris akun aktor/target dalam urutan UUID stabil, memeriksa ulang otorisasi dan aktivitas target, lalu memvalidasi/memperbarui resource dan menyimpan audit sebelum commit.
Perubahan akun/peran mengikuti aturan lock yang sama; penghapusan/demotion yang tersimpan bersamaan tidak dapat dilewati oleh mutasi dengan pemeriksaan lama.
Gunakan metode bernama seperti CreateTransaction(scope), SoftDeleteTransaction(scope,version), dan RestoreTransaction(scope,version); jangan memakai assignment owner atau role yang dikendalikan client secara generik.
Tindakan admin mencatat actor versus owner, operasi, metadata resource/version yang aman, dan hasil. Mutasi admin serta insert audit atomik; kegagalan audit melakukan rollback. Baca admin wajib menyimpan audit sebelum respons.
Catatan audit hanya append-only. Admin dapat memeriksanya melalui endpoint terlindungi; kredensial mentah/payload keuangan tidak dikembalikan atau dicatat.

## Penghapusan lunak

Field API isDelete dipetakan ke Go IsDelete dan SQL is_delete BOOLEAN NOT NULL DEFAULT FALSE.
Users/categories/transactions/templates menggunakan flag ini; flag tersebut menggantikan sinyal penghapusan berbasis timestamp dan usulan archive kategori sebelumnya.
Create default false. Delete mengatur true dan memperbarui versi/waktu audit. Restore mengatur false setelah pemeriksaan ownership, versi, dan kategori terkait.
Query aktif secara eksplisit memfilter transactions.is_delete=false. Query Trash secara eksplisit meminta true; otorisasi owner/admin tetap sama.
Kategori terhapus disembunyikan dari selector, tetapi label kategori historis tetap tersedia melalui join berscope. Jangan mengeluarkan transaksi historis hanya karena kategori hasil join terhapus.
Pengguna terhapus tidak dapat login/job tanpa mengubah flag baris anak; admin dapat memeriksa data yang dipertahankan dan memulihkan akun. Tidak ada hard delete aplikasi untuk entitas ini.

## Job dan operasi provider

Simpan owner_user_id, requested_by, request_mode, dan filter yang tidak dapat diubah pada job ekspor/backup.
Pengguna biasa hanya beroperasi pada scope owner sendiri; admin aktif dapat mengelola job dan unduhan owner terpilih.
Sebelum eksekusi, periksa ulang state akun owner/requester dan role admin saat request_mode=admin. Job yang dibatalkan/tidak berwenang tidak memanggil provider.
Jadwal berulang menyimpan owner, aktor pemberi otorisasi, dan mode; jadwal jeda memerlukan resume yang terotorisasi.
Panggilan eksternal dilakukan setelah job+audit tersimpan; jalankan di luar lock DB dan catat sukses/gagal/retry remote secara eksplisit.
Operasi Drive memerlukan koneksi/consent provider valid milik owner terpilih. Izin admin tidak membuat otorisasi OAuth.
Backup/restore mempertahankan flag isDelete, memvalidasi pemetaan owner terpilih, dan mengecualikan role/kredensial/riwayat audit. Restore tidak dapat mempromosikan user; operasi role admin khusus yang dapat melakukannya.

## Konsistensi dan client

Uang: NUMERIC(14,2), decimal Go eksak/minor unit, string desimal JSON. Jangan gunakan floating point atau kolom saldo tersimpan.
Versi record menjaga edit/delete/restore. Idempotensi create dibatasi oleh owner terotorisasi dan client_request_id; actor sebenarnya tetap dicatat terpisah.
Report/ekspor memakai agregasi dan filter owner/baris aktif yang sama.
Ikat cursor dan key cache/request pada actor, mode, owner, endpoint, filter, dan state penghapusan.
Semua respons terautentikasi memakai Cache-Control: no-store. Jangan menyimpan data pengguna lain pada cache offline.
UI admin menampilkan owner terpilih dan mengaktifkan pengelolaan. Simpan/buang sebelum mengganti form yang belum tersimpan; operasi yang telah dikirim tetap terikat target awal. Buang respons terlambat setelah target/sesi berubah.

## Operasi dan validasi

Pisahkan development/staging/production, rollout migrasi, dan rencana recovery yang direview.
Backup/recovery bencana database dan backup personal Drive adalah dua hal terpisah.
Log terstruktur menyamarkan payload keuangan, password, token, dan secret provider; health/live dan health/ready tidak membocorkan konfigurasi.
CI: formatting, vet/lint, pengujian bermakna, drift kode generate, build frontend, dan pemeriksaan dependensi.
Commit kode sqlc/OpenAPI hasil generate; kecualikan .env, signing key, dump, dan secret provisioning.
Validasi kebutuhan toolchain saat ini serta ketersediaan macOS/Xcode/signing pada milestone mobile.
