package service

import (
	"errors"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/models/dto"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
)

type ChannelService interface {
	CreateChannel(userID uuid.UUID, workspaceID uuid.UUID, req *dto.CreateChannelRequest) (*models.Channel, error)
	GetChannel(channelID uuid.UUID, userID uuid.UUID) (*models.Channel, error)
	ListWorkspaceChannels(workspaceID uuid.UUID, userID uuid.UUID) ([]*models.Channel, error)
	UpdateChannel(userID uuid.UUID, channelID uuid.UUID, req *dto.UpdateChannelRequest) (*models.Channel, error)
	DeleteChannel(userID uuid.UUID, channelID uuid.UUID) error
}

type channelService struct {
	channelRepo   repository.ChannelRepository
	workspaceRepo repository.WorkspaceRepository
}

func NewChannelService(channelRepo repository.ChannelRepository, workspaceRepo repository.WorkspaceRepository) ChannelService {
	return &channelService{
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *channelService) CreateChannel(userID uuid.UUID, workspaceID uuid.UUID, req *dto.CreateChannelRequest) (*models.Channel, error) {
	// Verify user is member of workspace
	member, err := s.workspaceRepo.GetMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	channel := &models.Channel{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: &req.Description,
		IsPrivate:   req.IsPrivate,
		CreatedBy:   &userID,
	}

	if err := s.channelRepo.Create(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

func (s *channelService) GetChannel(channelID uuid.UUID, userID uuid.UUID) (*models.Channel, error) {
	channel, err := s.channelRepo.FindByID(channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	// If private, verify membership
	if channel.IsPrivate {
		isMember, err := s.channelRepo.IsMember(channelID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
	} else {
		// If public, verify workspace membership
		isWSMember, err := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
		if err != nil {
			return nil, err
		}
		if isWSMember == nil {
			return nil, ErrUnauthorized
		}
	}

	return channel, nil
}

func (s *channelService) ListWorkspaceChannels(workspaceID uuid.UUID, userID uuid.UUID) ([]*models.Channel, error) {
	// Verify workspace membership
	member, err := s.workspaceRepo.GetMember(workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	return s.channelRepo.ListByWorkspaceID(workspaceID, userID)
}

func (s *channelService) UpdateChannel(userID uuid.UUID, channelID uuid.UUID, req *dto.UpdateChannelRequest) (*models.Channel, error) {
	channel, err := s.channelRepo.FindByID(channelID)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	// Verify permissions (only creator or workspace admin/owner can update)
	// For simplicity, we'll check if user is channel creator OR workspace admin/owner
	isWSMember, err := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
	if err != nil {
		return nil, err
	}

	canUpdate := false
	if channel.CreatedBy != nil && *channel.CreatedBy == userID {
		canUpdate = true
	} else if isWSMember != nil && (isWSMember.Role == "owner" || isWSMember.Role == "admin") {
		canUpdate = true
	}

	if !canUpdate {
		return nil, ErrUnauthorized
	}

	if req.Name != nil {
		channel.Name = *req.Name
	}
	if req.Description != nil {
		channel.Description = req.Description
	}
	if req.IsPrivate != nil {
		channel.IsPrivate = *req.IsPrivate
	}

	if err := s.channelRepo.Update(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

func (s *channelService) DeleteChannel(userID uuid.UUID, channelID uuid.UUID) error {
	channel, err := s.channelRepo.FindByID(channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// Verify permissions (only creator or workspace owner can delete)
	isWSMember, err := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
	if err != nil {
		return err
	}

	canDelete := false
	if channel.CreatedBy != nil && *channel.CreatedBy == userID {
		canDelete = true
	} else if isWSMember != nil && isWSMember.Role == "owner" {
		canDelete = true
	}

	if !canDelete {
		return ErrUnauthorized
	}

	return s.channelRepo.Delete(channelID)
}
