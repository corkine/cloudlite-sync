package database

import (
	"database/sql"
	"fmt"
	"time"

	"chchma.com/cloudlite-sync/internal/models"
)

// CreateDatabaseVersion 创建数据库版本
func (db *DB) CreateDatabaseVersion(version *models.DatabaseVersion) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 如果设置为最新版本，先取消其他版本的最新标记
	if version.IsLatest {
		_, err = tx.Exec(`UPDATE database_versions SET is_latest = 0 WHERE project_id = ?`, version.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to update latest flags: %w", err)
		}
	}

	// 插入新版本
	query := `INSERT INTO database_versions (id, project_id, version, file_hash, file_name, file_size, oss_key, description, is_latest, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err = tx.Exec(query, version.ID, version.ProjectID, version.Version, version.FileHash,
		version.FileName, version.FileSize, version.OSSKey, version.Description, version.IsLatest, now)
	if err != nil {
		return fmt.Errorf("failed to create database version: %w", err)
	}

	version.CreatedAt = now

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDatabaseVersion 获取数据库版本
func (db *DB) GetDatabaseVersion(id string) (*models.DatabaseVersion, error) {
	query := `SELECT id, project_id, version, file_hash, file_name, file_size, oss_key, description, is_latest, created_at 
			  FROM database_versions WHERE id = ?`

	version := &models.DatabaseVersion{}
	err := db.QueryRow(query, id).Scan(
		&version.ID,
		&version.ProjectID,
		&version.Version,
		&version.FileHash,
		&version.FileName,
		&version.FileSize,
		&version.OSSKey,
		&version.Description,
		&version.IsLatest,
		&version.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get database version: %w", err)
	}

	return version, nil
}

// GetLatestVersion 获取项目的最新版本
func (db *DB) GetLatestVersion(projectID string) (*models.DatabaseVersion, error) {
	query := `SELECT id, project_id, version, file_hash, file_name, file_size, oss_key, description, is_latest, created_at 
			  FROM database_versions WHERE project_id = ? AND is_latest = 1`

	version := &models.DatabaseVersion{}
	err := db.QueryRow(query, projectID).Scan(
		&version.ID,
		&version.ProjectID,
		&version.Version,
		&version.FileHash,
		&version.FileName,
		&version.FileSize,
		&version.OSSKey,
		&version.Description,
		&version.IsLatest,
		&version.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return version, nil
}

// GetVersionByHash 通过文件哈希获取版本
func (db *DB) GetVersionByHash(projectID, fileHash string) (*models.DatabaseVersion, error) {
	query := `SELECT id, project_id, version, file_hash, file_name, file_size, oss_key, description, is_latest, created_at 
			  FROM database_versions WHERE project_id = ? AND file_hash = ?`

	version := &models.DatabaseVersion{}
	err := db.QueryRow(query, projectID, fileHash).Scan(
		&version.ID,
		&version.ProjectID,
		&version.Version,
		&version.FileHash,
		&version.FileName,
		&version.FileSize,
		&version.OSSKey,
		&version.Description,
		&version.IsLatest,
		&version.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get version by hash: %w", err)
	}

	return version, nil
}

// ListDatabaseVersions 获取数据库版本列表（分页）
func (db *DB) ListDatabaseVersions(projectID string, page, pageSize int) ([]*models.DatabaseVersion, int, error) {
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM database_versions WHERE project_id = ?`
	err := db.QueryRow(countQuery, projectID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count database versions: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	query := `SELECT id, project_id, version, file_hash, file_name, file_size, oss_key, description, is_latest, created_at 
			  FROM database_versions WHERE project_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := db.Query(query, projectID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query database versions: %w", err)
	}
	defer rows.Close()

	var versions []*models.DatabaseVersion
	for rows.Next() {
		version := &models.DatabaseVersion{}
		err := rows.Scan(
			&version.ID,
			&version.ProjectID,
			&version.Version,
			&version.FileHash,
			&version.FileName,
			&version.FileSize,
			&version.OSSKey,
			&version.Description,
			&version.IsLatest,
			&version.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan database version: %w", err)
		}
		versions = append(versions, version)
	}

	return versions, total, nil
}

// DeleteDatabaseVersion 删除数据库版本
func (db *DB) DeleteDatabaseVersion(id string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 先获取版本信息，包括项目ID和是否是最新版本
	var projectID string
	var isLatest bool
	err = tx.QueryRow(`SELECT project_id, is_latest FROM database_versions WHERE id = ?`, id).Scan(&projectID, &isLatest)
	if err != nil {
		return fmt.Errorf("failed to get version info: %w", err)
	}

	// 删除版本
	_, err = tx.Exec(`DELETE FROM database_versions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete database version: %w", err)
	}

	// 如果删除的是最新版本，设置最近创建的版本为最新
	if isLatest {
		// 检查是否还有其他版本
		var count int
		err = tx.QueryRow(`SELECT COUNT(*) FROM database_versions WHERE project_id = ?`, projectID).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to count remaining versions: %w", err)
		}

		// 如果还有其他版本，设置最近创建的版本为最新
		if count > 0 {
			_, err = tx.Exec(`UPDATE database_versions SET is_latest = 1 
							  WHERE id = (SELECT id FROM database_versions 
							  WHERE project_id = ? 
							  ORDER BY created_at DESC LIMIT 1)`, projectID)
			if err != nil {
				return fmt.Errorf("failed to update latest version: %w", err)
			}
		}
		// 如果没有其他版本，不需要设置最新版本（项目将没有最新版本）
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetLatestVersion 设置指定版本为最新版本
func (db *DB) SetLatestVersion(projectID, versionID string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 取消所有版本的最新标记
	_, err = tx.Exec(`UPDATE database_versions SET is_latest = 0 WHERE project_id = ?`, projectID)
	if err != nil {
		return fmt.Errorf("failed to clear latest flags: %w", err)
	}

	// 设置指定版本为最新
	_, err = tx.Exec(`UPDATE database_versions SET is_latest = 1 WHERE id = ?`, versionID)
	if err != nil {
		return fmt.Errorf("failed to set latest version: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
