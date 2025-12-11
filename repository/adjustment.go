package repository

import (
	"database/sql"

	"github.com/Kamekure-Maisuke/maiyumi/model"
)

type AdjustmentRepository interface {
	Create(adj *model.Adjustment) error
	FindByTalentID(talentID int) ([]model.Adjustment, error)
	CalculateTotalScore(talentID, baseScore int, adjustmentType string) (int, error)
	CalculateTotalScores(talentIDs []int) (map[int]map[string]int, error)
}

type adjustmentRepository struct {
	db *sql.DB
}

func NewAdjustmentRepository(db *sql.DB) AdjustmentRepository {
	return &adjustmentRepository{db: db}
}

func (r *adjustmentRepository) Create(adj *model.Adjustment) error {
	_, err := r.db.Exec(`
		INSERT INTO adjustments (talent_id, adjustment_type, points, reason)
		VALUES (?, ?, ?, ?)`,
		adj.TalentID, adj.AdjustmentType, adj.Points, adj.Reason)
	return err
}

func (r *adjustmentRepository) FindByTalentID(talentID int) ([]model.Adjustment, error) {
	rows, err := r.db.Query(`
		SELECT id, talent_id, adjustment_type, points, reason, created_at
		FROM adjustments
		WHERE talent_id = ?
		ORDER BY created_at DESC`, talentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []model.Adjustment
	for rows.Next() {
		var a model.Adjustment
		if err := rows.Scan(&a.ID, &a.TalentID, &a.AdjustmentType, &a.Points, &a.Reason, &a.CreatedAt); err == nil {
			adjustments = append(adjustments, a)
		}
	}

	return adjustments, nil
}

func (r *adjustmentRepository) CalculateTotalScore(talentID, baseScore int, adjustmentType string) (int, error) {
	var total int
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(points), 0)
		FROM adjustments
		WHERE talent_id = ? AND adjustment_type = ?`,
		talentID, adjustmentType).Scan(&total)
	return baseScore + total, err
}

func (r *adjustmentRepository) CalculateTotalScores(talentIDs []int) (map[int]map[string]int, error) {
	if len(talentIDs) == 0 {
		return make(map[int]map[string]int), nil
	}

	result := make(map[int]map[string]int)
	for _, id := range talentIDs {
		result[id] = map[string]int{
			"beauty":   0,
			"cuteness": 0,
			"talent":   0,
		}
	}

	query := `
		SELECT talent_id, adjustment_type, COALESCE(SUM(points), 0)
		FROM adjustments
		WHERE talent_id IN (`

	args := make([]any, len(talentIDs))
	for i, id := range talentIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += `) GROUP BY talent_id, adjustment_type`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var talentID int
		var adjType string
		var points int
		if err := rows.Scan(&talentID, &adjType, &points); err == nil {
			result[talentID][adjType] = points
		}
	}

	return result, nil
}
