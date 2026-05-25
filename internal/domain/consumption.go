package domain

import "time"

// Consumption represents a single energy consumption record from the database.
type Consumption struct {
	ID                  string    `db:"id"`
	MeterID             int64     `db:"meter_id"`
	ActiveEnergy        float64   `db:"active_energy"`
	ReactiveInductive   float64   `db:"reactive_inductive"`
	ReactiveCapacitive  float64   `db:"reactive_capacitive"`
	ExportedEnergy      float64   `db:"exported"`
	Timestamp           time.Time `db:"timestamp"`
}

// KindPeriod represents the aggregation period type.
type KindPeriod string

const (
	KindPeriodDaily   KindPeriod = "daily"
	KindPeriodWeekly  KindPeriod = "weekly"
	KindPeriodMonthly KindPeriod = "monthly"
)

// ConsumptionQuery holds the parameters for querying consumption data.
type ConsumptionQuery struct {
	MeterIDs  []int64
	StartDate time.Time
	EndDate   time.Time
	Period    KindPeriod
}

// MeterConsumption holds aggregated consumption data for a single meter.
type MeterConsumption struct {
	MeterID            int64     `json:"meter_id"`
	Address            string    `json:"address"`
	Active             []float64 `json:"active"`
	ReactiveInductive  []float64 `json:"reactive_inductive"`
	ReactiveCapacitive []float64 `json:"reactive_capacitive"`
	Exported           []float64 `json:"exported"`
}

// ConsumptionResponse is the final JSON response returned by the API.
type ConsumptionResponse struct {
	Period    []string           `json:"period"`
	DataGraph []MeterConsumption `json:"data_graph"`
}
