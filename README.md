# Distributed Rate Limiter

## Overview
This project implements a **distributed rate limiter** using Go and Redis.  
It enforces request limits per user using a **sliding window algorithm** for better accuracy.

The system is designed to work across multiple backend instances and handles high concurrency using Redis and Lua scripting.

---

## Features
- Sliding window rate limiting (more accurate than fixed window)
- Distributed architecture using Redis
- Atomic operations using Redis Lua script (prevents race conditions)
- Configurable limits via environment variables
- REST API endpoint: `/request`
- Load testing with k6

---

## Architecture
- **Go API** handles incoming requests
- **Redis** stores request counters (shared across all instances)
- **Lua script** ensures atomic rate-limit checks

All backend instances share Redis, ensuring consistent limits across servers.

---

## Configuration
Environment variables:

```bash
RATE_LIMIT=10       # number of requests allowed
RATE_WINDOW=10      # time window in seconds
REDIS_ADDR=redis:6379
PORT=8080