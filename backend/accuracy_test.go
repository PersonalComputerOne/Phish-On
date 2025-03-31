package main

import (
	"encoding/csv"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/PersonalComputerOne/Phish-On/db"
)

func TestPhishingAccuracy(t *testing.T) {
	const testLimit = 100 // Change this value to control how many URLs to test

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
	defer file.Close()

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

	var urls []string
	testCases := make([]struct {
		url        string
		isPhishing bool
		host       string
	}, 0, len(records))

	for _, record := range records {
		url := record[0]
		status, err := strconv.Atoi(record[1])
		if err != nil {
			t.Fatalf("Failed to parse status value: %v", err)
		}
		isPhishing := status == 0

		host, err := getHost(url)
		if err != nil {
			t.Logf("Skipping URL due to host error: %s | Error: %v", url, err)
			continue
		}

		urls = append(urls, url)
		testCases = append(testCases, struct {
			url        string
			isPhishing bool
			host       string
		}{url, isPhishing, host})
	}

	phishingSet, err := batchPhishingCheck(pool, urls)
	if err != nil {
		t.Fatalf("Phishing check failed: %v", err)
	}

	domains, err := fetchLegitimateDomains(pool)
	if err != nil {
		t.Fatalf("Failed to fetch legitimate domains: %v", err)
	}

	tp, fp, tn, fn := 0, 0, 0, 0
	for _, tc := range testCases {
		result := computeResultForUrl(tc.url, tc.host, phishingSet, domains)

		switch {
		case tc.isPhishing && result.IsPhishing:
			tp++
		case !tc.isPhishing && result.IsPhishing:
			fp++
		case !tc.isPhishing && !result.IsPhishing:
			tn++
		case tc.isPhishing && !result.IsPhishing:
			fn++
		}
	}

	var precision, recall, f1 float64
	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	} else {
		precision = 0.0
	}
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	} else {
		recall = 0.0
	}
	if precision+recall > 0 {
		f1 = 2 * (precision * recall) / (precision + recall)
	} else {
		f1 = 0.0
	}

	t.Logf("True Positives: %d", tp)
	t.Logf("False Positives: %d", fp)
	t.Logf("True Negatives: %d", tn)
	t.Logf("False Negatives: %d", fn)
	t.Logf("Precision: %.4f", precision)
	t.Logf("Recall: %.4f", recall)
	t.Logf("F1 Score: %.4f", f1)
}
