package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"sam-app-ca-07-covid.root/pkg"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) (string, error) {
	// Load db authentification data
	db_auth := pkg.Load_db_auth()

	// Open db
	sql_source := db_auth.User + ":" + db_auth.Pass + "@(" + db_auth.Host + ":3306)/" + db_auth.Name
	db, err := sql.Open("mysql", sql_source)
	if err != nil {
		log.Print("データベースのオープンに失敗")
		panic(err)
	}
	defer db.Close()

	// Get latest
	latest_date := pkg.Get_latest_date(db, pkg.Table_list[0])

	// Get 30 days npatients data
	days_back := 30
	npatients_all, err := pkg.Db_get_npatients_all(db, latest_date, pkg.Table_content, days_back, pkg.Table_list[0])
	if err != nil {
		fmt.Println(err)
	}
	npatients_new := pkg.Calc_diff(npatients_all)

	// 週平均を計算
	average_npatients_new, err := pkg.Calc_week_average(npatients_new)
	if err != nil {
		fmt.Println(err)
	}

	// 前日比を計算
	ratio_npatients := pkg.Calc_ratio(average_npatients_new)

	// 前日比平均を計算
	days_calc := 3     // days_calc日分の平均を計算
	days_geomean := 14 // 平均に使う日数
	var geomean_ratio []float64
	for i := 0; i < days_calc; i++ {
		v := pkg.Calc_geometric_mean(ratio_npatients[i : days_geomean+i])
		geomean_ratio = append(geomean_ratio, v)
	}

	// 閾値で切って状態判定
	// state_infection{1: increase, 0: stable, -1: decrease}
	threshold := 0.01
	var state_infection []int
	for i := 0; i < days_calc; i++ {
		tmp_state := calc_state(geomean_ratio[i], threshold)
		state_infection = append(state_infection, tmp_state)
	}
	is_state_change := state_infection[0] - state_infection[1]

	// 通知
	if is_state_change >= 1 || is_state_change <= -1 {
		req_vals := map[string]string{
			"date":    latest_date,
			"name_jp": "全国",
		}
		text := pkg.MakeLineMessageNotification(req_vals, npatients_new[0], average_npatients_new[0], geomean_ratio, is_state_change)
		userid := ""
		line_auth := pkg.Load_line_auth()
		// 送信元ユーザにメッセージを返信する関数
		pkg.PostLineMessage(userid, text, line_auth.Secret, line_auth.Token)
	}
	return "", nil
}

func calc_state(geomean_ratio float64, threshold float64) int {
	var state_infection int
	is_increase := geomean_ratio > (1 + threshold)
	is_decrease := geomean_ratio < (1 - threshold)
	if is_increase {
		state_infection = 1
	} else if is_decrease {
		state_infection = -1
	} else {
		state_infection = 0
	}
	return state_infection
}
