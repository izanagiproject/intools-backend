package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool" // Correct import path for v5
)

var db *pgxpool.Pool

var	connString = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=10", "postgres", "eicdev", "localhost", "15432", "electra")

type Material struct {
	ID       int    `json:"id"`
	QCode    string `json:"qcode"`
	Plant    string `json:"plant"`
	Area     string `json:"area"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Specifications struct {
		Capacity int `json:"capacity"`
		Voltage  int `json:"voltage"`
		Current  int `json:"current"`
		RPM      int `json:"rpm"`
	} `json:"specifications"`
	Size struct {
		ShaftDiameter int `json:"shaft_diameter"`
		BaseWidth     int `json:"base_width"`
		BaseLength    int `json:"base_length"`
		C             int `json:"c"`
		E             int `json:"e"`
		H             int `json:"h"`
	} `json:"size"`
	Maker string `json:"maker"`
	PIC   struct {
		Team   string `json:"team"`
		Name   string `json:"name"`
		Phone  string `json:"phone"`
		Email  string `json:"email"`
	} `json:"pic"`
}

type APIResponse struct {
	Request struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"request"`
	Response struct {
		Count   int        `json:"count"`
		Success bool       `json:"success"`
		Data    []Material `json:"data"`
	} `json:"response"`
}

// QueryParams represents the query parameters
type QueryParams struct {
	Capacity int `json:"capacity"`
	Voltage  int `json:"voltage"`
	Current  int `json:"current"`
	RPM      int `json:"rpm"`
	ShaftDiameter int `json:"shaft_diameter"`
	BaseWidth     int `json:"base_width"`
	BaseLength    int `json:"base_length"`
	C             int `json:"c"`
	E             int `json:"e"`
	H             int `json:"h"`
}

func main() {
	var err error

	// Create a connection pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatal("Error parsing connection string:", err)
	}

	db, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Unable to connect to the database:", err)
	}
	defer db.Close()

	http.HandleFunc("/api/v1/intools/electra/materials/motor/high-voltage-all", getMaterials)
	http.HandleFunc("/api/v1/intools/electra/materials/motor/high-voltage", getMaterialsByParams)

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil, // Use the default ServeMux
	}

	// Use a goroutine to handle server shutdown
	go func() {
		// Create a channel to receive signals
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		// Wait for the signal to stop the server
		<-stop
		fmt.Println("\nServer is shutting down...")

		// Create a context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Shutdown the server
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server shutdown error:", err)
		}
	}()

	// Print a message indicating that the server is starting
	fmt.Println("Server is starting and listening on :8080")

	// Start the server
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
	}
}

func getMaterials(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers to the response
	w.Header().Add("Access-Control-Allow-Origin", "*")
    w.Header().Add("Access-Control-Allow-Credentials", "true")
    w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
    w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract limit and offset parameters from query string
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Convert parameters to integers with default values if not provided
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 0
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Perform raw query
	rows, err := db.Query(context.Background(),
		`SELECT plant, area, category, name, capacity, voltage, current, rpm, shaft_diameter, base_width, base_length, c, e, h, maker, id, qcode
		FROM public.list_materials`)
	if err != nil {
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Paginate the results in the code
	var materials []Material
	for rows.Next() {
		var material Material
		err := rows.Scan(
			&material.Plant, &material.Area, &material.Category, &material.Name,
			&material.Specifications.Capacity, &material.Specifications.Voltage, &material.Specifications.Current, 
			&material.Specifications.RPM, &material.Size.ShaftDiameter, &material.Size.BaseWidth, 
			&material.Size.BaseLength, &material.Size.C, &material.Size.E, &material.Size.H,
			&material.Maker, &material.ID, &material.QCode,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row %v", err.Error()), http.StatusInternalServerError)
			return
		}
		materials = append(materials, material)
	}

	// Apply pagination in the code
	startIndex := offset
	endIndex := offset + limit
	if endIndex > len(materials) || limit == 0 {
		endIndex = len(materials)
	}
	paginatedData := materials[startIndex:endIndex]

	// Construct response
	response := APIResponse{
		Request: struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		}{Limit: limit, Offset: offset},
		Response: struct {
			Count   int        `json:"count"`
			Success bool       `json:"success"`
			Data    []Material `json:"data"`
		}{Count: len(paginatedData), Success: true, Data: paginatedData},
	}

	// Marshal response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func getMaterialsByParams(w http.ResponseWriter, r *http.Request)  {
	// Add CORS headers to the response
	w.Header().Add("Access-Control-Allow-Origin", "*")
    w.Header().Add("Access-Control-Allow-Credentials", "true")
    w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
    w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract limit and offset parameters from query string
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Convert parameters to integers with default values if not provided
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 0
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse query parameters from the request URL
	params := QueryParams{
		Capacity:      parseFloatQueryParam(r, "capacity"),
		Voltage:       parseFloatQueryParam(r, "voltage"),
		Current:       parseFloatQueryParam(r, "current"),
		RPM:           parseFloatQueryParam(r, "rpm"),
		ShaftDiameter: parseFloatQueryParam(r, "shaft_diameter"),
		BaseWidth:     parseFloatQueryParam(r, "base_width"),
		BaseLength:    parseFloatQueryParam(r, "base_length"),
		C:             parseFloatQueryParam(r, "c"),
		E:             parseFloatQueryParam(r, "e"),
		H:             parseFloatQueryParam(r, "h"),
	}


	// Execute the dynamic SELECT query
	materials, err := selectMaterialsByParams(context.Background(), db, params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error selecting materials: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if there are no matching records
    if len(materials) == 0 {
        // Construct a default response
        defaultResponse := APIResponse{
            Request: struct {
                Limit  int `json:"limit"`
                Offset int `json:"offset"`
            }{Limit: limit, Offset: offset},
            Response: struct {
                Count   int        `json:"count"`
                Success bool       `json:"success"`
                Data    []Material `json:"data"`
            }{Count: 0, Success: true, Data: []Material{}},
        }

        // Marshal default response to JSON
        jsonResponse, err := json.Marshal(defaultResponse)
        if err != nil {
            http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
            return
        }

        // Set response headers and write JSON response
        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonResponse)
        return
    }

	// Apply pagination in the code
	startIndex := offset
	endIndex := offset + limit
	if endIndex > len(materials) || limit == 0 {
		endIndex = len(materials)
	}
	paginatedData := materials[startIndex:endIndex]

	// Construct response
	response := APIResponse{
		Request: struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		}{Limit: limit, Offset: offset},
		Response: struct {
			Count   int        `json:"count"`
			Success bool       `json:"success"`
			Data    []Material `json:"data"`
		}{Count: len(paginatedData), Success: true, Data: paginatedData},
	}

	// Marshal response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Function to parse a float query parameter from the request
