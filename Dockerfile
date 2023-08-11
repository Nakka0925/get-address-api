# ベースイメージを指定
FROM golang:1.17-alpine

RUN apk add --no-cache gcc libc-dev

# 作業ディレクトリを設定
WORKDIR /get-address-api

# Goモジュールを初期化
COPY go.mod step2/go.mod step3/go.mod go.sum ./
RUN go mod download

# プロジェクトファイルをコンテナ内にコピー
COPY . .

# アプリケーションのビルド
RUN go build -o app

# ポートを公開
EXPOSE 8080

# コンテナ起動時に実行されるコマンド
CMD ["./app"]
