package service

import (
	"io"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/DoDuy2004/slack-clone/backend/pkg/storage"
	"github.com/google/uuid"
)

type FileService interface {
	UploadFile(fileName string, fileSize int64, fileType string, reader io.Reader) (*models.Attachment, error)
	GetAttachment(id uuid.UUID) (*models.Attachment, error)
	LinkToMessage(attachmentID, messageID uuid.UUID) error
}

type fileService struct {
	attachmentRepo repository.AttachmentRepository
	storage        storage.Storage
}

func NewFileService(attachmentRepo repository.AttachmentRepository, storage storage.Storage) FileService {
	return &fileService{
		attachmentRepo: attachmentRepo,
		storage:        storage,
	}
}

func (s *fileService) UploadFile(fileName string, fileSize int64, fileType string, reader io.Reader) (*models.Attachment, error) {
	// 1. Save to storage
	storagePath, err := s.storage.Save(fileName, reader)
	if err != nil {
		return nil, err
	}

	// 2. Create attachment metadata
	attachment := &models.Attachment{
		ID:       uuid.New(),
		FileName: fileName,
		FileURL:  storagePath, // We store the path/name, GetURL will be called in mapping
		FileType: &fileType,
		FileSize: &fileSize,
	}

	if err := s.attachmentRepo.Create(attachment); err != nil {
		// Cleanup storage if DB fails
		s.storage.Delete(storagePath)
		return nil, err
	}

	// Update URL to full URL for the response
	attachment.FileURL = s.storage.GetURL(attachment.FileURL)

	return attachment, nil
}

func (s *fileService) GetAttachment(id uuid.UUID) (*models.Attachment, error) {
	return s.attachmentRepo.GetByID(id)
}

func (s *fileService) LinkToMessage(attachmentID, messageID uuid.UUID) error {
	return s.attachmentRepo.LinkToMessage(attachmentID, messageID)
}
