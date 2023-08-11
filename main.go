package main

import (
	// "encoding/json"
	"fmt"
	// "math"
	"net/http"
	"log"
	// "strconv"
	"get-address-api/step2"
	"get-address-api/step3"
	"database/sql"
    _ "github.com/mattn/go-sqlite3"
)


func main() {
	// SQLite3 データベースに接続
    db, err := sql.Open("sqlite3", "./access_logs.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // access_logs テーブルが存在しない場合は作成
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS access_logs (
            id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            postal_code VARCHAR(8) NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, API Server!")
	})

	http.HandleFunc("/address", step2.AddressHandler)
	http.HandleFunc("/address/access_logs", step3.AccessLogsHandler)

	http.ListenAndServe(":8080", nil)
}
