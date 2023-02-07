package pkg

type Errors struct {
	ErrorFlag    string `json:"errorFlag"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type DataNpatients struct {
	ErrorInfo Errors          `json:"errorInfo"`
	ItemList  []ItemNpatients `json:"itemList"`
}

type ItemNpatients struct {
	Date      string `json:"date"`
	Name_jp   string `json:"name_jp"`
	Npatients string `json:"npatients"`
}

type DataNdeaths struct {
	ErrorInfo Errors        `json:"errorInfo"`
	ItemList  []ItemNdeaths `json:"itemList"`
}

type ItemNdeaths struct {
	Date    string `json:"date"`
	Ndeaths string `json:"ndeaths"`
}

type DataOverseas struct {
	ErrorInfo Errors         `json:"errorInfo"`
	ItemList  []ItemOverseas `json:"itemList"`
}

type ItemOverseas struct {
	Date        string `json:"date"`
	DataName    string `json:"dataName"`
	InfectedNum string `json:"infectedNum"`
	DeceasedNum string `json:"deceasedNum"`
}

var (
	Defalt_vals = map[string]string{
		Table_content[0]: "20230101",
		Table_content[1]: "東京都",
	}
)

func Load_name_jp_list() []string {
	var name_jp_list = []string{
		"北海道",
		"青森県",
		"岩手県",
		"宮城県",
		"秋田県",
		"山形県",
		"福島県",
		"茨城県",
		"栃木県",
		"群馬県",
		"埼玉県",
		"千葉県",
		"東京都",
		"神奈川県",
		"新潟県",
		"富山県",
		"石川県",
		"福井県",
		"山梨県",
		"長野県",
		"岐阜県",
		"静岡県",
		"愛知県",
		"三重県",
		"滋賀県",
		"京都府",
		"大阪府",
		"兵庫県",
		"奈良県",
		"和歌山県",
		"鳥取県",
		"島根県",
		"岡山県",
		"広島県",
		"山口県",
		"徳島県",
		"香川県",
		"愛媛県",
		"高知県",
		"福岡県",
		"佐賀県",
		"長崎県",
		"熊本県",
		"大分県",
		"宮崎県",
		"鹿児島県",
		"沖縄県",
	}
	return name_jp_list
}
