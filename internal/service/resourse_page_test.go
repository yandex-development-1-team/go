package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service/api/mocks"
)

func TestResourcePageService_UploadFile(t *testing.T) {
	validSlug := "spec-projects"
	pdfHeader := makePDFReader()

	t.Run("invalid file type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockResourcePageRepository(ctrl)
		tx := mocks.NewMockTxRepository(ctrl)
		fs := mocks.NewMockfileUploader(ctrl)

		svc := &ResourcePageService{repo: repo, fileService: fs, txRepo: tx}

		src := bytes.NewReader([]byte("not a pdf"))
		_, err := svc.UploadFile(context.Background(), validSlug, src, "file.pdf", 9)

		assert.ErrorIs(t, err, models.ErrInvalidFileType)
	})

	t.Run("upload without old file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockResourcePageRepository(ctrl)
		tx := mocks.NewMockTxRepository(ctrl)
		fs := mocks.NewMockfileUploader(ctrl)

		uploaded := &dto.FileUploadResponse{UUID: "uuid1", URL: "http://s3/new.pdf", Filename: "file.pdf"}

		fs.EXPECT().Upload(gomock.Any(), gomock.Any(), "file.pdf", "application/pdf", gomock.Any()).Return(uploaded, nil)
		repo.EXPECT().GetBySlug(gomock.Any(), validSlug).Return(&models.ResourcePage{Links: nil}, nil)
		tx.EXPECT().RunToTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		fs.EXPECT().ActivateByURL(gomock.Any(), uploaded.URL).Return(nil)
		repo.EXPECT().Update(gomock.Any(), validSlug, gomock.Any()).Return(&models.ResourcePage{}, nil)

		svc := &ResourcePageService{repo: repo, fileService: fs, txRepo: tx}
		result, err := svc.UploadFile(context.Background(), validSlug, pdfHeader(), "file.pdf", 1024)

		assert.NoError(t, err)
		assert.Equal(t, uploaded, result)
	})

	t.Run("upload replaces old file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockResourcePageRepository(ctrl)
		tx := mocks.NewMockTxRepository(ctrl)
		fs := mocks.NewMockfileUploader(ctrl)

		oldURL := "http://s3/old.pdf"
		uploaded := &dto.FileUploadResponse{UUID: "uuid2", URL: "http://s3/new.pdf", Filename: "file.pdf"}

		fs.EXPECT().Upload(gomock.Any(), gomock.Any(), "file.pdf", "application/pdf", gomock.Any()).Return(uploaded, nil)
		repo.EXPECT().GetBySlug(gomock.Any(), validSlug).Return(&models.ResourcePage{
			Links: []models.ResourcePageLink{{URL: oldURL}},
		}, nil)
		tx.EXPECT().RunToTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		fs.EXPECT().DeactivateByURL(gomock.Any(), oldURL).Return(nil)
		fs.EXPECT().ActivateByURL(gomock.Any(), uploaded.URL).Return(nil)
		repo.EXPECT().Update(gomock.Any(), validSlug, gomock.Any()).Return(&models.ResourcePage{}, nil)

		svc := &ResourcePageService{repo: repo, fileService: fs, txRepo: tx}
		result, err := svc.UploadFile(context.Background(), validSlug, pdfHeader(), "file.pdf", 1024)

		assert.NoError(t, err)
		assert.Equal(t, uploaded, result)
	})

	t.Run("same url - no deactivation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockResourcePageRepository(ctrl)
		tx := mocks.NewMockTxRepository(ctrl)
		fs := mocks.NewMockfileUploader(ctrl)

		sameURL := "http://s3/same.pdf"
		uploaded := &dto.FileUploadResponse{UUID: "uuid3", URL: sameURL, Filename: "file.pdf"}

		fs.EXPECT().Upload(gomock.Any(), gomock.Any(), "file.pdf", "application/pdf", gomock.Any()).Return(uploaded, nil)
		repo.EXPECT().GetBySlug(gomock.Any(), validSlug).Return(&models.ResourcePage{
			Links: []models.ResourcePageLink{{URL: sameURL}},
		}, nil)
		tx.EXPECT().RunToTx(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			},
		)
		// DeactivateByURL не должен вызваться
		fs.EXPECT().ActivateByURL(gomock.Any(), sameURL).Return(nil)
		repo.EXPECT().Update(gomock.Any(), validSlug, gomock.Any()).Return(&models.ResourcePage{}, nil)

		svc := &ResourcePageService{repo: repo, fileService: fs, txRepo: tx}
		result, err := svc.UploadFile(context.Background(), validSlug, pdfHeader(), "file.pdf", 1024)

		assert.NoError(t, err)
		assert.Equal(t, uploaded, result)
	})
}

// makePDFReader возвращает фабрику ридеров с валидным PDF-заголовком
func makePDFReader() func() *bytes.Reader {
	return func() *bytes.Reader {
		buf := make([]byte, 512)
		copy(buf, []byte("%PDF-"))
		return bytes.NewReader(buf)
	}
}
