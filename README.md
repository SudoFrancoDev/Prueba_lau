# BIA Energy – Consumption Microservice

A clean-architecture Golang microservice that exposes energy consumption data for electricity meters, aggregated by **daily**, **weekly**, or **monthly** periods.

---

## Architecture

```
cmd/api/               → entrypoint (main.go)
internal/
  domain/              → entities & repository interfaces (no deps)
  repository/          → PostgreSQL implementation + mock address service
  service/             → business logic & aggregation
  handler/             → HTTP handlers (Gin)
  middleware/          → logger, recovery
pkg/
  database/            → PostgreSQL connection helper
migrations/            → SQL DDL scripts
scripts/               → CSV import helper
```

The dependency flow is strictly inward:

```
handler → service → repository → domain
```

---

## Requirements

| Tool | Version |
|------|---------|
| Go | 1.22+ |
| PostgreSQL | 14+ |
| Docker (optional) | any |

---

## Quick Start

### 1 – Start PostgreSQL with Docker

```bash
docker-compose up -d postgres
```

### 2 – Run migrations

```bash
make migrate
```

### 3 – Import CSV data

Place the CSV file (no header, columns: `id,meter_id,active_energy,reactive_inductive,reactive_capacitive,exported,timestamp`) in `data/consumptions.csv`, then:

```bash
make import-csv
# or with a custom path:
make import-csv CSV=/path/to/file.csv
```

### 4 – Run the service

```bash
make run
# or with Docker:
make docker-up
```

---

## Configuration

All settings are read from environment variables with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | DB username |
| `DB_PASSWORD` | `postgres` | DB password |
| `DB_NAME` | `bia_energy` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |

---

## API Reference

### `GET /consumption`

Returns aggregated energy consumption for one or more meters.

#### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `meters_ids` | string | ✅ | Comma-separated meter IDs, e.g. `1,2,3` |
| `start_date` | string | ✅ | ISO date `YYYY-MM-DD` |
| `end_date` | string | ✅ | ISO date `YYYY-MM-DD` |
| `kind_period` | string | ✅ | `daily` \| `weekly` \| `monthly` |

#### Example – Monthly

```bash
curl 'http://localhost:8080/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-07-10&kind_period=monthly'
```

```json
{
  "period": ["Jun 2023", "Jul 2023"],
  "data_graph": [
    {
      "meter_id": 1,
      "address": "Calle 123 # 45-67, Bogotá",
      "active": [1247.77, 0],
      "reactive_inductive": [483.07, 0],
      "reactive_capacitive": [35.68, 0],
      "exported": [33.22, 0]
    }
  ]
}
```

#### Example – Weekly

```bash
curl 'http://localhost:8080/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-06-26&kind_period=weekly'
```

#### Example – Daily

```bash
curl 'http://localhost:8080/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-06-10&kind_period=daily'
```

#### Example – Multiple meters

```bash
curl 'http://localhost:8080/consumption?meters_ids=1,2,3&start_date=2023-06-01&end_date=2023-06-30&kind_period=monthly'
```

---

## CSV Format

```
67c129ab-97ec-43cb-9abe-d2adb9405f70,1,6365.34521,2023-06-27 07:59:55+00
```

Columns (no header row):
1. `id` – UUID
2. `meter_id` – integer
3. `active_energy` – float
4. `reactive_inductive` – float
5. `reactive_capacitive` – float
6. `exported` – float
7. `timestamp` – `YYYY-MM-DD HH:MM:SS+TZ`

---

## Running Tests

```bash
# All unit tests
make test-unit

# With verbose output
go test ./internal/... -v
```

---

## Git Flow

```
main          → production-ready code
develop       → integration branch
feature/*     → new features  (branched from develop)
fix/*         → bug fixes     (branched from develop)
release/*     → release prep  (branched from develop → merged to main + develop)
```

Example:
```bash
git checkout develop
git checkout -b feature/weekly-aggregation
# ... code, commit ...
git push origin feature/weekly-aggregation
# open PR → develop
```
