package database

import (
	"context"
	"encoding/json"
	"github.com/yandex-development-1-team/go/internal/database/db_models"
	"os"
)

const mockLocalDir = "./internal/database/mocks/mock.json"

type MockClient struct {
	mockLocalDir string
}

func NewMockClient(ctx context.Context) *MockClient {
	return &MockClient{
		mockLocalDir: mockLocalDir,
	}
}

func (m MockClient) GetBoxes(ctx context.Context) ([]db_models.Box, error) {
	var boxes []db_models.Box

	data, err := os.ReadFile(m.mockLocalDir)
	if err != nil {
		return []db_models.Box{}, err
	}

	err = json.Unmarshal(data, &boxes)
	if err != nil {
		return []db_models.Box{}, err
	}

	return boxes, nil
}
