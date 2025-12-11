package repository

import (
	"database/sql"
	"testing"

	"github.com/Kamekure-Maisuke/maiyumi/model"
	_ "modernc.org/sqlite"
)

func setupTalentTestDB(t *testing.T) (*sql.DB, AdjustmentRepository) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE talents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			affiliation TEXT,
			beauty INTEGER NOT NULL,
			cuteness INTEGER NOT NULL,
			talent INTEGER NOT NULL,
			is_favorite BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE adjustments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			talent_id INTEGER NOT NULL,
			adjustment_type TEXT NOT NULL,
			points INTEGER NOT NULL,
			reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	adjRepo := NewAdjustmentRepository(db)
	return db, adjRepo
}

func TestTalentRepository_Create(t *testing.T) {
	db, adjRepo := setupTalentTestDB(t)
	defer db.Close()

	repo := NewTalentRepository(db, adjRepo)

	talent := &model.Talent{
		UserID:   1,
		Name:     "テストタレント",
		Beauty:   80,
		Cuteness: 75,
		Talent:   90,
	}

	err := repo.Create(talent)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM talents WHERE name = ?", talent.Name).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("Create() did not insert talent, count = %d", count)
	}
}

func TestTalentRepository_Update(t *testing.T) {
	db, adjRepo := setupTalentTestDB(t)
	defer db.Close()

	repo := NewTalentRepository(db, adjRepo)

	talent := &model.Talent{
		UserID:   1,
		Name:     "元の名前",
		Beauty:   80,
		Cuteness: 75,
		Talent:   90,
	}
	err := repo.Create(talent)
	if err != nil {
		t.Fatal(err)
	}

	var id int
	err = db.QueryRow("SELECT id FROM talents WHERE name = ?", talent.Name).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	updatedTalent := &model.Talent{
		ID:       id,
		UserID:   1,
		Name:     "更新後の名前",
		Beauty:   85,
		Cuteness: 80,
		Talent:   95,
	}

	err = repo.Update(updatedTalent)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	var name string
	var beauty int
	err = db.QueryRow("SELECT name, beauty FROM talents WHERE id = ?", id).Scan(&name, &beauty)
	if err != nil {
		t.Fatal(err)
	}
	if name != "更新後の名前" {
		t.Errorf("Update() name = %v, want %v", name, "更新後の名前")
	}
	if beauty != 85 {
		t.Errorf("Update() beauty = %v, want %v", beauty, 85)
	}
}

func TestTalentRepository_FindByUserID(t *testing.T) {
	db, adjRepo := setupTalentTestDB(t)
	defer db.Close()

	repo := NewTalentRepository(db, adjRepo)

	talents := []*model.Talent{
		{UserID: 1, Name: "タレント1", Beauty: 80, Cuteness: 75, Talent: 90},
		{UserID: 1, Name: "タレント2", Beauty: 85, Cuteness: 80, Talent: 85},
		{UserID: 2, Name: "タレント3", Beauty: 70, Cuteness: 85, Talent: 80},
	}

	for _, talent := range talents {
		err := repo.Create(talent)
		if err != nil {
			t.Fatal(err)
		}
	}

	found, err := repo.FindByUserID(1)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}
	if len(found) != 2 {
		t.Errorf("FindByUserID() returned %d talents, want 2", len(found))
	}
}

func TestTalentRepository_SearchByUserID(t *testing.T) {
	db, adjRepo := setupTalentTestDB(t)
	defer db.Close()

	repo := NewTalentRepository(db, adjRepo)

	talents := []*model.Talent{
		{UserID: 1, Name: "山田花子", Beauty: 80, Cuteness: 75, Talent: 90},
		{UserID: 1, Name: "佐藤太郎", Beauty: 85, Cuteness: 80, Talent: 85},
		{UserID: 1, Name: "田中次郎", Beauty: 70, Cuteness: 85, Talent: 80},
	}

	for _, talent := range talents {
		err := repo.Create(talent)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
	}{
		{
			name:      "部分一致検索_山田",
			query:     "山田",
			wantCount: 1,
		},
		{
			name:      "部分一致検索_太郎",
			query:     "太郎",
			wantCount: 1,
		},
		{
			name:      "部分一致検索_田",
			query:     "田",
			wantCount: 2,
		},
		{
			name:      "一致なし",
			query:     "存在しない",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.SearchByUserID(1, tt.query)
			if err != nil {
				t.Errorf("SearchByUserID() error = %v", err)
			}
			if len(found) != tt.wantCount {
				t.Errorf("SearchByUserID() returned %d talents, want %d", len(found), tt.wantCount)
			}
		})
	}
}

func TestTalentRepository_ToggleFavorite(t *testing.T) {
	db, adjRepo := setupTalentTestDB(t)
	defer db.Close()

	repo := NewTalentRepository(db, adjRepo)

	talent := &model.Talent{
		UserID:   1,
		Name:     "お気に入りテスト",
		Beauty:   80,
		Cuteness: 75,
		Talent:   90,
	}
	err := repo.Create(talent)
	if err != nil {
		t.Fatal(err)
	}

	var id int
	var isFavorite bool
	err = db.QueryRow("SELECT id, is_favorite FROM talents WHERE name = ?", talent.Name).Scan(&id, &isFavorite)
	if err != nil {
		t.Fatal(err)
	}
	if isFavorite {
		t.Errorf("Initial is_favorite should be false")
	}

	err = repo.ToggleFavorite(id, 1)
	if err != nil {
		t.Errorf("ToggleFavorite() error = %v", err)
	}

	err = db.QueryRow("SELECT is_favorite FROM talents WHERE id = ?", id).Scan(&isFavorite)
	if err != nil {
		t.Fatal(err)
	}
	if !isFavorite {
		t.Errorf("is_favorite should be true after toggle")
	}

	err = repo.ToggleFavorite(id, 1)
	if err != nil {
		t.Errorf("ToggleFavorite() error = %v", err)
	}

	err = db.QueryRow("SELECT is_favorite FROM talents WHERE id = ?", id).Scan(&isFavorite)
	if err != nil {
		t.Fatal(err)
	}
	if isFavorite {
		t.Errorf("is_favorite should be false after second toggle")
	}
}

func TestTalentRepository_Delete(t *testing.T) {
	db, adjRepo := setupTalentTestDB(t)
	defer db.Close()

	repo := NewTalentRepository(db, adjRepo)

	talent := &model.Talent{
		UserID:   1,
		Name:     "削除テスト",
		Beauty:   80,
		Cuteness: 75,
		Talent:   90,
	}
	err := repo.Create(talent)
	if err != nil {
		t.Fatal(err)
	}

	var id int
	err = db.QueryRow("SELECT id FROM talents WHERE name = ?", talent.Name).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	err = repo.Delete(id, 1)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM talents WHERE id = ?", id).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Delete() did not remove talent, count = %d", count)
	}

	err = repo.Delete(id, 2)
	if err != nil {
		t.Errorf("Delete() with wrong user_id should not error = %v", err)
	}
}
