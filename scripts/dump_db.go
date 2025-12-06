package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "data.db"
	dumpDir := "dump"

	timestamp := time.Now().Format("20060102_150405")
	dumpFile := filepath.Join(dumpDir, fmt.Sprintf("dump_%s.sql", timestamp))

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer db.Close()

	f, err := os.Create(dumpFile)
	if err != nil {
		log.Fatalf("ダンプファイル作成エラー: %v", err)
	}
	defer f.Close()

	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("テーブル一覧取得エラー: %v", err)
	}

	for _, table := range tables {
		if err := dumpTable(db, f, table); err != nil {
			log.Fatalf("テーブル %s のダンプエラー: %v", table, err)
		}
	}

	log.Printf("ダンプ完了: %s", dumpFile)
}

func getTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func dumpTable(db *sql.DB, f *os.File, table string) error {
	var schema string
	err := db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&schema)
	if err != nil {
		return err
	}

	fmt.Fprintf(f, "%s;\n\n", schema)

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		var sqlValues []string
		for _, v := range values {
			if v == nil {
				sqlValues = append(sqlValues, "NULL")
			} else {
				switch val := v.(type) {
				case []byte:
					sqlValues = append(sqlValues, fmt.Sprintf("'%s'", strings.ReplaceAll(string(val), "'", "''")))
				case string:
					sqlValues = append(sqlValues, fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''")))
				case int64:
					sqlValues = append(sqlValues, fmt.Sprintf("%d", val))
				default:
					sqlValues = append(sqlValues, fmt.Sprintf("'%v'", v))
				}
			}
		}

		fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s);\n",
			table,
			strings.Join(cols, ", "),
			strings.Join(sqlValues, ", "))
	}

	fmt.Fprintln(f)
	return nil
}
