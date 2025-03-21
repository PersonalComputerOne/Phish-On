package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"runtime"

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
			levenshteinSequentialHandler(c, pool)
		})
		api.POST("/levenshtein/parallel", func(c *gin.Context) {
			levenshteinParallelHandler(c, pool)
		})
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.Run(":8080")
}

func levenshteinSequentialHandler(c *gin.Context, pool *pgxpool.Pool) {
	var jsonData RequestBody

	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domains, err := fetchDomains(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve domains"})
		return
	}

	const maxDistance = 2
	var results []LevenshteinResult

	for _, inputUrl := range jsonData.Urls {
		host, err := getHost(inputUrl)
		if err != nil {
			host = inputUrl
		}

		minDistance := -1
		closestUrl := ""

		for _, d := range domains {
			distance := algorithms.ComputeDistance(host, d)
			if minDistance == -1 || distance < minDistance {
				minDistance = distance
				closestUrl = d
			}
		}

		if minDistance > maxDistance {
			closestUrl = ""
		}
		isReal := minDistance == 0

		results = append(results, LevenshteinResult{
			InputUrl:   inputUrl,
			Distance:   minDistance,
			IsReal:     isReal,
			ClosestUrl: closestUrl,
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"results": results})
}

func levenshteinParallelHandler(c *gin.Context, pool *pgxpool.Pool) {
	var jsonData RequestBody

	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domains, err := fetchDomains(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve domains"})
		return
	}

	const maxDistance = 2
	var results []LevenshteinResult

	for _, inputUrl := range jsonData.Urls {
		host, err := getHost(inputUrl)
		if err != nil {
			host = inputUrl
		}

		numCPUs := runtime.NumCPU()
		numDomains := len(domains)
		chunkSize := (numDomains + numCPUs - 1) / numCPUs

		type localMinResult struct {
			MinDistance int
			ClosestUrl  string
		}

		localMins := make(chan localMinResult, numCPUs)
		var wg sync.WaitGroup

		for i := 0; i < numCPUs; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if end > numDomains {
				end = numDomains
			}

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				localMin := localMinResult{MinDistance: -1, ClosestUrl: ""}
				for j := start; j < end; j++ {
					distance := algorithms.ComputeDistance(host, domains[j])
					if localMin.MinDistance == -1 || distance < localMin.MinDistance {
						localMin.MinDistance = distance
						localMin.ClosestUrl = domains[j]
					}
				}
				localMins <- localMin
			}(start, end)
		}

		go func() {
			wg.Wait()
			close(localMins)
		}()

		minDistance := -1
		closestUrl := ""

		for localMin := range localMins {
			if minDistance == -1 || localMin.MinDistance < minDistance {
				minDistance = localMin.MinDistance
				closestUrl = localMin.ClosestUrl
			}
		}

		if minDistance > maxDistance {
			closestUrl = ""
		}
		isReal := minDistance == 0

		results = append(results, LevenshteinResult{
			InputUrl:   inputUrl,
			Distance:   minDistance,
			IsReal:     isReal,
			ClosestUrl: closestUrl,
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"results": results})
}

func fetchDomains(pool *pgxpool.Pool) ([]string, error) {
	var domains []string

	rows, err := pool.Query(context.Background(), `SELECT url FROM domain`)
	if err != nil {
		log.Printf("Query error: %v", err)
		return nil, err
	}
	defer rows.Close()

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
	host := u.Hostname()
	host = strings.ToLower(host)
	host = strings.TrimSuffix(host, ".")
	return host, nil
}
