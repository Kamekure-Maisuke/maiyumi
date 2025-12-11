package repository

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupUserTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "正常なユーザー作成",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "重複ユーザー名",
			username: "testuser",
			password: "password456",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	username := "testuser"
	password := "password123"
	err := repo.Create(username, password)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "存在するユーザー",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "存在しないユーザー",
			username: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.FindByUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Username != username {
					t.Errorf("FindByUsername() username = %v, want %v", user.Username, username)
				}
				if user.Password != password {
					t.Errorf("FindByUsername() password = %v, want %v", user.Password, password)
				}
			}
		})
	}
}

func TestUserRepository_GetID(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	username := "testuser"
	err := repo.Create(username, "password123")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "存在するユーザーのID取得",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "存在しないユーザーのID取得",
			username: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := repo.GetID(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id <= 0 {
				t.Errorf("GetID() id = %v, want > 0", id)
			}
		})
	}
}

func TestUserRepository_UpdateUsername(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	username := "testuser"
	err := repo.Create(username, "password123")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := repo.GetID(username)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		userID      int
		newUsername string
		wantErr     bool
	}{
		{
			name:        "正常なユーザー名更新",
			userID:      userID,
			newUsername: "newusername",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateUsername(tt.userID, tt.newUsername)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				user, err := repo.FindByID(tt.userID)
				if err != nil {
					t.Errorf("FindByID() error = %v", err)
					return
				}
				if user.Username != tt.newUsername {
					t.Errorf("UpdateUsername() username = %v, want %v", user.Username, tt.newUsername)
				}
			}
		})
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	username := "testuser"
	err := repo.Create(username, "password123")
	if err != nil {
		t.Fatal(err)
	}

	userID, err := repo.GetID(username)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		userID      int
		newPassword string
		wantErr     bool
	}{
		{
			name:        "正常なパスワード更新",
			userID:      userID,
			newPassword: "newpassword456",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdatePassword(tt.userID, tt.newPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				user, err := repo.FindByID(tt.userID)
				if err != nil {
					t.Errorf("FindByID() error = %v", err)
					return
				}
				if user.Password != tt.newPassword {
					t.Errorf("UpdatePassword() password = %v, want %v", user.Password, tt.newPassword)
				}
			}
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupUserTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	username := "testuser"
	password := "password123"
	err := repo.Create(username, password)
	if err != nil {
		t.Fatal(err)
	}

	userID, err := repo.GetID(username)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		userID  int
		wantErr bool
	}{
		{
			name:    "存在するユーザーID",
			userID:  userID,
			wantErr: false,
		},
		{
			name:    "存在しないユーザーID",
			userID:  9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.FindByID(tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Username != username {
					t.Errorf("FindByID() username = %v, want %v", user.Username, username)
				}
				if user.Password != password {
					t.Errorf("FindByID() password = %v, want %v", user.Password, password)
				}
			}
		})
	}
}
