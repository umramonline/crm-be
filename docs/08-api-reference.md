# API Referansı

**Base path:** `/api/v1`  
**Content-Type:** `application/json` (follow-up create/update: `multipart/form-data`)

## Ortak response

```json
{
  "success": true,
  "message": "string",
  "data": {},
  "errors": {}
}
```

## Auth etiketleri

| Etiket | Anlam |
|--------|-------|
| Public | Middleware yok |
| Session | Access veya refresh cookie doğrulama |
| Session+Perm | `access_token` + `RoleHasAccess(method, route path)` |
| API Key | `Authorization: Bearer {CONSUME_API_KEY}` |

---

## Root / static

| Method | Path | Auth | Açıklama |
|--------|------|------|----------|
| GET | `/` | Public | Greeting / health benzeri |
| GET | `/storage/follow-ups/*` | Public | Follow-up görselleri |

---

## Auth

| Method | Path | Auth | Body / not |
|--------|------|------|------------|
| POST | `/api/v1/auth/otp/request` | Public | `{ "phone" }` |
| POST | `/api/v1/auth/otp/verify` | Public | `{ "phone", "otp_code" }` — cookie basmaz |
| POST | `/api/v1/auth/password/login` | Public | `{ "phone", "password" }` — cookie basar |
| POST | `/api/v1/auth/refresh` | Session (refresh cookie) | Access yeniler |
| POST | `/api/v1/auth/logout` | Public | Cookie temizler |
| GET | `/api/v1/auth/session` | Session (access cookie) | Session + permissions |

Cookie’ler: `access_token`, `refresh_token` (HttpOnly).

---

## Authorization

Tümü **Session+Perm**.

| Method | Path | Not |
|--------|------|-----|
| GET | `/api/v1/authorization/roles` | Umramonline roller |
| GET | `/api/v1/authorization/modules` | |
| POST | `/api/v1/authorization/modules` | |
| PUT | `/api/v1/authorization/modules/:id` | |
| DELETE | `/api/v1/authorization/modules/:id` | |
| GET | `/api/v1/authorization/module-methods` | Query: `module_id` |
| POST | `/api/v1/authorization/module-methods` | |
| PUT | `/api/v1/authorization/module-methods/:id` | |
| DELETE | `/api/v1/authorization/module-methods/:id` | |
| GET | `/api/v1/authorization/role-permissions` | Query: `role_id` |
| PUT | `/api/v1/authorization/role-permissions/:role_id` | Tüm izinleri replace |

---

## Customer

Tümü **Session+Perm**.

| Method | Path | Not |
|--------|------|-----|
| GET | `/api/v1/customers` | Query: `data_source`, filtreler, pagination |
| POST | `/api/v1/customers` | Create |
| GET | `/api/v1/customers/search` | Query: `q` |
| GET | `/api/v1/customers/:id` | Query: `data_source` |
| GET | `/api/v1/customers/backend` | |
| GET | `/api/v1/customers/backend/my-branches` | Session branch scope |
| GET | `/api/v1/customers/backend/:id` | |
| GET | `/api/v1/customers/umramonline` | |
| GET | `/api/v1/customers/umramonline/my-branches` | Session branch scope |
| GET | `/api/v1/customers/umramonline/:id` | |
| GET | `/api/v1/customers/full-registration/:id` | |
| GET | `/api/v1/customers/full-registration/:id/phone-exists` | Query: `cep` |
| PUT | `/api/v1/customers/full-registration/:id` | Full registration body |
| GET | `/api/v1/zones` | |
| GET | `/api/v1/cities` | |
| GET | `/api/v1/towns` | Query: `city_id` |
| GET | `/api/v1/branches` | |
| GET | `/api/v1/branches/:id/users` | |

---

## Task

Tümü **Session+Perm**.

| Method | Path | Not |
|--------|------|-----|
| GET | `/api/v1/tasks` | Filtre + pagination |
| GET | `/api/v1/tasks/assigned-to-me` | |
| GET | `/api/v1/tasks/:uuid` | Query: `tasks_customer_uuid` |
| PATCH | `/api/v1/tasks/:uuid/cancel` | Query/body: `tasks_customer_uuid` |
| POST | `/api/v1/tasks` | Create (+ SMS) |

---

## Follow-up

Tümü **Session+Perm**.

| Method | Path | Not |
|--------|------|-----|
| GET | `/api/v1/follow-ups` | |
| GET | `/api/v1/follow-ups/assigned-to-me` | |
| GET | `/api/v1/follow-ups/:uuid` | |
| POST | `/api/v1/follow-ups` | multipart |
| POST | `/api/v1/follow-ups/standalone` | multipart / branch-restricted |
| PUT | `/api/v1/follow-ups/:uuid` | multipart |

---

## IETTS

Tümü **Session+Perm**.

| Method | Path | Not |
|--------|------|-----|
| GET | `/api/v1/ietts` | Pagination |
| POST | `/api/v1/ietts/:uuid/convert-to-customer` | → `{ customer_id }` |

---

## Dashboard

| Method | Path | Auth | Not |
|--------|------|------|-----|
| GET | `/api/v1/dashboard` | Session+Perm | Query: `start_date`, `end_date` |

Response `data` alanları: `potential_customer_count`, `total_customer_count`, `customer_visit_count`, `new_customer_count`, `vehicle_entry_count`, `total_amount`, `loaded_credit_amount`, `vehicle_stock_count`, task count alanları, `overdue_task_count`.

---

## Consume

| Method | Path | Auth | Not |
|--------|------|------|-----|
| POST | `/api/v1/consume` | API Key | Customer events |

Detay ve örnek payload: [05-modules/consume.md](./05-modules/consume.md)

---

## Sayılar

| Kategori | Adet |
|----------|------|
| Public | 5 (+ static) |
| Session cookie only | 2 (refresh, session) |
| Session + permission | ~43 |
| API key | 1 |
| **Toplam kayıtlı route** | **~52** (+ static mount) |

Route wiring: `internal/infrastructure/http/server.go`
