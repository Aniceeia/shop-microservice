package load

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestLoadCreateOrders(t *testing.T) {
	baseURL := "http://localhost:8081"
	concurrency := 500
	requestsTotal := 100000

	var wg sync.WaitGroup
	var mu sync.Mutex
	processedRequests := 0
	failedRequests := 0

	tryPaths := []string{
		filepath.Clean(filepath.Join("..", "fixtures", "orders")),
		filepath.Clean(filepath.Join("tests", "fixtures", "orders")),
	}
	var fixturesDir string
	var files []os.DirEntry
	var err error
	for _, p := range tryPaths {
		files, err = os.ReadDir(p)
		if err == nil {
			fixturesDir = p
			break
		}
	}
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No test data found")
	}

	fmt.Printf("Found %d fixture files\n", len(files))

	startTime := time.Now()

	workChan := make(chan int, requestsTotal)
	for i := 0; i < requestsTotal; i++ {
		workChan <- i
	}
	close(workChan)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for range workChan {
				fileIndex := processedRequests % len(files)
				file := files[fileIndex]

				data, err := os.ReadFile(filepath.Join(fixturesDir, file.Name()))
				if err != nil {
					mu.Lock()
					failedRequests++
					processedRequests++
					mu.Unlock()
					continue
				}

				resp, err := http.Post(baseURL+"/api/orders", "application/json", bytes.NewReader(data))
				if err != nil {
					mu.Lock()
					failedRequests++
					processedRequests++
					mu.Unlock()
					continue
				}
				resp.Body.Close()

				mu.Lock()
				processedRequests++
				currentCount := processedRequests
				mu.Unlock()

				if currentCount%5000 == 0 {
					fmt.Printf("Processed %d requests\n", currentCount)
				}
			}
		}(i)
	}

	wg.Wait()

	duration := time.Since(startTime)
	throughput := float64(processedRequests) / duration.Seconds()

	fmt.Printf("\n=== Load Test Results ===\n")
	fmt.Printf("Total requests: %d\n", processedRequests)
	fmt.Printf("Failed requests: %d\n", failedRequests)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Throughput: %.2f requests/second\n", throughput)
	fmt.Printf("Success rate: %.2f%%\n",
		float64(processedRequests-failedRequests)/float64(processedRequests)*100)
}
