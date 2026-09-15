# Kontrol akses dan penghapusan lunak

Versi 0.3 • 14 September 2026
Dikonfirmasi: admin dapat menjalankan semua operasi aplikasi pada data pengguna mana pun. Pengguna biasa hanya dapat melakukan CRUD atas datanya sendiri.
Dokumen ini menggantikan pembatasan lama terhadap perubahan oleh admin.

## Matriks izin

| Operasi | Pengguna biasa | Admin |
| --- | --- | --- |
| Lihat/buat/edit pemasukan dan pengeluaran | Hanya milik sendiri | Pemilik mana pun yang dipilih |
| Hapus/pulihkan pemasukan dan pengeluaran | Hanya milik sendiri, hapus lunak | Pemilik mana pun yang dipilih, hapus lunak |
| CRUD/pulihkan kategori | Hanya milik sendiri | Pemilik mana pun yang dipilih |
| Laporan dan ekspor Excel/PDF, termasuk status/unduhan job | Hanya milik sendiri | Pemilik mana pun yang dipilih |
| CRUD/pulihkan/terapkan templat | Hanya milik sendiri | Pemilik mana pun yang dipilih |
| Pengaturan Drive, job backup, dan restore | Hanya milik sendiri | Pemilik mana pun yang dipilih, dengan otorisasi provider yang valid |
| Perbarui profil/hapus akun sendiri | Hanya milik sendiri | Pengguna mana pun |
| Daftar/buat/kelola akun dan tetapkan peran user/admin | Registrasi publik tidak dapat menetapkan admin | Diizinkan melalui operasi admin terlindungi |
| Lihat catatan bisnis di Trash | Hanya milik sendiri | Pemilik mana pun yang dipilih |
| Lihat catatan audit admin | Tidak dibuka | Diizinkan melalui endpoint audit admin |

Semua operasi bisnis memakai validasi, konkurensi optimistis, dan penghapusan yang mempertahankan baris. Password/token/secret provider tidak pernah dikembalikan sebagai data aplikasi biasa; alur reset/disconnect bekerja atas kredensial tanpa membukanya.
Admin dapat memilih akun mana pun, termasuk admin lain atau dirinya sendiri. My Data tetap menjadi scope diri yang praktis. Pengelolaan lintas pengguna menggunakan route admin eksplisit.

## Aktor dan pemilik

actor_user_id adalah akun terautentikasi dan tidak berubah saat memilih pengguna lain.
owner_user_id adalah pemilik data keuangan: aktor pada route personal; target eksplisit pada route admin yang diotorisasi.
Catatan baru yang dibuat admin tetap dimiliki pengguna terpilih. created_by/updated_by dan event audit admin mengidentifikasi admin yang bertindak.
Semua query/write keuangan mempertahankan predicate pemilik; jangan membiarkan is_admin menghapus pembatasan scope. ID catatan, kategori, konteks form, cursor, dan job harus sesuai dengan pemilik terpilih.

## Otorisasi

- Autentikasi dan muat state akun aktor saat ini pada setiap request terlindungi. Aktor terhapus mendapat 401 meskipun JWT belum kedaluwarsa.
- Setiap baca/tulis admin memeriksa peran di database saat ini. Pengguna biasa mendapat 403 sebelum target dicari.
- Registrasi publik menetapkan role=user. Parameter profil/body personal tidak boleh memasok override role atau owner.
- Pembuatan akun dan pembaruan peran admin diizinkan pada operasi terlindungi /admin/users. Bootstrap admin awal adalah tindakan operator eksplisit, bukan promosi pengguna pertama otomatis.
- Perubahan peran atau penghapusan lunak akun mencabut sesi refresh. Request admin berikutnya setelah demotion tersimpan ditolak. Otorisasi yang telah diberikan pada request yang sedang berjalan boleh selesai; mutasi menyerialkan pemeriksaan akun/peran sesuai database.md.
- Ketidakcocokan target atau resource terkait mengembalikan 404. Scope yang hilang/rusak mengembalikan 400, bukan operasi keuangan global.
- Route biasa selalu memakai scope aktor sendiri, termasuk untuk admin. Untuk mengelola B, admin C memanggil route admin B.
- Pengguna biasa tidak dapat mengakses Trash, ekspor, templat, atau backup pengguna lain. Admin dapat menjalankan operasi itu untuk target yang dipilih.
- Write normal memerlukan akun target aktif. Admin dapat memeriksa data target terhapus dan memulihkan akun sebelum perubahan keuangan berikutnya.

## Kontrak isDelete

| Lapisan | Nama | Nilai |
| --- | --- | --- |
| Field respons JSON | isDelete | boolean false / true |
| Field Go | IsDelete | bool |
| Kolom PostgreSQL | is_delete | BOOLEAN NOT NULL DEFAULT FALSE |

Berlaku untuk users, categories, transactions, dan templates. Sesi/koneksi tetap memakai siklus pencabutan token, sedangkan status/kedaluwarsa job terpisah.
Create → false. Delete → true. Restore → false. Endpoint aplikasi ini tidak menghapus baris secara fisik.
List/detail normal memerlukan false. List/detail Trash secara eksplisit memilih isDelete=true. Filter boolean tidak pernah mengubah otorisasi. Report dan ekspor report selalu memakai transaksi aktif.
Delete/restore memakai versi baris wajib, predicate pemilik, dan state siklus hidup yang diharapkan. Versi lama atau operasi ganda mengembalikan 409; catatan hilang atau milik pihak lain mengembalikan 404.
Kategori terhapus tetap tersedia sebagai label historis, tetapi transaksi dan templat baru/dipulihkan/diedit memerlukan kategori aktif yang cocok.
Pengguna dapat memulihkan data keuangannya saat akunnya aktif. Pemulihan akun terhapus memerlukan admin karena akun itu tidak dapat login.

