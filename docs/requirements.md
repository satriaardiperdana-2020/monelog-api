# Persyaratan

Draf v0.3 • 14 September 2026
Produk: Monelog. Dokumentasi: Bahasa Indonesia. UI yang diusulkan: Bahasa Indonesia.
Repositori: monelog-api (backend Go), monelog-app (Vue 3 + JavaScript; browser dan Capacitor Android/iOS).
Koreksi yang dikonfirmasi: admin memiliki izin aplikasi penuh atas data pengguna mana pun; pengguna biasa hanya dapat melakukan CRUD atas datanya sendiri. Penghapusan menggunakan boolean isDelete.

## Bukti dan ruang lingkup

Tangkapan layar menampilkan Home, detail hari, Add Transaction, dan Reports. Tangkapan layar launcher tidak menambah persyaratan fungsional.
Home menampilkan total hari ini, Add, ringkasan hari sebelumnya, serta tab Home/Reports/Settings.
Detail hari menampilkan transaksi bertanggal, total, serta kontrol yang tampak seperti templat/berbagi; perilaku kontrol sekunder belum dikonfirmasi.
Entry menampilkan tanggal, tipe pemasukan/pengeluaran, kategori, jumlah positif, judul, Cancel/Save, serta pintasan templat/kalkulator.
Reports menampilkan 7/30 hari terakhir atau rentang kustom, total/selisih, peringkat kategori, dan rincian mingguan/bulanan/kategori/transaksi.
Layar Settings, autentikasi, dan admin tidak ditampilkan. Alurnya di bawah adalah keputusan desain untuk mendukung persyaratan pengguna.

## Persyaratan fungsional

| ID | Persyaratan | Pengiriman / sumber |
| --- | --- | --- |
| FR-01 | Pengguna biasa terautentikasi dapat membuat, membaca, memperbarui, menghapus lunak, dan memulihkan hanya datanya sendiri | MVP, dikonfirmasi |
| FR-02 | Entri pemasukan/pengeluaran dengan tanggal, kategori, jumlah positif, dan judul | MVP, tangkapan layar |
| FR-03 | Edit/delete dengan konfirmasi dan restore melalui Trash | MVP; delete/restore memakai FR-17 |
| FR-04 | Total hari ini, riwayat harian berpaginasi, dan detail hari | MVP, tangkapan layar |
| FR-05 | Buat/baca/perbarui/hapus/pulihkan kategori per tipe transaksi | MVP; delete menggantikan usulan archive sebelumnya |
| FR-06 | Laporan 7 hari, 30 hari, dan rentang kustom inklusif | MVP, anotasi tangkapan layar |
| FR-07 | Laporan pemasukan, pengeluaran, selisih, mingguan/bulanan/kategori/transaksi | MVP, tangkapan layar |
| FR-08 | Ekspor catatan terfilter milik pemilik terpilih ke Excel/PDF | Rilis 1 |
| FR-09 | UI browser responsif dan paket Android/iOS | Dikonfirmasi |
| FR-10 | Templat transaksi yang dapat digunakan ulang, dengan form dapat diedit dan belum tersimpan saat diterapkan | Tahap berikutnya; detail masih diusulkan |
| FR-11 | Zona waktu profil, tampilan IDR, dan pengaturan kategori | Usulan MVP |
| FR-12 | Backup Google Drive terjadwal dan restore tervalidasi untuk pemilik terotorisasi yang dipilih | Tahap berikutnya |
| FR-13 | Entri dan sinkronisasi offline | Ditunda |
| FR-14 | Grafik, kalkulator, dan berbagi native | Ditunda; perilaku kontrol belum dikonfirmasi |
| FR-15 | Admin dapat mengelola akun, peran, kategori, pemasukan/pengeluaran, laporan, ekspor, templat, dan backup pengguna mana pun | Dikonfirmasi; setiap kemampuan dikirim bersama milestone fiturnya |
| FR-16 | Otorisasi backend, pemeriksaan akun/peran aktif, dan audit tindakan admin | Kontrol teknis MVP |
| FR-17 | Penghapusan lunak memakai isDelete=false untuk aktif dan true untuk terhapus; baris yang dipertahankan dapat dipulihkan | Dikonfirmasi |

## Izin

Lihat access-control.md. My Data dibatasi pada akun terautentikasi. Admin Management memilih pemilik eksplisit dan mendukung pembacaan maupun penulisan untuk pemilik tersebut.
CRUD admin diizinkan. Tidak ada pembatasan berbasis peran yang mencegah admin mengedit/menghapus catatan pengguna lain, membuat/mengunduh ekspor mereka, atau mengelola templat/backup mereka.
Identitas pemilik diturunkan dari otorisasi, bukan dipercaya dari body yang dikirim. Pengguna biasa tidak dapat memilih pemilik lain atau mengubah peran.
Admin dapat membuat pengguna, mengubah profil/peran, menghapus lunak, dan memulihkan akun melalui operasi admin yang dilindungi. Registrasi publik selalu membuat pengguna biasa.
Izin tidak melewati field wajib, kepemilikan/tipe kategori, pemeriksaan versi, atau otorisasi OAuth provider.

## Aturan bisnis

