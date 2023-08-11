package step3


import (
	"encoding/json"
	"net/http"
	"sort"
	"database/sql"
    _ "github.com/mattn/go-sqlite3"
)


type AccessLog struct {
    ID         int    `json:"id"`
    PostalCode string `json:"postal_code"`
    CreatedAt  string `json:"created_at"`
}

type AccessLogAggregate struct {
    PostalCode   string `json:"postal_code"`
    RequestCount int    `json:"request_count"`
}


func aggregateAccessLogs(accessLogs []*AccessLog) map[string]int {
    requestCountMap := make(map[string]int)
    for _, log := range accessLogs {
        requestCountMap[log.PostalCode]++
    }
    return requestCountMap
}

func sortAccessLogsByRequestCount(accessLogs []*AccessLog) []AccessLogAggregate {
    requestCountMap := aggregateAccessLogs(accessLogs)

    var aggregates []AccessLogAggregate
    for postalCode, requestCount := range requestCountMap {
        aggregates = append(aggregates, AccessLogAggregate{
            PostalCode:   postalCode,
            RequestCount: requestCount,
        })
    }

    // リクエスト回数の降順でソート
    sort.Slice(aggregates, func(i, j int) bool {
        return aggregates[i].RequestCount > aggregates[j].RequestCount
    })

    return aggregates
}

func getAccessLogsFromDatabase() ([]*AccessLog, error) {
    db, err := sql.Open("sqlite3", "./access_logs.db")
    if err != nil {
        return nil, err
    }
    defer db.Close()

    rows, err := db.Query("SELECT id, postal_code, created_at FROM access_logs")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var accessLogs []*AccessLog
    for rows.Next() {
        var log AccessLog
        if err := rows.Scan(&log.ID, &log.PostalCode, &log.CreatedAt); err != nil {
            return nil, err
        }
        accessLogs = append(accessLogs, &log)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return accessLogs, nil
}


func AccessLogsHandler(w http.ResponseWriter, r *http.Request) {
    // データベースからアクセスログを取得（省略）
	accessLogs, err := getAccessLogsFromDatabase()
    if err != nil {
        http.Error(w, "Failed to fetch access logs", http.StatusInternalServerError)
        return
    }

    // アクセスログをリクエスト回数ごとに集計し、ソート
    sortedAggregates := sortAccessLogsByRequestCount(accessLogs)

    // レスポンスを構築
    response := struct {
        AccessLogs []AccessLogAggregate `json:"access_logs"`
    }{
        AccessLogs: sortedAggregates,
    }

    // JSON形式でレスポンスを返す
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
