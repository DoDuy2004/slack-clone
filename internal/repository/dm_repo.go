package repository

import (
	"database/sql"

	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/google/uuid"
)

type DMRepository interface {
	Create(dm *models.DirectMessage, participantIDs []uuid.UUID) error
	FindByParticipants(workspaceID uuid.UUID, userIDs []uuid.UUID) (*models.DirectMessage, error)
	ListByUserID(workspaceID, userID uuid.UUID) ([]*models.DirectMessage, error)
	GetByID(id uuid.UUID) (*models.DirectMessage, error)
	IsParticipant(dmID, userID uuid.UUID) (bool, error)
}

type postgresDMRepository struct {
	db *database.DB
}

func NewDMRepository(db *database.DB) DMRepository {
	return &postgresDMRepository{db: db}
}

func (r *postgresDMRepository) Create(dm *models.DirectMessage, participantIDs []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO direct_messages (id, workspace_id) VALUES ($1, $2) RETURNING created_at`
	if err := tx.QueryRow(query, dm.ID, dm.WorkspaceID).Scan(&dm.CreatedAt); err != nil {
		return err
	}

	for _, userID := range participantIDs {
		participantQuery := `INSERT INTO dm_participants (dm_id, user_id) VALUES ($1, $2)`
		if _, err := tx.Exec(participantQuery, dm.ID, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *postgresDMRepository) FindByParticipants(workspaceID uuid.UUID, userIDs []uuid.UUID) (*models.DirectMessage, error) {
	// Find a DM that has exactly these participants in this workspace
	query := `
		SELECT dm.id, dm.workspace_id, dm.created_at
		FROM direct_messages dm
		JOIN dm_participants dp1 ON dm.id = dp1.dm_id
		JOIN dm_participants dp2 ON dm.id = dp2.dm_id
		WHERE dm.workspace_id = $1
		AND dp1.user_id = $2
		AND dp2.user_id = $3
		AND (SELECT COUNT(*) FROM dm_participants WHERE dm_id = dm.id) = 2
	`
	dm := &models.DirectMessage{}
	err := r.db.QueryRow(query, workspaceID, userIDs[0], userIDs[1]).Scan(&dm.ID, &dm.WorkspaceID, &dm.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return dm, nil
}

func (r *postgresDMRepository) ListByUserID(workspaceID, userID uuid.UUID) ([]*models.DirectMessage, error) {
	query := `
		SELECT dm.id, dm.workspace_id, dm.created_at
		FROM direct_messages dm
		JOIN dm_participants dp ON dm.id = dp.dm_id
		WHERE dm.workspace_id = $1 AND dp.user_id = $2
		ORDER BY dm.created_at DESC
	`
	rows, err := r.db.Query(query, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dms []*models.DirectMessage
	for rows.Next() {
		dm := &models.DirectMessage{}
		if err := rows.Scan(&dm.ID, &dm.WorkspaceID, &dm.CreatedAt); err != nil {
			return nil, err
		}
		// Participant loading would usually happen in service or with a separate query
		dms = append(dms, dm)
	}
	return dms, nil
}

func (r *postgresDMRepository) GetByID(id uuid.UUID) (*models.DirectMessage, error) {
	dm := &models.DirectMessage{}
	query := `SELECT id, workspace_id, created_at FROM direct_messages WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&dm.ID, &dm.WorkspaceID, &dm.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return dm, nil
}

func (r *postgresDMRepository) IsParticipant(dmID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM dm_participants WHERE dm_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRow(query, dmID, userID).Scan(&exists)
	return exists, err
}
