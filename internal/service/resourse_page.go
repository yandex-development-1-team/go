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

//go:generate mockgen -source=resourse_page.go -destination=api/mocks/mock_resource_page_deps.go -package=mocks
type fileUploader interface {
	Upload(ctx context.Context, src io.Reader, name, contentType string, size int64) (*dto.FileUploadResponse, error)
	ActivateByURL(ctx context.Context, url string) error
	DeactivateByURL(ctx context.Context, url string) error
}

type ResourcePageService struct {
	repo        repository.ResourcePageRepository
	fileService fileUploader
	txRepo      repository.TxRepository
}

func NewResourcePageService(repo repository.ResourcePageRepository, fs *svc.FileService, txRepo repository.TxRepository) *ResourcePageService {
	return &ResourcePageService{
		repo:        repo,
		fileService: fs,
		txRepo:      txRepo,
	}
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

	page, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get resource page: %w", err)
	}

	err = s.txRepo.RunToTx(ctx, func(txCtx context.Context) error {
		if len(page.Links) > 0 && page.Links[0].URL != uploaded.URL {
			if err := s.fileService.DeactivateByURL(txCtx, page.Links[0].URL); err != nil {
				return fmt.Errorf("deactivate old file: %w", err)
			}
		}

		if err := s.fileService.ActivateByURL(txCtx, uploaded.URL); err != nil {
			return fmt.Errorf("activate new file: %w", err)
		}

		page.Links = []models.ResourcePageLink{{
			Title: name,
			URL:   uploaded.URL,
		}}
		if _, err := s.repo.Update(txCtx, slug, *page); err != nil {
			return fmt.Errorf("update resource page: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return uploaded, nil
}
