package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/PersonalComputerOne/Phish-On/pkg/db"
	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	datasetURL := "https://www.kaggle.com/api/v1/datasets/download/cheedcheed/top1m"
	var sourceID int
	err = conn.QueryRow(context.Background(),
		`INSERT INTO source (name, url) VALUES ($1, $2)
		ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url
		RETURNING id`,
		"Kaggle - Alexa Top 1 Million Sites", datasetURL,
	).Scan(&sourceID)
	if err != nil {
		log.Fatal("Error inserting source:", err)
	}

	body, err := downloadKaggleDataset()
	if err != nil {
		log.Fatal("Error downloading dataset:", err)
	}

	rows, err := processCSV(body, sourceID)
	if err != nil {
		log.Fatal("Error processing CSV:", err)
	}

	copyCount, err := conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"domain"},
		[]string{"url", "source_id"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		log.Fatal("Error inserting domains:", err)
	}

	log.Printf("Successfully inserted %d domains into the database", copyCount)
}

func downloadKaggleDataset() ([]byte, error) {
	req, err := http.NewRequest("GET",
		"https://www.kaggle.com/api/v1/datasets/download/cheedcheed/top1m", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func processCSV(zipData []byte, sourceID int) ([][]interface{}, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, err
	}

	var csvFile io.ReadCloser
	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() {
			csvFile, err = f.Open()
			if err != nil {
				return nil, err
			}
			defer csvFile.Close()
			break
		}
	}

	if csvFile == nil {
		return nil, fmt.Errorf("no CSV file found in zip archive")
	}

	csvReader := csv.NewReader(csvFile)
	rows := [][]interface{}{}

	if _, err := csvReader.Read(); err != nil {
		return nil, err
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) < 2 {
			continue
		}

		rows = append(rows, []interface{}{record[1], sourceID})
	}

	return rows, nil
}
