# Desain API

Draf v0.3 • 14 September 2026
Base /api/v1. Operasi terlindungi memakai access JWT Bearer dan otorisasi sisi server.
Ini adalah panduan desain; Issue 004 menghasilkan api/openapi.yaml tervalidasi dan interface hasil generate sebelum implementasi domain.
Nama JSON snake_case kecuali boolean isDelete yang diminta. Go IsDelete dipetakan ke SQL is_delete.
Uang memakai string desimal eksak, misalnya "43500.00"; jumlah unsigned cocok dengan ^[0-9]+\.[0-9]{2}$, sedangkan difference boleh negatif.
Tanggal YYYY-MM-DD, timestamp UTC RFC3339, ID UUID opak.

## Otorisasi

Path personal selalu membatasi owner ke actor. Path admin memerlukan admin aktif saat ini dan membatasi data ke user target eksplisit.
CRUD admin, ekspor, templat, dan backup diizinkan. Pengguna biasa tidak dapat memakai path admin.
Field body yang tidak dikenal dan override owner ditolak; target admin berasal dari path, bukan override body.
Registrasi publik default ke user. role hanya dapat ditulis melalui operasi akun admin terlindungi.
Actor terhapus saat ini mendapat 401; non-admin pada path admin mendapat 403 sebelum target dicari.
Untuk admin saat ini: target rusak mendapat 400; resource tidak ada/salah target mendapat 404. Membaca data tersimpan milik target terhapus diizinkan; write keuangan memerlukan pemulihan akun lebih dahulu.
Semua respons/error terautentikasi memakai Cache-Control: no-store. Mutasi admin memerlukan penyimpanan data+audit atomik; storage gagal mengembalikan 503 tanpa mutasi tersimpan.

## Route akun dan autentikasi

| Method | Path | Input / hasil |
| --- | --- | --- |
| POST | /auth/register | email,password,timezone → 201 user biasa; override role/isDelete ditolak |
| POST | /auth/login | email,password,client_type:web/native → access token dan transport refresh |
| POST | /auth/refresh | cookie refresh browser+CSRF atau secret refresh native → sesi berotasi |
| POST | /auth/logout | transport refresh → cabut family, 204 |
| GET | /me | id,email,role,timezone,currency,isDelete,version sendiri |
| PATCH | /me | timezone,version → profil sendiri diperbarui |
| DELETE | /me | If-Match version → hapus lunak akun sendiri dan cabut sesi, 204 |
| GET | /admin/users | q?,isDelete=false,limit,cursor → direktori user berpaginasi |
| POST | /admin/users | email,password,timezone,role:user/admin → user aktif baru; role kosong default user |
| GET | /admin/users/{user_id} | isDelete=false atau true → metadata user terpilih |
| PATCH | /admin/users/{user_id} | timezone dan/atau role, version wajib → user aktif diperbarui |
| DELETE | /admin/users/{user_id} | If-Match version → hapus lunak user terpilih dan cabut sesi, 204 |
| POST | /admin/users/{user_id}/restore | version → user dipulihkan, isDelete=false |
| GET | /admin/audit-events | actor_user_id?,target_user_id?,action?,start_time,end_time,limit,cursor → metadata audit aman |

DTO direktori admin: id,email,role,timezone,currency,isDelete,version. Email immutable pada kontrak profil awal; perubahan identitas/auth memerlukan alur tervalidasi sendiri.
q direktori adalah substring email literal yang di-trim, panjang 1–100 jika ada; parameterize dan escape wildcard. Urutan email ASC,id ASC; limit default 30, maksimum 100.
Audit memakai created_at DESC,id DESC dan rentang maksimal 366 hari; tidak memuat password, token provider, atau payload finansial. Event audit bukan baris bisnis yang dapat diedit.
sub session/JWT tetap actor, bukan target. Perubahan role mencabut refresh session dan berlaku pada request admin berikutnya.
Akun terhapus tidak dapat memakai route self; admin aktif memulihkannya.
Registrasi publik dan pembuatan admin melakukan hash password yang diberikan; password tidak pernah dikembalikan atau dicatat.

`POST /auth/login` mengembalikan `401 AUTHENTICATION_FAILED` untuk email yang tidak dikenal maupun password yang salah. Route terlindungi mengembalikan kode yang sama untuk Bearer token yang hilang, kedaluwarsa, memiliki signature/algoritma/issuer/audience tidak valid, atau membawa claim `role`; detailnya tidak dibedakan agar tidak membocorkan kredensial atau mekanisme validasi.

## Route keuangan personal

