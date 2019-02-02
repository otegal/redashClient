# redashClient
RedashのAPIをつかってcsvを取得するためのツール

# 利用方法
configに設定値を入れて実行するだけ。

# configの説明
`config.example.json`を`config.json`にして使うこと。


### 基本設定
```json
{
    "apiKey": "Set your API Key", //RedashのUserAPIKeyをセットすること
    "baseUrl": "http://hogehoge.example.com", //利用する　RedashのURLを設定すること
    "exportPath": "/Users/h_hiroki/Desktop", //取得したCSVの出力先を指定できる
    "query": [
        {
            "queryId": "1106" // queryキーは後述する
        }
    ]
}
```

### queryキーについて
queryキーで対象となるクエリとパラメタの設定を行う。
以下にいくつか設定例を記す。

クエリパラメタが存在しないクエリのcsvを取得する
```json
{
    "apiKey": "Set your API Key",
    "baseUrl": "http://hogehoge.example.com",
    "exportPath": "/Users/h_hiroki/Desktop",
    "query": [
        {
            "queryId": "1106"
        }
    ]
}
```

クエリパラメタが1つだけ存在するクエリのcsvを取得する
```json
{
    "apiKey": "Set your API Key",
    "baseUrl": "http://hogehoge.example.com",
    "exportPath": "/Users/h_hiroki/Desktop",
    "query": [
        {
            "queryId": "1080",
            "baseParams": [
                "p_yyyy-mm"
            ],
            "setParams": [
                ["2018-12"],
                ["2018-11"],
                ["2018-10"]
            ]
        }
    ]
}
```

クエリパラメタが2つ存在するクエリのcsvを取得する
```json
{
    "apiKey": "Set your API Key",
    "baseUrl": "http://hogehoge.example.com",
    "exportPath": "/Users/h_hiroki/Desktop",
    "query": [
        {
            "queryId": "1107",
            "baseParams": [
                "p_start",
                "p_end"
            ],
            "setParams": [
                ["2018-12-01", "2018-12-01"],
                ["2018-12-02", "2018-12-02"],
                ["2018-12-03", "2018-12-03"]
            ]
        }
    ]
}
```


複数のクエリに対してcsvを取得する。
以下の場合、3つのクエリのCSVを取得できる。
```json
{
    "apiKey": "Set your API Key",
    "baseUrl": "http://hogehoge.example.com",
    "exportPath": "/Users/h_hiroki/Desktop",
    "query": [
        {
            "queryId": "1106"
        },
        {
            "queryId": "1107",
            "baseParams": [
                "p_start",
                "p_end"
            ],
            "setParams": [
                ["2018-12-01", "2018-12-01"],
                ["2018-12-02", "2018-12-02"],
                ["2018-12-03", "2018-12-03"]
            ]
        },
        {
            "queryId": "1080",
            "baseParams": [
                "p_yyyy-mm"
            ],
            "setParams": [
                ["2018-12"],
                ["2018-11"],
                ["2018-10"]
            ]
        }
    ]
}
```



# 参考記事
設定ファイルをjsonで書く記事　http://m-shige1979.hatenablog.com/entry/2017/10/30/080000
