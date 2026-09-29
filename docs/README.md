# CRM Backend Dokümantasyonu

Umran CRM backend API’sinin teknik dokümantasyonu.

## İçindekiler

| Bölüm | Dosya | Açıklama |
|-------|-------|----------|
| Genel bakış | [00-overview.md](./00-overview.md) | Proje amacı, teknoloji yığını, sınırlar |
| Mimari | [01-architecture.md](./01-architecture.md) | Katmanlar, prensipler, dizin yapısı |
| Başlangıç | [02-getting-started.md](./02-getting-started.md) | Kurulum, çalıştırma, sync CLI |
| Konfigürasyon | [03-configuration.md](./03-configuration.md) | Ortam değişkenleri |
| Auth & RBAC | [04-auth-and-rbac.md](./04-auth-and-rbac.md) | Oturum, cookie, izin modeli |
| Modüller | [05-modules/](./05-modules/) | Customer, task, follow-up, ietts, dashboard, consume |
| Veri modeli | [06-data-model.md](./06-data-model.md) | Tablolar ve ilişkiler |
| Umramonline | [07-umramonline-integration.md](./07-umramonline-integration.md) | Dış API entegrasyonu |
| API referansı | [08-api-reference.md](./08-api-reference.md) | Tüm HTTP endpoint’ler |
| Ops & deploy | [09-ops-deploy.md](./09-ops-deploy.md) | Deploy, startup, shutdown |
| Güvenlik | [10-security.md](./10-security.md) | Güvenlik notları ve dikkat edilecekler |

## Agent / Cursor standartları

- Clean Architecture (tüm repo): `../../.cursor/rules/clean-architecture.mdc`
- crm-be Go katmanları: `../.cursor/rules/clean-architecture-be.mdc`

## Hızlı özet

- **Dil:** Go 1.25
- **HTTP:** Fiber v2
- **DB:** MySQL + GORM
- **Entry point:** `cmd/api`
- **API prefix:** `/api/v1`
- **Auth:** Cookie tabanlı custom HMAC-JWT + Umramonline kimlik doğrulama
- **Yetki:** Rol → module_method (HTTP method + path)
