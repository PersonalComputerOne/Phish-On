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

type testPayload struct {
	Urls []string `json:"urls"`
}

func testRouter() (*gin.Engine, *pgxpool.Pool, error) {
	pool, err := db.Init()
	if err != nil {
		return nil, nil, err
	}

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
	const testLimit = 5 // Change this value to control how many URLs to test

	router, _, err := testRouter()
	if err != nil {
		t.Fatalf("Failed to initialize test router: %v", err)
	}

	urls, err := loadTestUrls(testLimit)
	if err != nil {
		t.Fatalf("Failed to load test URLs: %v", err)
	}

	payload := testPayload{
		Urls: urls,
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

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", w.Code)
	}

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

	invalidJSON := []byte(`{"urls": ["http://example.com"]`)

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
	const testLimit = 5 // Change this value to control how many URLs to test

	router, pool, err := testRouter()
	if err != nil {
		t.Fatalf("Failed to initialize test router: %v", err)
	}
	pool.Close()

	urls, err := loadTestUrls(testLimit)
	if err != nil {
		t.Fatalf("Failed to load test URLs: %v", err)
	}

	payload := testPayload{
		Urls: urls,
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

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 Internal Server Error due to DB failure, got %d", w.Code)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
