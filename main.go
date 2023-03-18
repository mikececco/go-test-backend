package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

type Spot struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Rating    float64 `json:"rating"`
}

type SpotsResponse struct {
	Spots []Spot `json:"spots"`
}

func main() {
	// Define the database connection string.
	dbConnectionString := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	// Connect to the database.
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Define the router.
	r := mux.NewRouter()

	// Define the spots endpoint.
	r.HandleFunc("/spots", func(w http.ResponseWriter, r *http.Request) {

		// Parse and validate the query parameters.
		lat, err := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
		if err != nil {
			http.Error(w, "Invalid latitude", http.StatusBadRequest)
			return
		}
		lon, err := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
		if err != nil {
			http.Error(w, "Invalid longitude", http.StatusBadRequest)
			return
		}
		radius, err := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
		if err != nil {
			http.Error(w, "Invalid radius", http.StatusBadRequest)
			return
		}
		areaType := r.URL.Query().Get("type")

		// Construct SQL query based on areaType parameter
		var query string
		if areaType == "circle" {
			query = `
				SELECT id, name, ST_X(geom) AS longitude, ST_Y(geom) AS latitude, rating
				FROM spots
				WHERE ST_DWithin(geom, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3 * 1000)
			`
		} else if areaType == "square" {
			query = `
			SELECT id, name, ST_X(geom) AS longitude, ST_Y(geom) AS latitude, rating
			FROM spots
			WHERE ST_DWithin(geom, ST_MakeEnvelope($1 - ($3 / (111.32 * COS(RADIANS($2)))), $2 - ($3 / 111.32), $1 + ($3 / (111.32 * COS(RADIANS($2)))), $2 + ($3 / 111.32), 4326), $3 * 1000)`
		} else {
			http.Error(w, "Invalid area type", http.StatusBadRequest)
			return
		}
			// Execute the SQL query.
		rows, err := db.Query(query, lon, lat, radius)
		if err != nil {
			http.Error(w, "Failed to execute query", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Parse the query result and construct the response.
		spots := make([]Spot, 0)
		for rows.Next() {
			var spot Spot
			err := rows.Scan(&spot.ID, &spot.Name, &spot.Longitude, &spot.Latitude, &spot.Rating)
			if err != nil {
				http.Error(w, "Failed to scan row", http.StatusInternalServerError)
				return
			}
			spots = append(spots, spot)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to iterate over rows", http.StatusInternalServerError)
			return
		}
		sort.Slice(spots, func(i, j int) bool {
			return spots[i].Rating > spots[j].Rating
		})
		response := SpotsResponse{Spots: spots}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

		// Write the response.
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	})
}

	// Start the server.
	log.Fatal(http.ListenAndServe(":8080", r))
