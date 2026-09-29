# Operasyon ve Deploy

## Process modeli

- Tek process HTTP sunucusu (`cmd/api`)
- Message consumer / worker process yok
- Background job framework yok (dashboard metrikleri request içi parallel)

## Startup checklist

1. Env değişkenleri yüklü mü? (`DATABASE_DSN`, Umramonline secret’ları, `SESSION_TOKEN_SECRET`, `CONSUME_API_KEY`)
2. MySQL erişilebilir mi?
3. Umramonline API erişilebilir mi?
4. `storage/follow-ups` yazılabilir mi? (follow-up upload için)
5. CORS origin frontend ile uyumlu mu?

## Graceful shutdown

SIGINT / SIGTERM alındığında:

1. Yeni istek kabulü durur
2. `SHUTDOWN_TIMEOUT_SECONDS` (default 10s) içinde mevcut istekler bitmeye çalışır
3. Fiber `ShutdownWithContext` çağrılır

## Build

```bash
go build -o app ./cmd/api
```

Not: Deploy workflow örneği `./cmd/api/main.go` path’ini kullanır; paket olarak `./cmd/api` tercih edilebilir.

## GitHub Actions deploy

Dosya: `.github/workflows/deploy.yml`

Durum: **yorum satırında (disabled)**.

Planlanan akış:

1. Self-hosted runner
2. SSH agent + known_hosts
3. Sunucuda `git pull origin main`
4. `go mod download`
5. `go build -o app ./cmd/api/main.go`
6. Binary’yi çalıştır

Production’da process manager (systemd, supervisord, docker) kullanılması önerilir; workflow örneği process yönetimini detaylandırmaz.

## Sync job

Toplu müşteri aktarımı:

```bash
go run ./cmd/sync-umramonline-customers
```

- Kaynak DSN kodda hardcoded (`127.0.0.1:33007/umramdb`)
- Operasyonel ortamda env’e taşınması gerekir
- Sürekli sync için consume webhook tercih edilmeli

## Migration stratejisi

- Schema: GORM AutoMigrate (uygulama açılışında)
- Permission: seed upsert (uygulama açılışında)
- Manuel migration tool (golang-migrate vb.) yok

> AutoMigrate production’da riskli olabilir (beklenmeyen kolon değişiklikleri). Kritik ortamlarda migration politikası netleştirilmelidir.

## Loglama

Standart `log` paketi kullanılır (structured logging / request ID middleware yok).

## Monitoring önerileri (henüz yok)

- Health endpoint genişletmesi (`/` dışında DB/UO check)
- Request latency / error rate metrikleri
- Consume idempotency / stale event oranları
- Disk kullanımı (`storage/follow-ups`)
