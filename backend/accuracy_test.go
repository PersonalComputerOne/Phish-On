package main

import (
	"log"
	"testing"

	"github.com/PersonalComputerOne/Phish-On/db"
)

func TestPhishingAccuracy(t *testing.T) {
	pool, err := db.Init()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	testData := []struct {
		url        string
		isPhishing bool
	}{
		{"http://github.com", false},
		{"http://githun.com", true},
		{"http://linkedin.com", false},
		{"http://linkedin.org", true},
		{"http://example.org", true},
	}

	tp, fp, tn, fn := 0, 0, 0, 0

	for _, data := range testData {
		host, err := getHost(data.url)
		if err != nil {
			log.Printf("Error getting host: %v", err)
			continue
		}

		phishingSet, err := batchPhishingCheck(pool, []string{host})
		if err != nil {
			t.Fatalf("Phishing check failed: %v", err)
		}

		domains, err := fetchLegitimateDomains(pool)
		if err != nil {
			t.Fatalf("Failed to fetch legitimate domains: %v", err)
		}

		result := computeResultForUrl(data.url, host, phishingSet, domains)

		if data.isPhishing && !result.IsReal {
			tp++
		} else if !data.isPhishing && !result.IsReal {
			fp++
		} else if !data.isPhishing && result.IsReal {
			tn++
		} else if data.isPhishing && result.IsReal {
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
