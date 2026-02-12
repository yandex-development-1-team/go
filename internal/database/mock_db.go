package database

import (
	"context"
	"encoding/json"
	"github.com/yandex-development-1-team/go/internal/repository"
	"os"
)

// Этот путь лучше получать из энвов
const mockLocalDir = "./internal/database/mocks/mock.json"

type MockClient struct {
	mockLocalDir string
}

func NewMockClient(ctx context.Context) *MockClient {
	return &MockClient{
		mockLocalDir: mockLocalDir,
	}
}

func (m MockClient) GetBoxSolutions(ctx context.Context) ([]repository.BoxSolution, error) {
	var boxes []repository.BoxSolution

	data, err := os.ReadFile(m.mockLocalDir)
	if err != nil {
		return []repository.BoxSolution{}, err
	}

	err = json.Unmarshal(data, &boxes)
	if err != nil {
		return []repository.BoxSolution{}, err
	}

	return boxes, nil
}
