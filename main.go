package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient   *redis.Client
	limiterScript *redis.Script
	LIMIT         int
	WINDOW        int
)

func init() {
	// Config
	LIMIT, _ = strconv.Atoi(getEnv("RATE_LIMIT", "10"))
	WINDOW, _ = strconv.Atoi(getEnv("RATE_WINDOW", "10"))

	// Redis Connection
	redisClient = redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
		PoolSize: 100,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✓ Connected to Redis successfully")

	// Load Lua Script
	scriptData, err := os.ReadFile("limiter.lua")
	if err != nil {
		log.Fatalf("Failed to read lua script: %v", err)
	}
	log.Printf("✓ Loaded Lua script (%d bytes)", len(scriptData))
	
	limiterScript = redis.NewScript(string(scriptData))
	log.Printf("✓ Rate limiter initialized: %d requests per %d seconds", LIMIT, WINDOW)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.RemoteAddr
	}

	key := fmt.Sprintf("rl:%s", userID)
	now := time.Now().Unix()
	
	// Debug: calculate windows
	currentWindow := now / int64(WINDOW)
	prevWindow := currentWindow - 1
	windowElapsed := now % int64(WINDOW)
	weight := float64(WINDOW-int(windowElapsed)) / float64(WINDOW)
	
	log.Printf("🔍 [%s] now=%d, currWin=%d, prevWin=%d, elapsed=%d, weight=%.2f", 
		userID, now, currentWindow, prevWindow, windowElapsed, weight)

	// Execute Lua Script
	result, err := limiterScript.Run(ctx, redisClient, []string{key}, LIMIT, WINDOW, now).Result()
	
	if err != nil {
		log.Printf("✗ Redis error: %v", err)
		http.Error(w, fmt.Sprintf("Redis error: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [%s] Lua result: %v (type: %T)", userID, result, result)

	if result == int64(1) {
		log.Printf("✓ [%s] ALLOWED", userID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request Allowed"))
	} else {
		log.Printf("✗ [%s] BLOCKED", userID)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate Limit Exceeded"))
	}
}
func main() {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Distributed Rate Limiter is running. Use /request endpoint."))
    })

	port := getEnv("PORT", "8080")
	fmt.Printf("🚀 Server starting on port %s (Limit: %d reqs / %d sec)\n", port, LIMIT, WINDOW)
	log.Fatal(http.ListenAndServe(":"+port, r))
}