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
		QueryID    string     `json:"queryId"`
		BaseParams []string   `json:"baseParams"`
		SetParams  [][]string `json;"setParams"`
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
		log.Fatalln(err)
	}

	conf := new(Config) // new()ではConfigのアドレスである*config型の値（つまりポインタ）を返却する
	err = json.Unmarshal(jsonString, conf)
	if err != nil {
		log.Fatalln(err)
	}

	return conf
}

// クエリ更新のAPIをコールする
func callRefreshAPI(Conf Config, queryID string, requestParam string) string {
	targetURL := fmt.Sprintf("%s/api/queries/%s/refresh?api_key=%s%s",
		Conf.BaseURL,
		queryID,
		Conf.APIKey,
		requestParam,
	)
	fmt.Println("targetURL is ... " + targetURL)

	client := &http.Client{}
	req, err := http.NewRequest("POST", targetURL, nil)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// レスポンスを構造体にパースする
	respBody := new(RedashAPIResponse)
	err = json.Unmarshal(body, respBody)
	if err != nil {
		log.Fatalln(err)
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
			log.Fatalln(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		// レスポンスを構造体にパースする
		err = json.Unmarshal(body, respBody)
		if err != nil {
			log.Fatalln(err)
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
func callResultAPIAndWriteFile(Conf Config, queryID string, requestParam string, queryResultID int) {
	targetURL := fmt.Sprintf("%s/api/queries/%s/results/%s.csv?api_key=%s",
		Conf.BaseURL,
		queryID,
		strconv.Itoa(queryResultID),
		Conf.APIKey,
	)

	resp, err := http.Get(targetURL)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// 以下はファイル出力処理
	exportDir := fmt.Sprintf("%s/%s",
		Conf.ExportPath,
		queryID,
	)

	// export先のディレクトリが存在しない場合は作成する
	if _, err := os.Stat(exportDir); err != nil {
		os.Mkdir(exportDir, 0777)
	}

	// ファイル書き込み
	exportFileName := fmt.Sprintf("%s/%s.csv",
		exportDir,
		requestParam,
	)

	file, err := os.Create(exportFileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	file.Write(body)
}

func main() {
	// configの読み込み
	Conf := getConfig()

	// Query単位で処理する
	for _, query := range Conf.Query {
		for _, setParam := range query.SetParams {
			// リクエストパラメタを作る
			requestParam := ""
			for index, baseParam := range query.BaseParams {
				requestParam += "&" + baseParam + "=" + setParam[index]
			}

			// クエリをリフレッシュするAPIをコール
			jobID := callRefreshAPI(*Conf, query.QueryID, requestParam)
			fmt.Println(jobID)

			// リフレッシュジョブの状況を確認するAPIをコール
			queryResultID := callJobStatusAPI(*Conf, jobID)
			fmt.Println(queryResultID)

			// リフレッシュ結果を取得するAPIをコールして結果をファイル書き出し
			callResultAPIAndWriteFile(*Conf, query.QueryID, requestParam, queryResultID)
		}
	}
}
