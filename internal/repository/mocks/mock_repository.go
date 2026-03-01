package mocks

import (
	"context"
	"encoding/json"
	"github.com/yandex-development-1-team/go/internal/repository/models"
	"os"
)

// Этот путь лучше получать из энвов
const mockLocalDir = "./internal/repository/mocks/mock.json"

type MockClient struct {
	mockLocalDir string
}

func NewMockClient() *MockClient {
	return &MockClient{
		mockLocalDir: mockLocalDir,
	}
}

func (m MockClient) GetBoxSolutions(ctx context.Context) ([]models.BoxSolution, error) {
	var boxes []models.BoxSolution

	data, err := os.ReadFile(m.mockLocalDir)
	if err != nil {
		return []models.BoxSolution{}, err
	}

	err = json.Unmarshal(data, &boxes)
	if err != nil {
		return []models.BoxSolution{}, err
	}

	return boxes, nil
}
