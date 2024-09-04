# pgMinRO

## 概要

* PostgreSQLの参照専用クライアント
* ブラウザ上から任意のSELECT文を実行して、結果を表示する軽量ツール

## 動作要件

* Go（version 1.23以降）がインストールされていること

例

```ShellSession
$ go version
go version go1.23.0 linux/amd64
```

## インストール

```Shell
go install github.com/ystkg/pgminro@latest
```

## 使い方

### 起動

```Shell
pgminro
```

* `PATH` に `$GOPATH/bin` が設定されていること
* プロセスの停止は Ctrl + C

### ブラウザからアクセス

<http://localhost:8432/>
