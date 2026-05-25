package domain

import "context"

// ConsumptionRepository defines the contract for consumption data access.
type ConsumptionRepository interface {
	GetByMeterIDsAndDateRange(ctx context.Context, meterIDs []int64, query ConsumptionQuery) ([]Consumption, error)
}

// AddressService defines the contract for the external address microservice.
type AddressService interface {
	GetAddressByMeterID(ctx context.Context, meterID int64) (string, error)
}
