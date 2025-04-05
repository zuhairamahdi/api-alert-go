# API Monitor

A robust API health monitoring system built with Go and PostgreSQL. Monitor multiple endpoints with customizable check intervals, receive real-time status updates, and get alerts for persistent failures.

## Features

- **Real-time Monitoring**: Continuous health checks of your API endpoints
- **Customizable Intervals**: Support for multiple check intervals:
  - 5 seconds (for testing)
  - 1 minute
  - 5 minutes
  - 15 minutes
  - 30 minutes
- **Smart Scheduling**: Automatic grouping of endpoints by interval for efficient monitoring
- **Failure Detection**: Alerts for persistent failures (3 consecutive non-2xx responses)
- **User Management**:
  - JWT-based authentication
  - User registration and login
  - Subscription-based access control
- **Endpoint Management**:
  - Add/remove endpoints
  - Set custom check intervals
  - View endpoint status and history
- **Subscription System**:
  - Free tier with trial period
  - Endpoint limits
  - Interval restrictions based on subscription level

## Tech Stack

- **Backend**: Go
- **Database**: PostgreSQL
- **Authentication**: JWT
- **ORM**: GORM
- **Web Framework**: Echo
- **Frontend**: HTML, JavaScript, Tabler UI

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL
- Docker (optional)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/zuhairamahdi/api-alert-go.git
   cd api-alert-go
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=api_monitor
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

### Docker Setup

1. Build and run with Docker Compose:
   ```bash
   docker-compose up --build
   ```

## Usage

1. Register a new account at `/register`
2. Log in at `/login`
3. Access the dashboard at `/dashboard`
4. Add endpoints with your desired check intervals
5. Monitor endpoint health in real-time

## API Endpoints

### Public Endpoints
- `POST /register` - Register a new user
- `POST /login` - User login

### Protected Endpoints
- `GET /api/user` - Get user information
- `PUT /api/user` - Update user profile
- `GET /api/subscription` - Get subscription details
- `POST /api/endpoints` - Create a new endpoint
- `GET /api/endpoints` - List all endpoints
- `GET /api/endpoints/:id` - Get endpoint details
- `PUT /api/endpoints/:id` - Update endpoint
- `DELETE /api/endpoints/:id` - Delete endpoint

## Health Monitoring

The system performs health checks by:
1. Grouping endpoints by interval
2. Creating schedules for each interval group
3. Running health checks at the specified intervals
4. Updating endpoint status in real-time
5. Alerting on persistent failures (3 consecutive non-2xx responses)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Echo](https://echo.labstack.com/) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [Tabler](https://tabler.io/) - UI components 
