# Genel Bakış

## Amaç

Bu proje, Umran CRM uygulamasının backend API’sidir. Frontend’e REST endpoint’ler sunar; kimlik doğrulama ve bir kısım müşteri/referans verisini **Umramonline** sisteminden alır; görev, takip (follow-up), IETTS, yetkilendirme ve CRM müşteri kayıtlarını kendi **MySQL** veritabanında tutar.

**Go module:** `github.com/umran/new.crm/backend`

## Teknoloji yığını

| Alan | Teknoloji | Not |
|------|-----------|-----|
| Dil | Go 1.25 | |
| HTTP framework | Fiber v2 | fasthttp tabanlı |
| ORM / veritabanı | GORM + MySQL | Startup’ta AutoMigrate |
| Config | `godotenv` + `caarlos0/env` | `.env` dosyasından |
| Oturum token | Custom HMAC-SHA256 JWT | 3rd-party JWT kütüphanesi yok |
| Concurrent işler | `golang.org/x/sync/errgroup` | Dashboard metrikleri |
| ID üretimi | `google/uuid` | Task / follow-up |
| CI/CD | GitHub Actions SSH deploy | Workflow şu an yorum satırında |

## Binary’ler

| Binary | Path | Açıklama |
|--------|------|----------|
| API sunucusu | `cmd/api` | Ana HTTP API |
| Müşteri sync | `cmd/sync-umramonline-customers` | Umramonline MySQL → CRM MySQL toplu backfill |

## Mimari özeti (tek cümle)

Fiber REST API + GORM/MySQL + Umramonline HTTP entegrasyonu; modüler Clean Architecture; cookie tabanlı oturum + RBAC; HTTP webhook ile müşteri event tüketimi.

## Ne yapar / ne yapmaz

### Yapar

- OTP / password login (Umramonline üzerinden)
- Cookie tabanlı session + rol bazlı yetkilendirme
- Müşteri CRUD (local DB + Umramonline kaynağı)
- Görev atama, iptal, SMS bildirimi
- Ziyaret / follow-up kayıtları ve görsel yükleme
- IETTS kayıtlarını müşteriye dönüştürme
- Dashboard KPI’ları (local + Umramonline)
- Dış sistemden müşteri event’lerini tüketme (`/consume`)

### Yapmaz

- Kafka / RabbitMQ / NATS gibi message broker kullanmaz
- Event üretmez / Event Sourcing uygulamaz
- GraphQL veya gRPC sunmaz
- Redis cache kullanmaz
- Standart OpenAPI/Swagger üretimi yoktur
- Server-side token revoke (blacklist) yoktur — logout cookie siler

## DDD / Event-driven durumu

| Soru | Cevap |
|------|-------|
| DDD mi? | **Kısmen** — modül bazlı bounded context + domain paketleri var; zengin aggregate / domain event publish yok (lightweight DDD) |
| Event-driven mı? | **Hayır (full EDA değil)** — yalnızca inbound HTTP webhook ile `customer.*` event tüketimi var |

## İlişkili sistemler

```
CRM Frontend ──cookie session──► CRM Backend (bu repo)
                                      │
                                      ├── MySQL (CRM)
                                      │
                                      └── HTTP ──► Umramonline API
                                                      │
Umramonline ──POST /api/v1/consume──► CRM Backend ◄───┘
```
