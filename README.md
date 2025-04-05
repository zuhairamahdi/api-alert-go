# API Monitor

application built with Go and Echo framework for monitoring API endpoints' health. The application allows you to register endpoints and automatically checks their health status at specified intervals.

## Features

- Register and manage API endpoints to monitor
- Create and manage monitoring schedules
- Automatic health checks at configurable intervals
- Track endpoint status and response times
- RESTful API for managing endpoints and schedules
- Concurrent health checking for multiple endpoints
- PostgreSQL database for persistent storage

## Getting Started

### Prerequisites

- Go 1.21 or higher (for local development)
- Docker and Docker Compose (for containerized deployment)
- PostgreSQL (for local development)

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/yourusername/api-monitor.git
cd api-monitor
```

2. Install dependencies:
```bash
go mod download
```

3. Set up the database:
```bash
createdb api_monitor
```

4. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Docker Deployment

1. Clone the repository:
```bash
git clone https://github.com/zuhairamahdi/api-alert-go.git
cd api-alert-go
```

2. Build and run with Docker Compose:
```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`

## API Endpoints

### Endpoints Management

- `POST /endpoints` - Create a new endpoint to monitor
  ```json
  {
    "url": "https://api.example.com/health",
    "interval": 30
  }
  ```

- `GET /endpoints` - List all monitored endpoints
- `GET /endpoints/:id` - Get details of a specific endpoint
- `PUT /endpoints/:id` - Update an endpoint's configuration
- `DELETE /endpoints/:id` - Remove an endpoint from monitoring

### Schedule Management

- `POST /schedules` - Create a new monitoring schedule
  ```json
  {
    "name": "Daily Check",
    "interval": 3600,
    "endpoints": [1, 2, 3]
  }
  ```

- `GET /schedules` - List all monitoring schedules
- `GET /schedules/:id` - Get details of a specific schedule
- `PUT /schedules/:id` - Update a schedule's configuration
- `DELETE /schedules/:id` - Remove a schedule

## Example Usage

1. Create a new endpoint to monitor:
```bash
curl -X POST http://localhost:8080/endpoints \
  -H "Content-Type: application/json" \
  -d '{"url": "https://api.example.com/health", "interval": 30}'
```

2. Create a new schedule:
```bash
curl -X POST http://localhost:8080/schedules \
  -H "Content-Type: application/json" \
  -d '{"name": "Hourly Check", "interval": 3600, "endpoints": [1]}'
```

3. List all schedules:
```bash
curl http://localhost:8080/schedules
```

## Environment Variables

The following environment variables can be configured:

- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL user (default: postgres)
- `DB_PASSWORD` - PostgreSQL password (default: postgres)
- `DB_NAME` - PostgreSQL database name (default: api_monitor)

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
