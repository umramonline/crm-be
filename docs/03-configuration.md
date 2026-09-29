# Konfigürasyon

Config `internal/infrastructure/config/config.go` içinde tanımlanır. Yükleme:

1. `godotenv.Load()` (`.env` varsa)
2. `env.Parse` ile struct tag’lerinden okuma

## Ortam değişkenleri

### Sunucu / CORS

| Değişken | Default | Açıklama |
|----------|---------|----------|
| `PORT` | `8080` | Listen port |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | İzin verilen origin’ler |
| `CORS_ALLOW_CREDENTIALS` | `true` | Cookie ile CORS |
| `SHUTDOWN_TIMEOUT_SECONDS` | `10` | Graceful shutdown timeout |

### Veritabanı

| Değişken | Default | Açıklama |
|----------|---------|----------|
| `DATABASE_DSN` | — | CRM MySQL DSN (zorunlu pratikte) |

Örnek DSN:

```
user:password@tcp(127.0.0.1:3306)/crm?charset=utf8mb4&parseTime=True&loc=Local
```

### Oturum / cookie

| Değişken | Default | Açıklama |
|----------|---------|----------|
| `SESSION_TOKEN_SECRET` | `dev-session-token-secret-change-me` | HMAC secret — **prod’da mutlaka değiştir** |
| `ACCESS_TOKEN_TTL_MINUTES` | `15` | Access cookie TTL |
| `REFRESH_TOKEN_TTL_DAYS` | `30` | Refresh cookie TTL |
| `AUTH_COOKIE_SECURE` | `false` | Secure flag (HTTPS’te `true`) |
| `AUTH_COOKIE_SAME_SITE` | `Lax` | SameSite politikası |

### Consume

| Değişken | Default | Açıklama |
|----------|---------|----------|
| `CONSUME_API_KEY` | — | `POST /api/v1/consume` Bearer key; boşsa endpoint 503 döner |

### Umramonline

| Değişken | Default | Açıklama |
|----------|---------|----------|
| `UMRAMONLINE_BASE_URL` | — | API base URL |
| `UMRAMONLINE_API_KEY` | — | `X-API-KEY` header |
| `UMRAMONLINE_API_TOKEN` | — | `Authorization: Bearer` |
| `UMRAMONLINE_TIMEOUT_SECONDS` | `10` | HTTP client timeout |
| `UMRAMONLINE_OTP_REQUEST_PATH` | `/api/v1/crm/auth/otp/request` | |
| `UMRAMONLINE_OTP_VERIFY_PATH` | `/api/v1/crm/auth/otp/verify` | |
| `UMRAMONLINE_PASSWORD_LOGIN_PATH` | `/api/v1/crm/auth/password/login` | |
| `UMRAMONLINE_USER_ROLES_PATH` | `/api/v1/crm/auth/user-roles` | |
| `UMRAMONLINE_CUSTOMERS_PATH` | `/api/v1/crm/customers` | |
| `UMRAMONLINE_CUSTOMER_SEARCH_PATH` | `/api/v1/crm/customers/search` | |
| `UMRAMONLINE_CUSTOMER_PHONE_EXISTS_PATH` | `/api/v1/crm/customers/phone-exists` | config.go’da; `.env.example`’da yok |
| `UMRAMONLINE_ZONES_PATH` | `/api/v1/crm/zones` | |
| `UMRAMONLINE_CITIES_PATH` | `/api/v1/crm/cities` | |
| `UMRAMONLINE_TOWNS_PATH` | `/api/v1/crm/towns` | |
| `UMRAMONLINE_BRANCHES_PATH` | `/api/v1/crm/branches` | |
| `UMRAMONLINE_TASK_SMS_PATH` | `/api/v1/crm/tasks/sms-created` | config.go’da; `.env.example`’da yok |
| `UMRAMONLINE_DASHBOARD_VEHICLE_ENTRY_PATH` | `/api/v1/crm/dashboard/vehicle-entry-count` | config.go’da |
| `UMRAMONLINE_DASHBOARD_TOTAL_AMOUNT_PATH` | `/api/v1/crm/dashboard/total-amount` | config.go’da |
| `UMRAMONLINE_DASHBOARD_LOADED_CREDIT_PATH` | `/api/v1/crm/dashboard/loaded-credit` | config.go’da |

## Helper metodlar

Config struct üzerinde:

- `Addr()` → `:{PORT}`
- `UmramonlineTimeout()` → duration
- `AccessTokenTTL()` → duration
- `RefreshTokenTTL()` → duration
- `ShutdownTimeout()` → duration

## `.env.example` vs kod

`.env.example` temel değişkenleri listeler. SMS, dashboard path’leri ve phone-exists path’i örnek dosyada olmayabilir; kod default path’lerle çalışır. Production’da tüm kritik secret’ların env’de tanımlı olduğundan emin olun.
