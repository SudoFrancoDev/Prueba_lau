#!/usr/bin/env bash
# scripts/import_csv.sh
# Imports the energy consumption CSV into MySQL (XAMPP compatible).
#
# Usage:
#   DB_HOST=localhost DB_USER=root DB_PASSWORD= \
#   DB_NAME=bia_energy ./scripts/import_csv.sh data/consumptions.csv

set -euo pipefail

CSV_FILE="${1:-data/consumptions.csv}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-bia_energy}"

if [ ! -f "$CSV_FILE" ]; then
  echo "ERROR: CSV file not found at '$CSV_FILE'"
  exit 1
fi

# Build mysql auth string
MYSQL_CMD="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER"
if [ -n "$DB_PASSWORD" ]; then
  MYSQL_CMD="$MYSQL_CMD -p$DB_PASSWORD"
fi

echo "==> Running migrations..."
$MYSQL_CMD < migrations/001_create_consumptions.sql

echo "==> Importing CSV: $CSV_FILE"
# Convert the file to an absolute path (MySQL LOAD DATA needs it)
ABS_CSV=$(realpath "$CSV_FILE")

$MYSQL_CMD "$DB_NAME" <<SQL
LOAD DATA LOCAL INFILE '$ABS_CSV'
INTO TABLE consumptions
FIELDS TERMINATED BY ','
LINES TERMINATED BY '\n'
(id, meter_id, active_energy, reactive_inductive, reactive_capacitive, exported, timestamp);
SQL

echo "==> Done. Row count:"
$MYSQL_CMD "$DB_NAME" -e "SELECT COUNT(*) AS total_rows FROM consumptions;"
