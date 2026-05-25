package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bia-energy/consumption-service/internal/domain"
)

// ConsumptionService handles business logic for energy consumption queries.
type ConsumptionService struct {
	repo    domain.ConsumptionRepository
	address domain.AddressService
}

// NewConsumptionService creates a new ConsumptionService.
func NewConsumptionService(repo domain.ConsumptionRepository, address domain.AddressService) *ConsumptionService {
	return &ConsumptionService{repo: repo, address: address}
}

// GetConsumption fetches and aggregates consumption data based on the query parameters.
func (s *ConsumptionService) GetConsumption(ctx context.Context, query domain.ConsumptionQuery) (*domain.ConsumptionResponse, error) {
	records, err := s.repo.GetByMeterIDsAndDateRange(ctx, query.MeterIDs, query)
	if err != nil {
		return nil, fmt.Errorf("fetch records: %w", err)
	}

	periods := buildPeriods(query.StartDate, query.EndDate, query.Period)

	// Group records by meter ID
	byMeter := make(map[int64][]domain.Consumption)
	for _, r := range records {
		byMeter[r.MeterID] = append(byMeter[r.MeterID], r)
	}

	dataGraph := make([]domain.MeterConsumption, 0, len(query.MeterIDs))
	for _, meterID := range query.MeterIDs {
		addr, err := s.address.GetAddressByMeterID(ctx, meterID)
		if err != nil {
			addr = "Dirección desconocida"
		}

		mc := domain.MeterConsumption{
			MeterID:            meterID,
			Address:            addr,
			Active:             make([]float64, len(periods)),
			ReactiveInductive:  make([]float64, len(periods)),
			ReactiveCapacitive: make([]float64, len(periods)),
			Exported:           make([]float64, len(periods)),
		}

		meterRecords := byMeter[meterID]
		aggregateIntoPeriods(meterRecords, periods, query.Period, &mc)
		dataGraph = append(dataGraph, mc)
	}

	periodLabels := make([]string, len(periods))
	for i, p := range periods {
		periodLabels[i] = p.label
	}

	return &domain.ConsumptionResponse{
		Period:    periodLabels,
		DataGraph: dataGraph,
	}, nil
}

// period represents a named time bucket with a start and end.
type period struct {
	label string
	start time.Time
	end   time.Time
}

// buildPeriods generates the list of time buckets based on the kind of period.
func buildPeriods(start, end time.Time, kind domain.KindPeriod) []period {
	switch kind {
	case domain.KindPeriodDaily:
		return buildDailyPeriods(start, end)
	case domain.KindPeriodWeekly:
		return buildWeeklyPeriods(start, end)
	case domain.KindPeriodMonthly:
		return buildMonthlyPeriods(start, end)
	default:
		return buildDailyPeriods(start, end)
	}
}

func buildDailyPeriods(start, end time.Time) []period {
	var periods []period
	cur := truncateToDay(start)
	endDay := truncateToDay(end)
	for !cur.After(endDay) {
		periods = append(periods, period{
			label: cur.Format("Jan 2"),
			start: cur,
			end:   cur.Add(24*time.Hour - time.Second),
		})
		cur = cur.AddDate(0, 0, 1)
	}
	return periods
}

func buildWeeklyPeriods(start, end time.Time) []period {
	var periods []period
	cur := truncateToDay(start)
	for cur.Before(end) || cur.Equal(truncateToDay(end)) {
		weekEnd := cur.AddDate(0, 0, 6)
		if weekEnd.After(end) {
			weekEnd = truncateToDay(end)
		}
		label := fmt.Sprintf("%s - %s",
			cur.Format("Jan 2"),
			weekEnd.Format("Jan 2"),
		)
		periods = append(periods, period{
			label: label,
			start: cur,
			end:   weekEnd.Add(24*time.Hour - time.Second),
		})
		cur = cur.AddDate(0, 0, 7)
	}
	return periods
}

func buildMonthlyPeriods(start, end time.Time) []period {
	var periods []period
	cur := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())
	for !cur.After(endMonth) {
		nextMonth := cur.AddDate(0, 1, 0)
		periods = append(periods, period{
			label: cur.Format("Jan 2006"),
			start: cur,
			end:   nextMonth.Add(-time.Second),
		})
		cur = nextMonth
	}
	return periods
}

// aggregateIntoPeriods sums consumption records into the correct period slots.
func aggregateIntoPeriods(records []domain.Consumption, periods []period, kind domain.KindPeriod, mc *domain.MeterConsumption) {
	for _, rec := range records {
		idx := findPeriodIndex(rec.Timestamp, periods)
		if idx < 0 {
			continue
		}
		mc.Active[idx] += rec.ActiveEnergy
		mc.ReactiveInductive[idx] += rec.ReactiveInductive
		mc.ReactiveCapacitive[idx] += rec.ReactiveCapacitive
		mc.Exported[idx] += rec.ExportedEnergy
	}
}

// findPeriodIndex returns the index of the period that contains the given timestamp.
func findPeriodIndex(t time.Time, periods []period) int {
	for i, p := range periods {
		if (t.Equal(p.start) || t.After(p.start)) && (t.Equal(p.end) || t.Before(p.end)) {
			return i
		}
	}
	return -1
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
