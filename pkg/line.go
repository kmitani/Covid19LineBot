package pkg

import (
	"fmt"
	"os"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"
)

type LineAuth struct {
	Secret string
	Token  string
}

type LineMessage struct {
	Destination string `json:"destination"`
	Events      []struct {
		ReplyToken string `json:"replyToken"`
		Type       string `json:"type"`
		Mode       string `json:"mode"`
		Timestamp  int64  `json:"timestamp"`
		Source     struct {
			Type   string `json:"type"`
			UserID string `json:"userId"`
		} `json:"source"`
		WebhookEventID  string `json:"webhookEventId"`
		DeliveryContext struct {
			IsRedelivery bool `json:"isRedelivery"`
		} `json:"deliveryContext"`
		Message struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Text   string `json:"text"`
			Emojis []struct {
				Index     int    `json:"index"`
				Length    int    `json:"length"`
				ProductID string `json:"productId"`
				EmojiID   string `json:"emojiId"`
			} `json:"emojis"`
			Mention struct {
				Mentionees []struct {
					Index  int    `json:"index"`
					Length int    `json:"length"`
					UserID string `json:"userId"`
				} `json:"mentionees"`
			} `json:"mention"`
		} `json:"message"`
	} `json:"events"`
}

func Load_line_auth() LineAuth {
	var line_auth LineAuth
	line_auth.Secret = os.Getenv("CHANNEL_SECRET")
	line_auth.Token = os.Getenv("CHANNEL_TOKEN")
	return line_auth
}

func MakeLineMessage(req_vals map[string]string, npatients_new int, average_npatients_new int, geomean_ratio float64) string {
	res_date := req_vals["date"][0:4] + "年" + req_vals["date"][4:6] + "月" + req_vals["date"][6:] + "日"
	res_name_jp := req_vals["name_jp"]
	var isIncrease string
	if geomean_ratio >= 1 {
		isIncrease = "増加"
	} else if geomean_ratio < 1 {
		isIncrease = "減少"
	}
	geomean_ratio_str := strconv.FormatFloat(geomean_ratio, 'f', 3, 64)
	text := fmt.Sprintf(
		"[Request]%s, %sの感染状況をお知らせします。新規感染者数が%sしています。新規感染者数は%d人、週平均は%d人、感染拡大率は%sでした。",
		res_date, res_name_jp, isIncrease, npatients_new, average_npatients_new, geomean_ratio_str)

	return text
}
func MakeLineMessageNotification(req_vals map[string]string, npatients_new int, average_npatients_new int, geomean_ratio []float64, isChangeChange int) string {
	res_date := req_vals["date"][0:4] + "年" + req_vals["date"][4:6] + "月" + req_vals["date"][6:] + "日"
	res_name_jp := req_vals["name_jp"]
	var isIncrease string
	if isChangeChange >= 1 {
		isIncrease = "増加"
	} else if isChangeChange < -1 {
		isIncrease = "減少"
	}
	geomean_ratio_str := strconv.FormatFloat(geomean_ratio[0], 'f', 3, 64)
	geomean_ratio_str_yesterday := strconv.FormatFloat(geomean_ratio[1], 'f', 3, 64)
	text := fmt.Sprintf("[Notification]%s, %sの新規感染者数が%sする兆しがあります。前日の感染拡大率は%sでしたが、本日は%sでした。今後の感染状況に注意しましょう。新規感染者数は%d人、週平均は%d人でした。各都道府県の感染状況を知りたい方は都道府県名を入力してください。",
		res_date, res_name_jp, isIncrease, geomean_ratio_str_yesterday, geomean_ratio_str, npatients_new, average_npatients_new)

	return text
}

func PostLineMessage(userid string, text string, CHANNEL_SECRET string, CHANNEL_TOKEN string) error {

	bot, err := linebot.New(CHANNEL_SECRET, CHANNEL_TOKEN)
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
