package step2


import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type ExternalAddressResponse struct {
	Response struct {
		Location []struct {
			City       string `json:"city"`
			Town       string `json:"town"`
			Prefecture string `json:"prefecture"`
			Postal     string `json:"postal"`
			X          string `json:"x"`
			Y          string `json:"y"`
		} `json:"location"`
	} `json:"response"`
}

type AddressResponse struct {
	PostalCode       string  `json:"postal_code"`
	HitCount         int     `json:"hit_count"`
	Address          string  `json:"address"`
	TokyoStationDist float64 `json:"tokyo_sta_distance"`
}

func HaversineDistance(y, x, yt, xt float64) float64 {
	// 地球の半径 (km)
	R := 6371.0

	distance := (math.Pi * R / 180) * math.Sqrt(math.Pow((x-xt)*math.Cos(math.Pi*(y+yt)/360), 2) + math.Pow(y-yt, 2))

	return distance
}

func GetAddressFromExternalAPI(postalCode string) (*ExternalAddressResponse, error) {
	apiURL := fmt.Sprintf("https://geoapi.heartrails.com/api/json?method=searchByPostal&postal=%s", postalCode)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var addressResponse ExternalAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&addressResponse); err != nil {
		return nil, err
	}

	return &addressResponse, nil
}

func FindCommonAddress(strings []string) string {
	if len(strings) == 0 {
		return ""
	}

	shortest := strings[0]
	for _, str := range strings {
		if len(str) < len(shortest) {
			shortest = str
		}
	}


	// 共通部分を見つける
	for i := 0; i < len(shortest); i += 3 {
		address := shortest[:i+3]
		for _, str := range strings {
			if !IsAddress(address, str) {
				return shortest[:i]
			}
		}
	}

	return shortest
}

func IsAddress(address, str string) bool {
	return str[:len(address)] == address
}

func AddressHandler(w http.ResponseWriter, r *http.Request) {
	postalCode := r.URL.Query().Get("postal_code")
	if postalCode == "" {
		http.Error(w, "Postal code is required", http.StatusBadRequest)
		return
	}

	// SQLite3 データベースに接続
    db, err := sql.Open("sqlite3", "./access_logs.db")
    if err != nil {
        http.Error(w, "Failed to log access", http.StatusInternalServerError)
        return
    }
    defer db.Close()

    // アクセスログをデータベースに保存
    _, err = db.Exec("INSERT INTO access_logs (postal_code) VALUES (?)", postalCode)
    if err != nil {
        http.Error(w, "Failed to log access", http.StatusInternalServerError)
        return
    }

	externalAddress, err := GetAddressFromExternalAPI(postalCode)
	if err != nil {
		http.Error(w, "Failed to fetch address from external API", http.StatusInternalServerError)
		return
	}

	var maxDistance float64
	maxDistance = 0.0

	for _, loc := range externalAddress.Response.Location {
		y, _ := strconv.ParseFloat(loc.Y, 64)
		x, _ := strconv.ParseFloat(loc.X, 64)

		distance := HaversineDistance(y, x, 35.6809591, 139.7673068)

		if distance > maxDistance {
			maxDistance = distance
		}
	}

	var townNames []string
	for _, loc := range externalAddress.Response.Location {
		townNames = append(townNames, loc.Town)
	}
	
	commonTown := FindCommonAddress(townNames)

	address := &AddressResponse{
		PostalCode:       postalCode,
		HitCount:         len(externalAddress.Response.Location),
		Address:          externalAddress.Response.Location[0].Prefecture + externalAddress.Response.Location[0].City + commonTown,
		TokyoStationDist: math.Round(maxDistance*10) / 10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(address)
}