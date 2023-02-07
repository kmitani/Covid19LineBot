package main

// リクエストに応じて返信するAPI

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/line/line-bot-sdk-go/linebot"
	"sam-app-ca-07-covid.root/pkg"
)

// 送信データ
type SendData struct {
	Date          string  `json:"date"`
	Name_jp       string  `json:"name_jp"`
	NpatientsNew  int     `json:"npatientsNew"`
	RatioPatients float64 `json:"ratioPatients"`
}

var (
	req_vals   = pkg.Defalt_vals
	line_event pkg.LineMessage

	name_jp_list = pkg.Load_name_jp_list()
)

const (
	YYYYMMDD   = "20060102"   //DBに渡す時はこの形式
	YYYY_MM_DD = "2006-01-02" //Opendata側はこっちの形式
)

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load db authentification data
	db_auth := pkg.Load_db_auth()

	// Open db
	sql_source := db_auth.User + ":" + db_auth.Pass + "@(" + db_auth.Host + ":3306)/" + db_auth.Name
	db, err := sql.Open("mysql", sql_source)
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 500,
		}, err
	}
	defer db.Close()

	// Get data of latest date
	latest_date := pkg.Get_latest_date(db, pkg.Table_list[0])
	req_vals["date"] = latest_date
	err = json.Unmarshal([]byte(req.Body), &line_event) // LineのJSONをstructに
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 500,
		}, err
	}
	req_vals = line_message_interpret(line_event.Events[0].Message.Text, db)

	// リクエストに応じてデータ読み出し
	// データの存在チェック.
	is_exist, err := pkg.Db_check_exit(db, pkg.Table_list[0], pkg.Table_content, req_vals)
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 500,
		}, err
	}
	if is_exist != 1 {
		err = errors.New("no data")
		return events.APIGatewayProxyResponse{
			Body:       "No data", //エラーメッセージを含むJSONの方がいいかも
			StatusCode: 500,
		}, err
	}

	// days_back日分のデータを取得
	days_back := 30
	npatients, err := pkg.Db_get_npatients(db, req_vals["date"], pkg.Table_content, req_vals, days_back, pkg.Table_list[0])
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 500,
		}, err
	}
	npatients_new := pkg.Calc_diff(npatients)

	// 週平均を計算
	average_npatients_new, err := pkg.Calc_week_average(npatients_new)
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 500,
		}, err
	}

	// 前日比を計算
	ratio_npatients := pkg.Calc_ratio(average_npatients_new)

	// 前日比平均を計算
	days_geomean := 14
	geomean_ratio := pkg.Calc_geometric_mean(ratio_npatients[0:days_geomean])

	if req.HTTPMethod == "POST" { // POSTの時にLINEメッセージを返す
		// text作成
		text := pkg.MakeLineMessage(req_vals, npatients_new[0], average_npatients_new[0], geomean_ratio)

		// destinationが存在しなかった場合の例外処理. 不要?
		var userid string
		if line_event.Destination != "" {
			userid = fmt.Sprintf("%v", line_event.Events[0].Source.UserID)
		} else {
			userid = ""
		}

		// 送信元ユーザにメッセージを返信
		postLineMessage(userid, text)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 200,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 500,
	}, nil
}

func postLineMessage(userid string, text string) error {
	line_auth := pkg.Load_line_auth()
	bot, err := linebot.New(line_auth.Secret, line_auth.Token)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if userid != "" {
		_, err = bot.PushMessage(userid, linebot.NewTextMessage(text)).Do()
	} else {
		_, err = bot.BroadcastMessage(linebot.NewTextMessage(text)).Do()
	}

	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func line_message_interpret(text string, db *sql.DB) map[string]string {
	// 都道府県リストとメッセージを照合
	for _, v := range name_jp_list {
		if strings.Contains(text, v) {
			req_vals["name_jp"] = v
		}
	}
	// @todo 時間照合

	return req_vals
}
