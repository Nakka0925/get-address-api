FROM golang:1.17-alpine

RUN apk add --no-cache gcc libc-dev

WORKDIR /get-address-api

COPY go.mod step2/go.mod step3/go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app

EXPOSE 8080

CMD ["./app"]
