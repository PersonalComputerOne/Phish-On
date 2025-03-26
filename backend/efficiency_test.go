package main

import (
	"log"
	"runtime"
	"testing"
	"time"
)

func benchmarkPerformance(urls []string, iterations int) {
	// Sequential performance
	startSeq := time.Now()
	for i := 0; i < iterations; i++ {
		computeResultsSequential(urls, extractHosts(urls), map[string]bool{}, []string{"example.com", "test.org", "another.net"}) // Mock phishingSet and domains
	}
	seqDuration := time.Since(startSeq)

	// Parallel performance
	startPar := time.Now()
	for i := 0; i < iterations; i++ {
		computeResultsParallel(urls, extractHosts(urls), map[string]bool{}, []string{"example.com", "test.org", "another.net"}) // Mock phishingSet and domains
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
	urls := []string{
		"http://github.com",
		"http://githun.com",
		"http://linkedin.com",
		"http://linkedin.org",
		"http://example.org",
		"http://google.com",
		"http://googgle.com",
		"http://facebook.com",
		"http://facebok.com",
		"http://twitter.com",
		"http://twiter.com",
		"http://amazon.com",
		"http://amozon.com",
		"http://apple.com",
		"http://aplle.com",
		"http://microsoft.com",
		"http://microsft.com",
		"http://yahoo.com",
		"http://yaho.com",
		"http://reddit.com",
	}
	iterations := 5 // Run the test 5 times.
	benchmarkPerformance(urls, iterations)
}
