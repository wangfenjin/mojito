# Mojito

[![Go Report Card](https://goreportcard.com/badge/github.com/wangfenjin/mojito)](https://goreportcard.com/report/github.com/wangfenjin/mojito)
[![License](https://img.shields.io/github/license/wangfenjin/mojito)](https://github.com/wangfenjin/mojito/blob/main/LICENSE)

Mojito is a production-ready HTTP server template in Go, designed to be compatible with the [FastAPI Full Stack Template](https://github.com/fastapi/full-stack-fastapi-template). It provides a robust foundation for building scalable web applications.

## Features

- 📦 Standard Go project layout following [golang-standards](https://github.com/golang-standards/project-layout)
- 🔧 Flexible configuration management with [Viper](https://github.com/spf13/viper)
- 💾 Database operations using [GORM](https://github.com/go-gorm/gorm)
- 🔄 Live reload during development with [Air](https://github.com/air-verse/air)
- 🧪 API testing made easy with [Hurl](https://github.com/Orange-OpenSource/hurl)

## Roadmap

The following features are planned for future releases:

- 🔧 Enhanced configuration management with environment-specific configs and secrets handling
- 📚 OpenAPI/Swagger documentation auto-generation from code
- 🗃️ Database migration tool integration for version-controlled schema changes
- 🐳 Docker support with multi-stage builds and optimized images
- 🔄 GitHub Actions CI/CD pipeline with automated testing and deployment

## Getting Started

1. Run the server:
   ```bash
   make watch
   ```
2. Run unit tests:
   ```bash
   make test
   ```
3. Run API tests:
   ```bash
   make test-api
   ```
4. Generate database migration sqls:
   ```bash
   # start the database, generation only support postgres
   docker compose up -d
   # generate sqls
   make run-migrate
   ```
5. Access the API docs: http://localhost:8080/docs/swagger/

## License
This project is licensed under the [MIT License](LICENSE).