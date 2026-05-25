package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bia-energy/consumption-service/internal/domain"
	"github.com/bia-energy/consumption-service/internal/service"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockConsumptionRepo struct{ mock.Mock }

func (m *mockConsumptionRepo) GetByMeterIDsAndDateRange(
	ctx context.Context,
	meterIDs []int64,
	query domain.ConsumptionQuery,
) ([]domain.Consumption, error) {
	args := m.Called(ctx, meterIDs, query)
	return args.Get(0).([]domain.Consumption), args.Error(1)
}

type mockAddressSvc struct{ mock.Mock }

func (m *mockAddressSvc) GetAddressByMeterID(ctx context.Context, meterID int64) (string, error) {
	args := m.Called(ctx, meterID)
	return args.String(0), args.Error(1)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func date(y int, mo time.Month, d int) time.Time {
	return time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)
}

func consumptionAt(meterID int64, t time.Time, active, ri, rc, exp float64) domain.Consumption {
	return domain.Consumption{
		MeterID:            meterID,
		ActiveEnergy:       active,
		ReactiveInductive:  ri,
		ReactiveCapacitive: rc,
		ExportedEnergy:     exp,
		Timestamp:          t,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGetConsumption_Daily(t *testing.T) {
	repo := new(mockConsumptionRepo)
	addrSvc := new(mockAddressSvc)
	svc := service.NewConsumptionService(repo, addrSvc)

	start := date(2023, 6, 1)
	end := date(2023, 6, 3)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{1},
		StartDate: start,
		EndDate:   end.Add(24*time.Hour - time.Second),
		Period:    domain.KindPeriodDaily,
	}

	records := []domain.Consumption{
		consumptionAt(1, date(2023, 6, 1).Add(10*time.Hour), 100, 20, 5, 3),
		consumptionAt(1, date(2023, 6, 1).Add(20*time.Hour), 50, 10, 2, 1),
		consumptionAt(1, date(2023, 6, 2).Add(8*time.Hour), 200, 30, 8, 4),
		consumptionAt(1, date(2023, 6, 3).Add(12*time.Hour), 75, 15, 3, 2),
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{1}, query).Return(records, nil)
	addrSvc.On("GetAddressByMeterID", mock.Anything, int64(1)).Return("Calle 1 # 2-3", nil)

	resp, err := svc.GetConsumption(context.Background(), query)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Jun 1", "Jun 2", "Jun 3"}, resp.Period)
	assert.Len(t, resp.DataGraph, 1)

	mc := resp.DataGraph[0]
	assert.Equal(t, int64(1), mc.MeterID)
	assert.Equal(t, "Calle 1 # 2-3", mc.Address)
	assert.InDelta(t, 150.0, mc.Active[0], 0.001) // Jun 1: 100+50
	assert.InDelta(t, 200.0, mc.Active[1], 0.001) // Jun 2
	assert.InDelta(t, 75.0, mc.Active[2], 0.001)  // Jun 3

	repo.AssertExpectations(t)
	addrSvc.AssertExpectations(t)
}

func TestGetConsumption_Weekly(t *testing.T) {
	repo := new(mockConsumptionRepo)
	addrSvc := new(mockAddressSvc)
	svc := service.NewConsumptionService(repo, addrSvc)

	start := date(2023, 6, 1)
	end := date(2023, 6, 14)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{1},
		StartDate: start,
		EndDate:   end.Add(24*time.Hour - time.Second),
		Period:    domain.KindPeriodWeekly,
	}

	records := []domain.Consumption{
		// Week 1: Jun 1–7
		consumptionAt(1, date(2023, 6, 3).Add(10*time.Hour), 300, 50, 10, 5),
		// Week 2: Jun 8–14
		consumptionAt(1, date(2023, 6, 10).Add(10*time.Hour), 400, 60, 12, 6),
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{1}, query).Return(records, nil)
	addrSvc.On("GetAddressByMeterID", mock.Anything, int64(1)).Return("Av Principal", nil)

	resp, err := svc.GetConsumption(context.Background(), query)

	assert.NoError(t, err)
	assert.Len(t, resp.Period, 2)
	assert.InDelta(t, 300.0, resp.DataGraph[0].Active[0], 0.001)
	assert.InDelta(t, 400.0, resp.DataGraph[0].Active[1], 0.001)

	repo.AssertExpectations(t)
}

