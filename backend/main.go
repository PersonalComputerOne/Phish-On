package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/gin-gonic/gin"
)

type domain struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

var domains = []domain{
	{ID: "1", Url: "github.com"},
	{ID: "2", Url: "google.com"},
	{ID: "3", Url: "facebook.com"},
	{ID: "4", Url: "linkedin.com"},
	{ID: "5", Url: "x.com"},
	{ID: "6", Url: "instagram.com"},
}

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

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/levenshtein", func(c *gin.Context) {
		var jsonData struct {
			Urls []string `json:"urls"`
		}

		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var results []urlResult

		const maxDistance = 2

		for _, inputUrl := range jsonData.Urls {
			host, err := getHost(inputUrl)
			if err != nil {
				host = inputUrl
			}

			minDistance := -1
			closestUrl := ""

			for _, d := range domains {
				distance := levenshtein.ComputeDistance(host, d.Url)

				if minDistance == -1 || distance < minDistance {
					minDistance = distance
					closestUrl = d.Url
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
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
