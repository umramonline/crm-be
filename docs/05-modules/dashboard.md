# Dashboard modülü

**Path:** `internal/dashboard`

## Amaç

Tarih aralığı ve şube kapsamına göre KPI dashboard verisi üretmek.

## Veri kaynakları (hibrit)

### Local MySQL (repository)

- Potential customers (`uo_id = 0`)
- Total customers
- Customer visits
- New customers
- Vehicle stock sum
- Task stats (pending / in_progress / completed)
- Overdue tasks

### Umramonline (provider)

- Vehicle entry count
- Total amount
- Loaded credit

Metrikler `errgroup` ile paralel çekilir.

## Domain

`Stats` alanları:

| JSON field | Kaynak |
|------------|--------|
| `potential_customer_count` | Local |
| `total_customer_count` | Local |
| `customer_visit_count` | Local |
| `new_customer_count` | Local |
| `vehicle_stock_count` | Local |
| `pending_task_count` | Local |
| `in_progress_task_count` | Local |
| `completed_task_count` | Local |
| `overdue_task_count` | Local |
| `vehicle_entry_count` | Umramonline |
| `total_amount` | Umramonline |
| `loaded_credit_amount` | Umramonline |

Filter: `start_date`, `end_date`, `branch_ids`, `AllowAllBranches` (admin).

## Endpoint

| Method | Path | Açıklama |
|--------|------|----------|
| GET | `/api/v1/dashboard` | Query: `start_date`, `end_date` |

## İş kuralları

- Admin tüm şubeleri görebilir (`AllowAllBranches`)
- Diğer roller session branch’leriyle sınırlanır
- Geçersiz tarih filtresi validation error döner
