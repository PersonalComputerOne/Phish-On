package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/PersonalComputerOne/Phish-On/db"
)

func loadTestUrls(limit int) ([]string, error) {
	file, err := os.Open("../datasets/new_data_urls.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) > limit {
		rand.Shuffle(len(records), func(i, j int) {
			records[i], records[j] = records[j], records[i]
		})
		records = records[:limit]
	}

	urls := make([]string, len(records))
	for i, record := range records {
		urls[i] = record[0]
	}

	return urls, nil
}

func benchmarkPerformance(urls []string) { // Removed iterations parameter
	pool, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	domains, err := fetchLegitimateDomains(pool)
	if err != nil {
		log.Fatalf("Failed to fetch legitimate domains: %v", err)
	}
	// Sequential performance
	startSeq := time.Now()
	computeResultsSequential(urls, extractHosts(urls), map[string]bool{}, domains) // Removed loop
	seqDuration := time.Since(startSeq)

	// Parallel performance
	startPar := time.Now()
	computeResultsParallel(urls, extractHosts(urls), map[string]bool{}, domains) // Removed loop
	parDuration := time.Since(startPar)

	// Calculate metrics
	avgSeqTime := seqDuration.Seconds()
	avgParTime := parDuration.Seconds()
	speedup := avgSeqTime / avgParTime

	log.Printf("Sequential Time: %.4f seconds", avgSeqTime) // Changed to Time
	log.Printf("Parallel Time: %.4f seconds", avgParTime)   // Changed to Time
	log.Printf("Speedup: %.2fx", speedup)
}

func TestPerformance(t *testing.T) {
	const testLimit = 15 // Change this value to control how many URLs to test

	urls, err := loadTestUrls(testLimit)
	if err != nil {
		t.Fatalf("Failed to load test URLs: %v", err)
	}

	t.Logf("Testing performance with %d URLs", len(urls))
	benchmarkPerformance(urls) // Removed iterations parameter
}
