# Task modülü

**Path:** `internal/task`

## Amaç

Şube / kullanıcıya görev atama, müşteri bağlantıları (`tasks_customers`), görev iptali ve oluşturma sonrası SMS bildirimi.

## Veri kaynakları

- **MySQL:** görevler ve müşteri link’leri
- **Umramonline:** `GetBranch`, `GetBranchUser` (doğrulama), `SendTaskCreatedSMS`

## Domain kavramları

- `Task` / `TaskListItem`
- `CreateTaskInput`
- `TaskCustomer` — görev-müşteri satırı + status
- `TaskCreatedSMSInput`
- Priority: `high | medium | low`
- Status: `pending | in_progress | cancelled | completed`

## Persistence

Tablolar: `tasks`, `tasks_customers`

## Endpoint’ler

| Method | Path | Açıklama |
|--------|------|----------|
| GET | `/api/v1/tasks` | Filtreli liste |
| GET | `/api/v1/tasks/assigned-to-me` | Oturum kullanıcısına atananlar |
| GET | `/api/v1/tasks/:uuid` | Detay (`tasks_customer_uuid` query gerekebilir) |
| PATCH | `/api/v1/tasks/:uuid/cancel` | İptal (`tasks_customer_uuid` zorunlu) |
| POST | `/api/v1/tasks` | Oluştur (+ opsiyonel SMS) |

## İş kuralları

- Oluştururken branch ve assigned user Umramonline’dan doğrulanır
- Müşterilerin ilgili şubeye ait olup olmadığı kontrol edilir (`InvalidCustomerIDsForBranch`)
- Cancel, ilgili `tasks_customers.status` değerini `cancelled` yapar
- SMS başarısız olsa bile görev oluşmuş olabilir (implementasyon detayına göre handler davranışı)

## İlişkiler

```
tasks 1──* tasks_customers *──1 customers
tasks_customers 1──* tasks_follow_ups
```
