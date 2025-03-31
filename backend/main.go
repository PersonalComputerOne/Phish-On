package main

import (
	"context"
	"log"
	"math"
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
	InputUrl      string             `json:"input_url"`
	IsReal        bool               `json:"is_real"`
	SimilarityMap map[string]float64 `json:"similarity_map"`
	IsPhishing    bool               `json:"is_phishing"`
}

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

	phishingSet, err := batchPhishingCheck(pool, append(jsonData.Urls, hosts...))
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

	var suspiciousLinks []string
	for _, result := range results {
		if !result.IsReal && !result.IsPhishing {
			suspiciousLinks = append(suspiciousLinks, result.InputUrl)
		}
	}

	if err := insertSuspiciousLinks(pool, suspiciousLinks); err != nil {
		log.Printf("Failed to insert suspicious links: %v", err)
	}

	c.IndentedJSON(http.StatusOK, gin.H{"results": results})
}

func insertSuspiciousLinks(pool *pgxpool.Pool, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	_, err := pool.Exec(context.Background(),
		"INSERT INTO suspicious_links (url) SELECT unnest($1::text[]) ON CONFLICT (url) DO NOTHING",
		urls)
	return err
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

	for i := range urls {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, url, host string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			results[idx] = computeResultForUrl(url, host, phishingSet, domains)

		}(i, urls[i], hosts[i])
	}

	wg.Wait()
	return results
}

func computeResultForUrl(inputUrl, host string, phishingSet map[string]bool, domains []string) LevenshteinResult {
	if phishingSet[inputUrl] || phishingSet[host] {
		return LevenshteinResult{
			InputUrl:   inputUrl,
			IsPhishing: true,
		}
	}

	numDomains := len(domains)
	if numDomains == 0 {
		return LevenshteinResult{
			InputUrl:   inputUrl,
			IsReal:     false,
			IsPhishing: false,
		}
	}

	var exactMatch string
	similarityMap := make(map[string]float64)
	threshold := 0.85

	for _, d := range domains {
		distance := algorithms.ComputeDistance(host, d)
		similarity := similarityIndex(host, d)

		if distance == 0 {
			exactMatch = d
			break
		} else if similarity >= threshold {
			similarityMap[d] = similarity
		}
	}

	isReal := exactMatch != "" // isReal is true only if there's an exact match

	return LevenshteinResult{
		InputUrl:      inputUrl,
		IsReal:        isReal,
		SimilarityMap: similarityMap,
		IsPhishing:    false,
	}
}

// similarityIndex calculates the similarity index based on Levenshtein distance.
func similarityIndex(a, b string) float64 {
	distance := algorithms.ComputeDistance(a, b)
	maxLen := math.Max(float64(len(a)), float64(len(b)))
	normalizedDistance := float64(distance) / maxLen
	return 1 - normalizedDistance
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
