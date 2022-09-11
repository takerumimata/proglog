# proglog

Go 言語による分散サービス

```zsh
# ログにレコードを追加
$ curl -X POST localhost:8080 -d '{"record": {"value": "hoge"}}'
# ログを取得. -Xでmethod指定せずにrequestを送るとcurlの仕様的にpostになってしまうので注意（-dをつけてbodyを入れるとそうなる）
$ curl -X GET localhost:8080 -d '{"offset": 0}'
```
