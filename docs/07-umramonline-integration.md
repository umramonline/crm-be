# Umramonline Entegrasyonu

## Genel

CRM backend, Umramonline’ı hem kimlik kaynağı hem de referans/operasyonel veri kaynağı olarak kullanır.

**Client:** `internal/umramonline/client.go`

### Ortak header’lar

Her istekte:

- `X-API-KEY: {UMRAMONLINE_API_KEY}`
- `Authorization: Bearer {UMRAMONLINE_API_TOKEN}`
- `Accept` / `Content-Type: application/json`

`BaseURL`, API key ve token boş olmamalıdır.

## Client metodları

| Client method | HTTP | Default path |
|---------------|------|--------------|
| `RequestOTP` | POST | `/api/v1/admin/login` |
| `VerifyOTP` | POST | `/api/v1/admin/login/verify` |
| `LoginWithPassword` | POST | `/api/v1/admin/login` |
| `ListRoles` | GET | `/api/v1/admin/crm/user-roles` |
| `ListCustomers` | GET | `/api/v1/admin/crm/customer-list` |
| `GetCustomer` | GET | `/api/v1/crm/customers/{id}` |
| `SearchCustomer` | GET | `/api/v1/admin/customer/search?q=` |
| `CustomerPhoneExists` | GET | *(env; UO endpoint yoksa kapalı)* |
| `ListZones` | GET | `/api/v1/admin/crm/zones` (+ `branch_ids[]`) |
| `ListCities` | GET | `/api/v1/admin/cities` |
| `ListTowns` | GET | `/api/v1/admin/towns?city_id=` |
| `ListBranches` | GET | `/api/v1/admin/crm/branches` |
| `GetBranch` | GET | `{UMRAMONLINE_BRANCHES_PATH}/{id}` |
| `ListBranchUsers` | GET | `{UMRAMONLINE_BRANCHES_PATH}/{id}/users` |
| `GetBranchUser` | GET | `{UMRAMONLINE_BRANCHES_PATH}/{id}/users/{userId}` |
| `SendTaskCreatedSMS` | POST | *(env; UO endpoint yoksa kapalı)* |
| `GetDashboardStats` | GET | `/api/v1/admin/dashboard` |

Dashboard query: zorunlu `start_date`, `end_date` (`Y-m-d`); yanıtta `vehicle_entry_count`, `total_amount`, `loaded_credit`.

## Modül kullanım matrisi

| CRM modülü | Adapter | Umramonline kullanımı |
|------------|---------|------------------------|
| auth | Doğrudan client | OTP request/verify, password login |
| authorization | `authorization/.../umramonline` | `ListRoles` |
| customer | `customer/.../umramonline` | customers, search, phone-exists, zones, cities, towns, branches, users |
| task | `task/.../umramonline` | GetBranch, GetBranchUser, SendTaskCreatedSMS |
| dashboard | `dashboard/.../umramonline` | `GET /api/v1/admin/dashboard` (3 metrik) |
| followup | — | HTTP çağrısı yok |
| ietts | — | HTTP çağrısı yok |
| consume | — | HTTP çağrısı yok (UO CRM’e event gönderir) |

## Veri akış yönleri

```mermaid
flowchart TB
  subgraph outbound [CRM → Umramonline]
    Auth[OTP / Login]
    Ref[Customers / Zones / Branches]
    SMS[Task SMS]
    Dash[Dashboard stats]
  end

  subgraph inbound [Umramonline → CRM]
    Consume[POST /api/v1/consume events]
    SyncCLI[sync CLI - direct MySQL]
  end

  Auth --> UO[Umramonline]
  Ref --> UO
  SMS --> UO
  Dash --> UO
  UO --> Consume
  UODB[(UO MySQL)] --> SyncCLI
```

## Anti-Corruption Layer

Her modül kendi `Provider` tipi ile Umramonline cevabını domain DTO’larına map’ler. Application katmanı ham UO JSON’u görmez; interface üzerinden çalışır. Bu sayede UO şema değişiklikleri adapter’da izole edilebilir.

## Hata yaklaşımları

- UO 422 (login): kimlik bilgileri hatalı
- Timeout: `UMRAMONLINE_TIMEOUT_SECONDS`
- Downstream unavailable: application katmanında `*Unavailable` hataları → HTTP 5xx / anlamlı mesaj

## Sync CLI ayrımı

`cmd/sync-umramonline-customers` Umramonline **HTTP API** kullanmaz; kaynak MySQL’e doğrudan bağlanır. Bu, API entegrasyonundan ayrı bir operasyonel araçtır.
