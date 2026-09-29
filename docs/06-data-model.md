# Veri modeli

Bağlantı: `internal/infrastructure/persistence/database.go` → `OpenMySQL(dsn)`.

Startup AutoMigrate sırası: authorization → customer → task → followup → ietts → consume.

## ER özeti

```
modules 1──* module_methods 1──* role_permissions

customers 1──* customer_telephones
customers 1──* tasks_customers *──1 tasks
tasks_customers 1──* tasks_follow_ups
tasks_follow_ups 1──* tasks_follow_up_images
tasks_follow_ups 1──* follows_meet_people

ietts_records.customer_id → customers.id (opsiyonel)
processed_events (consume idempotency ledger)
```

Çoğu tabloda GORM soft-delete (`deleted_at`) kullanılır.

## Authorization

### `modules`

| Kolon | Tip / not |
|-------|-----------|
| `id` | PK |
| `name` | Unique (ör. `customers`) |
| timestamps + soft-delete | |

### `module_methods`

| Kolon | Tip / not |
|-------|-----------|
| `id` | PK |
| `module_id` | FK → modules (CASCADE) |
| `name` | Örn. `customers.list` |
| `description` | |
| `method` | HTTP method (opsiyonel) |
| `path` | Fiber route path (opsiyonel) |
| timestamps + soft-delete | |

### `role_permissions`

| Kolon | Tip / not |
|-------|-----------|
| `id` | PK |
| `role_id` | Umramonline rol ID |
| `module_method_id` | Unique pair with role_id |
| timestamps + soft-delete | |

## Customer

### `customers`

Önemli kolonlar:

| Kolon | Not |
|-------|-----|
| `id` | PK |
| `uo_id` | Umramonline müşteri ID (0 = potansiyel/local) |
| `branch_id` | Şube |
| Kimlik / iletişim | `unvan`, `ad`, `soyad`, `yetkili_adi`, `cep`, `telefon`, `eposta`, … |
| Adres | `mahalle`, `cadde`, `sokak`, `semt`, `il_kodu`, `ilce_kodu`, `kapi_no`, `address_detail` |
| Vergi / kimlik | `vergi_no`, `vergi_dairesi`, `tc_no`, … |
| CRM alanları | `type` (default `bireysel`), `vehicle_stock_count`, `corporate_sector`, `website`, `google_map_link`, … |
| Soft-delete | `deleted_at` |

### `customer_telephones`

| Kolon | Not |
|-------|-----|
| `id` | PK |
| `customer_id` | Index |
| `phone_number` | |
| `title` | |

## Task

### `tasks`

| Kolon | Not |
|-------|-----|
| `id` | PK |
| `uuid` | Unique char(36) |
| `title`, `description` | |
| `created_by_user_id`, `created_by_user_full_name` | |
| `assigned_user_id`, `assigned_user_full_name` | |
| `branch_id`, `branch_name` | |
| `visit_date`, `due_date` | date |
| `priority` | enum `high\|medium\|low` (default medium) |

### `tasks_customers`

| Kolon | Not |
|-------|-----|
| `id` | PK |
| `uuid` | Unique |
| `task_id` | FK → tasks (CASCADE); standalone follow-up’ta null olabilir |
| `customer_id` | FK → customers (CASCADE) |
| `status` | enum `pending\|in_progress\|cancelled\|completed` |

## Follow-up

### `tasks_follow_ups`

| Kolon | Not |
|-------|-----|
| `id` | PK |
| `uuid` | Unique |
| `tasks_customer_id` | FK → tasks_customers (RESTRICT) |
| `visit_type` | enum (`Yerinde Ziyaret`) |
| `visit_date`, `next_visit_date` | |
| `agreement_reached` | bool |
| `agreement_failure_reason` | enum (fiyat, mesafe, …) |
| `note` | max 150 |
| `assigned_user_id`, `assigned_user_full_name` | |

### `tasks_follow_up_images`

| Kolon | Not |
|-------|-----|
| `uuid` | Unique |
| `tasks_follow_up_id` | FK CASCADE |
| `path`, `url` | Local storage path + public URL |

### `follows_meet_people`

| Kolon | Not |
|-------|-----|
| `uuid` | Unique |
| `tasks_follow_up_id` | FK CASCADE |
| `title` | Unvan enum (Genel Müdür, Sahibi, …) |
| `name`, `surname`, `phone`, `email` | |

## IETTS

### `ietts_records`

| Kolon | Not |
|-------|-----|
| `uuid` | Unique |
| `document_number` | |
| `company_name`, `business_name`, `business_address` | |
| `document_issue_date`, `document_status` | |
| `city`, `district` | |
| `customer_id` | Convert sonrası dolar |

## Consume

### `processed_events`

| Kolon | Not |
|-------|-----|
| `event_uuid` | PK char(36) |
| `uo_id` | |
| `event_type` | `customer.created` vb. |
| `occurred_at` | Composite index `(uo_id, event_type, occurred_at)` |
| `processed_at` | |

## Dashboard

Ayrı tablo yok. Mevcut `customers`, `tasks`, `tasks_customers`, follow-up tabloları üzerinde aggregate sorgular çalışır.

## Seed verisi

Seed’ler **iş kaydı** (müşteri/görev) üretmez. Yalnızca authorization module/method/permission upsert eder (`internal/authorization/infrastructure/persistence/*_seed.go`).
