package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PersonalComputerOne/Phish-On/db"
	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PhishTankResp struct {
	URL string `json:"url"`
}

func main() {
	conn, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	if err := seedPhishTankData(conn); err != nil {
		log.Fatalf("PhishTank seeding failed: %v", err)
	}
}

func seedPhishTankData(conn *pgxpool.Pool) error {
	c := colly.NewCollector(
		colly.AllowedDomains("phishtank.org"),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"),
	)

	var sourceID int
	baseURL := "https://phishtank.org/phish_search.php?valid=y&active=All&Search=Search"
	err := conn.QueryRow(context.Background(),
		`INSERT INTO source (name, url) VALUES ($1, $2)
			ON CONFLICT (url) DO UPDATE SET 
					last_crawled_at = CURRENT_TIMESTAMP
			RETURNING id`,
		"PhishTank", baseURL,
	).Scan(&sourceID)
	if err != nil {
		return fmt.Errorf("error inserting source: %w", err)
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 10, // Increased parallelism
		Delay:       1 * time.Second,
	})

	uniqueDomains := make(map[string]bool)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	batch := pgx.Batch{}
	batchSize := 0
	maxBatchSize := 100

	commitBatch := func() error {
		if batchSize == 0 {
			return nil
		}

		br := conn.SendBatch(ctx, &batch)
		defer br.Close()

		successCount := 0
		for i := 0; i < batchSize; i++ {
			_, err := br.Exec()
			if err != nil {
				log.Printf("Error inserting domain: %v", err)
			} else {
				successCount++
			}
		}

		batch = pgx.Batch{}
		log.Printf("Committed batch of %d domains (%d successful)", batchSize, successCount)
		batchSize = 0
		return nil
	}

	// Extract phishing URLs from table rows
	c.OnHTML("tr[style='background: #ffffcc;']", func(e *colly.HTMLElement) {
		tds := e.DOM.Find("td.value")
		if tds.Length() < 2 {
			return
		}

		// Extract raw URL from second column
		rawURL := strings.SplitN(tds.Eq(1).Text(), "\n", 2)[0]
		rawURL = strings.TrimSpace(rawURL)

		// Clean timestamp from URL if present
		if idx := strings.Index(rawURL, "added on"); idx > 0 {
			rawURL = strings.TrimSpace(rawURL[:idx])
		}

		// Parse and normalize domain
		domain, err := extractDomain(rawURL)
		if err != nil {
			log.Printf("Skipping invalid URL %q: %v", rawURL, err)
			return
		}

		cleanDomain := normalizeDomain(domain)
		if cleanDomain == "" || uniqueDomains[cleanDomain] {
			return
		}

		uniqueDomains[cleanDomain] = true

		batch.Queue(
			`INSERT INTO domain (url, is_phishing, source_id)
					VALUES ($1, $2, $3)
					ON CONFLICT DO NOTHING`,
			cleanDomain, true, sourceID,
		)
		batchSize++

		// Commit batch when it reaches max size
		if batchSize >= maxBatchSize {
			if err := commitBatch(); err != nil {
				log.Printf("Error committing batch: %v", err)
			}
		}
	})

	// After each page is processed
	c.OnScraped(func(r *colly.Response) {
		log.Printf("Finished processing page: %s", r.Request.URL.String())

		// Commit any pending inserts
		if batchSize > 0 {
			if err := commitBatch(); err != nil {
				log.Printf("Error committing batch after page: %v", err)
			}
		}
	})

	pageCount := 206000

	for page := 0; page < pageCount; page++ {
		pageURL := baseURL
		if page > 0 {
			pageURL = fmt.Sprintf("%s&page=%d", baseURL, page)
		}

		log.Printf("Visiting page %d: %s", page, pageURL)
		err := c.Visit(pageURL)
		if err != nil {
			log.Printf("Error visiting page %d: %v", page, err)
			// Continue to next page if one fails
			continue
		}
		c.Wait()
	}

	// Wait for all requests to finish
	c.Wait()

	// Final commit for any remaining items
	if batchSize > 0 {
		if err := commitBatch(); err != nil {
			log.Printf("Error committing final batch: %v", err)
		}
	}

	log.Printf("Successfully processed PhishTank domains: %d unique domains found", len(uniqueDomains))
	return nil
}

func extractDomain(rawURL string) (string, error) {
	if idx := strings.Index(rawURL, "added on"); idx > 0 {
		rawURL = strings.TrimSpace(rawURL[:idx])
	}

	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Host == "" && u.Opaque != "" {
		parts := strings.SplitN(u.Opaque, "/", 2)
		return parts[0], nil
	}

	host := u.Hostname()
	if host == "" {
		return "", fmt.Errorf("empty host in URL")
	}

	return host, nil
}

func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}
