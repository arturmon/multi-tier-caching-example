package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arturmon/multi-tier-caching"
	"github.com/arturmon/multi-tier-caching-example/config"
	"github.com/arturmon/multi-tier-caching-example/logger"
	"github.com/arturmon/multi-tier-caching/storage"
)

func main() {

	cfg := config.LoadConfig()

	logger.InitLogger(cfg.LogLevel)
	logger.Info("Launching the cache system", "memoryCacheSize", cfg.MemoryCacheSize)

	logger.Info("Launching the cache system", "memoryCacheSize", cfg.DatabaseDSN)

	memoryStorage := multi_tier_caching.NewCacheLayer(cfg.UseMemcached, cfg.MemcachedADDR)

	dbStorage, err := storage.NewDatabaseStorage(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("Failed to connect to the database", "error", err)
		return
	}
	defer dbStorage.Close()

	redisStorage, err := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return
	}

	cache := multi_tier_caching.NewMultiTierCache(
		[]multi_tier_caching.CacheLayer{
			memoryStorage,
			multi_tier_caching.NewRedisCache(redisStorage),
		},
		multi_tier_caching.NewDatabaseCache(dbStorage),
	)

	err = cache.Set(context.Background(), "key1", "value1")
	if err != nil {
		return
	}
	val, _ := cache.Get(context.Background(), "key1")
	fmt.Println("Cached Value:", val)

	// Waiting for the program to complete
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down...")
}
