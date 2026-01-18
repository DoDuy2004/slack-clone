package service

import (
	"errors"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrReactionAlreadyExists = errors.New("reaction already exists")
	ErrReactionNotFound      = errors.New("reaction not found")
)

type ReactionService interface {
	AddReaction(userID, messageID uuid.UUID, emoji string) (*models.Reaction, *models.Message, error)
	RemoveReaction(userID, messageID uuid.UUID, emoji string) (*models.Message, error)
	GetReactions(messageID uuid.UUID) ([]*models.Reaction, error)
}

type reactionService struct {
	reactionRepo  repository.ReactionRepository
	messageRepo   repository.MessageRepository
	channelRepo   repository.ChannelRepository
	dmRepo        repository.DMRepository
	workspaceRepo repository.WorkspaceRepository
}

func NewReactionService(
	reactionRepo repository.ReactionRepository,
	messageRepo repository.MessageRepository,
	channelRepo repository.ChannelRepository,
	dmRepo repository.DMRepository,
	workspaceRepo repository.WorkspaceRepository,
) ReactionService {
	return &reactionService{
		reactionRepo:  reactionRepo,
		messageRepo:   messageRepo,
		channelRepo:   channelRepo,
		dmRepo:        dmRepo,
		workspaceRepo: workspaceRepo,
	}
}

func (s *reactionService) AddReaction(userID, messageID uuid.UUID, emoji string) (*models.Reaction, *models.Message, error) {
	// 1. Verify message exists
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, nil, err
	}
	if message == nil {
		return nil, nil, ErrMessageNotFound
	}

	// 2. Verify access
	if err := s.verifyAccess(userID, message); err != nil {
		return nil, nil, err
	}

	// 3. Check if already exists
	existing, err := s.reactionRepo.GetByMessageUserEmoji(messageID, userID, emoji)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, ErrReactionAlreadyExists
	}

	// 4. Create reaction
	reaction := &models.Reaction{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	if err := s.reactionRepo.Add(reaction); err != nil {
		return nil, nil, err
	}

	return reaction, message, nil
}

func (s *reactionService) RemoveReaction(userID, messageID uuid.UUID, emoji string) (*models.Message, error) {
	// 1. Verify message exists
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, err
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// 2. Verify access
	if err := s.verifyAccess(userID, message); err != nil {
		return nil, err
	}

	// 3. Remove
	if err := s.reactionRepo.Remove(messageID, userID, emoji); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *reactionService) GetReactions(messageID uuid.UUID) ([]*models.Reaction, error) {
	return s.reactionRepo.ListByMessageID(messageID)
}

func (s *reactionService) verifyAccess(userID uuid.UUID, message *models.Message) error {
	if message.ChannelID != nil {
		// Check channel access
		isMember, err := s.channelRepo.IsMember(*message.ChannelID, userID)
		if err != nil {
			return err
		}
		if !isMember {
			channel, err := s.channelRepo.FindByID(*message.ChannelID)
			if err != nil {
				return err
			}
			if channel == nil {
				return ErrChannelNotFound
			}
			if channel.IsPrivate {
				return ErrUnauthorized
			}
			// Check workspace membership for public channels
			wsMember, err := s.workspaceRepo.GetMember(channel.WorkspaceID, userID)
			if err != nil {
				return err
			}
			if wsMember == nil {
				return ErrUnauthorized
			}
		}
		return nil
	}

	if message.DMID != nil {
		// Check DM access
		isParticipant, err := s.dmRepo.IsParticipant(*message.DMID, userID)
		if err != nil {
			return err
		}
		if !isParticipant {
			return ErrUnauthorized
		}
		return nil
	}

	return ErrUnauthorized
}
