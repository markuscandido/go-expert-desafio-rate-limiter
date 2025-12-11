# Project Overview

This project is a high-performance rate limiter middleware for Go web servers. It is designed to be easily integrated into any Go HTTP server to protect services from excessive requests. The rate limiter supports both IP-based and token-based limiting, with token-based limiting taking precedence. It uses Redis for distributed and persistent storage of rate limit data.

The architecture is composed of the following layers:
- **Config Layer (`config/`)**: Loads configuration from environment variables or a `.env` file.
- **Storage Layer (`storage/`)**: Defines a `Strategy` interface for storage backends and provides a Redis implementation. This design allows for easy extension to other storage systems.
- **Limiter Logic (`limiter/`)**: Contains the core rate limiting logic.
- **Middleware (`middleware/`)**: Provides an HTTP middleware for easy integration with Go web servers.
- **Main (`main.go`)**: A sample web server that demonstrates how to use the rate limiter middleware.

# Building and Running

## Using Docker Compose (Recommended)

The easiest way to get the project running is by using Docker Compose.

```bash
# Start the services in detached mode
docker-compose up -d
```

To stop the services:

```bash
docker-compose down
```

## Manual Setup

**Prerequisites:**
- Go 1.21 or higher
- Redis 6.0 or higher

**Steps:**

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd rate-limiter
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Configure the environment:**
    Create a `.env` file by copying the example and customizing it if needed.
    ```bash
    cp .env.example .env
    ```

4.  **Build the application:**
    ```bash
    go build -o rate-limiter .
    ```

5.  **Run the application:**
    Make sure your Redis server is running, then execute the following command:
    ```bash
    ./rate-limiter
    ```

# Testing

To run the test suite, use the following command:

```bash
go test ./...
```

To run tests with coverage:

```bash
go test -cover ./...
```

# Development Conventions

- **Configuration**: All configuration is handled through environment variables, with support for `.env` files for local development.
- **Storage**: The storage backend is abstracted through the `storage.Strategy` interface, making it possible to swap out Redis for another storage system.
- **Middleware**: The rate limiter is implemented as a standard Go HTTP middleware, making it easy to integrate with any Go web server.
- **Testing**: The project has a suite of unit tests that cover the core logic and middleware.
