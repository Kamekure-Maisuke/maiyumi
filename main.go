package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Kamekure-Maisuke/maiyumi/model"
	"github.com/Kamekure-Maisuke/maiyumi/repository"
	_ "modernc.org/sqlite"
)

type App struct {
	userRepo       repository.UserRepository
	talentRepo     repository.TalentRepository
	adjustmentRepo repository.AdjustmentRepository
	sessionRepo    repository.SessionRepository
	tmpl           *template.Template
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "data.db")
	if err != nil {
		return nil, err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS talents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		affiliation TEXT,
		beauty INTEGER NOT NULL CHECK(beauty >= 1 AND beauty <= 10),
		cuteness INTEGER NOT NULL CHECK(cuteness >= 1 AND cuteness <= 10),
		talent INTEGER NOT NULL CHECK(talent >= 1 AND talent <= 10),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE TABLE IF NOT EXISTS adjustments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		talent_id INTEGER NOT NULL,
		adjustment_type TEXT NOT NULL CHECK(adjustment_type IN ('beauty', 'cuteness', 'talent')),
		points INTEGER NOT NULL CHECK(points >= -10 AND points <= 10),
		reason TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (talent_id) REFERENCES talents(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(createTableSQL)
	return db, err
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (app *App) getUsername(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	username, _ := app.sessionRepo.Get(cookie.Value)
	return username
}

func (app *App) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	username := app.getUsername(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	username := app.getUsername(r)
	data := struct {
		Name string
	}{
		Name: username,
	}
	app.tmpl.ExecuteTemplate(w, "index.tmpl", data)
}

func (app *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		app.tmpl.ExecuteTemplate(w, "register.tmpl", nil)
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			http.Error(w, "ユーザー名とパスワードを入力してください", http.StatusBadRequest)
			return
		}

		hashedPassword := hashPassword(password)
		if err := app.userRepo.Create(username, hashedPassword); err != nil {
			http.Error(w, "ユーザー登録に失敗しました", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		app.tmpl.ExecuteTemplate(w, "login.tmpl", nil)
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := app.userRepo.FindByUsername(username)
		if err != nil {
			http.Error(w, "ログインに失敗しました", http.StatusUnauthorized)
			return
		}

		hashedPassword := hashPassword(password)
		if hashedPassword != user.Password {
			http.Error(w, "ログインに失敗しました", http.StatusUnauthorized)
			return
		}

		sessionID := generateSessionID()
		app.sessionRepo.Set(sessionID, username)

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		app.sessionRepo.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *App) handleTalents(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	username := app.getUsername(r)
	userID, err := app.userRepo.GetID(username)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	talents, err := app.talentRepo.FindByUserID(userID)
	if err != nil {
		http.Error(w, "タレント一覧の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	app.tmpl.ExecuteTemplate(w, "talents.tmpl", map[string]any{
		"Talents": talents,
	})
}

func (app *App) handleTalentNew(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	if r.Method == http.MethodGet {
		app.tmpl.ExecuteTemplate(w, "talent_form.tmpl", map[string]any{
			"IsEdit": false,
		})
		return
	}

	if r.Method == http.MethodPost {
		username := app.getUsername(r)
		userID, err := app.userRepo.GetID(username)
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		name := r.FormValue("name")
		affiliation := r.FormValue("affiliation")
		beauty, _ := strconv.Atoi(r.FormValue("beauty"))
		cuteness, _ := strconv.Atoi(r.FormValue("cuteness"))
		talent, _ := strconv.Atoi(r.FormValue("talent"))

		if name == "" || beauty < 1 || beauty > 10 || cuteness < 1 || cuteness > 10 || talent < 1 || talent > 10 {
			http.Error(w, "入力値が不正です", http.StatusBadRequest)
			return
		}

		newTalent := &model.Talent{
			UserID:   userID,
			Name:     name,
			Beauty:   beauty,
			Cuteness: cuteness,
			Talent:   talent,
		}

		if affiliation != "" {
			newTalent.Affiliation = sql.NullString{String: affiliation, Valid: true}
		}

		if err := app.talentRepo.Create(newTalent); err != nil {
			http.Error(w, "タレントの登録に失敗しました", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/talents", http.StatusSeeOther)
	}
}

func (app *App) handleTalentEdit(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	talentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	username := app.getUsername(r)
	userID, err := app.userRepo.GetID(username)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		talent, err := app.talentRepo.FindByID(talentID, userID)
		if err != nil {
			http.Error(w, "タレント情報の取得に失敗しました", http.StatusNotFound)
			return
		}

		app.tmpl.ExecuteTemplate(w, "talent_form.tmpl", map[string]any{
			"IsEdit": true,
			"Talent": talent,
		})
		return
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		affiliation := r.FormValue("affiliation")
		beauty, _ := strconv.Atoi(r.FormValue("beauty"))
		cuteness, _ := strconv.Atoi(r.FormValue("cuteness"))
		talent, _ := strconv.Atoi(r.FormValue("talent"))

		if name == "" || beauty < 1 || beauty > 10 || cuteness < 1 || cuteness > 10 || talent < 1 || talent > 10 {
			http.Error(w, "入力値が不正です", http.StatusBadRequest)
			return
		}

		updateTalent := &model.Talent{
			ID:       talentID,
			UserID:   userID,
			Name:     name,
			Beauty:   beauty,
			Cuteness: cuteness,
			Talent:   talent,
		}

		if affiliation != "" {
			updateTalent.Affiliation = sql.NullString{String: affiliation, Valid: true}
		}

		if err := app.talentRepo.Update(updateTalent); err != nil {
			http.Error(w, "タレント情報の更新に失敗しました", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/talents/detail?id="+strconv.Itoa(talentID), http.StatusSeeOther)
	}
}

func (app *App) handleTalentDelete(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "無効なメソッドです", http.StatusMethodNotAllowed)
		return
	}

	talentID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	username := app.getUsername(r)
	userID, err := app.userRepo.GetID(username)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	if err := app.talentRepo.Delete(talentID, userID); err != nil {
		http.Error(w, "タレントの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/talents", http.StatusSeeOther)
}

func (app *App) handleTalentAdjust(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "無効なメソッドです", http.StatusMethodNotAllowed)
		return
	}

	talentID, err := strconv.Atoi(r.FormValue("talent_id"))
	if err != nil {
		http.Error(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	username := app.getUsername(r)
	userID, err := app.userRepo.GetID(username)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	exists, err := app.talentRepo.Exists(talentID, userID)
	if err != nil || !exists {
		http.Error(w, "タレント情報が見つかりません", http.StatusNotFound)
		return
	}

	adjustmentType := r.FormValue("adjustment_type")
	points, _ := strconv.Atoi(r.FormValue("points"))
	reason := r.FormValue("reason")

	if (adjustmentType != "beauty" && adjustmentType != "cuteness" && adjustmentType != "talent") ||
		points < -10 || points > 10 || reason == "" {
		http.Error(w, "入力値が不正です", http.StatusBadRequest)
		return
	}

	adjustment := &model.Adjustment{
		TalentID:       talentID,
		AdjustmentType: adjustmentType,
		Points:         points,
		Reason:         reason,
	}

	if err := app.adjustmentRepo.Create(adjustment); err != nil {
		http.Error(w, "調整の追加に失敗しました", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/talents/detail?id="+strconv.Itoa(talentID), http.StatusSeeOther)
}

func (app *App) handleTalentDetail(w http.ResponseWriter, r *http.Request) {
	if !app.requireAuth(w, r) {
		return
	}

	talentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "無効なIDです", http.StatusBadRequest)
		return
	}

	username := app.getUsername(r)
	userID, err := app.userRepo.GetID(username)
	if err != nil {
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	talent, err := app.talentRepo.FindByID(talentID, userID)
	if err != nil {
		http.Error(w, "タレント情報の取得に失敗しました", http.StatusNotFound)
		return
	}

	adjustments, err := app.adjustmentRepo.FindByTalentID(talentID)
	if err != nil {
		http.Error(w, "履歴の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	app.tmpl.ExecuteTemplate(w, "talent_detail.tmpl", map[string]any{
		"Talent":      talent,
		"Adjustments": adjustments,
	})
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	adjustmentRepo := repository.NewAdjustmentRepository(db)
	talentRepo := repository.NewTalentRepository(db, adjustmentRepo)

	app := &App{
		userRepo:       repository.NewUserRepository(db),
		talentRepo:     talentRepo,
		adjustmentRepo: adjustmentRepo,
		sessionRepo:    repository.NewSessionRepository(),
		tmpl: template.Must(template.ParseFiles(
			"templates/index.tmpl",
			"templates/login.tmpl",
			"templates/register.tmpl",
			"templates/talents.tmpl",
			"templates/talent_detail.tmpl",
			"templates/talent_form.tmpl",
		)),
	}

	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/register", app.handleRegister)
	http.HandleFunc("/login", app.handleLogin)
	http.HandleFunc("/logout", app.handleLogout)
	http.HandleFunc("/talents", app.handleTalents)
	http.HandleFunc("/talents/new", app.handleTalentNew)
	http.HandleFunc("/talents/edit", app.handleTalentEdit)
	http.HandleFunc("/talents/delete", app.handleTalentDelete)
	http.HandleFunc("/talents/adjust", app.handleTalentAdjust)
	http.HandleFunc("/talents/detail", app.handleTalentDetail)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
