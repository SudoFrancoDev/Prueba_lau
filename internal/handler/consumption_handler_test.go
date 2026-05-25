package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/bia-energy/consumption-service/internal/domain"
	"github.com/bia-energy/consumption-service/internal/handler"
	"github.com/bia-energy/consumption-service/internal/service"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockRepo struct{ mock.Mock }

func (m *mockRepo) GetByMeterIDsAndDateRange(
	ctx context.Context,
	ids []int64,
	query domain.ConsumptionQuery,
) ([]domain.Consumption, error) {
	args := m.Called(ctx, ids, query)
	return args.Get(0).([]domain.Consumption), args.Error(1)
}

type mockAddr struct{ mock.Mock }

func (m *mockAddr) GetAddressByMeterID(ctx context.Context, meterID int64) (string, error) {
	args := m.Called(ctx, meterID)
	return args.String(0), args.Error(1)
}

// ---------------------------------------------------------------------------
// Test setup helper
// ---------------------------------------------------------------------------

func setupServer(repo domain.ConsumptionRepository, addr domain.AddressService) http.Handler {
	svc := service.NewConsumptionService(repo, addr)
	h := handler.NewConsumptionHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGetConsumption_BadRequest_MissingParams(t *testing.T) {
	srv := setupServer(new(mockRepo), new(mockAddr))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/consumption", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConsumption_BadRequest_InvalidKindPeriod(t *testing.T) {
	srv := setupServer(new(mockRepo), new(mockAddr))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet,
		"/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-06-30&kind_period=quarterly", nil)
	srv.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConsumption_OK_Daily(t *testing.T) {
	repo := new(mockRepo)
	addr := new(mockAddr)

	start := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 6, 1, 23, 59, 59, 0, time.UTC)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{1},
		StartDate: start,
		EndDate:   end,
		Period:    domain.KindPeriodDaily,
	}
	records := []domain.Consumption{
		{MeterID: 1, ActiveEnergy: 100, ReactiveInductive: 20,
			ReactiveCapacitive: 5, ExportedEnergy: 3,
			Timestamp: start.Add(10 * time.Hour)},
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{1}, query).Return(records, nil)
	addr.On("GetAddressByMeterID", mock.Anything, int64(1)).Return("Calle 1", nil)

	srv := setupServer(repo, addr)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet,
		"/consumption?meters_ids=1&start_date=2023-06-01&end_date=2023-06-01&kind_period=daily", nil)
	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp domain.ConsumptionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Period, 1)
	assert.Len(t, resp.DataGraph, 1)
	assert.Equal(t, int64(1), resp.DataGraph[0].MeterID)
	assert.InDelta(t, 100.0, resp.DataGraph[0].Active[0], 0.001)
}

func TestGetConsumption_OK_MultipleMeters(t *testing.T) {
	repo := new(mockRepo)
	addr := new(mockAddr)

	start := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 6, 1, 23, 59, 59, 0, time.UTC)
	query := domain.ConsumptionQuery{
		MeterIDs:  []int64{1, 2},
		StartDate: start,
		EndDate:   end,
		Period:    domain.KindPeriodDaily,
	}
	records := []domain.Consumption{
		{MeterID: 1, ActiveEnergy: 100, Timestamp: start.Add(8 * time.Hour)},
		{MeterID: 2, ActiveEnergy: 200, Timestamp: start.Add(9 * time.Hour)},
	}

	repo.On("GetByMeterIDsAndDateRange", mock.Anything, []int64{1, 2}, query).Return(records, nil)
	addr.On("GetAddressByMeterID", mock.Anything, int64(1)).Return("Calle 1", nil)
	addr.On("GetAddressByMeterID", mock.Anything, int64(2)).Return("Calle 2", nil)

	srv := setupServer(repo, addr)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet,
		"/consumption?meters_ids=1,2&start_date=2023-06-01&end_date=2023-06-01&kind_period=daily", nil)
	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp domain.ConsumptionResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.DataGraph, 2)
}
