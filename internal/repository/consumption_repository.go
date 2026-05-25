package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/bia-energy/consumption-service/internal/domain"
)

type consumptionRepository struct {
	db *sql.DB
}

// NewConsumptionRepository creates a new MySQL-backed consumption repository.
func NewConsumptionRepository(db *sql.DB) domain.ConsumptionRepository {
	return &consumptionRepository{db: db}
}

// GetByMeterIDsAndDateRange fetches raw consumption records for the given meters and date range.
// MySQL uses ? placeholders (not $1, $2 like PostgreSQL).
func (r *consumptionRepository) GetByMeterIDsAndDateRange(
	ctx context.Context,
	meterIDs []int64,
	query domain.ConsumptionQuery,
) ([]domain.Consumption, error) {
	if len(meterIDs) == 0 {
		return nil, fmt.Errorf("no meter IDs provided")
	}

	// Build ? placeholders for the IN clause: (?, ?, ...)
	placeholders := strings.Repeat("?,", len(meterIDs))
	placeholders = placeholders[:len(placeholders)-1] // trim trailing comma

	args := make([]interface{}, 0, len(meterIDs)+2)
	for _, id := range meterIDs {
		args = append(args, id)
	}
	args = append(args, query.StartDate, query.EndDate)

	q := fmt.Sprintf(`
		SELECT id, meter_id, active_energy, reactive_inductive, reactive_capacitive, exported, timestamp
		FROM consumptions
		WHERE meter_id IN (%s)
		  AND timestamp >= ?
		  AND timestamp <= ?
		ORDER BY meter_id, timestamp ASC
	`, placeholders)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query consumptions: %w", err)
	}
	defer rows.Close()

	var results []domain.Consumption
	for rows.Next() {
		var c domain.Consumption
		if err := rows.Scan(
			&c.ID,
			&c.MeterID,
			&c.ActiveEnergy,
			&c.ReactiveInductive,
			&c.ReactiveCapacitive,
			&c.ExportedEnergy,
			&c.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan consumption row: %w", err)
		}
		results = append(results, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate consumption rows: %w", err)
	}

	return results, nil
}
