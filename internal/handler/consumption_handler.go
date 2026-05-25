package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bia-energy/consumption-service/internal/domain"
	"github.com/bia-energy/consumption-service/internal/service"
)

const dateLayout = "2006-01-02"

// ConsumptionHandler handles HTTP requests for consumption data.
type ConsumptionHandler struct {
	svc *service.ConsumptionService
}

// NewConsumptionHandler creates a new ConsumptionHandler.
func NewConsumptionHandler(svc *service.ConsumptionService) *ConsumptionHandler {
	return &ConsumptionHandler{svc: svc}
}

// RegisterRoutes attaches routes to the given ServeMux.
func (h *ConsumptionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/consumption", h.GetConsumption)
}

// GetConsumption returns aggregated energy consumption for one or more meters.
// GET /consumption?meters_ids=1,2&start_date=2023-06-01&end_date=2023-06-30&kind_period=monthly
func (h *ConsumptionHandler) GetConsumption(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	q := r.URL.Query()

	meterIDs, err := parseMeterIDs(q.Get("meters_ids"))
	if err != nil || len(meterIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or missing meters_ids"})
		return
	}

	startDate, err := time.Parse(dateLayout, q.Get("start_date"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid start_date, expected YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse(dateLayout, q.Get("end_date"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid end_date, expected YYYY-MM-DD"})
		return
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	kindPeriod := domain.KindPeriod(q.Get("kind_period"))
	if !isValidKindPeriod(kindPeriod) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind_period must be daily, weekly, or monthly"})
		return
	}

	query := domain.ConsumptionQuery{
		MeterIDs:  meterIDs,
		StartDate: startDate,
		EndDate:   endDate,
		Period:    kindPeriod,
	}

	resp, err := h.svc.GetConsumption(r.Context(), query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// parseMeterIDs parses a comma-separated string of meter IDs into a slice of int64.
func parseMeterIDs(raw string) ([]int64, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func isValidKindPeriod(k domain.KindPeriod) bool {
	return k == domain.KindPeriodDaily || k == domain.KindPeriodWeekly || k == domain.KindPeriodMonthly
}
