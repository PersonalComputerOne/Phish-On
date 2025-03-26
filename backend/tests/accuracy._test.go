package tests

import (
	"log"
	"testing"

	"github.com/PersonalComputerOne/Phish-On/algorithms"
	"github.com/PersonalComputerOne/Phish-On/db"
)

func TestPhishingAccuracy(t *testing.T) {
	pool, err := db.Init()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Prepare your test data: known phishing and legitimate URLs
	testData := []struct {
		url        string
		isPhishing bool
	}{
		{"http://phishing-example.com", true},
		{"http://legitimate-example.com", false},
		// Add more test URLs here
		{"http://another-phishing-site.com", true},
		{"http://safe-site.org", false},
		// ...
	}

	tp, fp, tn, fn := 0, 0, 0, 0

	for _, data := range testData {
		host, err := algorithms.GetHost(data.url)
		if err != nil {
			log.Printf("Error getting host: %v", err)
			continue
		}

		phishingSet, err := algorithms.BatchPhishingCheck(pool, []string{host})
		if err != nil {
			t.Fatalf("Phishing check failed: %v", err)
		}

		domains, err := algorithms.FetchLegitimateDomains(pool)
		if err != nil {
			t.Fatalf("Failed to fetch legitimate domains: %v", err)
		}

		result := algorithms.ComputeResultForUrl(data.url, host, phishingSet, domains)

		if data.isPhishing && result.IsPhishing {
			tp++
		} else if !data.isPhishing && result.IsPhishing {
			fp++
		} else if !data.isPhishing && !result.IsPhishing {
			tn++
		} else if data.isPhishing && !result.IsPhishing {
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
