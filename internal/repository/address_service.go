package repository

import (
	"context"
	"fmt"

	"github.com/bia-energy/consumption-service/internal/domain"
)

// mockAddressService returns hardcoded addresses per meter ID.
// Replace with a real HTTP client when the address microservice is available.
type mockAddressService struct{}

// NewMockAddressService creates a mock implementation of the AddressService.
func NewMockAddressService() domain.AddressService {
	return &mockAddressService{}
}

func (m *mockAddressService) GetAddressByMeterID(_ context.Context, meterID int64) (string, error) {
	addresses := map[int64]string{
		1: "Calle 123 # 45-67, Bogotá",
		2: "Carrera 80 # 12-34, Medellín",
		3: "Av. El Dorado # 89-10, Bogotá",
	}
	if addr, ok := addresses[meterID]; ok {
		return addr, nil
	}
	return fmt.Sprintf("Dirección medidor %d", meterID), nil
}