| Method | Path | Input / hasil |
| --- | --- | --- |
| GET | /categories | type?,isDelete=false,limit,cursor → kategori sendiri |
| POST | /categories | name,type → 201 kategori aktif |
| GET | /categories/{id} | isDelete=false atau true → kategori sendiri |
| PATCH | /categories/{id} | name,version → kategori aktif diperbarui; type immutable |
| DELETE | /categories/{id} | If-Match version → isDelete=true, 204 |
| POST | /categories/{id}/restore | version → kategori isDelete=false |
| GET | /transactions | start_date,end_date,type?,category_id?,isDelete=false,limit,cursor → transaksi sendiri |
| POST | /transactions | object create → 201; replay idempoten yang sama 200 |
| GET | /transactions/{id} | isDelete=false atau true → transaksi sendiri |
| PATCH | /transactions/{id} | field date/type/category/amount/title dan version → transaksi aktif diperbarui |
| DELETE | /transactions/{id} | If-Match version → isDelete=true, 204 |
| POST | /transactions/{id}/restore | version → transaksi isDelete=false |
| GET | /daily-summaries | start_date,end_date,limit,cursor → total harian aktif |
| GET | /reports/summary | start_date,end_date → income/expense/difference aktif, kategori teratas |
| GET | /reports/breakdown | start_date,end_date,group_by:week/month/category → total aktif terkelompok |

Admin mendukung method, payload, dan status yang sama untuk setiap route keuangan dengan mengganti slash awal menjadi /admin/users/{user_id}/.

Contoh:
- POST /admin/users/{user_id}/transactions membuat record milik user tersebut.
- PATCH /admin/users/{user_id}/transactions/{id} mengedit record user tersebut.
- DELETE /admin/users/{user_id}/transactions/{id} mengatur flag record menjadi true.
- POST /admin/users/{user_id}/transactions/{id}/restore mengaturnya kembali menjadi false.
- GET /admin/users/{user_id}/reports/summary melaporkan baris aktif user tersebut.

Issue 004 harus memperluas pemetaan menjadi operasi OpenAPI terpisah untuk setiap method/path; Issue 013 membuat adapter admin setelah service domain inti.

## Kontrak soft delete

isDelete adalah field respons wajib pada DTO user/category/transaction/template dan hanya menerima boolean JSON. Baris baru default false di server.
Body create/PATCH biasa tidak boleh memuat isDelete. DELETE dan /restore adalah transisi siklus hidup yang diotorisasi.
Pada list/detail resource, query isDelete hanya menerima false atau true, default false. true memilih Trash, termasuk detail yang diperlukan untuk restore.
Detail normal record terhapus mengembalikan 404. Detail Trash terotorisasi mengembalikan versi saat ini.
DELETE memerlukan If-Match: "3" untuk versi 3. Delete berhasil menaikkan versi ke 4 dan mengembalikan 204. Ambil detail Trash dengan versi terbaru sebelum restore.
POST /restore dengan body {"version":4} mengembalikan resource aktif versi 5. Jangan mengatur isDelete secara manual melalui update generik.
Version hilang → 400; version lama/sudah dihapus/sudah dipulihkan → 409; record asing/hilang → 404.
Flag tidak mengubah otorisasi. Tidak ada endpoint penghapusan fisik baris bisnis.
Report/ekspor/daily summary selalu memakai transaksi aktif; isDelete pada route tersebut ditolak dengan 400, bukan memasukkan Trash ke total.
Label kategori terhapus tetap muncul pada transaksi historis; record baru/edit/restore memerlukan kategori aktif.

## Contoh admin membuat transaksi owner

POST /admin/users/6b3ab04b-22cb-4777-b3ec-15f3247b123d/transactions
```json
{
  "transaction_date": "2026-09-08",
  "type": "expense",
  "category_id": "bda088ec-2694-473c-997d-cb93161456f1",
  "amount": "43500.00",
  "title": "Belanja",
  "client_request_id": "22f60b83-db97-48f7-b591-b403b93c4c12"
}
```

Respons 201 ilustratif:
```json
{
  "data": {
    "id": "a1e29497-1328-496c-9c29-542815d5c2fd",
    "user_id": "6b3ab04b-22cb-4777-b3ec-15f3247b123d",
    "transaction_date": "2026-09-08",
    "type": "expense",
    "category_id": "bda088ec-2694-473c-997d-cb93161456f1",
    "amount": "43500.00",
    "title": "Belanja",
    "isDelete": false,
    "version": 1
  },
  "scope": {
    "mode": "admin",
    "owner_user_id": "6b3ab04b-22cb-4777-b3ec-15f3247b123d"
  }
}
```

DTO transaksi nyata juga memuat client_request_id, created_by, updated_by, created_at, dan updated_at. Field actor berasal dari autentikasi.
Payload resource sukses memakai data; payload list menambahkan page.next_cursor (null di akhir).
Respons admin target juga memuat scope.mode="admin" dan scope.owner_user_id; respons personal mempertahankan envelope. 204 tidak memiliki body.
request_hash, password_hash, session hash, dan secret provider tidak pernah disertakan.

