// main_integration_test.go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/PersonalComputerOne/Phish-On/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testPayload is the request structure expected by the API.
type testPayload struct {
	Urls []string `json:"urls"`
}

// testRouter initializes a Gin engine with your routes, similar to your main function.
// It returns the router and the database pool so that tests can seed or clean up data.
func testRouter() (*gin.Engine, *pgxpool.Pool, error) {
	// Initialize DB (you can set an environment variable for test DB connection string)
	pool, err := db.Init()
	if err != nil {
		return nil, nil, err
	}

	// Create a new Gin engine (using Test Mode)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	{
		api.POST("/levenshtein/sequential", func(c *gin.Context) {
			levenshteinHandler(c, pool, false)
		})
		api.POST("/levenshtein/parallel", func(c *gin.Context) {
			levenshteinHandler(c, pool, true)
		})
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "runtime": "test"})
	})

	return router, pool, nil
}

func TestLevenshteinParallel_Success(t *testing.T) {
	router, _, err := testRouter()
	if err != nil {
		t.Fatalf("Failed to initialize test router: %v", err)
	}

	// Build a valid payload.
	payload := testPayload{
		Urls: []string{"http://githun.com", "http://github.com", "https://giiiiithdub.com", "https://linkeddin.com", "https://twitter.com"},
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create a test request.
	req, err := http.NewRequest(http.MethodPost, "/api/v1/levenshtein/parallel", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a recorder.
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", w.Code)
	}

	// Optionally, check response JSON.
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	t.Logf("Response: %+v", resp)
}

func TestLevenshteinParallel_InvalidJSON(t *testing.T) {
	router, _, err := testRouter()
	if err != nil {
		t.Fatalf("Failed to initialize test router: %v", err)
	}

	// Prepare an invalid JSON payload.
	invalidJSON := []byte(`{"urls": ["http://example.com"]`) // missing closing bracket

	req, err := http.NewRequest(http.MethodPost, "/api/v1/levenshtein/parallel", bytes.NewBuffer(invalidJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got %d", w.Code)
	}
}

func TestLevenshteinParallel_DBFailure(t *testing.T) {
	// Initialize the router and DB as usual.
	router, pool, err := testRouter()
	if err != nil {
		t.Fatalf("Failed to initialize test router: %v", err)
	}
	// Simulate DB failure by closing the pool.
	pool.Close()

	payload := testPayload{
		Urls: []string{"http://phishy.com/malicious"},
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/v1/levenshtein/parallel", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Expect an internal server error because the DB connection is closed.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 Internal Server Error due to DB failure, got %d", w.Code)
	}
}

// Optionally, you can use TestMain to setup and teardown global test configuration.
func TestMain(m *testing.M) {
	// Example: set environment variables for test DB.
	// os.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/testdb")
	code := m.Run()
	os.Exit(code)
}
