package service

import (
	"encoding/json"
	"log"

	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/DoDuy2004/slack-clone/backend/internal/websocket"
	"github.com/google/uuid"
)

type PresenceService interface {
	SetOnline(userID uuid.UUID) error
	SetOffline(userID uuid.UUID) error
	UpdateCustomStatus(userID uuid.UUID, status string) error
}

type presenceService struct {
	userRepo repository.UserRepository
	hub      *websocket.Hub
}

func NewPresenceService(userRepo repository.UserRepository, hub *websocket.Hub) PresenceService {
	return &presenceService{
		userRepo: userRepo,
		hub:      hub,
	}
}

func (s *presenceService) SetOnline(userID uuid.UUID) error {
	if err := s.userRepo.UpdateStatus(userID, "online"); err != nil {
		return err
	}

	s.broadcastPresence(userID, "online")
	return nil
}

func (s *presenceService) SetOffline(userID uuid.UUID) error {
	if err := s.userRepo.UpdateStatus(userID, "offline"); err != nil {
		return err
	}

	s.broadcastPresence(userID, "offline")
	return nil
}

func (s *presenceService) UpdateCustomStatus(userID uuid.UUID, status string) error {
	// For away or other custom states
	if err := s.userRepo.UpdateStatus(userID, status); err != nil {
		return err
	}

	s.broadcastPresence(userID, status)
	return nil
}

func (s *presenceService) broadcastPresence(userID uuid.UUID, status string) {
	payload, err := json.Marshal(websocket.PresencePayload{
		UserID: userID,
		Status: status,
	})
	if err != nil {
		log.Printf("error marshaling presence payload: %v", err)
		return
	}

	s.hub.Broadcast(&websocket.WSMessage{
		Type:    websocket.EventUserPresence,
		Payload: payload,
		UserID:  &userID, // Optional: tag with user ID
	})
}
