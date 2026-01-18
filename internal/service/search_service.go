package service

import (
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/google/uuid"
)

type SearchService interface {
	SearchMessages(userID, workspaceID uuid.UUID, query string, limit, offset int) ([]*models.Message, error)
}

type searchService struct {
	messageRepo   repository.MessageRepository
	workspaceRepo repository.WorkspaceRepository
}

func NewSearchService(messageRepo repository.MessageRepository, workspaceRepo repository.WorkspaceRepository) SearchService {
	return &searchService{
		messageRepo:   messageRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *searchService) SearchMessages(userID, workspaceID uuid.UUID, query string, limit, offset int) ([]*models.Message, error) {
	// 1. Verify workspace membership
	member, err := s.workspaceRepo.GetMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// 2. Perform search
	return s.messageRepo.Search(workspaceID, query, limit, offset)
}
