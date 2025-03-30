package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PersonalComputerOne/Phish-On/db"
)

type ProcessorFunc func([]byte) ([]string, error)

func main() {
	conn, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	sources := []struct {
		Name      string
		URL       string
		Processor ProcessorFunc
	}{
		{
			Name: "Kaggle - Alexa Top 1 Million Sites",
			URL:  "https://www.kaggle.com/api/v1/datasets/download/cheedcheed/top1m",
			Processor: func(data []byte) ([]string, error) {
				return processZipFile(data, "", 1)
			},
		},
		{
			Name: "Umbrella Top 1 Million Sites",
			URL:  "https://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip",
			Processor: func(data []byte) ([]string, error) {
				return processZipFile(data, "top-1m.csv", 1)
			},
		},
		{
			Name: "Majestic Million",
			URL:  "https://downloads.majestic.com/majestic_million.csv",
			Processor: func(data []byte) ([]string, error) {
				return parseCSV(bytes.NewReader(data), 2)
			},
		},
		{
			Name: "Tranco Top 1 Million",
			URL:  "https://tranco-list.eu/top-1m.csv.zip",
			Processor: func(data []byte) ([]string, error) {
				return processZipFile(data, "top-1m.csv", 1)
			},
		},
		{
			Name: "BuiltWith Top 1 Million",
			URL:  "https://builtwith.com/dl/builtwith-top1m.zip",
			Processor: func(data []byte) ([]string, error) {
				return processZipFile(data, ".csv", 1)
			},
		},
	}

	for _, source := range sources {
		var sourceID int
		err := conn.QueryRow(context.Background(),
			`INSERT INTO source (name, url) VALUES ($1, $2)
			ON CONFLICT (url) DO UPDATE SET name = EXCLUDED.name
			RETURNING id`,
			source.Name, source.URL,
		).Scan(&sourceID)
		if err != nil {
			log.Printf("Error inserting source %s: %v", source.Name, err)
			continue
		}

		data, err := downloadDataset(source.URL)
		if err != nil {
			log.Printf("Error downloading %s: %v", source.URL, err)
			continue
		}

		domains, err := source.Processor(data)
		if err != nil {
			log.Printf("Error processing %s: %v", source.Name, err)
			continue
		}

		// Use batch insert with conflict handling
		batch := &pgx.Batch{}
		for _, domain := range domains {
			batch.Queue(
				`INSERT INTO domain (url, source_id) 
        VALUES ($1, $2) 
        ON CONFLICT (url) DO NOTHING`, // Conflict on url column
				strings.ToLower(domain),
				sourceID,
			)
		}

		results := conn.SendBatch(context.Background(), batch)
		defer results.Close()

		var totalInserted int
		for range domains {
			_, err := results.Exec()
			if err == nil {
				totalInserted++
			}
		}

		if err := results.Close(); err != nil {
			log.Printf("Error finalizing batch insert for %s: %v", source.Name, err)
			continue
		}

		log.Printf("Processed %s: %d new domains inserted", source.Name, totalInserted)
	}

	seedPhishtank(conn)
}

func downloadDataset(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func processZipFile(zipData []byte, filePattern string, domainColumn int) ([]string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, err
	}

	var csvFile io.ReadCloser
	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() && (filePattern == "" || strings.Contains(f.Name, filePattern)) {
			csvFile, err = f.Open()
			if err != nil {
				return nil, err
			}
			defer csvFile.Close()
			break
		}
	}

	if csvFile == nil {
		return nil, fmt.Errorf("no matching CSV file found in zip")
	}

	return parseCSV(csvFile, domainColumn)
}

func parseCSV(r io.Reader, domainColumn int) ([]string, error) {
	csvReader := csv.NewReader(r)
	var domains []string

	// Skip header
	_, _ = csvReader.Read()

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) <= domainColumn {
			continue
		}

		domain := strings.TrimSpace(record[domainColumn])
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	return domains, nil
}

func getSourceID(ctx context.Context, pool *pgxpool.Pool, name, urlStr string) (int, error) {
	var id int
	err := pool.QueryRow(ctx, "SELECT id FROM source WHERE name = $1", name).Scan(&id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = pool.QueryRow(ctx,
				"INSERT INTO source(name, url, added_at, last_crawled_at) VALUES($1, $2, $3, $4) RETURNING id",
				name, urlStr, time.Now(), time.Now(),
			).Scan(&id)
			if err != nil {
				return 0, fmt.Errorf("failed to insert source: %w", err)
			}
		} else {
			return 0, fmt.Errorf("failed to query source: %w", err)
		}
	}
	return id, nil
}

func seedPhishtank(pool *pgxpool.Pool) {
	ctx := context.Background()

	sourceName := "phishtank"
	sourceURL := "https://www.phishtank.org"
	sourceID, err := getSourceID(ctx, pool, sourceName, sourceURL)
	if err != nil {
		log.Fatalf("Error getting source ID: %v", err)
	}

	file, err := os.Open("seed/phishtank.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	insertSQL := `
		INSERT INTO domain(url, is_phishing, source_id, added_at)
		VALUES($1, $2, $3, $4)
		ON CONFLICT (url) DO NOTHING
	`

	total := len(lines)
	if total == 0 {
		fmt.Println("No lines to process.")
		return
	}

	insertedCount := 0

	for _, line := range lines {
		_, err = pool.Exec(ctx, insertSQL, line, true, sourceID, time.Now())
		if err != nil {
			continue
		}

		insertedCount++
		fmt.Printf("\rInserted %d/%d", insertedCount, total)
	}

	log.Println("Finished processing file.")
}
