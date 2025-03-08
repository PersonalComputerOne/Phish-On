package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PersonalComputerOne/Phish-On/db"
	"github.com/agnivade/levenshtein"
	"github.com/gin-gonic/gin"
)

func main() {
	pool, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/levenshtein", func(c *gin.Context) {
			var jsonData struct {
				Urls []string `json:"urls"`
			}

			if err := c.ShouldBindJSON(&jsonData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			const maxDistance = 2
			var domains []string

			rows, err := pool.Query(context.Background(), `SELECT url FROM domain`)
			if err != nil {
				log.Printf("Query error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Query error"})
				return
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

			var results []struct {
				InputUrl   string `json:"input_url"`
				Distance   int    `json:"distance"`
				IsReal     bool   `json:"is_real"`
				ClosestUrl string `json:"closest_url"`
			}

			for _, inputUrl := range jsonData.Urls {
				host, err := getHost(inputUrl)
				if err != nil {
					host = inputUrl
				}

				minDistance := -1
				closestUrl := ""

				for _, d := range domains {
					distance := levenshtein.ComputeDistance(host, d)
					if minDistance == -1 || distance < minDistance {
						minDistance = distance
						closestUrl = d
					}
				}

				if minDistance > maxDistance {
					closestUrl = ""
				}
				isReal := minDistance == 0

				results = append(results, struct {
					InputUrl   string `json:"input_url"`
					Distance   int    `json:"distance"`
					IsReal     bool   `json:"is_real"`
					ClosestUrl string `json:"closest_url"`
				}{
					InputUrl:   inputUrl,
					Distance:   minDistance,
					IsReal:     isReal,
					ClosestUrl: closestUrl,
				})
			}

			c.IndentedJSON(http.StatusOK, gin.H{"results": results})
		})
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.Run(":8080")
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
