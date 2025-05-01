# Mojito

[![Go Report Card](https://goreportcard.com/badge/github.com/wangfenjin/mojito)](https://goreportcard.com/report/github.com/wangfenjin/mojito)
[![License](https://img.shields.io/github/license/wangfenjin/mojito)](https://github.com/wangfenjin/mojito/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/wangfenjin/mojito/graph/badge.svg?token=Id3axA9TgY)](https://codecov.io/gh/wangfenjin/mojito)

Mojito is a production-ready HTTP server template written in Go (Golang), inspired by the [FastAPI Full Stack Template](https://github.com/fastapi/full-stack-fastapi-template). It provides a solid foundation for building scalable and maintainable web applications and APIs using Go.

## Features

*   **Modern Go:** Built with Go 1.24.
*   **High-Performance Routing:** Uses [chi/v5](https://github.com/go-chi/chi) for flexible and fast routing.
*   **Configuration Management:** Leverages [Viper](https://github.com/spf13/viper) for handling configuration from files, environment variables, etc.
*   **Database Integration:** Uses [pgx/v5](https://github.com/jackc/pgx) for efficient PostgreSQL interaction. Includes a basic structure for models and queries (`/models`).
*   **Authentication:** Implements JWT-based authentication (`/common`, `/middleware`).
*   **Request Handling & Validation:** Generic request/response handling middleware with validation using [validator/v10](https://github.com/go-playground/validator).
*   **Middleware:** Includes standard middleware for logging, request ID, recovery, CORS, and authentication.
*   **API Documentation:** Automatic OpenAPI (Swagger) spec generation and Swagger UI endpoint (`/docs/swagger/`).
*   **Dockerized:** Comes with `Dockerfile` for building container images and `docker-compose.yml` for local development database setup.
*   **Development Workflow:**
    *   `Makefile` with commands for common tasks (build, run, test, lint, etc.).
    *   Live reload during development using [Air](https://github.com/air-verse/air) (`make watch`).
*   **Testing:** Includes API tests using [Hurl](https://github.com/Orange-OpenSource/hurl).
*   **Structured Layout:** Follows standard Go project layout conventions.

## Project Structure

```
.
├── api/              # OpenAPI specs, JSON schemas
├── build/            # Packaging and Continuous Integration scripts
│   └── package/      # Dockerfile
├── cmd/              # Main application entrypoints
│   └── mojito/       # Main web server application
├── common/           # Shared utilities (config, JWT, logging, password hashing)
├── config/           # Configuration files (e.g., config.yaml.example)
├── docs/             # Project documentation
├── middleware/       # HTTP middleware (auth, error handling, request parsing)
├── models/           # Database models, queries (using pgx), and schema
│   └── gen/          # Generated code (e.g., from sqlc)
├── openapi/          # OpenAPI generation logic
├── routes/           # API route handlers and definitions
├── tests/            # API tests with hurl
├── .air.toml         # Configuration for Air live reload
├── .gitignore        # Git ignore file
├── docker-compose.yml # Docker Compose for local development (e.g., database)
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
├── LICENSE           # Project License
├── Makefile          # Make commands for development tasks
└── README.md         # This file
```


## Getting Started

### Prerequisites

*   Go >= 1.24
*   Docker & Docker Compose
*   Make

### Setup & Running

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/wangfenjin/mojito.git
    cd mojito
    ```

2.  **Set up Configuration:**
    *   Adjust `config/config.yaml` as needed (especially database credentials if not using default Docker Compose setup).

3.  **Start the Database:**
    ```bash
    docker compose up -d postgres
    ```
    This will start a PostgreSQL container based on the `docker-compose.yml` file. The schema in `models/schema.sql` will be applied automatically on initialization.

4.  **Run the Application:**
    *   **With Live Reload (Recommended for Development):**
        ```bash
        make watch
        ```
        This uses Air to automatically rebuild and restart the server when code changes are detected.
    *   **Without Live Reload:**
        ```bash
        make run
        ```
        Or build and run the binary:
        ```bash
        make build
        ./bin/mojito
        ```

5.  **Access the API:** The server will typically start on `http://localhost:8080` (or as configured).

### Running Tests

*   **Run all tests:**
    ```bash
    make test-api
    ```
*   **Run tests with verbose output:**
    ```bash
    make test-verbose
    ```
*   **Run tests with coverage:**
    ```bash
    make test-coverage
    ```
*   **Generate HTML coverage report:**
    ```bash
    make test-coverage-html
    ```

## Configuration

Configuration is managed by Viper and loaded primarily from `config/config.yaml`. Environment variables can also be used to override settings (refer to `common/config.go` for details).

## API Documentation

Once the server is running, API documentation (Swagger UI) is available at:
`http://localhost:8080/docs/swagger/`

The OpenAPI specification JSON is served at:
`http://localhost:8080/docs/openapi.json`

## License

This project is licensed under the [MIT License](LICENSE).