package mocks

import (
	"context"
	"encoding/json"
	"github.com/yandex-development-1-team/go/internal/database/repository/models"
	"os"
)

// Этот путь лучше получать из энвов
const mockLocalDir = "./internal/repository/mocks/mock.json"

type MockClient struct {
	mockLocalDir string
}

func NewMockClient(mockLocalDir string) *MockClient {
	return &MockClient{
		mockLocalDir: mockLocalDir,
	}
}

func (m MockClient) GetServices(ctx context.Context, telegramID int64) ([]models.Service, error) {
	var boxes []models.Service

	data, err := os.ReadFile(m.mockLocalDir)
	if err != nil {
		return []models.Service{}, err
	}

	err = json.Unmarshal(data, &boxes)
	if err != nil {
		return []models.Service{}, err
	}

	return boxes, nil
}
