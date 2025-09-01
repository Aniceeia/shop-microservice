package load

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func TestLoadCreateOrders(t *testing.T) {
	baseURL := "http://localhost:8081"
	concurrency := 50
	requestsPerWorker := 20

	var wg sync.WaitGroup
	var mu sync.Mutex
	totalRequests := 0
	failedRequests := 0

	// Load test data
	files, err := os.ReadDir("tests/fixtures/orders")
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No test data found")
	}

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				fileIndex := (workerID*requestsPerWorker + j) % len(files)
				file := files[fileIndex]

				data, err := os.ReadFile("tests/fixtures/orders/" + file.Name())
				if err != nil {
					mu.Lock()
					failedRequests++
					mu.Unlock()
					continue
				}

				resp, err := http.Post(baseURL+"/api/orders", "application/json", bytes.NewReader(data))
				if err != nil {
					mu.Lock()
					failedRequests++
					mu.Unlock()
					continue
				}
				resp.Body.Close()

				mu.Lock()
				totalRequests++
				mu.Unlock()

				if totalRequests%100 == 0 {
					fmt.Printf("Processed %d requests\n", totalRequests)
				}
			}
		}(i)
	}

	wg.Wait()

	duration := time.Since(startTime)
	throughput := float64(totalRequests) / duration.Seconds()

	fmt.Printf("\n=== Load Test Results ===\n")
	fmt.Printf("Total requests: %d\n", totalRequests)
	fmt.Printf("Failed requests: %d\n", failedRequests)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Throughput: %.2f requests/second\n", throughput)
	fmt.Printf("Success rate: %.2f%%\n",
		float64(totalRequests-failedRequests)/float64(totalRequests)*100)
}
