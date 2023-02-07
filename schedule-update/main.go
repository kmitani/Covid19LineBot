package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
	"sam-app-ca-07-covid.root/pkg"
)

const (
	YYYYMMDD   = "20060102"   // リクエスト送信はこの形式
	YYYY_MM_DD = "2006-01-02" // 受信データはこの形式
	hour_JST   = 9
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) (string, error) {
	// Load db authentification data
	dbAuth := pkg.Load_db_auth()

	// Open db
	sql_source := dbAuth.User + ":" + dbAuth.Pass + "@(" + dbAuth.Host + ":3306)/" + dbAuth.Name
	db, err := sql.Open("mysql", sql_source)
	if err != nil {
		log.Print("データベースのオープンに失敗")
		panic(err)
	}
	defer db.Close()

	// Get latest date
	latest_date := pkg.Get_latest_date(db, pkg.Table_list[0])
	fmt.Println(latest_date)
	latest_date_time, err := time.Parse(YYYYMMDD, latest_date)
	if err != nil {
		panic(err)
	}

	//// Get opendata of unloaded date
	//　now - latest -> potential unloaded date interval
	now := time.Now().UTC().Add(time.Hour * hour_JST)
	diff_day := int(math.Floor(now.Sub(latest_date_time).Hours() / 24))

	for _, address := range pkg.Address_list {
		for i_day := 1; i_day < diff_day; i_day++ {
			req_date := now.Add(-time.Hour * 24 * time.Duration(i_day)).Format(YYYYMMDD)
			full_address := address + "?date=" + req_date
			req, err := http.NewRequest(http.MethodGet, full_address, nil)
			if err != nil {
				panic(err)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			body, _ := io.ReadAll(res.Body)

			insert_body_npatients(body, db)

		}
	}

	return "", nil
}

func insert_body_npatients(body []byte, db *sql.DB) {
	var data pkg.DataNpatients
	err := json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)

	err = checkErrorFlag(data.ErrorInfo)
	if err != nil {
		panic(err)
	}

	// 時間が余れば Bulk Insertに変更
	insert := "INSERT INTO npatients(date, name_jp, npatients) VALUES (?, ?, ?)"
	for _, k := range data.ItemList {
		prep, err := db.Prepare(insert)
		if err != nil {
			panic(err)
		}
		prep.Exec(k.Date, k.Name_jp, k.Npatients)
		fmt.Println(k.Date, k.Name_jp, k.Npatients)
		prep.Close()
	}
}

func checkErrorFlag(info pkg.Errors) error {
	if info.ErrorFlag != "0" {
		return fmt.Errorf("flag: %s, code: %s, msg: %s ", info.ErrorFlag, info.ErrorCode, info.ErrorMessage)
	}
	return nil
}
