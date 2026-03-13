package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/service/models"
)

type ResourcePageService struct {
	repo *repository.ResourcePageRepository
}

func NewResourcePageService(repo *repo.ResourcePageRepo) *ResourcePageService {
	return &ResourcePageService{repo: repo}
}

func (s *ResourcePageService) GetResourcePage(ctx context.Context, slug string) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ResourcePageService) GetResourcePagePublic(ctx context.Context, slug string) (*dto.ResourcePagePublic, error) {
	page, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err // Передаем ошибку выше (например, NotFound)
	}

	// Преобразование из models.ResourcePage в dto.ResourcePagePublic
	publicPage := &dto.ResourcePagePublic{
		Title:   page.Title,
		Content: page.Content,
		Links:   page.Links, // models.Link совпадает с форматом для публичного DTO
	}
	return publicPage, nil
}

func (s *ResourcePageService) UpdateResourcePage(ctx context.Context, slug string, updateReq *dto.ResourcePageUpdateRequest) (*models.ResourcePage, error) {
	// 1. Валидация URL в links
	if updateReq.Links != nil {
		for _, link := range *updateReq.Links {
			parsedURL, err := url.ParseRequestURI(link.URL)
			if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
				return nil, fmt.Errorf("validation error: invalid URI '%s' for link title '%s'", link.URL, link.Title)
			}
		}
	}

	// 2. Получение текущей страницы
	currentPage, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		// Если страница не найдена, возвращаем ошибку, как указано в спецификации API для PUT.
		return nil, err // Это будет обработано хендлером как 404
	}

	// 3. Мерж значений: используем новое значение, если оно предоставлено, иначе старое
	if updateReq.Title != nil {
		currentPage.Title = *updateReq.Title
	}
	if updateReq.Content != nil {
		currentPage.Content = *updateReq.Content
	}
	if updateReq.Links != nil {
		// Заменяем весь массив links новым из запроса
		currentPage.Links = *updateReq.Links
	}
	// updated_at обновится автоматически в репо при Update

	// 4. Обновление страницы (вместо Upsert)
	err = s.repo.UpdatePage(ctx, currentPage)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource page: %w", err)
	}

	// 5. Возврат обновленной страницы
	return currentPage, nil
}

func (s *ResourcePageService) GetAllSummaries(ctx context.Context) ([]*models.ResourcePageSummary, error) {
	return s.repo.GetAllSummaries(ctx)
}
