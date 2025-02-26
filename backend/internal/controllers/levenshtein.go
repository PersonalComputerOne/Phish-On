package controllers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PersonalComputerOne/Phish-On/pkg/db"
	"github.com/agnivade/levenshtein"
	"github.com/gin-gonic/gin"
)

type urlResult struct {
	InputUrl   string `json:"input_url"`
	Distance   int    `json:"distance"`
	IsReal     bool   `json:"is_real"`
	ClosestUrl string `json:"closest_url"`
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

func Levenshtein(c *gin.Context) {
	conn, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer conn.Close()

	var jsonData struct {
		Urls []string `json:"urls"`
	}

	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var results []urlResult

	const maxDistance = 2

	var domains []string

	rows, _ := conn.Query(context.Background(), `SELECT url FROM domain`)

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			log.Fatalf("Failed to scan URL: %v", err)
		}
		domains = append(domains, url)
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

		results = append(results, urlResult{
			InputUrl:   inputUrl,
			Distance:   minDistance,
			IsReal:     isReal,
			ClosestUrl: closestUrl,
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"results": results})
}
