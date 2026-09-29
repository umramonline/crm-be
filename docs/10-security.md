# Güvenlik

## Kimlik ve oturum

| Konu | Durum | Öneri |
|------|-------|-------|
| Token imza | Custom HMAC-SHA256 JWT | `SESSION_TOKEN_SECRET` prod’da güçlü ve gizli olmalı |
| Cookie | HttpOnly | HTTPS’te `AUTH_COOKIE_SECURE=true` |
| SameSite | Default `Lax` | Cross-site ihtiyaç yoksa uygun |
| Logout | Cookie silme | Server-side revoke / blacklist yok — access token TTL kısa tutulmalı |
| OTP verify | Session açmaz | Bilinçli ürün kararı mı netleştirilsin |

## Yetkilendirme

- Route koruması `RoleHasAccess` ile method + Fiber route path exact match
- Path değişince seed/DB permission güncellenmezse 403 oluşur
- UI-only permission’lar (`*.menu`) API’yi korumaz
- Admin `role_id = 30` hardcoded — değişirse `shared/auth` + seed’ler güncellenmeli

## Consume endpoint

- Ayrı Bearer API key ile korunur
- Key boşsa 503
- Session auth ile karıştırılmamalı
- Key rotasyonu için env güncelleme + deploy gerekir
- Rate limiting / IP allowlist şu an yok

## Public yüzeyler

| Yüzey | Risk |
|-------|------|
| `GET /` | Düşük |
| Auth OTP/login | Brute-force koruması bu katmanda yok (UO’ya bağlı) |
| `/storage/follow-ups/*` | **Public dosya servisi** — hassas görseller için erişim kontrolü yok |
| CORS | `CORS_ALLOWED_ORIGINS` sıkı tutulmalı |

## Secret yönetimi

Env’de tutulan secret’lar:

- `SESSION_TOKEN_SECRET`
- `UMRAMONLINE_API_KEY` / `UMRAMONLINE_API_TOKEN`
- `DATABASE_DSN`
- `CONSUME_API_KEY`

`.env` commit edilmemeli (`.gitignore` ile). `.env.example` secret değer içermemeli.

## Veri

- Soft-delete birçok tabloda aktif; hard delete consume/delete akışında da soft olabilir
- PII (telefon, TC, vergi no) local DB’de — erişim logları ve yedekleme politikası operasyonel olarak tanımlanmalı
- Sync CLI hardcoded DB credential içerir — production için riskli

## Bağımlılık / transport

- Umramonline TLS doğrulaması standart `net/http` client davranışına bağlı
- Timeout varsayılan 10s
- Retry/circuit breaker yok

## Checklist (prod)

- [ ] `SESSION_TOKEN_SECRET` değiştirildi
- [ ] `AUTH_COOKIE_SECURE=true`
- [ ] CORS origin whitelist doğru
- [ ] `CONSUME_API_KEY` set ve güçlü
- [ ] Umramonline key/token rotasyonu planı var
- [ ] Follow-up static dosya erişim politikası gözden geçirildi
- [ ] Sync CLI hardcoded DSN kullanılmıyor / kaldırıldı
- [ ] HTTPS termination (reverse proxy) aktif
