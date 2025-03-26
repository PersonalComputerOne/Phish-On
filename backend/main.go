package main

import (
	"log"
	"net/http"
	"runtime"
	"sync"

	"github.com/PersonalComputerOne/Phish-On/algorithms"
	"github.com/PersonalComputerOne/Phish-On/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RequestBody struct {
	Urls []string `json:"urls"`
}

func main() {
	pool, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	router := gin.Default()

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
		c.JSON(http.StatusOK, gin.H{"status": "ok", "runtime": runtime.NumCPU()})
	})

	router.Run(":8080")
}

func levenshteinHandler(c *gin.Context, pool *pgxpool.Pool, parallel bool) {
	var jsonData RequestBody
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hosts := extractHosts(jsonData.Urls)

	phishingSet, err := algorithms.BatchPhishingCheck(pool, hosts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Phishing check failed"})
		return
	}

	domains, err := algorithms.FetchLegitimateDomains(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get domains"})
		return
	}

	var results []algorithms.LevenshteinResult
	if parallel {
		results = computeResultsParallel(jsonData.Urls, hosts, phishingSet, domains)
	} else {
		results = computeResultsSequential(jsonData.Urls, hosts, phishingSet, domains)
	}

	c.IndentedJSON(http.StatusOK, gin.H{"results": results})
}

func extractHosts(urls []string) []string {
	hosts := make([]string, len(urls))
	for i, inputUrl := range urls {
		host, err := algorithms.GetHost(inputUrl)
		if err != nil {
			host = inputUrl
		}
		hosts[i] = host
	}
	return hosts
}

func computeResultsSequential(urls, hosts []string, phishingSet map[string]bool, domains []string) []algorithms.LevenshteinResult {
	results := make([]algorithms.LevenshteinResult, len(urls))

	for i, inputUrl := range urls {
		results[i] = algorithms.ComputeResultForUrl(inputUrl, hosts[i], phishingSet, domains)
	}

	return results
}

func computeResultsParallel(urls, hosts []string, phishingSet map[string]bool, domains []string) []algorithms.LevenshteinResult {
	results := make([]algorithms.LevenshteinResult, len(urls))

	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())

	for i := range urls {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, url, host string) {
			defer func() {
				<-sem
				wg.Done()
				<-sem // Release semaphore slot
			}()

			results[idx] = algorithms.ComputeResultForUrl(url, hosts[i], phishingSet, domains)
		}(i, urls[i], hosts[i])
	}

	wg.Wait()
	return results
}
