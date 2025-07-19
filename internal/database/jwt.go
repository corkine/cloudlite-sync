package database

import (
	"math/rand"
	"time"

	"chchma.com/cloudlite-sync/internal/models"
)

// JWTProject 相关操作

func (db *DB) CreateJWTProject(project *models.JWTProject) error {
	query := `INSERT INTO jwt_projects (id, name, description, public_key, private_key, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, project.ID, project.Name, project.Description,
		project.PublicKey, project.PrivateKey, now, now)
	return err
}

func (db *DB) GetJWTProject(id string) (*models.JWTProject, error) {
	query := `SELECT id, name, description, public_key, private_key, created_at, updated_at 
			  FROM jwt_projects WHERE id = ?`

	project := &models.JWTProject{}
	err := db.QueryRow(query, id).Scan(
		&project.ID, &project.Name, &project.Description,
		&project.PublicKey, &project.PrivateKey, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (db *DB) ListJWTProjects() ([]*models.JWTProject, error) {
	query := `SELECT id, name, description, public_key, private_key, created_at, updated_at 
			  FROM jwt_projects ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.JWTProject
	for rows.Next() {
		project := &models.JWTProject{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description,
			&project.PublicKey, &project.PrivateKey, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (db *DB) UpdateJWTProject(project *models.JWTProject) error {
	query := `UPDATE jwt_projects SET name = ?, description = ?, public_key = ?, private_key = ?, updated_at = ? 
			  WHERE id = ?`

	now := time.Now()
	_, err := db.Exec(query, project.Name, project.Description,
		project.PublicKey, project.PrivateKey, now, project.ID)
	return err
}

func (db *DB) DeleteJWTProject(id string) error {
	// 先删除相关的tokens
	_, err := db.Exec("DELETE FROM jwt_tokens WHERE project_id = ?", id)
	if err != nil {
		return err
	}

	// 再删除项目
	_, err = db.Exec("DELETE FROM jwt_projects WHERE id = ?", id)
	return err
}

// JWTToken 相关操作

func (db *DB) CreateJWTToken(token *models.JWTToken) error {
	query := `INSERT INTO jwt_tokens (id, project_id, purpose, username, role, token, is_active, expires_at, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, token.ID, token.ProjectID, token.Purpose, token.Username,
		token.Role, token.Token, token.IsActive, token.ExpiresAt, now, now)
	return err
}

func (db *DB) GetJWTToken(id string) (*models.JWTToken, error) {
	query := `SELECT id, project_id, purpose, username, role, token, is_active, expires_at, created_at, updated_at 
			  FROM jwt_tokens WHERE id = ?`

	token := &models.JWTToken{}
	err := db.QueryRow(query, id).Scan(
		&token.ID, &token.ProjectID, &token.Purpose, &token.Username, &token.Role,
		&token.Token, &token.IsActive, &token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (db *DB) ListJWTTokens(projectID string) ([]*models.JWTToken, error) {
	query := `SELECT id, project_id, purpose, username, role, token, is_active, expires_at, created_at, updated_at 
			  FROM jwt_tokens WHERE project_id = ? ORDER BY created_at DESC`

	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.JWTToken
	for rows.Next() {
		token := &models.JWTToken{}
		err := rows.Scan(
			&token.ID, &token.ProjectID, &token.Purpose, &token.Username, &token.Role,
			&token.Token, &token.IsActive, &token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (db *DB) UpdateJWTToken(token *models.JWTToken) error {
	query := `UPDATE jwt_tokens SET purpose = ?, username = ?, role = ?, is_active = ?, expires_at = ?, updated_at = ? 
			  WHERE id = ?`

	now := time.Now()
	_, err := db.Exec(query, token.Purpose, token.Username, token.Role,
		token.IsActive, token.ExpiresAt, now, token.ID)
	return err
}

func (db *DB) DeleteJWTToken(id string) error {
	_, err := db.Exec("DELETE FROM jwt_tokens WHERE id = ?", id)
	return err
}

func (db *DB) DeleteExpiredJWTTokens(projectID string) (int64, error) {
	query := `DELETE FROM jwt_tokens WHERE project_id = ? AND expires_at < ?`

	result, err := db.Exec(query, projectID, time.Now())
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (db *DB) GetJWTTokenByToken(tokenStr string) (*models.JWTToken, error) {
	query := `SELECT id, project_id, purpose, username, role, token, is_active, expires_at, created_at, updated_at 
			  FROM jwt_tokens WHERE token = ?`

	token := &models.JWTToken{}
	err := db.QueryRow(query, tokenStr).Scan(
		&token.ID, &token.ProjectID, &token.Purpose, &token.Username, &token.Role,
		&token.Token, &token.IsActive, &token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// 生成JWT Token ID
func GenerateJWTTokenID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	id := make([]byte, 12)
	for i := range id {
		id[i] = letters[r.Intn(len(letters))]
	}
	return string(id)
}

// 生成JWT Token字符串
func GenerateJWTTokenString() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	token := make([]byte, 64)
	for i := range token {
		token[i] = letters[r.Intn(len(letters))]
	}
	return string(token)
}
