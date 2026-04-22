package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
	svc "github.com/yandex-development-1-team/go/internal/service/api"
)

type ResourcePageService struct {
	repo        repository.ResourcePageRepository
	fileService *svc.FileService
}

func NewResourcePageService(repo repository.ResourcePageRepository, fs *svc.FileService) *ResourcePageService {
	return &ResourcePageService{repo: repo, fileService: fs}
}

func (s *ResourcePageService) GetResourcePage(ctx context.Context, slug string) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ResourcePageService) GetAllResourcePages(ctx context.Context) ([]models.ResourcePage, error) {
	return s.repo.GetAll(ctx)
}

func (s *ResourcePageService) UpdateResourcePage(ctx context.Context, slug string, page models.ResourcePage) (*models.ResourcePage, error) {
	updated, err := s.repo.Update(ctx, slug, page)
	if err != nil {
		return nil, fmt.Errorf("update resource page: %w", err)
	}

	return updated, nil
}

func (s *ResourcePageService) ClearResourcePage(ctx context.Context, slug string) (*models.ResourcePage, error) {
	cleared, err := s.repo.Clear(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("clear resource page: %w", err)
	}

	return cleared, nil
}

func (s *ResourcePageService) DeleteResourcePageLink(ctx context.Context, slug string, id string) (*models.ResourcePage, error) {
	page, err := s.repo.DeleteLink(ctx, slug, id)
	if err != nil {
		return nil, fmt.Errorf("delete resource page link: %w", err)
	}
	return page, nil
}

func (s *ResourcePageService) UploadFile(
	ctx context.Context,
	slug string,
	src io.ReadSeeker,
	name string,
	size int64,
) (*dto.FileUploadResponse, error) {
	if _, ok := models.AllowedSlugs[slug]; !ok {
		return nil, models.ErrInvalidSlug
	}

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read file header: %w", err)
	}
	if http.DetectContentType(buf[:n]) != "application/pdf" {
		return nil, models.ErrInvalidFileType
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}

	uploaded, err := s.fileService.Upload(ctx, src, name, "application/pdf", size)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	// берём текущую страницу чтобы не потерять title и content
	page, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get resource page: %w", err)
	}

	// деактивируем старый если есть и отличается
	if len(page.Links) > 0 && page.Links[0].URL != uploaded.URL {
		if err := s.fileService.DeactivateByURL(ctx, page.Links[0].URL); err != nil {
			return nil, fmt.Errorf("deactivate old file: %w", err)
		}
	}

	// активируем новый
	if err := s.fileService.ActivateByURL(ctx, uploaded.URL); err != nil {
		return nil, fmt.Errorf("activate new file: %w", err)
	}

	// обновляем только links, title и content берём из текущей страницы
	page.Links = []models.ResourcePageLink{{
		Title: name,
		URL:   uploaded.URL,
	}}
	if _, err := s.repo.Update(ctx, slug, *page); err != nil {
		return nil, fmt.Errorf("update resource page: %w", err)
	}

	return uploaded, nil
}
