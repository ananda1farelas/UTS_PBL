# SIAKAD Mini — RESTful API (UTS Praktikum Back End)

RESTful API back end untuk **SIAKAD Mini**, layanan akademik sederhana yang mengelola data
mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS), sesuai studi kasus UTS.

- **Bahasa**: Go 1.24
- **Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: PostgreSQL (driver `lib/pq`, raw SQL, tanpa ORM)
- **Autentikasi**: JWT (`golang-jwt/jwt/v5`), password di-hash dengan bcrypt

## 1. Struktur Project

```
UTS/
├── app/
│   ├── handler/      # HTTP handler (terima request, panggil service, balas JSON)
│   ├── model/         # Struct representasi tabel database
│   ├── repository/    # Query SQL mentah ke PostgreSQL
│   └── service/       # Business logic & rules
├── config/             # Loader .env, struct config, logger
├── database/           # Koneksi DB, migration runner, seeder
├── helper/             # JWT, hashing, response envelope, pagination, dsb
├── logs/                # Log file aplikasi (app.log)
├── middleware/          # Auth JWT, role guard, rate limiter login
├── migration/            # File SQL migration bernomor urut
├── route/                 # Wiring semua endpoint
├── main.go
├── go.mod / go.sum
└── .env.example
```

Alur request: `route` → `middleware` (auth/role) → `handler` (validasi input, format response)
→ `service` (business rules) → `repository` (query SQL) → PostgreSQL.

## 2. Cara Menjalankan

### Prasyarat
- Go 1.24+
- PostgreSQL 14+ yang sudah running

### Langkah

```bash
cd UTS
cp .env.example .env
# sesuaikan DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, JWT_SECRET di .env

go mod tidy

# jalankan aplikasi — otomatis connect DB + jalankan semua migration
go run main.go
```

Saat pertama kali start, aplikasi otomatis:
1. Connect ke PostgreSQL sesuai `.env`
2. Menjalankan semua file di `migration/*.sql` yang belum pernah dijalankan
   (tracked lewat tabel `schema_migrations`, jadi aman dijalankan berulang)

### Seeding data awal

```bash
go run main.go seed
```

Mengisi database dengan:
- 1 akun **admin** — `admin@siakad.test` / `admin12345`
- 20 akun **mahasiswa** — email `mahasiswaNN@siakad.test`, password awal = **NIM** masing-masing
  (NIM mengikuti pola `202200000001` s/d `202200000020`)
- 10 **mata kuliah** (kode `IF101`–`IF502`)

### Menjalankan unit test

```bash
go test ./...
```

## 3. Model Data

| Tabel | Kolom utama | Relasi |
|---|---|---|
| `users` | id, email, password (hash), role (admin/mahasiswa) | 1–1 ke `students` |
| `students` | id, user_id, nim (unik), nama, prodi, angkatan, ipk_terakhir, deleted_at | 1–N ke `enrollments` |
| `courses` | id, kode_mk (unik), nama_mk, sks, semester, kuota | 1–N ke `enrollments` |
| `enrollments` | id, student_id, course_id, tahun_akademik, created_at | unique(student_id, course_id, tahun_akademik) |

Soft delete diterapkan pada `students` lewat kolom `deleted_at` — data tidak pernah benar-benar
dihapus, hanya disembunyikan dari query aktif, supaya riwayat KRS tetap konsisten.

## 4. Business Rules

1. **Batas SKS per semester** ditentukan dari `ipk_terakhir` mahasiswa:
   - IPK ≥ 3.00 → maksimal 24 SKS
   - IPK 2.50–2.99 → maksimal 21 SKS
   - IPK < 2.50 → maksimal 18 SKS
   - Dihitung ulang setiap kali mahasiswa mengambil mata kuliah baru, dibatasi per `tahun_akademik`.
2. **Tidak boleh mengambil mata kuliah yang sama dua kali** pada `tahun_akademik` yang sama
   (dijaga lewat unique constraint di database + pengecekan eksplisit sebelum insert → `409`).
3. **Kuota mata kuliah** dicek dengan row locking (`SELECT ... FOR UPDATE`) di dalam transaction,
   supaya dua request bersamaan tidak bisa sama-sama lolos saat kuota tersisa 1.
4. **Mahasiswa hanya bisa mengakses/mengubah KRS miliknya sendiri** — dicek di service layer
   dengan membandingkan `student_id` pemilik enrollment dengan mahasiswa yang sedang login.

Semua pengecekan di atas untuk `POST /enrollments` dan `DELETE /enrollments/{id}` dibungkus dalam
satu database transaction agar atomic.

## 5. Autentikasi & Otorisasi

- Login (`POST /auth/login`) mengembalikan JWT yang menyimpan `user_id`, `email`, dan `role`.
- Semua endpoint (kecuali login) mewajibkan header `Authorization: Bearer <token>`.
- Middleware `Auth` memvalidasi token lalu menyimpan klaim ke request context.
- Middleware `RequireRole` membatasi endpoint tertentu hanya untuk `admin` atau `mahasiswa`.
- Rate limiting login: maksimal 5 percobaan **gagal** per menit per IP, lewat batas → `429`.

## 6. Daftar Endpoint

| No | Method | Endpoint | Akses | Status Sukses |
|---|---|---|---|---|
| 1 | POST | `/api/v1/auth/login` | Publik | 200 |
| 2 | GET | `/api/v1/auth/me` | Semua role | 200 |
| 3 | GET | `/api/v1/students` | Admin | 200 |
| 4 | POST | `/api/v1/students` | Admin | 201 |
| 5 | GET | `/api/v1/students/{id}` | Admin, mahasiswa (data sendiri) | 200 |
| 6 | PUT | `/api/v1/students/{id}` | Admin | 200 |
| 7 | DELETE | `/api/v1/students/{id}` | Admin | 204 |
| 8 | GET | `/api/v1/courses` | Semua role | 200 |
| 9 | POST | `/api/v1/enrollments` | Mahasiswa | 201 |
| 10 | DELETE | `/api/v1/enrollments/{id}` | Mahasiswa (milik sendiri) | 204 |

## 7. Format Response

Semua response JSON mengikuti struktur seragam:

```json
{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [ ... ],
  "meta": { "current_page": 1, "per_page": 10, "total": 20, "last_page": 2 }
}
```

Error validasi (422):

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": { "nim": ["NIM sudah terdaftar"] }
}
```

Error lain (401/403/404/409/429/500) memakai bentuk `{ "success": false, "message": "..." }`.
Error 500 tidak pernah membocorkan stack trace — ditangani oleh `helper.GlobalErrorHandler`
yang didaftarkan sebagai `ErrorHandler` Fiber.

## 8. Testing Manual (sudah dilakukan selama development)

Semua 10 endpoint sudah diuji end-to-end terhadap PostgreSQL asli, termasuk:
- Login sukses & gagal, rate limiting (6x gagal → 429 di percobaan ke-6)
- CRUD mahasiswa lengkap + validasi + duplikasi NIM/email
- Akses silang data mahasiswa lain → 403
- Soft delete → data hilang dari listing & detail (404)
- Pengambilan KRS sampai menyentuh batas SKS persis (18/21/24) → ditolak di SKS ke-N+1 dengan
  pesan sisa SKS
- Duplikasi pengambilan mata kuliah yang sama → 409
- Kuota penuh → 422
- Pembatalan KRS milik sendiri vs milik orang lain → 204 vs 403

Unit test otomatis tersedia untuk helper password (`helper/password_validator_test.go`),
dijalankan dengan `go test ./...`.
