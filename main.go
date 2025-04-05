package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"api-monitor/database"
	"api-monitor/handlers"
	"api-monitor/middleware"
	"api-monitor/models"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/lib/pq"
)

// Endpoint represents an API endpoint to monitor
type Endpoint struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Interval    int       `json:"interval"` // in seconds
	LastChecked time.Time `json:"last_checked"`
	Status      string    `json:"status"`
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID         int       `json:"id"`
	EndpointID int       `json:"endpoint_id"`
	Status     int       `json:"status"`
	Response   string    `json:"response"`
	CheckedAt  time.Time `json:"checked_at"`
}

var endpoints []Endpoint

// Track consecutive failures for each endpoint
var consecutiveFailures = make(map[int]int)

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Load endpoints from database
	if err := handlers.LoadEndpoints(); err != nil {
		log.Fatalf("Failed to load endpoints: %v", err)
	}

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Serve static files
	e.Static("/static", "static")

	// Web routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})
	e.GET("/login", func(c echo.Context) error {
		return c.File("static/login.html")
	})
	e.GET("/register", func(c echo.Context) error {
		return c.File("static/register.html")
	})
	e.GET("/dashboard", func(c echo.Context) error {
		return c.File("static/index.html")
	})

	// Public API routes
	e.POST("/register", handlers.CreateUser)
	e.POST("/login", handlers.Login)

	// Protected API routes
	api := e.Group("/api")
	api.Use(middleware.JWT([]byte("your-secret-key"))) // Replace with your secret key

	// User routes
	api.GET("/user", handlers.GetUser)
	api.PUT("/user", handlers.UpdateUser)
	api.GET("/subscription", handlers.GetSubscription)

	// Endpoint routes
	api.POST("/endpoints", handlers.CreateEndpoint)
	api.GET("/endpoints", handlers.GetEndpoints)
	api.GET("/endpoints/:id", handlers.GetEndpoint)
	api.PUT("/endpoints/:id", handlers.UpdateEndpoint)
	api.DELETE("/endpoints/:id", handlers.DeleteEndpoint)

	// Schedule routes
	api.POST("/schedules", handlers.CreateScheduleHandler)
	api.GET("/schedules", handlers.GetSchedulesHandler)
	api.GET("/schedules/:id", handlers.GetScheduleHandler)
	api.PUT("/schedules/:id", handlers.UpdateScheduleHandler)
	api.DELETE("/schedules/:id", handlers.DeleteScheduleHandler)

	// Start health monitoring in background
	go startHealthMonitoring()

	// Start expiry checker in background
	go startExpiryChecker()

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func startHealthMonitoring() {
	for {
		// Load all non-expired endpoints from database
		var dbEndpoints []database.Endpoint
		if err := database.DB.Where("expires_at > ? OR expires_at IS NULL", time.Now()).Find(&dbEndpoints).Error; err != nil {
			log.Printf("Failed to load endpoints: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		log.Printf("Starting health monitoring for %d endpoints", len(dbEndpoints))

		// Group endpoints by interval
		intervalGroups := make(map[int][]database.Endpoint)
		for _, endpoint := range dbEndpoints {
			intervalGroups[endpoint.Interval] = append(intervalGroups[endpoint.Interval], endpoint)
		}

		// Create schedules for each interval group
		tickers := make(map[int]*time.Ticker)
		for interval, endpoints := range intervalGroups {
			// Create or update schedule
			schedule := &database.Schedule{
				Name:      fmt.Sprintf("Schedule for %d second interval", interval),
				Interval:  interval,
				CreatedAt: time.Now(),
				Endpoints: make(pq.Int64Array, len(endpoints)),
			}
			for i, endpoint := range endpoints {
				schedule.Endpoints[i] = int64(endpoint.ID)
			}

			// Create or update schedule in database
			if err := database.DB.Where("interval = ?", interval).FirstOrCreate(schedule).Error; err != nil {
				log.Printf("Failed to create/update schedule for interval %d: %v", interval, err)
				continue
			}

			// Create ticker for this interval
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			tickers[interval] = ticker

			// Start goroutine for this schedule
			go func(s database.Schedule) {
				log.Printf("Started monitoring schedule: %s (ID: %d) with interval %d seconds", s.Name, s.ID, s.Interval)
				for range ticker.C {
					// Load endpoints for this schedule
					var currentEndpoints []database.Endpoint
					if err := database.DB.Where("id = ANY(?) AND (expires_at > ? OR expires_at IS NULL)", s.Endpoints, time.Now()).Find(&currentEndpoints).Error; err != nil {
						log.Printf("Failed to load endpoints for schedule %d: %v", s.ID, err)
						continue
					}

					log.Printf("Schedule %s (ID: %d) checking %d endpoints", s.Name, s.ID, len(currentEndpoints))
					// Check each endpoint
					for _, dbEndpoint := range currentEndpoints {
						endpoint := dbEndpoint.ToModel()
						go checkEndpoint(&endpoint)
					}
				}
			}(*schedule)
		}

		// Keep the function running until an error occurs
		select {
		case <-time.After(24 * time.Hour):
			// Restart every 24 hours to prevent memory leaks
			log.Println("Restarting health monitoring after 24 hours")
			return
		}
	}
}

func startExpiryChecker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		handlers.CheckExpiredEndpoints()
	}
}

func checkEndpoint(endpoint *models.Endpoint) {
	log.Printf("Checking endpoint: %s", endpoint.URL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try up to 3 times
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(endpoint.URL)
		if err != nil {
			if i < maxRetries-1 {
				log.Printf("Retry %d/%d for endpoint %s: %v", i+1, maxRetries, endpoint.URL, err)
				time.Sleep(2 * time.Second)
				continue
			}

			endpoint.Status = "error"
			endpoint.LastChecked = time.Now()
			if err := database.UpdateEndpointStatus(endpoint.ID, "error"); err != nil {
				log.Printf("Failed to update endpoint status: %v", err)
			}
			log.Printf("Endpoint check failed after %d retries: %s - Status: error", maxRetries, endpoint.URL)
			return
		}
		defer resp.Body.Close()

		// Check if status code is not in 2xx range
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			consecutiveFailures[endpoint.ID]++
			if consecutiveFailures[endpoint.ID] >= 3 {
				log.Printf("WARNING: Endpoint %s has returned non-2xx status code (%d) for 3 consecutive checks!", endpoint.URL, resp.StatusCode)
			}
		} else {
			// Reset consecutive failures counter on successful response
			consecutiveFailures[endpoint.ID] = 0
		}

		endpoint.Status = "ok"
		endpoint.LastChecked = time.Now()
		if err := database.UpdateEndpointStatus(endpoint.ID, "ok"); err != nil {
			log.Printf("Failed to update endpoint status: %v", err)
		}
		log.Printf("Endpoint check successful: %s - Status: ok (HTTP %d)", endpoint.URL, resp.StatusCode)
		return
	}
}
