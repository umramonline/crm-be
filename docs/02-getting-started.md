# Başlangıç

## Gereksinimler

- Go 1.25+
- MySQL (CRM veritabanı için `DATABASE_DSN`)
- Erişilebilir Umramonline API (auth, roller, referans verisi)
- Frontend origin’in CORS listesinde olması (varsayılan: `http://localhost:5173`)

## Kurulum

```bash
git clone <repo-url>
cd crm-be
cp .env.example .env
# .env dosyasını doldurun — özellikle DATABASE_DSN, UMRAMONLINE_*, SESSION_TOKEN_SECRET, CONSUME_API_KEY
```

## API’yi çalıştırma

```bash
go run ./cmd/api
```

veya

```bash
go build -o app ./cmd/api
./app
```

Varsayılan adres: `http://localhost:8080`

Sağlık kontrolü:

```bash
curl http://localhost:8080/
```

## Startup sırası

`cmd/api` açılışında sırayla:

1. Config yüklenir (`.env` + env parse)
2. Umramonline HTTP client oluşturulur
3. MySQL bağlantısı açılır
4. AutoMigrate çalışır:
   - authorization → customer → task → followup → ietts → consume
5. Permission seed’leri çalışır (admin `role_id = 30`)
6. Servisler / handler’lar bağlanır
7. Fiber listen başlar
8. SIGINT/SIGTERM ile graceful shutdown

## Toplu müşteri sync (CLI)

İlk kurulum veya backfill için:

```bash
go run ./cmd/sync-umramonline-customers
```

Davranış:

- Kaynak: lokal Umramonline MySQL (`127.0.0.1:33007/umramdb` — kodda hardcoded)
- Hedef: CRM `DATABASE_DSN`
- `status = 1` müşterileri `uo_id` ile upsert eder
- Batch boyutu: `SYNC_BATCH_SIZE` (default `500`)
- Log: `scanned` / `inserted` / `updated`

> **Not:** Sürekli senkronizasyon için `POST /api/v1/consume` webhook’u kullanılır. Sync CLI yalnızca toplu taşıma içindir. Hardcoded kaynak DSN production için uygun değildir.

## Test

```bash
go test ./...
```

Modül bazlı örnek:

```bash
go test ./internal/customer/application/...
go test ./internal/consume/application/...
```

## Tipik geliştirme akışı

1. `.env` ile local MySQL + Umramonline staging bağla
2. `go run ./cmd/api`
3. Frontend’i `CORS_ALLOWED_ORIGINS` ile eşleştir
4. Password login ile cookie al
5. Protected endpoint’leri cookie ile çağır
