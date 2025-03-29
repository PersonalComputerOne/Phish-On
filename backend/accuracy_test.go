package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/PersonalComputerOne/Phish-On/db"
)

func TestPhishingAccuracy(t *testing.T) {
	const testLimit = 5 // Change this value to control how many URLs to test

	pool, err := db.Init()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	file, err := os.Open("../datasets/new_data_urls.csv")
	if err != nil {
		t.Fatalf("Failed to open test data file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV records: %v", err)
	}

	if len(records) > testLimit {
		rand.Shuffle(len(records), func(i, j int) {
			records[i], records[j] = records[j], records[i]
		})
		records = records[:testLimit]
	}

	t.Logf("Testing %d URLs from the dataset", len(records))

	tp, fp, tn, fn := 0, 0, 0, 0

	domains, err := fetchLegitimateDomains(pool)
	if err != nil {
		t.Fatalf("Failed to fetch legitimate domains: %v", err)
	}

	for _, record := range records {
		url := record[0]
		status, err := strconv.Atoi(record[1])
		if err != nil {
			t.Fatalf("Failed to parse status value: %v", err)
		}
		isPhishing := status == 0

		host, err := getHost(url)
		if err != nil {
			log.Printf("Error getting host: %v", err)
			continue
		}

		phishingSet, err := batchPhishingCheck(pool, []string{host})
		if err != nil {
			t.Fatalf("Phishing check failed: %v", err)
		}

		result := computeResultForUrl(url, host, phishingSet, domains)

		if isPhishing && !result.IsReal {
			tp++
		} else if !isPhishing && !result.IsReal {
			fp++
		} else if !isPhishing && result.IsReal {
			tn++
		} else if isPhishing && result.IsReal {
			fn++
		}
	}

	precision := float64(tp) / float64(tp+fp)
	recall := float64(tp) / float64(tp+fn)
	f1 := 2 * (precision * recall) / (precision + recall)

	t.Logf("True Positives: %d", tp)
	t.Logf("False Positives: %d", fp)
	t.Logf("True Negatives: %d", tn)
	t.Logf("False Negatives: %d", fn)
	t.Logf("Precision: %.4f", precision)
	t.Logf("Recall: %.4f", recall)
	t.Logf("F1 Score: %.4f", f1)
}
