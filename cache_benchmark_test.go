package main

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/arturmon/multi-tier-caching"
	"github.com/arturmon/multi-tier-caching-example/config"
	"github.com/arturmon/multi-tier-caching-example/logger"
	"github.com/arturmon/multi-tier-caching/storage"
)

const maxRecords = 1000 // Number of records
const maxRepeats = 1000

func getRepeats(maxRepeats int) int {
	return rand.Intn(maxRepeats) + 1
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func setKey(cache *multi_tier_caching.MultiTierCache, key, value string, wg *sync.WaitGroup, mu *sync.Mutex, storedKeys map[string]string, setCount *int, totalSetDuration *time.Duration, totalRecords *int, totalRepeats *int, accessCounts map[string]int) {
	defer wg.Done()
	repeats := getRepeats(maxRepeats)

	for j := 0; j < repeats; j++ { // Record with repeat for one key
		start := time.Now()
		err := cache.Set(context.Background(), key, value)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("Failed to set key: %v\n", err)
		} else {
			mu.Lock()
			*setCount++
			storedKeys[key] = value
			*totalSetDuration += duration
			*totalRecords++
			*totalRepeats += repeats

			accessCounts[key]++ // Use accessCounts here

			mu.Unlock()
			fmt.Printf("[SET] Key: %s, Time: %v, Repeat: %d\n", key, duration, j+1)
		}
	}
}

func getKey(cache *multi_tier_caching.MultiTierCache, key, expectedValue string, wg *sync.WaitGroup, mu *sync.Mutex, getCount *int, missCount *int, totalGetDuration *time.Duration, firstAccessTimes map[string]time.Time, lastAccessTimes map[string]time.Time, accessCounts map[string]int) {
	defer wg.Done()
	repeats := getRepeats(maxRepeats)

	for j := 0; j < repeats; j++ { // Read with repeat for one key
		start := time.Now()
		val, err := cache.Get(context.Background(), key)
		duration := time.Since(start)

		mu.Lock()
		if err != nil || val == "" {
			*missCount++
			fmt.Printf("[MISS] Key: %s (not found), Repeat: %d\n", key, j+1)
		} else {
			*getCount++ // Increment getCount only for successful GET operations
			*totalGetDuration += duration
			if val != expectedValue {
				fmt.Printf("[ERROR] Key: %s, Expected: %s, Got: %s, Repeat: %d\n", key, expectedValue, val, j+1)
			} else {
				// We record the time of the first request only once (at the first request)
				if _, exists := firstAccessTimes[key]; !exists {
					firstAccessTimes[key] = time.Now()
				}

				// Update the last access time on each iteration
				lastAccessTimes[key] = time.Now()

				accessCounts[key]++ // Use accessCounts here

				fmt.Printf("[GET] Key: %s, Value: %s, Time: %v, Repeat: %d\n", key, val, duration, j+1)
			}
		}
		mu.Unlock()
	}
}

