# pgMinRO

## 概要

* PostgreSQLの参照向けクライアント
* ブラウザ上から任意のSELECT文を実行して結果を表示することを主目的とした軽量ツール
* 参照向けの調整になっているだけで更新も可能
* SQLドライバは画面上で選択（ `pq` or `pgx` ）

## 動作要件

* Go（version 1.23以降）がインストールされていること

例

```ShellSession
$ go version
go version go1.23.1 linux/amd64
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

もしくは

```Shell
`go env GOPATH`/bin/pgminro
```

### ブラウザからアクセス

<http://localhost:8432/>

### 停止

プロセスの停止は Ctrl + C
