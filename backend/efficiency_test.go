package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"runtime"
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

func benchmarkPerformance(urls []string, iterations int) {
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
	for range iterations {
		computeResultsSequential(urls, extractHosts(urls), map[string]bool{}, domains)
	}
	seqDuration := time.Since(startSeq)

	// Parallel performance
	startPar := time.Now()
	for range iterations {
		computeResultsParallel(urls, extractHosts(urls), map[string]bool{}, domains)
	}
	parDuration := time.Since(startPar)

	// Calculate metrics
	avgSeqTime := seqDuration.Seconds() / float64(iterations)
	avgParTime := parDuration.Seconds() / float64(iterations)
	speedup := avgSeqTime / avgParTime
	efficiency := speedup / float64(runtime.NumCPU())

	log.Printf("Sequential Avg Time: %.4f seconds", avgSeqTime)
	log.Printf("Parallel Avg Time: %.4f seconds", avgParTime)
	log.Printf("Speedup: %.2fx", speedup)
	log.Printf("Efficiency: %.2f", efficiency)
}

func TestPerformance(t *testing.T) {
	const testLimit = 5  // Change this value to control how many URLs to test
	const iterations = 5 // Number of times to run the benchmark

	urls, err := loadTestUrls(testLimit)
	if err != nil {
		t.Fatalf("Failed to load test URLs: %v", err)
	}

	t.Logf("Testing performance with %d URLs", len(urls))
	benchmarkPerformance(urls, iterations)
}
