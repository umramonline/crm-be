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

Workflow build komutu: `go build -o app ./cmd/api`.

## GitHub Actions

| Workflow | Tetikleyici | Amaç |
|----------|-------------|------|
| `ci.yml` | push (main, test, develop), PR | `go test ./...`, build doğrulama |
| `deploy-test.yml` | push → `test` | CI sonrası test sunucusu |
| `deploy-production.yml` | push → `main` | CI sonrası production |

Deploy: self-hosted runner, SSH, sunucuda `git pull`, `go build -o app ./cmd/api`, `systemctl restart` (varsayılan unit: `crm-api`).

### GitHub Environment ayarları

**Variables** (`test` / `production`):

| Variable | Örnek (test) |
|----------|----------------|
| `DEPLOY_APP_DIR` | `/var/www/testcrm.umram.online/crm-be` |
| `HEALTH_CHECK_URL` | API kökü (`GET /`) |
| `SYSTEMD_SERVICE` | `crm-api` (opsiyonel) |

**Secrets:** `TEST_SSH_PRIVATE_KEY`, `TEST_SSH_HOST`, `TEST_SSH_USER` ve production için `PROD_SSH_*`.

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
