package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/joho/godotenv"
	"sam-app-ca-07-covid.root/pkg"

	_ "github.com/go-sql-driver/mysql"
)

var (
	address_list = []string{
		"https://opendata.corona.go.jp/api/Covid19JapanAll",
		//Add opendata URL here
	}

	table_name_list = []string{
		"npatients",
		//Add table name here
	}

	id_sql   = "id INT NOT NULL AUTO_INCREMENT PRIMARY KEY"
	date_sql = "date DATE NOT NULL"

	table_content_npatients = []string{
		id_sql,
		date_sql,
		"name_jp VARCHAR(20) NOT NULL",
		"npatients INT NOT NULL",
	}
)

func main() {
	// Load db authentification data
	godotenv.Load("db.env")
	db_auth := pkg.Load_db_auth()

	// Open db
	sql_source := db_auth.User + ":" + db_auth.Pass + "@(" + db_auth.Host + ":3306)/" + db_auth.Name
	db, err := sql.Open("mysql", sql_source)
	if err != nil {
		log.Print("データベースのオープンに失敗")
		panic(err)
	}
	defer db.Close()

	// Create table
	create_table(db, table_name_list[0], table_content_npatients) // no reuse
	fmt.Println("table created")

	// Load and insert opendata with concurrent processing
	// @see https://qiita.com/suba-ru/items/4e7341a53142e005472a :並行処理と無名関数
	var wg sync.WaitGroup
	f := func(i int) { //@see https://qiita.com/sudix/items/67d4cad08fe88dcb9a6d
		defer wg.Done()
		fmt.Printf("%v: started  \n", i)

		body := load_opendata(address_list[i], db)
		fmt.Printf("%v: loaded  \n", i)

		insert_body_npatients(body, db)
		fmt.Printf("%v: inserted\n", i)
	}

	for i := 0; i < len(address_list); i++ {
		wg.Add(1)
		go f(i)
	}
	wg.Wait()
	fmt.Println("done")
}

func create_table(db *sql.DB, table_name string, table_content []string) {
	// クエリ生成. table_contentを入れていく
	sql_create := "CREATE TABLE IF NOT EXISTS " + table_name + "("
	for _, v := range table_content {
		sql_create = sql_create + v + ", "
	}
	sql_create = sql_create[:len(sql_create)-2]
	sql_create = sql_create + ")"
	fmt.Println(sql_create)

	// クエリ実行
	_, err := db.Exec(sql_create)
	if err != nil {
		panic(err)
	}
}

func load_opendata(full_address string, db *sql.DB) []byte {
	// リクエスト生成
	req, err := http.NewRequest(http.MethodGet, full_address, nil)
	if err != nil {
		panic(err)
	}

	// リクエスト実行
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// body解釈
	body, _ := io.ReadAll(resp.Body)

	return body
}

func insert_body_npatients(body []byte, db *sql.DB) {
	var data pkg.DataNpatients
	err := json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	err = checkErrorFlag(data.ErrorInfo)
	if err != nil {
		panic(err)
	}

	insert := "INSERT INTO npatients(date, name_jp, npatients) VALUES (?, ?, ?)"
	for _, k := range data.ItemList {
		prep, err := db.Prepare(insert)
		if err != nil {
			panic(err)
		}
		prep.Exec(k.Date, k.Name_jp, k.Npatients)
		prep.Close()
	}
}

func checkErrorFlag(info pkg.Errors) error {
	if info.ErrorFlag != "0" {
		return fmt.Errorf("flag: %s, code: %s, msg: %s ", info.ErrorFlag, info.ErrorCode, info.ErrorMessage)
	}
	return nil
}
