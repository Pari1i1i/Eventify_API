# Eventify API

Backend REST API untuk platform **Eventify** — ticketing & manajemen acara. Melayani Web Admin, Portal Panitia/Organizer, Portal Customer, dan aplikasi Flutter Mobile (QR Check-in).

Dibangun dengan **Go + Gin + GORM + MySQL**, lengkap dengan autentikasi JWT, Swagger documentation, cron worker, dan integrasi webhook pembayaran.

## Tech Stack

- **Go** 1.x — bahasa utama
- **Gin** — HTTP web framework
- **GORM** — ORM untuk MySQL
- **MySQL** — database
- **JWT (golang-jwt/v5)** — autentikasi berbasis token
- **Swagger (swaggo/gin-swagger)** — dokumentasi API otomatis
- **godotenv** — konfigurasi environment
- **Midtrans / Xendit** — webhook notifikasi pembayaran

## Fitur

- Autentikasi: register, login, forgot/reset password, profil, ganti password
- Manajemen event: CRUD, ticket tiers, upload banner, approval/status event
- Order & pembayaran: buat order, webhook payment, expire order otomatis (cron tiap 1 menit)
- Tiket & check-in: tiket customer, entry scanner / QR check-in
- Dashboard admin: statistik total revenue, jumlah tiket, kehadiran
- Role-based access control: `admin`, `panitia` (organizer), `customer`
- Middleware CORS untuk semua origin

## Struktur Proyek

```
EventifyApi/
├── config/         # Load config & koneksi database
├── cron/           # Background worker (expire unpaid orders)
├── dto/            # Data transfer object & validasi request
├── docs/           # Hasil generate Swagger
├── handlers/       # HTTP handler / controller
├── middlewares/    # Auth JWT, role guard, CORS
├── models/         # Model GORM (User, Event, Order, dll)
├── repositories/   # Akses database
├── services/       # Business logic
├── uploads/        # Upload banner event
├── utils/          # Helper response JSON, dsb.
└── main.go         # Entry point & pendaftaran route
```

## Menjalankan Secara Lokal

1. **Prasyarat**: Go terpasang dan MySQL berjalan (buat database, contoh: `eventify_db`).
2. **Siapkan environment**:

   ```bash
   cp .env.example .env
   ```

   Sesuaikan nilai di `.env`:

   ```env
   APP_NAME=EventifyApi
   APP_ENV=development
   APP_PORT=8080
   APP_URL=http://localhost:8080

   DB_HOST=127.0.0.1
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=
   DB_NAME=eventify_db

   JWT_SECRET=eventify_super_secret_jwt_key_2026_change_in_production
   JWT_EXPIRATION_HOURS=72

   UPLOAD_DIR=./uploads

   MIDTRANS_SERVER_KEY=your-midtrans-server-key-here
   MIDTRANS_CLIENT_KEY=your-midtrans-client-key-here
   ```

3. **Jalankan server**:

   ```bash
   go mod download
   go run main.go
   ```

4. **Akses**:
   - API: `http://localhost:8080/api/v1`
   - Health check: `http://localhost:8080/health`
   - Dokumentasi Swagger: `http://localhost:8080/swagger/index.html`

5. **Build biner (Windows)**:

   ```bash
   go build -o eventifyApi.exe main.go
   ```

## Environment Variables

| Variabel                   | Deskripsi                                              |
| -------------------------- | ------------------------------------------------------ |
| `APP_NAME`                 | Nama service                                           |
| `APP_ENV`                  | `development` / `production`                           |
| `APP_PORT`                 | Port HTTP server                                       |
| `APP_URL`                  | URL public service                                     |
| `DB_HOST` / `DB_PORT`      | Host & port MySQL                                      |
| `DB_USER` / `DB_PASSWORD`  | Kredensial MySQL                                       |
| `DB_NAME`                  | Nama database                                          |
| `JWT_SECRET`               | Secret key token JWT                                   |
| `JWT_EXPIRATION_HOURS`     | Umur token JWT dalam jam                               |
| `UPLOAD_DIR`               | Folder penyimpanan upload banner                       |
| `MIDTRANS_SERVER_KEY`      | Server key Midtrans                                    |
| `MIDTRANS_CLIENT_KEY`      | Client key Midtrans                                    |

## Ringkasan Endpoint

### Public
| Method | Endpoint                 | Deskripsi                         |
| ------ | ------------------------ | --------------------------------- |
| POST   | `/api/v1/auth/register`  | Registrasi customer               |
| POST   | `/api/v1/auth/login`     | Login, return JWT token           |
| GET    | `/api/v1/events`         | Daftar event published            |
| GET    | `/api/v1/events/:slug`   | Detail event by slug              |
| POST   | `/api/v1/payments/webhook` | Webhook pembayaran pessimist     |

### Authenticated (Customer)
| Method | Endpoint                          | Deskripsi                      |
| ------ | --------------------------------- | ------------------------------ |
| GET    | `/api/v1/auth/me`                 | Profil sendiri                 |
| POST   | `/api/v1/orders`                  | Buat order / checkout          |
| GET    | `/api/v1/orders/my-orders`        | Order milik sendiri            |
| GET    | `/api/v1/tickets/my-tickets`      | Tiket milik sendiri            |

### Organizer (Panitia)
| Method | Endpoint                             | Deskripsi                    |
| ------ | ------------------------------------ | ---------------------------- |
| POST   | `/api/v1/organizer/events`           | Buat event                   |
| PUT    | `/api/v1/organizer/events/:id`       | Update event                 |
| DELETE | `/api/v1/organizer/events/:id`       | Hapus event                  |
| POST   | `/api/v1/scanner/check-in`           | Check-in tiket via QR        |

### Admin
| Method | Endpoint                     | Deskripsi                          |
| ------ | ---------------------------- | ---------------------------------- |
| GET    | `/api/v1/admin/dashboard`    | Statistik dashboard                |
| GET    | `/api/v1/admin/users`        | Daftar semua user                  |
| GET    | `/api/v1/admin/orders`       | Daftar semua order                 |
| GET    | `/api/v1/admin/events`       | Daftar semua event                 |
| PUT    | `/api/v1/admin/events/:id/status` | Set status event (publish/draft) |

> Semua route yang butuh otorisasi wajib membawa header `Authorization: Bearer <token>`. Peran dibatasi: admin, panitia, customer.

## Cara Deploy (VPS + Docker)

1. Dockerfile di `EventifyApi`; build image dan jalankan pada port (contoh `8093`).
2. Backup tunnel `cloudflared` menghadapkan port ke domain HTTPS.
3. Pastikan `APP_ENV=production` dan `JWT_SECRET` diganti dengan nilai rahasia kuat.
4. Jadikan MySQL sebagai volume/service terpisah agar data tidak hilang saat redeploy.

## Lisensi

MIT