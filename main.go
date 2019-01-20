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
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseURL"`
	Query   []struct {
		QueryID string `json:"queryId"`
		Params  string `json:"params"`
	} `json:"query"`
}

// RefreshAPIResponse はクエリをリフレッシュするAPIのレスポンスを示す構造体
type RefreshAPIResponse struct {
	Job struct {
		Status        int    `json:"status"`
		Errorstatus   string `json:"error"`
		ID            string `json:"id"`
		QueryResultID int    `json:"query_result_id"`
		UpdatedAt     int    `json:"updated_at"`
	} `json:"job"`
}

// TODO 上記構造体と同じなのでRedashAPIRequestとかの名前で1つにまとめる
// JobStatusAPIResponse はクエリをリフレッシュするAPIのレスポンスを示す構造体
type JobStatusAPIResponse struct {
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
		log.Println(err)
	}

	conf := new(Config)
	err = json.Unmarshal(jsonString, conf)
	if err != nil {
		log.Println(err)
	}

	return conf
}

// クエリ更新のAPIをコールする
func callRefreshAPI(Conf Config) string {
	// redashのクエリパラメタ仮実装。
	targetYearMonth := time.Now().Format("2006-01")

	targetURL := fmt.Sprintf("%s/api/queries/%s/refresh?api_key=%s&%s=%s",
		Conf.BaseURL,
		Conf.Query[0].QueryID,
		Conf.APIKey,
		Conf.Query[0].Params,
		targetYearMonth,
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
	respBody := new(RefreshAPIResponse)
	err = json.Unmarshal(body, respBody)
	if err != nil {
		log.Println(err)
	}

	return respBody.Job.ID
}

// リフレッシュジョブの状況を確認するAPIをコールする
func callJobStatusAPI(Conf Config, jobID string) int {
	respBody := new(JobStatusAPIResponse)
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
func callResultAPIAndWriteFile(Conf Config, queryResultID int) {
	targetURL := fmt.Sprintf("%s/api/queries/%s/results/%s.csv?api_key=%s",
		Conf.BaseURL,
		Conf.Query[0].QueryID,
		strconv.Itoa(queryResultID),
		Conf.APIKey,
	)

	resp, err := http.Get(targetURL)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// ファイル書き込み。
	// TODO ファイル書き出し先のパスも指定できるようにする
	file, err := os.Create("./test.csv") // TODO 危ないのでgit管理外のディレクトリに出力するようにする
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	file.Write(body)
}

func main() {
	// configの読み込み
	Conf := getConfig()

	// クエリをリフレッシュするAPIをコール
	jobID := callRefreshAPI(*Conf)
	fmt.Println(jobID)

	// リフレッシュジョブの状況を確認するAPIをコール
	queryResultID := callJobStatusAPI(*Conf, jobID)
	fmt.Println(queryResultID)

	// リフレッシュ結果を取得するAPIをコールして結果をファイル書き出し
	callResultAPIAndWriteFile(*Conf, queryResultID)
}
