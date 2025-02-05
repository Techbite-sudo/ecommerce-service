package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type visitor struct {
	count     int
	lastSeen  time.Time
	resetTime time.Time
}

type RateLimiterConfig struct {
	Max      int           // Maximum number of requests allowed
	Duration time.Duration // Time window for the rate limit
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.RWMutex
)

// RateLimiter creates a middleware that limits requests based on IP address
func RateLimiter() fiber.Handler {
	// Clean up old entries every minute
	go cleanupVisitors()

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		mu.Lock()
		v, exists := visitors[ip]

		now := time.Now()

		if !exists {
			// First visit
			visitors[ip] = &visitor{
				count:     1,
				lastSeen:  now,
				resetTime: now.Add(1 * time.Minute),
			}
			mu.Unlock()
			return c.Next()
		}

		// Reset count if time window has passed
		if now.After(v.resetTime) {
			v.count = 1
			v.lastSeen = now
			v.resetTime = now.Add(1 * time.Minute)
			mu.Unlock()
			return c.Next()
		}

		// Increment count and update last seen
		v.count++
		v.lastSeen = now

		// Check if rate limit exceeded
		if v.count > 60 { // 60 requests per minute
			mu.Unlock()
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
				"reset": v.resetTime.Unix(),
			})
		}

		mu.Unlock()
		return c.Next()
	}
}

// cleanupVisitors removes old visitor entries periodically
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// Custom rate limiter for specific routes or different limits
func CustomRateLimiter(config RateLimiterConfig) fiber.Handler {
	visitors := make(map[string]*visitor)
	var mu sync.RWMutex

	// Cleanup routine
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > config.Duration*2 {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		mu.Lock()
		v, exists := visitors[ip]

		now := time.Now()

		if !exists {
			visitors[ip] = &visitor{
				count:     1,
				lastSeen:  now,
				resetTime: now.Add(config.Duration),
			}
			mu.Unlock()
			return c.Next()
		}

		if now.After(v.resetTime) {
			v.count = 1
			v.lastSeen = now
			v.resetTime = now.Add(config.Duration)
			mu.Unlock()
			return c.Next()
		}

		v.count++
		v.lastSeen = now

		if v.count > config.Max {
			mu.Unlock()
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
				"reset": v.resetTime.Unix(),
			})
		}

		mu.Unlock()
		return c.Next()
	}
}
