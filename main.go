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

	_ "modernc.org/sqlite"
)

var db *sql.DB
var sessions = make(map[string]string)

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", "data.db")
	if err != nil {
		return err
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
	return err
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

func getUsername(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return sessions[cookie.Value]
}

func requireAuth(w http.ResponseWriter, r *http.Request) bool {
	username := getUsername(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}

type Talent struct {
	ID            int
	Name          string
	Affiliation   sql.NullString
	Beauty        int
	Cuteness      int
	Talent        int
	TotalBeauty   int
	TotalCuteness int
	TotalTalent   int
	CreatedAt     string
}

type Adjustment struct {
	ID             int
	AdjustmentType string
	Points         int
	Reason         string
	CreatedAt      string
}

func getUserID(username string) (int, error) {
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	return userID, err
}

func calculateTotalScore(talentID int, baseScore int, adjustmentType string) (int, error) {
	var total int
	err := db.QueryRow(`
		SELECT COALESCE(SUM(points), 0)
		FROM adjustments
		WHERE talent_id = ? AND adjustment_type = ?`,
		talentID, adjustmentType).Scan(&total)
	return baseScore + total, err
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl := template.Must(template.ParseFiles(
		"templates/index.tmpl",
		"templates/login.tmpl",
		"templates/register.tmpl",
		"templates/talents.tmpl",
		"templates/talent_detail.tmpl",
		"templates/talent_form.tmpl",
	))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}

		username := getUsername(r)

		data := struct {
			Name string
		}{
			Name: username,
		}
		tmpl.ExecuteTemplate(w, "index.tmpl", data)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl.ExecuteTemplate(w, "register.tmpl", nil)
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
			_, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPassword)
			if err != nil {
				http.Error(w, "ユーザー登録に失敗しました", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tmpl.ExecuteTemplate(w, "login.tmpl", nil)
			return
		}

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			var storedPassword string
			err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)
			if err != nil {
				http.Error(w, "ログインに失敗しました", http.StatusUnauthorized)
				return
			}

			hashedPassword := hashPassword(password)
			if hashedPassword != storedPassword {
				http.Error(w, "ログインに失敗しました", http.StatusUnauthorized)
				return
			}

			sessionID := generateSessionID()
			sessions[sessionID] = username

			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err == nil {
			delete(sessions, cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	http.HandleFunc("/talents", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}

		username := getUsername(r)
		userID, err := getUserID(username)
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		rows, err := db.Query(`
			SELECT id, name, affiliation, beauty, cuteness, talent, created_at
			FROM talents
			WHERE user_id = ?
			ORDER BY created_at DESC`, userID)
		if err != nil {
			http.Error(w, "タレント一覧の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var talents []Talent
		for rows.Next() {
			var t Talent
			err := rows.Scan(&t.ID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent, &t.CreatedAt)
			if err != nil {
				continue
			}
			t.TotalBeauty, _ = calculateTotalScore(t.ID, t.Beauty, "beauty")
			t.TotalCuteness, _ = calculateTotalScore(t.ID, t.Cuteness, "cuteness")
			t.TotalTalent, _ = calculateTotalScore(t.ID, t.Talent, "talent")
			talents = append(talents, t)
		}

		tmpl.ExecuteTemplate(w, "talents.tmpl", map[string]any{
			"Talents": talents,
		})
	})

	http.HandleFunc("/talents/new", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}

		if r.Method == http.MethodGet {
			tmpl.ExecuteTemplate(w, "talent_form.tmpl", map[string]any{
				"IsEdit": false,
			})
			return
		}

		if r.Method == http.MethodPost {
			username := getUsername(r)
			userID, err := getUserID(username)
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

			var affiliationVal any
			if affiliation == "" {
				affiliationVal = nil
			} else {
				affiliationVal = affiliation
			}

			_, err = db.Exec(`
				INSERT INTO talents (user_id, name, affiliation, beauty, cuteness, talent)
				VALUES (?, ?, ?, ?, ?, ?)`,
				userID, name, affiliationVal, beauty, cuteness, talent)
			if err != nil {
				http.Error(w, "タレントの登録に失敗しました", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/talents", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/talents/edit", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}

		talentID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "無効なIDです", http.StatusBadRequest)
			return
		}

		username := getUsername(r)
		userID, err := getUserID(username)
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		if r.Method == http.MethodGet {
			var t Talent
			err = db.QueryRow(`
				SELECT id, name, affiliation, beauty, cuteness, talent
				FROM talents
				WHERE id = ? AND user_id = ?`, talentID, userID).Scan(
				&t.ID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent)
			if err != nil {
				http.Error(w, "タレント情報の取得に失敗しました", http.StatusNotFound)
				return
			}

			tmpl.ExecuteTemplate(w, "talent_form.tmpl", map[string]any{
				"IsEdit": true,
				"Talent": t,
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

			var affiliationVal any
			if affiliation == "" {
				affiliationVal = nil
			} else {
				affiliationVal = affiliation
			}

			_, err = db.Exec(`
				UPDATE talents
				SET name = ?, affiliation = ?, beauty = ?, cuteness = ?, talent = ?
				WHERE id = ? AND user_id = ?`,
				name, affiliationVal, beauty, cuteness, talent, talentID, userID)
			if err != nil {
				http.Error(w, "タレント情報の更新に失敗しました", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/talents/detail?id="+strconv.Itoa(talentID), http.StatusSeeOther)
		}
	})

	http.HandleFunc("/talents/delete", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
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

		username := getUsername(r)
		userID, err := getUserID(username)
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("DELETE FROM talents WHERE id = ? AND user_id = ?", talentID, userID)
		if err != nil {
			http.Error(w, "タレントの削除に失敗しました", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/talents", http.StatusSeeOther)
	})

	http.HandleFunc("/talents/adjust", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
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

		username := getUsername(r)
		userID, err := getUserID(username)
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		var exists int
		err = db.QueryRow("SELECT 1 FROM talents WHERE id = ? AND user_id = ?", talentID, userID).Scan(&exists)
		if err != nil {
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

		_, err = db.Exec(`
			INSERT INTO adjustments (talent_id, adjustment_type, points, reason)
			VALUES (?, ?, ?, ?)`,
			talentID, adjustmentType, points, reason)
		if err != nil {
			http.Error(w, "調整の追加に失敗しました", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/talents/detail?id="+strconv.Itoa(talentID), http.StatusSeeOther)
	})

	http.HandleFunc("/talents/detail", func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r) {
			return
		}

		talentID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, "無効なIDです", http.StatusBadRequest)
			return
		}

		username := getUsername(r)
		userID, err := getUserID(username)
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		var t Talent
		err = db.QueryRow(`
			SELECT id, name, affiliation, beauty, cuteness, talent, created_at
			FROM talents
			WHERE id = ? AND user_id = ?`, talentID, userID).Scan(
			&t.ID, &t.Name, &t.Affiliation, &t.Beauty, &t.Cuteness, &t.Talent, &t.CreatedAt)
		if err != nil {
			http.Error(w, "タレント情報の取得に失敗しました", http.StatusNotFound)
			return
		}

		t.TotalBeauty, _ = calculateTotalScore(t.ID, t.Beauty, "beauty")
		t.TotalCuteness, _ = calculateTotalScore(t.ID, t.Cuteness, "cuteness")
		t.TotalTalent, _ = calculateTotalScore(t.ID, t.Talent, "talent")

		rows, err := db.Query(`
			SELECT id, adjustment_type, points, reason, created_at
			FROM adjustments
			WHERE talent_id = ?
			ORDER BY created_at DESC`, talentID)
		if err != nil {
			http.Error(w, "履歴の取得に失敗しました", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var adjustments []Adjustment
		for rows.Next() {
			var a Adjustment
			err := rows.Scan(&a.ID, &a.AdjustmentType, &a.Points, &a.Reason, &a.CreatedAt)
			if err == nil {
				adjustments = append(adjustments, a)
			}
		}

		tmpl.ExecuteTemplate(w, "talent_detail.tmpl", map[string]any{
			"Talent":      t,
			"Adjustments": adjustments,
		})
	})

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
