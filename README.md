# pgMinRO

## 概要

* PostgreSQLの参照専用クライアント
* ブラウザ上から任意のSELECT文を実行して、結果を表示する軽量ツール

## 画面遷移

```mermaid
flowchart LR
    connect[DB接続] --> query[クエリー]
    query --> query
```

## 使い方

### 依存モジュールの準備

```Shell
go mod tidy
```

### 起動

```Shell
go run .
```

### ブラウザからアクセス

http://localhost:8432/