func TestMultiTierCache(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	cfg := config.LoadConfig()

	fmt.Println("===== CONFIGURATION =====")
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)
	fmt.Printf("Memory Cache Size: %d\n", cfg.MemoryCacheSize)
	fmt.Printf("Database DSN: %s\n", cfg.DatabaseDSN)
	fmt.Printf("Redis Addr: %s\n", cfg.RedisAddr)
	fmt.Printf("Redis Password: %s\n", cfg.RedisPassword)
	fmt.Println("=========================")

	logger.InitLogger(cfg.LogLevel)

	dbStorage, err := storage.NewDatabaseStorage(cfg.DatabaseDSN)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbStorage.Close()
	fmt.Println("Connected to Database")

	redisStorage, err := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis")
	fmt.Println("=========================")

	cache := multi_tier_caching.NewMultiTierCache(
		[]multi_tier_caching.CacheLayer{
			multi_tier_caching.NewMemoryCache(),
			multi_tier_caching.NewRedisCache(redisStorage),
		},
		multi_tier_caching.NewDatabaseCache(dbStorage),
	)

	var (
		wg               sync.WaitGroup
		setCount         int
		getCount         int
		missCount        int
		mu               sync.Mutex
		storedKeys       = make(map[string]string)
		totalSetDuration time.Duration
		totalGetDuration time.Duration
		totalRecords     int
		totalRepeats     int
		firstAccessTimes = make(map[string]time.Time)
		lastAccessTimes  = make(map[string]time.Time)
		accessCounts     = make(map[string]int)
	)

	// Start of time measurement
	startTime := time.Now()

	// Writing and reading to cache
	for i := 0; i < maxRecords; i++ {
		key := randomString(10)
		value := randomString(20)

		// We launch a separate goroutine for writing and reading
		wg.Add(2)

		go setKey(cache, key, value, &wg, &mu, storedKeys, &setCount, &totalSetDuration, &totalRecords, &totalRepeats, accessCounts)
		go getKey(cache, key, value, &wg, &mu, &getCount, &missCount, &totalGetDuration, firstAccessTimes, lastAccessTimes, accessCounts)
	}

	wg.Wait()

	// Calculating the total duration
	elapsedTime := time.Since(startTime)

	// Average write and read time
	var avgSetDuration, avgGetDuration time.Duration
	if setCount > 0 {
		avgSetDuration = totalSetDuration / time.Duration(setCount)
	}
	if getCount > 0 {
		avgGetDuration = totalGetDuration / time.Duration(getCount)
	}

	// Bandwidth
	var throughputSet, throughputGet float64
	if elapsedTime.Seconds() > 0 {
		throughputSet = float64(setCount) / elapsedTime.Seconds()
		throughputGet = float64(getCount) / elapsedTime.Seconds()
	}

	// Missing percentage
	missRate := 0.0
	if getCount > 0 {
		missRate = float64(missCount) / float64(getCount) * 100
	}

	// Results
	fmt.Println("\n===== TEST PERFORMANCE SUMMARY =====")
	fmt.Printf("Total SET operations: %d\n", setCount)
	fmt.Printf("Total GET operations: %d\n", getCount)
	fmt.Printf("Total MISS count: %d\n", missCount)
	fmt.Printf("Total records written: %d\n", totalRecords)
	fmt.Printf("Total repeats performed: %d\n", totalRepeats)
	fmt.Printf("Average SET duration: %v\n", avgSetDuration)
	fmt.Printf("Average GET duration: %v\n", avgGetDuration)
	fmt.Printf("Throughput SET: %.2f ops/sec\n", throughputSet)
	fmt.Printf("Throughput GET: %.2f ops/sec\n", throughputGet)
	fmt.Printf("Miss rate: %.2f%%\n", missRate)

	// Sorted by number of hits
	type keyAccess struct {
		key   string
		count int
		first time.Time
		last  time.Time
	}
	var keyAccesses []keyAccess
	for key, count := range accessCounts {
		keyAccesses = append(keyAccesses, keyAccess{
			key:   key,
			count: count,
			first: firstAccessTimes[key],
			last:  lastAccessTimes[key],
		})
	}

	// Sort by number of requests (descending)
	sort.Slice(keyAccesses, func(i, j int) bool {
		return keyAccesses[i].count > keyAccesses[j].count
	})

	// We bring out the top 10
	fmt.Println("\n===== Top 10 Keys by Access Count =====")
	for i, ka := range keyAccesses[:10] {
		firstAccessMs := ka.first.Sub(startTime).Milliseconds()
		lastAccessMs := ka.last.Sub(startTime).Milliseconds()
		fmt.Printf("Rank %d - Key: %s, Access Count: %d, First Access: %dms, Last Access: %dms\n", i+1, ka.key, ka.count, firstAccessMs, lastAccessMs)
	}

	fmt.Println("========================")
}
