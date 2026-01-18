package service

import (
	"errors"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/models/dto"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

type MessageService interface {
	GetChannelMessages(userID uuid.UUID, channelID uuid.UUID, limit, offset int) ([]*models.Message, error)
	GetDMMessages(userID uuid.UUID, dmID uuid.UUID, limit, offset int) ([]*models.Message, error)
	GetThreads(userID uuid.UUID, parentID uuid.UUID) ([]*models.Message, error)
	UpdateMessage(userID uuid.UUID, messageID uuid.UUID, req *dto.UpdateMessageRequest) (*models.Message, error)
	DeleteMessage(userID uuid.UUID, messageID uuid.UUID) error
	SendChannelMessage(userID uuid.UUID, channelID uuid.UUID, req *dto.CreateMessageRequest) (*models.Message, error)
	SendDMMessage(userID uuid.UUID, dmID uuid.UUID, req *dto.CreateMessageRequest) (*models.Message, error)
}

type messageService struct {
	messageRepo   repository.MessageRepository
	channelRepo   repository.ChannelRepository
	workspaceRepo repository.WorkspaceRepository
	dmRepo        repository.DMRepository
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	channelRepo repository.ChannelRepository,
	workspaceRepo repository.WorkspaceRepository,
	dmRepo repository.DMRepository,
) MessageService {
	return &messageService{
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		workspaceRepo: workspaceRepo,
		dmRepo:        dmRepo,
	}
}

func (s *messageService) SendChannelMessage(userID uuid.UUID, channelID uuid.UUID, req *dto.CreateMessageRequest) (*models.Message, error) {
	// Verify channel membership
	isMember, err := s.channelRepo.IsMember(channelID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		// If not direct member, check if it's a public channel and user is in workspace
		channel, err := s.channelRepo.FindByID(channelID)
		if err != nil {
			return nil, err
		}
		if channel == nil {
			return nil, ErrChannelNotFound
		}

		if channel.IsPrivate {
			return nil, ErrUnauthorized
		}

		// Check workspace membership
		wsMember, err := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
		if err != nil {
			return nil, err
		}
		if wsMember == nil {
			return nil, ErrUnauthorized
		}
	}

	// If it's a reply, verify parent exists and belongs to the same channel
	if req.ParentMessageID != nil {
		parent, err := s.messageRepo.FindByID(*req.ParentMessageID)
		if err != nil {
			return nil, err
		}
		if parent == nil || parent.ChannelID == nil || *parent.ChannelID != channelID {
			return nil, errors.New("invalid parent message")
		}
	}

	message := &models.Message{
		ID:              uuid.New(),
		Content:         req.Content,
		SenderID:        &userID,
		ChannelID:       &channelID,
		ParentMessageID: req.ParentMessageID,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *messageService) GetChannelMessages(userID uuid.UUID, channelID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	// Verify access
	isMember, err := s.channelRepo.IsMember(channelID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		channel, err := s.channelRepo.FindByID(channelID)
		if err != nil {
			return nil, err
		}
		if channel == nil {
			return nil, ErrChannelNotFound
		}

		if channel.IsPrivate {
			return nil, ErrUnauthorized
		}

		wsMember, err := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
		if err != nil {
			return nil, err
		}
		if wsMember == nil {
			return nil, ErrUnauthorized
		}
	}

	return s.messageRepo.ListByChannelID(channelID, limit, offset)
}

func (s *messageService) SendDMMessage(userID uuid.UUID, dmID uuid.UUID, req *dto.CreateMessageRequest) (*models.Message, error) {
	// 1. Verify user is participant in DM
	isParticipant, err := s.dmRepo.IsParticipant(dmID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrUnauthorized
	}

	message := &models.Message{
		ID:              uuid.New(),
		Content:         req.Content,
		SenderID:        &userID,
		DMID:            &dmID,
		ParentMessageID: req.ParentMessageID,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *messageService) GetDMMessages(userID uuid.UUID, dmID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	// 1. Verify user is participant
	isParticipant, err := s.dmRepo.IsParticipant(dmID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrUnauthorized
	}

	return s.messageRepo.ListByDMID(dmID, limit, offset)
}

func (s *messageService) GetThreads(userID uuid.UUID, parentID uuid.UUID) ([]*models.Message, error) {
	parent, err := s.messageRepo.FindByID(parentID)
	if err != nil {
		return nil, err
	}
	if parent == nil || parent.ChannelID == nil {
		return nil, ErrMessageNotFound
	}

	// Verify access to channel
	_, err = s.GetChannelMessages(userID, *parent.ChannelID, 1, 0)
	if err != nil {
		return nil, err
	}

	return s.messageRepo.ListReplies(parentID)
}

func (s *messageService) UpdateMessage(userID uuid.UUID, messageID uuid.UUID, req *dto.UpdateMessageRequest) (*models.Message, error) {
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// Only sender can update
	if message.SenderID == nil || *message.SenderID != userID {
		return nil, ErrUnauthorized
	}

	message.Content = req.Content
	if err := s.messageRepo.Update(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *messageService) DeleteMessage(userID uuid.UUID, messageID uuid.UUID) error {
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// Check permissions
	// Only sender can delete, OR workspace owner/admin (for management)
	isOwner := false
	if message.SenderID != nil && *message.SenderID == userID {
		isOwner = true
	} else if message.ChannelID != nil {
		channel, _ := s.channelRepo.FindByID(*message.ChannelID)
		if channel != nil {
			wsMember, _ := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
			if wsMember != nil && (wsMember.Role == "owner" || wsMember.Role == "admin") {
				isOwner = true
			}
		}
	}

	if !isOwner {
		return ErrUnauthorized
	}

	return s.messageRepo.SoftDelete(messageID)
}
