package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type ResourcePageService struct {
	repo repository.ResourcePageRepository
}

func NewResourcePageService(repo repository.ResourcePageRepository) *ResourcePageService {
	return &ResourcePageService{repo: repo}
}

func (s *ResourcePageService) GetResourcePage(ctx context.Context, slug string) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ResourcePageService) UpdateResourcePage(ctx context.Context, slug string, newPageData *models.ResourcePage) (*models.ResourcePage, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("ERROR: failed to rollback transaction: %v", err)
		}
	}()

	currentPage, err := s.repo.GetBySlugTx(ctx, tx, slug, true)
	if err != nil {
		return nil, err
	}

	if newPageData.Title != "" {
		currentPage.Title = newPageData.Title
	}
	if newPageData.Content != "" {
		currentPage.Content = newPageData.Content
	}
	if newPageData.Links != nil {
		newLinksMap := make(map[string]models.ResourcePageLink, len(newPageData.Links))
		for _, link := range newPageData.Links {
			newLinksMap[link.ID] = link
		}

		updatedLinks := make([]models.ResourcePageLink, 0, len(newLinksMap)+len(newPageData.Links))

		for _, link := range currentPage.Links {
			if v, exists := newLinksMap[link.ID]; exists {
				updatedLinks = append(updatedLinks, v)
				delete(newLinksMap, link.ID)
			} else {
				updatedLinks = append(updatedLinks, link)
			}
		}

		for _, link := range newPageData.Links {
			if _, exists := newLinksMap[link.ID]; exists {
				updatedLinks = append(updatedLinks, link)
			}
		}

		currentPage.Links = updatedLinks
	}

	err = s.repo.UpdatePageContentAndLinksTx(ctx, tx, currentPage.Slug, currentPage.Title, currentPage.Content, currentPage.Links)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource page in transaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return currentPage, nil
}

func (s *ResourcePageService) DeleteLink(ctx context.Context, slug string, linkID string) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("ERROR: failed to rollback transaction: %v", err)
		}
	}()

	currentPage, err := s.repo.GetBySlugTx(ctx, tx, slug, true)
	if err != nil {
		return err
	}

	found := false
	newLinks := make([]models.ResourcePageLink, 0, len(currentPage.Links))
	for _, link := range currentPage.Links {
		if link.ID != linkID {
			newLinks = append(newLinks, link)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("link with id '%s' not found in page '%s'", linkID, slug)
	}

	currentPage.Links = newLinks
	err = s.repo.UpdatePageContentAndLinksTx(ctx, tx, currentPage.Slug, currentPage.Title, currentPage.Content, currentPage.Links)
	if err != nil {
		return fmt.Errorf("failed to update page after deleting link: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *ResourcePageService) GetAllSummaries(ctx context.Context) ([]*models.ResourcePage, error) {
	return s.repo.GetAllSummaries(ctx)
}
