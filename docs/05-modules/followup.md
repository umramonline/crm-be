# Follow-up modülü

**Path:** `internal/followup`

## Amaç

Ziyaret / takip kayıtları: anlaşma durumu, not, sonraki ziyaret, görüşülen kişiler, görseller.

## Veri kaynakları

- **MySQL:** follow-up, images, meet people
- **Local disk:** `storage/follow-ups` → static URL `/storage/follow-ups` (**public**)

Umramonline HTTP çağrısı yok.

## Domain kavramları

- `FollowUp` / `FollowUpListItem`
- Create / standalone create / update input’ları
- `Image` / `ImageUpload` / `StoredImage`
- `MeetPerson` — unvan enum’lu kişi kaydı
- `visit_type`: `Yerinde Ziyaret`
- `agreement_failure_reason` enum değerleri (Fiyat yüksek, Mesafe Uzak, …)

## Persistence

Tablolar:

- `tasks_follow_ups`
- `tasks_follow_up_images`
- `follows_meet_people`

AutoMigrate sonrası assignee alanları için one-time SQL backfill çalışabilir.

## Endpoint’ler

| Method | Path | Açıklama |
|--------|------|----------|
| GET | `/api/v1/follow-ups` | Liste (branch-aware) |
| GET | `/api/v1/follow-ups/assigned-to-me` | Atananlar |
| GET | `/api/v1/follow-ups/:uuid` | Detay |
| POST | `/api/v1/follow-ups` | Oluştur (multipart) |
| POST | `/api/v1/follow-ups/standalone` | Task’sız oluştur (branch-restricted) |
| PUT | `/api/v1/follow-ups/:uuid` | Güncelle (multipart) |

## İş kuralları

- Normal create bir `tasks_customer` kaydına bağlanır
- Standalone create, `task_id` null olan bir `tasks_customers` satırı oluşturabilir
- Update’te görseller / meet people replace semantiği uygulanır; silinen dosyalar storage’dan temizlenir
- Static dosya mount’u public’tir — hassas içerik politikası gözden geçirilmelidir