## List dan laporan

Limit halaman default 30, maksimum 100. Urutan transaksi aktif transaction_date DESC,created_at DESC,id DESC; Trash updated_at DESC,id DESC.
Kategori diurutkan name,id; daily summary date DESC. Cursor mengikat actor/mode/owner/endpoint/filter/isDelete; penggunaan lintas scope mengembalikan 400.
Batas tanggal wajib untuk list transaksi/report/ekspor: start<=end, maksimum 366 hari inklusif (usulan). Filter kategori dari owner lain mengembalikan 404.
Preset memakai zona waktu owner, termasuk target admin. List kosong [], jumlah kosong "0.00".
Data total berisi start_date,end_date,income,expense,difference,top_income_categories,top_expense_categories.
Kategori teratas maksimal lima per tipe, urut amount DESC lalu category_id, tiap item {category_id,name,amount}.
Breakdown mengembalikan period_start untuk minggu/bulan atau category_id/name/type untuk kategori beserta total. Periode batas parsial setelah filter.
Semua kalkulasi hanya memasukkan transaksi is_delete=false. Delete/restore memengaruhi report dan ekspor owner secara konsisten.

## Ekspor (Issue 009)

Path personal /exports dan path sepadan /admin/users/{user_id}/exports:
| Method | Suffix | Perilaku |
| --- | --- | --- |
| POST | /exports | start_date,end_date,type?,category_id?,format:xlsx/pdf → job 202 |
| GET | /exports/{id} | status queued/running/succeeded/failed/canceled berscope |
| GET | /exports/{id}/download | otorisasi owner/admin saat ini dan stream artifact; 409 belum siap, 410 kedaluwarsa |

Job menyimpan owner_user_id, requested_by, dan request_mode. Pengguna biasa hanya mengakses job miliknya; admin dapat mengelola job owner terpilih.
Worker memeriksa ulang state requester/owner dan otoritas admin bila perlu; filter dan owner immutable.
Ekspor berisi semua transaksi aktif yang cocok, bukan hanya satu halaman. Kedaluwarsa artifact yang diusulkan 24 jam; unduhan selalu melakukan otorisasi ulang.

## Templat dan Drive (kontrak tahap berikutnya)

Issue 011 mendefinisikan /templates personal dan route /admin/users/{user_id}/templates untuk list/get/create/update/delete/restore/apply-to-form; gunakan isDelete/version dan kategori yang cocok dengan target.
Issue 012 mendefinisikan /drive/connection, /backup-schedules, /backups, /restores personal serta route admin target yang cocok.
Admin dapat memulai/memutus koneksi Drive target, mengubah jadwal, membuat/membatalkan/mengulang job backup, mengakses backup terotorisasi, dan menjalankan restore tervalidasi.
Callback OAuth mengikat actor, target, dan mode dalam state terlindungi lalu memeriksa ulang otorisasi. Koneksi target yang diotorisasi provider wajib ada.
Simpan owner/requester/mode untuk job. Jadwal berulang menyimpan actor/mode pemberi otorisasi; otorisasi dicabut atau akun terhapus menjeda eksekusi.
Data backup/restore mempertahankan isDelete tetapi mengecualikan role/password/session/secret provider/catatan audit; import ketat tidak boleh menyelundupkan perubahan role.
Snapshot restore mempertahankan hubungan kategori historis yang valid, termasuk kategori owner sama yang terhapus, tanpa mengubah flag; restore transaksi biasa tetap memerlukan kategori aktif.
Skema lengkap, konkurensi, dan state provider untuk route lanjutan ditentukan sebelum implementasi, bukan dianggap sudah tersedia.

## Kontrak error

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Periksa field yang ditandai.",
    "fields": {"amount": "Harus lebih besar dari nol"},
    "request_id": "opaque-id"
  }
}
```

400 untuk ID/cursor/boolean rusak, version hilang, atau field owner/role/isDelete yang dilarang; 401 token tidak ada/tidak valid atau actor terhapus; 403 role admin/CSRF ditolak; 404 baris hilang/di luar scope; 409 konflik version/lifecycle/idempotensi atau target/kategori tidak aktif; 422 validasi field; 429 throttle; 500 kegagalan internal yang disanitasi; 503 storage otorisasi/audit wajib tidak tersedia.
405 hanya untuk method yang tidak didukung, bukan CRUD admin yang didukung.
Health /health/live dan /health/ready berada di luar /api/v1 dan tidak menampilkan data/konfigurasi.
OpenAPI harus mendefinisikan setiap operasi konkret, envelope, nullability, header version, boolean lifecycle, kebutuhan auth, dan error; validasi contoh serta interface hasil generate.
