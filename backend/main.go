package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"

	"github.com/PersonalComputerOne/Phish-On/algorithms"
	"github.com/PersonalComputerOne/Phish-On/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LevenshteinResult struct {
	InputUrl   string `json:"input_url"`
	Distance   int    `json:"distance"`
	IsReal     bool   `json:"is_real"`
	ClosestUrl string `json:"closest_url"`
	IsPhishing bool   `json:"is_phishing"`
}

type RequestBody struct {
	Urls []string `json:"urls"`
}

const maxDistance = 2

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
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

	phishingSet, err := batchPhishingCheck(pool, hosts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Phishing check failed"})
		return
	}

	domains, err := fetchLegitimateDomains(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get domains"})
		return
	}

	var results []LevenshteinResult
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
		host, err := getHost(inputUrl)
		if err != nil {
			host = inputUrl
		}
		hosts[i] = host
	}
	return hosts
}

func computeResultsSequential(urls, hosts []string, phishingSet map[string]bool, domains []string) []LevenshteinResult {
	results := make([]LevenshteinResult, len(urls))

	for i, inputUrl := range urls {
		results[i] = computeResultForUrl(inputUrl, hosts[i], phishingSet, domains)
	}

	return results
}

func computeResultsParallel(urls, hosts []string, phishingSet map[string]bool, domains []string) []LevenshteinResult {
	results := make([]LevenshteinResult, len(urls))

	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())
	for i, inputUrl := range urls {
		wg.Add(1)
		sem <- struct{}{} // Acquire a semaphore slot
		go func(idx int, urlStr, host string) {
			defer func() {
				<-sem // Release semaphore slot
				wg.Done()
			}()
			results[idx] = computeResultForUrl(urlStr, host, phishingSet, domains)
		}(i, inputUrl, hosts[i])
	}
	wg.Wait()

	return results
}

func computeResultForUrl(inputUrl, host string, phishingSet map[string]bool, domains []string) LevenshteinResult {
	if phishingSet[host] {
		return LevenshteinResult{
			InputUrl:   inputUrl,
			IsPhishing: true,
			Distance:   -1,
		}
	}

	minDistance := -1
	closestUrl := ""
	for _, d := range domains {
		distance := algorithms.ComputeDistance(host, d)
		if minDistance == -1 || distance < minDistance {
			minDistance = distance
			closestUrl = d
			if distance == 0 {
				break // Exact match found
			}
		}
	}

	isReal := minDistance == 0
	if minDistance > maxDistance {
		closestUrl = ""
	}

	return LevenshteinResult{
		InputUrl:   inputUrl,
		Distance:   minDistance,
		IsReal:     isReal,
		ClosestUrl: closestUrl,
		IsPhishing: false,
	}
}

func batchPhishingCheck(pool *pgxpool.Pool, hosts []string) (map[string]bool, error) {
	phishingSet := make(map[string]bool)
	if len(hosts) == 0 {
		return phishingSet, nil
	}

	rows, err := pool.Query(context.Background(),
		"SELECT url FROM domain WHERE url = ANY($1) AND is_phishing = TRUE", hosts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		phishingSet[url] = true
	}
	return phishingSet, nil
}

func fetchLegitimateDomains(pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(context.Background(),
		"SELECT url FROM domain WHERE is_phishing = FALSE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func getHost(inputURL string) (string, error) {
	if !strings.Contains(inputURL, "://") && !strings.HasPrefix(inputURL, "//") {
		inputURL = "http://" + inputURL
	}
	u, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}
	host := strings.ToLower(u.Hostname())
	return strings.TrimSuffix(host, "."), nil
}
