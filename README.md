# maiyumi

## 実行

```shell
go build
./maiyumi
```

## ダンプ

```shell
go run scripts/dump_db.go
sqlite3 data.db < dump/xxxx.sql
```

## 本番

```shell
GOOS=linux GOARCH=amd64 go build -o maiyumi
tar czf maiyumi-deploy.tar.gz maiyumi templates/
```