- MVP hanya memakai IDR. Simpan tepat dua angka desimal; jangan gunakan floating point biner. Uang API adalah string desimal.
- Jumlah pemasukan/pengeluaran positif; selisih = pemasukan dikurangi pengeluaran, bukan saldo akun.
- Rentang jumlah 0.01–999999999999.99. Judul di-trim, wajib, panjang 1–200 karakter.
- Kategori milik pemilik transaksi dan cocok dengan tipenya. Catatan yang dibuat admin menjadi milik pengguna terpilih, sedangkan atribusi mencatat admin.
- transaction_date adalah tanggal kalender; timestamp audit UTC. Zona waktu default Asia/Jakarta dan dapat dikonfigurasi.
- Preset tanggal memakai zona waktu pemilik keuangan, termasuk di Admin Management. Tujuh hari berarti hari ini dan enam hari sebelumnya; 30 hari berarti hari ini dan 29 hari sebelumnya.
- Minggu dimulai Senin. Endpoint kustom mencakup kedua tanggal. Filter dilakukan sebelum pengelompokan agar bucket minggu/bulan batas hanya memuat tanggal terpilih.
- Tanggal transaksi masa depan ditolak pada MVP (usulan). Rentang laporan kosong/masa depan dapat mengembalikan nol.
- Pengguna, kategori, transaksi, dan templat baru dimulai dengan isDelete=false. Hanya operasi delete/restore yang mengubahnya; payload create/edit biasa tidak dapat melewati aturan siklus hidup.
- List normal, total harian, laporan, dan ekspor hanya memuat transaksi aktif. Trash secara eksplisit meminta isDelete=true; hal itu tidak mengubah total laporan.
- Penghapusan lunak kategori mengeluarkannya dari pilihan tetapi mempertahankan jumlah transaksi dan label kategori historis. Catatan lama dapat dibaca/dihapus; edit/restore memerlukan kategori aktif yang cocok.
- Nama kategori unik per pemilik/tipe di seluruh baris aktif/terhapus; pulihkan, jangan buat ulang nama yang telah dihapus.
- Penghapusan lunak pengguna menonaktifkan autentikasi/job dan mencabut sesi; baris terkait tetap dipertahankan. Restore tidak memulihkan sesi lama atau otomatis melanjutkan jadwal backup yang dijeda.
- Templat menggunakan aturan penghapusan boolean dan kepemilikan kategori yang sama.
- Tidak ada penghapusan fisik baris bisnis melalui operasi aplikasi biasa, termasuk Delete admin.
- MVP tidak memiliki dompet/transfer/saldo akun. Penarikan ATM bukan transfer otomatis; mencatat penarikan dan belanja berikutnya sebagai pengeluaran akan menghitung ganda.

## Alur layar

Pengguna biasa: login → My Data → CRUD personal, laporan, Trash/Restore, lalu ekspor/templat/backup.
Admin: login → My Data atau Admin Management → pemilih pengguna yang dapat dicari → CRUD, laporan, Trash/Restore, dan fitur lanjutan milik pengguna terpilih.
Email/identitas pengguna terpilih tetap terlihat di dekat aksi pengelolaan; form, dialog, request tertunda, dan job terikat pada pilihan tersebut.
Pergantian pengguna menghapus data/filter lama dan membuang respons terlambat. Jika form belum disimpan terbuka, minta save/discard sebelum berpindah. Request yang telah dikirim tetap memakai target awal; hasil terlambat tidak boleh memperbarui layar target baru.
Admin juga memiliki kontrol create/update/role/delete/restore pengguna. Identitas sesi login tetap menjadi aktor sepanjang waktu.

## Contoh penerimaan

- Pemasukan 1000000.00 dan pengeluaran 43500.00 menghasilkan selisih 956500.00.
- Pengeluaran 500000.00 + 43500.00 + 226000.00 menghasilkan pengeluaran harian 769500.00.
- Hari/laporan kosong mengembalikan total nol dan daftar kosong.
- Save yang gagal mempertahankan input; pengiriman ganda tidak membuat transaksi ganda.
- Pengguna biasa A tidak dapat mengakses/mengubah data B atau memanggil route admin, meskipun mengetahui UUID.
- Admin C dapat membuat/mengedit/menghapus/memulihkan catatan A dan B melalui route berscope masing-masing; laporan pengguna terpilih mengikuti setiap perubahan.
- Delete mengatur isDelete=true dan menaikkan versi tanpa menghapus baris. Detail aktif mengembalikan 404; Trash menampilkannya. Restore mengatur false dan memasukkan kembali jumlah tepat satu kali.
- Admin menargetkan A dengan catatan/kategori milik B mendapat 404, bukan pergantian pemilik diam-diam.
- Penghapusan akun memblokir token akses yang belum kedaluwarsa. Demotion memblokir operasi admin berikutnya setelah perubahan peran tersimpan.
- Operasi ekspor/templat/backup lintas pengguna berhasil untuk admin aktif dan gagal untuk pengguna biasa.

## Gerbang kualitas dan rilis

HTTPS, log yang aman dari rahasia, hashing password, sesi refresh berotasi, login yang dibatasi laju, dan pemeriksaan izin backend.
Form yang mudah diakses, layout mobile yang terbaca, label yang tidak hanya mengandalkan warna, serta state loading/kosong/error/Trash.
Target p95 list/summary yang diusulkan: di bawah 500 ms pada 20 pengguna bersamaan dan 100 ribu transaksi per pemilik di infrastruktur staging yang disepakati.
Uji matriks akses, constraint/migrasi PostgreSQL nyata, aturan uang/tanggal, delete/restore boolean, otorisasi job, alur browser, dan build perangkat.
MVP memerlukan internet; pertahankan input yang belum disimpan dan jelaskan kegagalan koneksi.

## Keputusan yang tersisa

Registrasi mandiri publik versus akun buatan admin; judul wajib/tanggal masa depan; satu templat versus bundle harian; detail settings, berbagi, grafik/kalkulator, online-first, dan usulan IDR-only.
Izin admin penuh dan aturan penghapusan lunak boolean di atas telah dikonfirmasi dan menggantikan interpretasi sebelumnya.
