package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	// "os"
	// "io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TickerResponse はAPIレスポンスの構造体です。
type TickerResponse struct {
	Ltp float64 `json:"ltp"`
}

// getLtp はAPIからLTPを取得します。
func getLtp() (float64, error) {
	resp, err := http.Get("https://api.bitflyer.com/v1/ticker?product_code=FX_BTC_JPY")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	// body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var ticker TickerResponse
	err = json.Unmarshal(body, &ticker)
	if err != nil {
		return 0, err
	}

	return ticker.Ltp, nil
}

// saveLtp はLTPをデータベースに保存します。
func saveLtp(db *sql.DB, ltp float64) error {
	_, err := db.Exec("INSERT INTO ltp_values (value, timestamp) VALUES (?, ?)", ltp, time.Now())
	return err
}

func main() {
	db, err := sql.Open("sqlite3", "./ltp.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// テーブルが存在しない場合は作成します。
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS ltp_values (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        value FLOAT NOT NULL,
        timestamp DATETIME NOT NULL
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}

	// 10秒ごとにLTPを取得してデータベースに保存します。
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		ltp, err := getLtp()
		if err != nil {
			log.Println("Error fetching LTP:", err)
			continue
		}

		err = saveLtp(db, ltp)
		if err != nil {
			log.Println("Error saving LTP:", err)
			continue
		}

		fmt.Println("Saved LTP:", ltp)
	}
}
