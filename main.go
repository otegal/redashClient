package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const configFile = "./config.json" // 設定ファイルのパス

// Config は設定ファイルを示す構造体
type Config struct {
	APIKey     string `json:"apiKey"`
	BaseURL    string `json:"baseURL"`
	ExportPath string `json:"exportPath"`
	Query      []struct {
		QueryID string `json:"queryId"`
		Params  []struct {
			BaseParam string   `json:"baseParam"`
			SetParams []string `json:"setParams"`
		} `json"params"`
	} `json:"query"`
}

// RedashAPIResponse はRedashAPIのレスポンスを示す構造体。
// RefreshAPIとJobStatusAPIのレスポンス兼用
type RedashAPIResponse struct {
	Job struct {
		Status        int    `json:"status"`
		Errorstatus   string `json:"error"`
		ID            string `json:"id"`
		QueryResultID int    `json:"query_result_id"`
		UpdatedAt     int    `json:"updated_at"`
	} `json:"job"`
}

func getConfig() *Config {
	jsonString, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println(err) // TODO errに入った場合、fatalの方が良さげ。処理止めたい
	}

	conf := new(Config) // new()ではConfigのアドレスである*config型の値（つまりポインタ）を返却する
	err = json.Unmarshal(jsonString, conf)
	if err != nil {
		log.Println(err)
	}

	return conf
}

// クエリ更新のAPIをコールする
func callRefreshAPI(Conf Config, queryID string, baseParam string, setParam string) string {
	fmt.Println(queryID, baseParam, setParam)
	targetURL := fmt.Sprintf("%s/api/queries/%s/refresh?api_key=%s&%s=%s",
		Conf.BaseURL,
		queryID,
		Conf.APIKey,
		baseParam,
		setParam,
	)
	fmt.Println("targetURL is ... " + targetURL)

	client := &http.Client{}
	req, err := http.NewRequest("POST", targetURL, nil)
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// レスポンスを構造体にパースする
	respBody := new(RedashAPIResponse)
	err = json.Unmarshal(body, respBody)
	if err != nil {
		log.Println(err)
	}

	return respBody.Job.ID
}

// リフレッシュジョブの状況を確認するAPIをコールする
func callJobStatusAPI(Conf Config, jobID string) int {
	respBody := new(RedashAPIResponse)
	targetURL := fmt.Sprintf("%s/api/jobs/%s?api_key=%s",
		Conf.BaseURL,
		jobID,
		Conf.APIKey,
	)

	// QueryResultIDがセットされるまでAPIコールし続ける
	for {
		resp, err := http.Get(targetURL)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		// レスポンスを構造体にパースする
		err = json.Unmarshal(body, respBody)
		if err != nil {
			log.Println(err)
		}

		if respBody.Job.QueryResultID != 0 {
			break
		} else {
			log.Println("まだ結果が出ていません。5秒待機した後に再度結果確認します。")
			time.Sleep(5 * time.Second)
		}
	}

	return respBody.Job.QueryResultID
}

// リフレッシュ結果を取得するAPIをコールして結果をファイル書き出す
func callResultAPIAndWriteFile(Conf Config, queryID string, setParam string, queryResultID int) {
	targetURL := fmt.Sprintf("%s/api/queries/%s/results/%s.csv?api_key=%s",
		Conf.BaseURL,
		queryID,
		strconv.Itoa(queryResultID),
		Conf.APIKey,
	)

	resp, err := http.Get(targetURL)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// ファイル書き込み
	exportFileName := fmt.Sprintf("%s/%s/%s.csv", // TODO すでにフォルダが存在しないと書き込みできないのでロジック追加
		Conf.ExportPath,
		queryID,
		setParam,
	)

	file, err := os.Create(exportFileName)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	file.Write(body)
}

func main() {
	// configの読み込み
	Conf := getConfig()

	// Query単位で処理する。 TODO 実装予定。まだ複数指定してもできないよ
	for _, query := range Conf.Query {
		// Params単位で処理する。
		for _, params := range query.Params {
			// SetParams単位で処理する。
			for _, setParam := range params.SetParams {
				// // クエリをリフレッシュするAPIをコール
				jobID := callRefreshAPI(*Conf, query.QueryID, params.BaseParam, setParam)
				fmt.Println(jobID)

				// リフレッシュジョブの状況を確認するAPIをコール
				queryResultID := callJobStatusAPI(*Conf, jobID)
				fmt.Println(queryResultID)

				// リフレッシュ結果を取得するAPIをコールして結果をファイル書き出し
				callResultAPIAndWriteFile(*Conf, query.QueryID, setParam, queryResultID)
			}
		}
	}
}