## UI Admin Management

Pencarian dan pemilihan pengguna dimulai kosong. Tampilkan identitas pemilik dengan jelas dan aktifkan aksi pengelolaan sesuai tahap pengiriman fitur.
Ikat pemilik form/dialog saat dibuat. Sebelum berpindah, simpan atau buang input yang belum disimpan; jangan mengarahkan payload form terbuka ke akun baru.
Hapus data dan cursor lama saat berpindah, batalkan pembacaan, dan abaikan generasi respons lama. Write/job yang sedang berjalan mempertahankan pemilik awal; penyelesaiannya tidak boleh mengubah layar baru.
Sediakan Trash/Restore per pemilik terpilih serta kontrol akun/peran admin. UI pemilik menyediakan CRUD keuangan yang sama untuk diri sendiri.
Gunakan key state terpisah untuk actor/mode/owner/filter/version; buang state saat logout, penolakan peran, atau penghapusan akun. Semua respons terautentikasi menggunakan Cache-Control: no-store.

## Ekspor, templat, dan backup

Operasi personal menetapkan owner=actor. Operasi admin menetapkan owner=target yang telah diverifikasi; simpan requester dan owner serta request_mode.
Pengguna biasa hanya dapat mengakses job/data dengan ID owner miliknya. Admin aktif dapat mengelola job target melalui route target tersebut, tanpa bergantung pada requester awal.
Worker memvalidasi ulang aktivitas requester/owner dan, untuk request admin, peran admin saat eksekusi. Batalkan job antre yang tidak berwenang; retry tidak pernah mengubah owner.
Jadwal berulang menyimpan owner, aktor pemberi otorisasi, dan mode. Aktor/peran tidak valid atau owner terhapus menjedakan jadwal; melanjutkan kembali memerlukan otorisasi baru.
Operasi Google juga memerlukan koneksi OAuth valid untuk akun Drive target. Status admin aplikasi tidak menyediakan kredensial atau consent provider.
Format backup mempertahankan flag isDelete, tetapi mengecualikan role, password/session, secret provider, dan catatan audit. Restore memvalidasi/memetakan semua baris ke target terotorisasi dan tidak dapat memberikan role. Perubahan role memakai endpoint admin khusus.

## Audit

Catat aktor, target, tindakan, UUID resource, hasil, request ID, waktu UTC, serta nama field/version yang aman.
Write admin yang berhasil melakukan commit mutasi data dan event audit dalam satu transaksi database. Kegagalan audit membatalkan write; pembacaan yang berhasil menyimpan event sebelum respons.
Event audit hanya dapat ditambahkan; admin boleh memeriksanya, bukan menulis ulang sejarah melalui CRUD bisnis. Catat nilai role lama/baru tanpa payload rahasia.
Job eksternal menyimpan otorisasi, job, dan audit secara atomik sebelum memanggil provider. Retry dan kegagalan eksternal memakai state job eksplisit, bukan klaim rollback database atas pekerjaan remote.
Jangan mencatat judul/jumlah keuangan, password, token, atau body respons lengkap.

## Matriks penerimaan wajib

| Pengujian | Hasil yang diharapkan |
| --- | --- |
| A melakukan CRUD/restore atas datanya | Berhasil dengan total/siklus hidup yang benar |
| A memakai ID resource B, filter Trash, route ekspor/templat/backup | 404 pada resource personal; 403 pada route admin |
| C membuat/mengedit/menghapus/memulihkan catatan B melalui route admin | Berhasil; owner B dipertahankan, actor C diaudit |
| C memakai ID B di dalam route admin A | 404; tidak ada owner yang berubah |
| C mengirim versi lama | 409, tidak ada lost update |
| Delete diikuti list/report/export normal | Baris dipertahankan dengan true dan dikeluarkan dari hasil aktif/total |
| Trash/restore | Baris true hanya terlihat di Trash terotorisasi; restore false dan total memasukkannya sekali |
| Race delete/edit/restore | Paling banyak satu operasi versi yang diharapkan menang; request lain konflik |
| Payload create/edit memasok isDelete atau owner sembarang | 400; siklus hidup/kepemilikan tidak dapat dilewati |
| C mengubah peran A melalui operasi terlindungi | Berhasil dan diaudit; mutasi role publik/profil ditolak |
| C didemote atau akun aktor dihapus | Request terlindungi/admin berikutnya ditolak sesuai kondisi |
| Target berubah saat form/write terbuka | Tidak ada write ke owner yang salah |
| Admin mengekspor/menyimpan backup/mengelola templat B | Berhasil dengan scope target; A tidak dapat menggunakannya |
| Job admin antre setelah requester didemote | Dibatalkan/dijeda, tanpa panggilan provider |
| Insert audit gagal saat write admin | Mutasi database dibatalkan |
| Kategori terhapus dengan transaksi aktif | Jumlah/label historis dipertahankan; kategori tidak tersedia untuk dipilih |
| Soft delete akun | Sesi dicabut, job dijeda, baris anak dipertahankan |
| Field role/owner berbahaya dalam backup | Ditolak; perlindungan owner dan role tetap berlaku |

Pengiriman: 002–004 skema/auth/kontrak; 005–007 CRUD inti dan laporan; 013 backend admin penuh; 008 UI; 009 ekspor; 010 mobile; 011 templat; 012 backup.