func parseFloatQueryParam(r *http.Request, paramName string) int {
	paramValue := r.URL.Query().Get(paramName)
	if paramValue == "" {
		return 0
	}

	floatValue, err := strconv.ParseFloat(paramValue, 32)
	if err != nil {
		return 0
	}

	return int(floatValue)
}

// Function to execute the dynamic SELECT query
func selectMaterialsByParams(ctx context.Context, db *pgxpool.Pool, params QueryParams) ([]Material, error) {
	query, values := buildSelectQuery(params)

	rows, err := db.Query(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("unable to execute query: %w", err)
	}
	defer rows.Close()

	// Paginate the results in the code
	var materials []Material
	for rows.Next() {
		var material Material
		err := rows.Scan(
			&material.Plant, &material.Area, &material.Category, &material.Name,
			&material.Specifications.Capacity, &material.Specifications.Voltage, &material.Specifications.Current, 
			&material.Specifications.RPM, &material.Size.ShaftDiameter, &material.Size.BaseWidth, 
			&material.Size.BaseLength, &material.Size.C, &material.Size.E, &material.Size.H,
			&material.Maker, &material.ID, &material.QCode,
		)
		if err != nil {
			fmt.Printf("Error scanning row %v", err.Error())
			return nil, err
		}
		materials = append(materials, material)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	return materials, nil
}

// Function to build a dynamic SELECT query based on the provided parameters
func buildSelectQuery(params QueryParams) (string, []interface{}) {
	query := "SELECT plant, area, category, name, capacity, voltage, current, rpm, shaft_diameter, base_width, base_length, c, e, h, maker, id, qcode FROM public.list_materials WHERE true"
	var values []interface{}

	// Check each parameter and add it to the query if it's not zero
	if params.Capacity != 0 {
		query += " AND capacity = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.Capacity)
	}
	if params.Voltage != 0 {
		query += " AND voltage = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.Voltage)
	}
	if params.Current != 0 {
		query += " AND current = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.Current)
	}
	if params.RPM != 0 {
		query += " AND rpm = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.RPM)
	}
	if params.ShaftDiameter != 0 {
		query += " AND shaft_diameter = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.ShaftDiameter)
	}
	if params.BaseWidth != 0 {
		query += " AND base_width = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.BaseWidth)
	}
	if params.BaseLength != 0 {
		query += " AND base_length = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.BaseLength)
	}
	if params.C != 0 {
		query += " AND c = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.C)
	}
	if params.E != 0 {
		query += " AND e = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.E)
	}
	if params.H != 0 {
		query += " AND h = $" + strconv.Itoa(len(values)+1)
		values = append(values, params.H)
	}

	return query, values
}
