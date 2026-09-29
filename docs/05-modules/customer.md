# Customer modülü

**Path:** `internal/customer`

## Amaç

Müşteri listeleme, arama, oluşturma, tam kayıt (full registration) ve referans verileri (zones, cities, towns, branches, branch users).

## Veri kaynakları

İki kaynak desteklenir (`data_source`):

| Kaynak | Anlamı |
|--------|--------|
| `backend` | CRM MySQL (`customers` tablosu) |
| `umramonline` | Umramonline HTTP API |

Bazı listelerde merge yapılır: Umramonline listesi `uo_id` üzerinden backend kayıtlarıyla zenginleştirilir (veya tersi, sorgunun yönüne göre).

## Domain kavramları

- `Customer` — liste satırı
- `CustomerDetail` — detay + telephones
- `CreateCustomerInput` — minimal oluşturma
- `FullRegistrationInput` — tam kayıt formu
- `ListQuery` / `ListResult` / `Pagination`
- `Zone`, `City`, `Town`, `Branch`, `BranchUser`

## Persistence

Tablolar: `customers`, `customer_telephones`  
Detay: [06-data-model.md](../06-data-model.md)

## Endpoint’ler

Hepsi session + permission gerektirir.

| Method | Path | Açıklama |
|--------|------|----------|
| GET | `/api/v1/customers` | Liste (`data_source` + filtreler) |
| POST | `/api/v1/customers` | Yeni müşteri |
| GET | `/api/v1/customers/search` | Arama (`q`) |
| GET | `/api/v1/customers/:id` | Detay (`data_source`) |
| GET | `/api/v1/customers/backend` | Backend liste |
| GET | `/api/v1/customers/backend/my-branches` | Session şube scoped backend |
| GET | `/api/v1/customers/backend/:id` | Backend detay |
| GET | `/api/v1/customers/umramonline` | UO liste |
| GET | `/api/v1/customers/umramonline/my-branches` | Session şube scoped UO |
| GET | `/api/v1/customers/umramonline/:id` | UO detay |
| GET | `/api/v1/customers/full-registration/:id` | Tam kayıt formu verisi |
| GET | `/api/v1/customers/full-registration/:id/phone-exists` | Cep uniqueness (`cep` query) |
| PUT | `/api/v1/customers/full-registration/:id` | Tam kayıt tamamla / güncelle |
| GET | `/api/v1/zones` | Bölgeler |
| GET | `/api/v1/cities` | Şehirler |
| GET | `/api/v1/towns` | İlçeler (`city_id`) |
| GET | `/api/v1/branches` | Bayiler |
| GET | `/api/v1/branches/:id/users` | Bayi kullanıcıları |

## İş kuralları

- Telefon formatı: `05XXXXXXXXX`
- Create sırasında hem backend hem Umramonline phone uniqueness kontrol edilir
- Branch scope: admin (`role_id=30`) hariç session `branch_ids` ile sınırlanır
- Full registration corporate sector sabit listeden seçilir (Teknoloji, İnşaat, Otomotiv, …)
- Metin alanlarında max length validasyonu vardır
- `uo_id = 0` kayıtlar potansiyel / henüz UO’ya bağlı olmayan müşteriler olarak dashboard’da sayılır
- Liste `sort_by` whitelist: `domain.NormalizeListSortBy` — `credit`, `point`, `created_at`, `vehicle_stock_count`; backend SQL sıralaması yalnızca `created_at`, `vehicle_stock_count`
- `page` / `per_page`: `NormalizeListQuery` (1–100); birleşik listede backend metin filtresi + UO sıralama/filtresi için merge katmanı sayfalar (`PaginateCustomers`, UO intersect taraması)

## Provider / Repository

- `CustomerProvider` — Umramonline adapter
- `CustomerRepository` — local MySQL

Application servisi her iki interface’i orkestre eder.
