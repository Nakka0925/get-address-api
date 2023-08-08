package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
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

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 地球の半径 (km)
	R := 6371.0

	// 緯度・経度の差をラジアンに変換
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	// 距離の計算
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

func getAddressFromExternalAPI(postalCode string) (*ExternalAddressResponse, error) {
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

func findCommonAddress(strings []string) string {
	if len(strings) == 0 {
		return ""
	}

	// Find the shortest string in the slice
	shortest := strings[0]
	for _, str := range strings {
		if len(str) < len(shortest) {
			shortest = str
		}
	}

	// Find the common Address
	for i := 0; i < len(shortest); i += 3 {
		address := shortest[:i+3]
		for _, str := range strings {
			if !isAddress(address, str) {
				return shortest[:i]
			}
		}
	}

	return shortest
}

func isAddress(address, str string) bool {
	return len(str) >= len(address) && str[:len(address)] == address
}

func addressHandler(w http.ResponseWriter, r *http.Request) {
	postalCode := r.URL.Query().Get("postal_code")
	if postalCode == "" {
		http.Error(w, "Postal code is required", http.StatusBadRequest)
		return
	}

	externalAddress, err := getAddressFromExternalAPI(postalCode)
	if err != nil {
		http.Error(w, "Failed to fetch address from external API", http.StatusInternalServerError)
		return
	}

	var maxDistance float64
	var farthestAddress ExternalAddressResponse

	for _, loc := range externalAddress.Response.Location {
		lat, _ := strconv.ParseFloat(loc.Y, 64)
		lon, _ := strconv.ParseFloat(loc.X, 64)

		distance := haversineDistance(lat, lon, 35.6809591, 139.7673068)
		if distance > maxDistance {
			maxDistance = distance
			farthestAddress = *externalAddress
		}
	}

	var townNames []string
	for _, loc := range externalAddress.Response.Location {
		townNames = append(townNames, loc.Town)
	}
	commonTown := findCommonAddress(townNames)
	fmt.Println("Common Prefix:", townNames)

	address := &AddressResponse{
		PostalCode:       postalCode,
		HitCount:         len(externalAddress.Response.Location),
		Address:          farthestAddress.Response.Location[0].Prefecture + farthestAddress.Response.Location[0].City + commonTown,
		TokyoStationDist: math.Round(maxDistance*10) / 10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(address)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, API Server!")
	})

	http.HandleFunc("/address", addressHandler)

	http.ListenAndServe(":8080", nil)
}