func TestGetConsumption_Monthly(t *testing.T) {
	repo := new(mockConsumptionRepo)
	addrSvc := new(mockAddressSvc)
	svc := service.NewConsumptionService(repo, addrSvc)

	start := date(2023, 6, 1)
	end := date(2023, 7, 31)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{1},
		StartDate: start,
		EndDate:   end.Add(24*time.Hour - time.Second),
		Period:    domain.KindPeriodMonthly,
	}

	records := []domain.Consumption{
		consumptionAt(1, date(2023, 6, 15).Add(12*time.Hour), 500, 80, 15, 7),
		consumptionAt(1, date(2023, 7, 20).Add(12*time.Hour), 600, 90, 18, 9),
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{1}, query).Return(records, nil)
	addrSvc.On("GetAddressByMeterID", mock.Anything, int64(1)).Return("Carrera 50", nil)

	resp, err := svc.GetConsumption(context.Background(), query)

	assert.NoError(t, err)
	assert.Equal(t, []string{"Jun 2023", "Jul 2023"}, resp.Period)
	assert.InDelta(t, 500.0, resp.DataGraph[0].Active[0], 0.001)
	assert.InDelta(t, 600.0, resp.DataGraph[0].Active[1], 0.001)

	repo.AssertExpectations(t)
}

func TestGetConsumption_MultipleMeters(t *testing.T) {
	repo := new(mockConsumptionRepo)
	addrSvc := new(mockAddressSvc)
	svc := service.NewConsumptionService(repo, addrSvc)

	start := date(2023, 6, 1)
	end := date(2023, 6, 1)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{1, 2},
		StartDate: start,
		EndDate:   end.Add(24*time.Hour - time.Second),
		Period:    domain.KindPeriodDaily,
	}

	records := []domain.Consumption{
		consumptionAt(1, date(2023, 6, 1).Add(10*time.Hour), 100, 20, 5, 3),
		consumptionAt(2, date(2023, 6, 1).Add(11*time.Hour), 200, 40, 10, 6),
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{1, 2}, query).Return(records, nil)
	addrSvc.On("GetAddressByMeterID", mock.Anything, int64(1)).Return("Calle 1", nil)
	addrSvc.On("GetAddressByMeterID", mock.Anything, int64(2)).Return("Calle 2", nil)

	resp, err := svc.GetConsumption(context.Background(), query)

	assert.NoError(t, err)
	assert.Len(t, resp.DataGraph, 2)
	assert.Equal(t, int64(1), resp.DataGraph[0].MeterID)
	assert.Equal(t, int64(2), resp.DataGraph[1].MeterID)
	assert.InDelta(t, 100.0, resp.DataGraph[0].Active[0], 0.001)
	assert.InDelta(t, 200.0, resp.DataGraph[1].Active[0], 0.001)

	repo.AssertExpectations(t)
}

func TestGetConsumption_EmptyRecords(t *testing.T) {
	repo := new(mockConsumptionRepo)
	addrSvc := new(mockAddressSvc)
	svc := service.NewConsumptionService(repo, addrSvc)

	start := date(2023, 6, 1)
	end := date(2023, 6, 1)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{99},
		StartDate: start,
		EndDate:   end.Add(24*time.Hour - time.Second),
		Period:    domain.KindPeriodDaily,
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{99}, query).Return([]domain.Consumption{}, nil)
	addrSvc.On("GetAddressByMeterID", mock.Anything, int64(99)).Return("Dirección medidor 99", nil)

	resp, err := svc.GetConsumption(context.Background(), query)

	assert.NoError(t, err)
	assert.Len(t, resp.DataGraph, 1)
	assert.Equal(t, 0.0, resp.DataGraph[0].Active[0]) // No data → zero values

	repo.AssertExpectations(t)
}
