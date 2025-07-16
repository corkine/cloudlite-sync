package database

import (
	"database/sql"
	"fmt"
	"time"

	"chchma.com/cloudlite-sync/internal/models"
)

// CreateProject 创建项目
func (db *DB) CreateProject(project *models.Project) error {
	query := `INSERT INTO projects (id, name, description, website, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, project.ID, project.Name, project.Description, project.Website, now, now)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	project.CreatedAt = now
	project.UpdatedAt = now
	return nil
}

// GetProject 获取项目
func (db *DB) GetProject(id string) (*models.Project, error) {
	query := `SELECT id, name, description, website, created_at, updated_at FROM projects WHERE id = ?`

	project := &models.Project{}
	err := db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Website,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// ListProjects 获取项目列表（分页）
func (db *DB) ListProjects(page, pageSize int) ([]*models.Project, int, error) {
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM projects`
	err := db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := `SELECT id, name, description, website, created_at, updated_at 
			  FROM projects ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Website,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, total, nil
}

// UpdateProject 更新项目
func (db *DB) UpdateProject(project *models.Project) error {
	query := `UPDATE projects SET name = ?, description = ?, website = ?, updated_at = ? WHERE id = ?`

	now := time.Now()
	_, err := db.Exec(query, project.Name, project.Description, project.Website, now, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	project.UpdatedAt = now
	return nil
}

// DeleteProject 删除项目
func (db *DB) DeleteProject(id string) error {
	// 首先删除相关的凭证和数据库版本
	queries := []string{
		`DELETE FROM database_versions WHERE project_id = ?`,
		`DELETE FROM credentials WHERE project_id = ?`,
		`DELETE FROM projects WHERE id = ?`,
	}

	for _, query := range queries {
		_, err := db.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to delete project data: %w", err)
		}
	}

	return nil
}
