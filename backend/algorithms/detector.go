package algorithms

import (
	"context"
	"log"
	"math"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LevenshteinResult struct {
	InputUrl      string             `json:"input_url"`
	IsReal        bool               `json:"is_real"`
	SimilarityMap map[string]float64 `json:"similarity_map"`
	IsPhishing    bool               `json:"is_phishing"`
}

func ComputeResultForUrl(inputUrl, host string, phishingSet map[string]bool, domains []string) LevenshteinResult {
	if phishingSet[host] {
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
		distance := ComputeDistance(host, d)
		similarity := SimilarityIndex(host, d)

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

// SimilarityIndex calculates the similarity index based on Levenshtein distance.
func SimilarityIndex(a, b string) float64 {
	distance := ComputeDistance(a, b)
	maxLen := math.Max(float64(len(a)), float64(len(b)))
	normalizedDistance := float64(distance) / maxLen
	return 1 - normalizedDistance
}

func BatchPhishingCheck(pool *pgxpool.Pool, hosts []string) (map[string]bool, error) {
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

func FetchLegitimateDomains(pool *pgxpool.Pool) ([]string, error) {
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

func GetHost(inputURL string) (string, error) {
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
