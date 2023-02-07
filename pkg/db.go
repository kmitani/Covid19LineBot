package pkg

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Address_list = []string{
		"https://opendata.corona.go.jp/api/Covid19JapanAll",
		//Add URL here
	}

	Table_list = []string{
		"npatients",
		//Add table name here
	}

	Table_content = []string{"date", "name_jp", "npatients"}
)

type DB struct {
	User string
	Pass string
	Host string
	Name string
}

func Load_db_auth() DB {
	var db DB
	db.User = os.Getenv("DB_USER")
	db.Pass = os.Getenv("DB_PASS")
	db.Host = os.Getenv("DB_HOST")
	db.Name = os.Getenv("DB_NAME")
	return db
}

func Db_check_exit(db *sql.DB, table_name string, param_name []string, params map[string]string) (int, error) {

	sql_exists := "SELECT EXISTS( SELECT * FROM " + table_name +
		" WHERE " + param_name[0] + "= ? AND " + param_name[1] + " = ?)"
	fmt.Println(sql_exists)
	rows, err := db.Query(sql_exists, params[param_name[0]], params[param_name[1]])
	if err != nil {
		return 0, err
	}

	var isExist int
	rows.Next()
	err = rows.Scan(&isExist)
	return isExist, nil
}

func Db_get_npatients(db *sql.DB, req_date string, param_name []string, req_vals map[string]string, days_back int, table_name string) ([]int, error) {

	YYYYMMDD := "20060102" //DBに渡す時はこの形式
	// YYYY_MM_DD := "2006-01-02" //DBから受ける時はこの形式

	req_date_time, err := time.Parse(YYYYMMDD, req_vals["date"])
	if err != nil {
		return nil, err
	}
	// 遡る日数を計算
	req_date_past := req_date_time.Add(-time.Hour * 24 * time.Duration(days_back)).Format(YYYYMMDD)

	// Queryを生成・実行
	sql_select := "SELECT " + param_name[2] + " FROM " + table_name +
		" WHERE (date <= ? AND date >= ?)" +
		" AND " + param_name[1] + " = ? ORDER BY date DESC" //tableごとにdateの形式が違うので注意
	rows, err := db.Query(sql_select, req_vals["date"], req_date_past, req_vals[param_name[1]])
	if err != nil {
		return nil, err
	}

	// sql.Rowsをスライス[data]に入れる
	var data []int
	for rows.Next() {
		var data_i int
		err = rows.Scan(&data_i)
		if err != nil {
			return nil, err
		}
		data = append(data, data_i)
	}
	if len(data) < 30 {
		err = errors.New("short data length")
		return nil, err
	}

	return data, nil
}

func Db_get_npatients_all(db *sql.DB, req_date string, param_name []string, days_back int, table_name string) ([]int, error) {

	YYYYMMDD := "20060102" //DBに渡す時はこの形式
	// YYYY_MM_DD := "2006-01-02" //DBから受ける時はこの形式

	req_date_time, err := time.Parse(YYYYMMDD, req_date)
	if err != nil {
		return nil, err
	}
	// 遡る日数を計算
	var req_date_list []string
	req_date_list = append(req_date_list, req_date)
	tmp_time := req_date_time
	for i := 1; i < days_back; i++ {
		tmp_time = tmp_time.Add(-24 * time.Hour)
		v := tmp_time.Format(YYYYMMDD)
		req_date_list = append(req_date_list, v)
	}

	// Queryを生成・実行
	var npatients []int
	for _, v := range req_date_list {
		var data []int
		sql_select := "SELECT " + param_name[2] + " FROM " + table_name +
			" WHERE (date = ?)" //tableごとにdateの形式が違うので注意
		rows, err := db.Query(sql_select, v)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var data_i int
			err = rows.Scan(&data_i)
			if err != nil {
				return nil, err
			}
			data = append(data, data_i)
		}
		v, _ := Calc_sum(data)
		npatients = append(npatients, v)
	}

	return npatients, nil
}

func Get_latest_date(db *sql.DB, table_name string) string {
	// YYYYMMDD形式で出力
	sql_recent := "SELECT MAX(date) FROM " + table_name
	rows, _ := db.Query(sql_recent)
	rows.Next()
	var recent_date string
	rows.Scan(&recent_date)
	recent_date = recent_date[:4] + recent_date[5:7] + recent_date[8:]
	return recent_date
}
