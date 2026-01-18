package service

import (
	"errors"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrDMNotFound = errors.New("dm session not found")
)

type DMService interface {
	GetOrCreateDM(userID uuid.UUID, workspaceID uuid.UUID, recipientID uuid.UUID) (*models.DirectMessage, error)
	ListUserDMs(userID, workspaceID uuid.UUID) ([]*models.DirectMessage, error)
}

type dmService struct {
	dmRepo        repository.DMRepository
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
}

func NewDMService(
	dmRepo repository.DMRepository,
	workspaceRepo repository.WorkspaceRepository,
	userRepo repository.UserRepository,
) DMService {
	return &dmService{
		dmRepo:        dmRepo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (s *dmService) GetOrCreateDM(userID uuid.UUID, workspaceID uuid.UUID, recipientID uuid.UUID) (*models.DirectMessage, error) {
	if userID == recipientID {
		return nil, errors.New("cannot create DM with yourself")
	}

	// 1. Verify both users are in the same workspace
	senderMember, err := s.workspaceRepo.GetMember(workspaceID, userID)
	if err != nil || senderMember == nil {
		return nil, ErrUnauthorized
	}

	recipientMember, err := s.workspaceRepo.GetMember(workspaceID, recipientID)
	if err != nil || recipientMember == nil {
		return nil, errors.New("recipient is not a member of this workspace")
	}

	// 2. Check if DM already exists
	participants := []uuid.UUID{userID, recipientID}
	dm, err := s.dmRepo.FindByParticipants(workspaceID, participants)
	if err != nil {
		return nil, err
	}

	if dm != nil {
		return dm, nil
	}

	// 3. Create new DM
	newDM := &models.DirectMessage{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
	}

	if err := s.dmRepo.Create(newDM, participants); err != nil {
		return nil, err
	}

	return newDM, nil
}

func (s *dmService) ListUserDMs(userID, workspaceID uuid.UUID) ([]*models.DirectMessage, error) {
	// Verify workspace membership
	member, err := s.workspaceRepo.GetMember(workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrUnauthorized
	}

	return s.dmRepo.ListByUserID(workspaceID, userID)
}
