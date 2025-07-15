package database

import (
	"database/sql"
	"fmt"
	"time"

	"chchma.com/cloudlite-sync/internal/models"
)

// CreateCredential 创建凭证
func (db *DB) CreateCredential(credential *models.Credential) error {
	query := `INSERT INTO credentials (id, project_id, token, is_active, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, credential.ID, credential.ProjectID, credential.Token, credential.IsActive, now, now)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	credential.CreatedAt = now
	credential.UpdatedAt = now
	return nil
}

// GetCredentialByToken 通过token获取凭证
func (db *DB) GetCredentialByToken(token string) (*models.Credential, error) {
	query := `SELECT id, project_id, token, is_active, created_at, updated_at 
			  FROM credentials WHERE token = ? AND is_active = 1`

	credential := &models.Credential{}
	err := db.QueryRow(query, token).Scan(
		&credential.ID,
		&credential.ProjectID,
		&credential.Token,
		&credential.IsActive,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	return credential, nil
}

// GetCredential 获取凭证
func (db *DB) GetCredential(id string) (*models.Credential, error) {
	query := `SELECT id, project_id, token, is_active, created_at, updated_at 
			  FROM credentials WHERE id = ?`

	credential := &models.Credential{}
	err := db.QueryRow(query, id).Scan(
		&credential.ID,
		&credential.ProjectID,
		&credential.Token,
		&credential.IsActive,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	return credential, nil
}

// ListCredentials 获取凭证列表（分页）
func (db *DB) ListCredentials(projectID string, page, pageSize int) ([]*models.Credential, int, error) {
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM credentials WHERE project_id = ?`
	err := db.QueryRow(countQuery, projectID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count credentials: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := `SELECT id, project_id, token, is_active, created_at, updated_at 
			  FROM credentials WHERE project_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := db.Query(query, projectID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.Credential
	for rows.Next() {
		credential := &models.Credential{}
		err := rows.Scan(
			&credential.ID,
			&credential.ProjectID,
			&credential.Token,
			&credential.IsActive,
			&credential.CreatedAt,
			&credential.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, credential)
	}

	return credentials, total, nil
}

// UpdateCredential 更新凭证
func (db *DB) UpdateCredential(credential *models.Credential) error {
	query := `UPDATE credentials SET is_active = ?, updated_at = ? WHERE id = ?`

	now := time.Now()
	_, err := db.Exec(query, credential.IsActive, now, credential.ID)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	credential.UpdatedAt = now
	return nil
}

// DeleteCredential 删除凭证
func (db *DB) DeleteCredential(id string) error {
	query := `DELETE FROM credentials WHERE id = ?`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	return nil
}

// DeactivateCredential 停用凭证
func (db *DB) DeactivateCredential(id string) error {
	query := `UPDATE credentials SET is_active = 0, updated_at = ? WHERE id = ?`
	_, err := db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to deactivate credential: %w", err)
	}
	return nil
}
