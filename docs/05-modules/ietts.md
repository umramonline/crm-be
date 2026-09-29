# IETTS modülü

**Path:** `internal/ietts`

## Amaç

IETTS belge kayıtlarını listelemek ve seçilen kaydı CRM müşterisine dönüştürmek.

## Veri kaynakları

Yalnızca MySQL (`ietts_records`). Umramonline HTTP yok.

Kayıtların nasıl import edildiği bu repoda tanımlı değildir (dış süreç / başka araç).

## Domain kavramları

- `RecordListItem` / `Record`
- `CustomerFromIettsInput` — müşteri oluşturma alanları
- `ConvertToCustomerResult` — `{ customer_id }`
- `business_name` yardımcıları — unvanı ad/soyad parçalarına ayırma / truncate

## Persistence

Tablo: `ietts_records`  
`customer_id` convert sonrası dolar.

## Endpoint’ler

| Method | Path | Açıklama |
|--------|------|----------|
| GET | `/api/v1/ietts` | Sayfalı liste (`sort_by`, `sort_order`, kolon filtreleri) |
| POST | `/api/v1/ietts/:uuid/convert-to-customer` | Müşteriye dönüştür |

## Liste sıralama

`sort_by` whitelist (`domain.NormalizeListSortBy`): `document_number`, `company_name`, `business_name`, `business_address`, `document_issue_date`, `document_status`, `city`, `district`, `created_at`. Geçersiz değer → varsayılan `id DESC`.

## İş kuralları

- Convert, customer repo üzerinden `CreateCustomerFromIetts` çağırır
- Ardından IETTS kaydına `customer_id` yazılır
- Business name parse kuralları domain helper’larında test edilmiştir
