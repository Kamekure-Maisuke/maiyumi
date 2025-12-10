package repository

import (
	"database/sql"

	"github.com/Kamekure-Maisuke/maiyumi/model"
)

type TalentRepository interface {
	Create(talent *model.Talent) error
	Update(talent *model.Talent) error
	Delete(id, userID int) error
	FindByID(id, userID int) (*model.Talent, error)
	FindByUserID(userID int) ([]model.Talent, error)
	SearchByUserID(userID int, query string) ([]model.Talent, error)
	FindFavoritesByUserID(userID int) ([]model.Talent, error)
	ToggleFavorite(id, userID int) error
	Exists(id, userID int) (bool, error)
}

type talentRepository struct {
	db       *sql.DB
	adjRepo  AdjustmentRepository
}

func NewTalentRepository(db *sql.DB, adjRepo AdjustmentRepository) TalentRepository {
	return &talentRepository{db: db, adjRepo: adjRepo}
}

func (r *talentRepository) Create(talent *model.Talent) error {
	var affiliation any
	if talent.Affiliation.Valid {
		affiliation = talent.Affiliation.String
	}

	_, err := r.db.Exec(`
		INSERT INTO talents (user_id, name, affiliation, beauty, cuteness, talent)
		VALUES (?, ?, ?, ?, ?, ?)`,
		talent.UserID, talent.Name, affiliation, talent.Beauty, talent.Cuteness, talent.Talent)
	return err
}

func (r *talentRepository) Update(talent *model.Talent) error {
	var affiliation any
	if talent.Affiliation.Valid {
		affiliation = talent.Affiliation.String
	}

	_, err := r.db.Exec(`
		UPDATE talents
		SET name = ?, affiliation = ?, beauty = ?, cuteness = ?, talent = ?
		WHERE id = ? AND user_id = ?`,
		talent.Name, affiliation, talent.Beauty, talent.Cuteness, talent.Talent, talent.ID, talent.UserID)
	return err
}

func (r *talentRepository) Delete(id, userID int) error {
	_, err := r.db.Exec("DELETE FROM talents WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (r *talentRepository) FindByID(id, userID int) (*model.Talent, error) {
	var t model.Talent
	err := r.db.QueryRow(`
		SELECT id, user_id, name, affiliation, beauty, cuteness, talent, is_favorite, created_at
		FROM talents
		WHERE id = ? AND user_id = ?`, id, userID).Scan(
		&t.ID, &t.UserID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent, &t.IsFavorite, &t.CreatedAt)
	if err != nil {
		return nil, err
	}

	t.TotalBeauty, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Beauty, "beauty")
	t.TotalCuteness, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Cuteness, "cuteness")
	t.TotalTalent, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Talent, "talent")

	return &t, nil
}

func (r *talentRepository) FindByUserID(userID int) ([]model.Talent, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, affiliation, beauty, cuteness, talent, is_favorite, created_at
		FROM talents
		WHERE user_id = ?
		ORDER BY is_favorite DESC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var talents []model.Talent
	for rows.Next() {
		var t model.Talent
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent, &t.IsFavorite, &t.CreatedAt); err != nil {
			continue
		}

		t.TotalBeauty, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Beauty, "beauty")
		t.TotalCuteness, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Cuteness, "cuteness")
		t.TotalTalent, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Talent, "talent")

		talents = append(talents, t)
	}

	return talents, nil
}

func (r *talentRepository) SearchByUserID(userID int, query string) ([]model.Talent, error) {
	searchQuery := "%" + query + "%"
	rows, err := r.db.Query(`
		SELECT id, user_id, name, affiliation, beauty, cuteness, talent, is_favorite, created_at
		FROM talents
		WHERE user_id = ? AND (name LIKE ? OR affiliation LIKE ?)
		ORDER BY is_favorite DESC, created_at DESC`, userID, searchQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var talents []model.Talent
	for rows.Next() {
		var t model.Talent
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent, &t.IsFavorite, &t.CreatedAt); err != nil {
			continue
		}

		t.TotalBeauty, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Beauty, "beauty")
		t.TotalCuteness, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Cuteness, "cuteness")
		t.TotalTalent, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Talent, "talent")

		talents = append(talents, t)
	}

	return talents, nil
}

func (r *talentRepository) FindFavoritesByUserID(userID int) ([]model.Talent, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, affiliation, beauty, cuteness, talent, is_favorite, created_at
		FROM talents
		WHERE user_id = ? AND is_favorite = 1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var talents []model.Talent
	for rows.Next() {
		var t model.Talent
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent, &t.IsFavorite, &t.CreatedAt); err != nil {
			continue
		}

		t.TotalBeauty, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Beauty, "beauty")
		t.TotalCuteness, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Cuteness, "cuteness")
		t.TotalTalent, _ = r.adjRepo.CalculateTotalScore(t.ID, t.Talent, "talent")

		talents = append(talents, t)
	}

	return talents, nil
}

func (r *talentRepository) ToggleFavorite(id, userID int) error {
	_, err := r.db.Exec(`
		UPDATE talents
		SET is_favorite = NOT is_favorite
		WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

func (r *talentRepository) Exists(id, userID int) (bool, error) {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM talents WHERE id = ? AND user_id = ?", id, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
