# AI Codebase Management Guide - Mojito Project

## 1. Project Overview

*   **Project Name**: Mojito
*   **Description**: Mojito is a production-ready HTTP server template written in Go (Golang), inspired by the FastAPI Full Stack Template. It provides a solid foundation for building scalable and maintainable web applications and APIs using Go.
*   **Primary Language**: Go (Golang)

## 2. Key Features

*   **API Documentation**: Automatically generates OpenAPI (Swagger) specifications and provides Swagger UI via `/docs/swagger/`.
*   **Dockerized**: Provides a `Dockerfile` for building container images and `docker-compose.yml` for local development database setup.
*   **Development Workflow**:
    *   `Makefile` contains common task commands (build, run, test, lint, etc.).
    *   Uses [Air](https://github.com/air-verse/air) for live reloading during development (`make watch`).
*   **Testing**: Includes API tests using [Hurl](https://github.com/Orange-OpenSource/hurl).
*   **Structured Layout**: Follows standard Go project layout conventions.

## 3. Project Structure and Key Directories

.
├── api/              # OpenAPI specifications, JSON schemas (e.g., fastapi.json, openapi.json)
├── assets/           # Other project-related resources (images, logos, etc.)
├── build/            # Packaging and continuous integration scripts
│   └── package/      # Dockerfile is located here
├── cmd/              # Main application entry points
│   └── mojito/       # Main web server application (main.go is here)
├── common/           # Shared utilities (config, JWT, logging, password hashing)
├── config/           # Configuration files (e.g., config.yaml.example)
├── docs/             # Project documentation (e.g., /docs/swagger/ provides API docs)
├── middleware/       # HTTP middleware (authentication, error handling, request parsing)
├── models/           # Database models, queries (using pgx), and schemas
│   └── gen/          # Generated code (e.g., from sqlc)
├── openapi/          # OpenAPI generation logic (e.g., registry.go)
├── routes/           # API route handlers and definitions (e.g., login.go, test.go, utils.go, docs.go)
├── tests/            # API tests (using hurl)
├── .air.toml         # Configuration for Air live reloading
├── .gitignore        # Git ignore file
├── docker-compose.yml # Docker Compose for local development (e.g., database)
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
├── LICENSE           # Project license
├── Makefile          # Make commands for development tasks
└── README.md         # Project README file

## 4. Core Components and Logic

*   **Entry Point**:
    *   The main entry point for the application is located in <mcfile name="main.go" path="./cmd/mojito/main.go"></mcfile>.
*   **Configuration Management**:
    *   Configuration is loaded via the `Load` function in <mcfile name="config.go" path="./common/config.go"></mcfile>.
    *   Configuration files are typically located in the `./config` directory and support environment variable overrides (prefixed with `MOJITO`).
*   **API Definition & Documentation**:
    *   OpenAPI specifications are usually stored in the <mcfolder name="api" path="./api"></mcfolder> directory (e.g., <mcfile name="fastapi.json" path="./api/fastapi.json"></mcfile> or `openapi.json`).
    *   <mcfile name="docs.go" path="./routes/docs.go"></mcfile> is responsible for serving the Swagger UI and OpenAPI specification.
    *   <mcfile name="registry.go" path="./openapi/registry.go"></mcfile> contains logic related to dynamic registration and generation of OpenAPI information.
*   **Routing**:
    *   Uses `chi` as the HTTP router.
    *   Route definitions are in various files under the <mcfolder name="routes" path="./routes"></mcfolder> directory, such as <mcfile name="login.go" path="./routes/login.go"></mcfile>, <mcfile name="test.go" path="./routes/test.go"></mcfile>, and <mcfile name="utils.go" path="./routes/utils.go"></mcfile>.
    *   The main route registration logic calls `routes.RegisterRoutes(r)` in <mcfile name="main.go" path="./cmd/mojito/main.go"></mcfile>.
*   **Middleware**:
    *   HTTP middleware is located in the <mcfolder name="middleware" path="./middleware"></mcfolder> directory.
    *   <mcfile name="handler.go" path="./middleware/handler.go"></mcfile> contains a generic middleware `WithHandler` for request handling, parameter parsing, validation, and response writing, as well as an authentication middleware `RequireAuth`.
*   **Database Interaction**:
    *   Database models and query logic are in the <mcfolder name="models" path="./models"></mcfolder> directory.
    *   Uses `pgx` to interact with a PostgreSQL database.
    *   `sqlc` might be used to generate type-safe code from SQL (typically in the `models/gen/` directory).
        *   **Schema Files**: All SQL schema definitions (DDL) must be placed in the `models/migrations` directory. Each schema change should be split into two files:
            - `XXX_schema.up.sql`: Contains the forward migration (creating/altering tables)
            - `XXX_schema.down.sql`: Contains the rollback migration (dropping/reverting changes)
            - up.sql and down.sql should always be updated together to ensure consistency.
            - Files should follow a migration-friendly naming convention, e.g., `001_create_users.up.sql`/`001_create_users.down.sql`, `002_add_email_to_users.up.sql`/`002_add_email_to_users.down.sql`
        *   **Query Files**: All SQLC query files (DML) must be placed in the `models/queries/` directory, e.g., `users.sql`, `products.sql`.
        *   **Timestamp Trigger Function**: The project uses a common trigger function `update_updated_at_column()` to automatically update `updated_at` timestamps. This function is already defined in the database (e.g., by an earlier migration script, subsequent schema files that require this functionality for new tables should apply the existing trigger and MUST NOT redefine the function.
        *   **UUID Primary Keys**: For tables using UUIDs as primary keys, the UUID should be generated in the application code (Go) before insertion. Do not set a default value (e.g., `uuid_generate_v4()`) for UUID primary key columns in the SQL schema.
    *   **Data Deletion**: For data deletion operations, always prefer soft deletes over hard deletes. This typically involves adding a `deleted_at` (timestamp, nullable) column to relevant tables and updating this field instead of physically removing the row. Ensure queries filter out soft-deleted records by default unless explicitly requested.
*   **Dependency Management**:
    *   Project dependencies are managed using Go Modules, defined in the <mcfile name="go.mod" path="./go.mod"></mcfile> file.

## 5. Development and Building

*   **Makefile**: The <mcfile name="Makefile" path="./Makefile"></mcfile> provides common development commands:
    *   `make build`: Builds the application.
    *   `make run`: Runs the application.
    *   `make build-run`: Builds and runs the application.
    *   `make watch`: Uses Air for live reloading.
    *   `make test`: Runs unit tests.
    *   `make test-api`: Runs Hurl API tests (test files are in the <mcfolder name="tests" path="./tests"></mcfolder> directory).
    *   `make test-all`: Runs all tests.
    *   `make clean`: Cleans build artifacts.
*   **Docker**:
    *   <mcfile name="Dockerfile" path="./build/package/Dockerfile"></mcfile> (typically at this path) is used to build Docker images.
    *   <mcfile name="docker-compose.yml" path="./docker-compose.yml"></mcfile> is used for the local development environment (e.g., starting a database service).

## 6. AI Interaction Suggestions

*   **Modifying Code**:
    *   Adding new routes: Create or modify the corresponding `*.go` file in the <mcfolder name="routes" path="./routes"></mcfolder> directory and ensure it's registered in `routes.RegisterRoutes` (usually called in `cmd/mojito/main.go`).
    *   Modifying models: Update SQL schemas or Go structs under <mcfolder name="models" path="./models"></mcfolder>. If `sqlc` is used, code regeneration might be needed.
    *   Adding middleware: Implement in the <mcfolder name="middleware" path="./middleware"></mcfolder> directory and apply it in `cmd/mojito/main.go`.
*   **Understanding the API**:
    *   Consult `/docs/swagger/` or <mcfile name="api/openapi.json" path="./api/openapi.json"></mcfile> (or <mcfile name="api/fastapi.json" path="./api/fastapi.json"></mcfile>) to understand API endpoints, request/response structures. The <mcfile name="api/openapi.json" path="./api/openapi.json"></mcfile> is generated based on the handler function signatures.
    *   A good way to understand our API design directly from the code is to look at the handler functions defined in the <mcfolder name="routes" path="./routes/"></mcfolder> folder. These functions generally follow the signature `handlerFunc(ctx context.Context, req RequestType) (ResponseType, error)`.
    *   For consistency and clarity, new handler functions should also adhere to this `handlerFunc(ctx, req) (resp, error)` signature.
*   **Running and Testing**:
    *   Use commands from the `Makefile` for building, running, and testing.
    *   API test cases are in the <mcfolder name="tests" path="./tests"></mcfolder> directory, using Hurl format.
