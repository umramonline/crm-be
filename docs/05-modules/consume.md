# Consume modülü

**Path:** `internal/consume`

## Amaç

Umramonline (veya eşdeğer sistem) tarafından gönderilen müşteri değişiklik event’lerini CRM MySQL’e uygulamak. Message broker yoktur; HTTP webhook kullanılır.

## Auth

`Authorization: Bearer {CONSUME_API_KEY}`

- Key boşsa middleware 503 döner
- Session cookie kullanılmaz

## Endpoint

`POST /api/v1/consume`

## Envelope

Top-level JSON:

```json
{
  "event_id": "<uuid>",
  "event_type": "customer.created|customer.updated|customer.deleted",
  "...": "payload alanları aynı seviyede"
}
```

Handler tüm body’yi payload olarak application’a iletir.

## Event tipleri

| `event_type` | Davranış | Zorunlu alanlar |
|--------------|----------|-----------------|
| `customer.created` | Upsert customer + telephones | `uo_id`, `occurred_at`, müşteri alanları |
| `customer.updated` | Upsert (stale check) | Aynı |
| `customer.deleted` | Soft-delete by `uo_id` | `uo_id`, `occurred_at` |

### Created / updated payload alanları

`uo_id`, `branch_id`, `unvan`, `ad`, `soyad`, `yetkili_adi`, `cep`, `telefon`, `fax`, `eposta`, `web`, adres alanları, vergi/kimlik alanları, `type`, `mersis`, pasaport alanları, `created_at`, `updated_at`, `telephones[{phone_number,title}]`, `occurred_at`.

## Idempotency ve sıralama

Tablo: `processed_events`

1. Aynı `event_uuid` daha önce işlendiyse → `already_processed`
2. Update’te aynı `uo_id` için daha yeni `occurred_at` varsa → `stale_event` (erken return; processed kaydı yazılmayabilir)
3. Upsert duplicate sezgisi: `(tc_no OR vergi_no)` ve `(telefon OR cep)` örtüşmesi
4. Delete: telephones + customer soft-delete; müşteri yoksa da processed kaydı `deleted` olabilir

## Sonuç action → HTTP

| Action | Status | Message |
|--------|--------|---------|
| `created` | 201 | Customer created. |
| `updated` | 200 | Customer updated. |
| `already_processed` | 200 | Event already processed. |
| `stale_event` | 200 | Stale event skipped. |
| `deleted` | 200 | Customer deleted. |

## Örnek: customer.created

```json
{
  "event_id": "11111111-1111-1111-1111-111111111111",
  "event_type": "customer.created",
  "uo_id": 12345,
  "branch_id": 1,
  "unvan": "Örnek Ticaret",
  "ad": "Ali",
  "soyad": "Veli",
  "cep": "05551234567",
  "telefon": "02121234567",
  "type": "kurumsal",
  "occurred_at": "2026-07-30T10:00:00Z",
  "telephones": [
    { "phone_number": "05559876543", "title": "Muhasebe" }
  ]
}
```

## Örnek: customer.deleted

```json
{
  "event_id": "22222222-2222-2222-2222-222222222222",
  "event_type": "customer.deleted",
  "uo_id": 12345,
  "occurred_at": "2026-07-30T12:00:00Z"
}
```

## Sync CLI ile ilişki

- **Sync CLI:** ilk/toplu backfill (direct MySQL)
- **Consume:** sürekli / incremental senkronizasyon (HTTP events)
