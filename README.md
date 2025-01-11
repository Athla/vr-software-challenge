# VR Software Challenge - Transaction Processing API

A robust HTTP REST API built in Go for processing purchase transactions with asynchronous queue handling. The application provides a reliable way to store and process financial transactions with proper validation and audit logging.

## Features

- ✅ RESTful API endpoints for transaction management
- ✅ Asynchronous processing using Kafka
- ✅ PostgreSQL for persistent storage
- ✅ Transaction audit logging
- ✅ Comprehensive validation
- ✅ Rate limiting
- ✅ Health checks
- ✅ Docker containerization
- ✅ Currency conversion using Treasury Reporting Rates API

## Technology Stack

- **Language:** Go 1.23
- **Database:** PostgreSQL 16
- **Message Queue:** Apache Kafka
- **Container:** Docker & Docker Compose

## Prerequisites

- Docker and Docker Compose
- Go 1.23 or higher (for local development)
- Make (optional, for using Makefile commands)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/Athla/vr-software-challenge.git
cd vr-software-challenge
```

2. Copy the environment file and adjust as needed:
```bash
cp .env.example .env
```
2.1. Copy your .env to the ./config and ./tests, so it can be used throughout the application.

3. Start the services:
```bash
make docker-run
```

4. Run database migrations:
```bash
make migrate
```

## API Documentation

### Endpoints

#### Create Transaction
```http
POST /api/v1/transactions

Request Body:
{
    "description": "Office Supplies",
    "transaction_date": "2024-01-20",
    "amount_usd": 123.45
}

Response (201 Created):
{
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "PENDING",
    "message": "Transaction created successfully."
}
```

#### Get Transaction
```http
GET /api/v1/transactions/{id}

Response (200 OK):
{
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "description": "Office Supplies",
    "transaction_date": "2024-01-20T00:00:00Z",
    "amount_usd": "123.45",
    "created_at": "2024-01-20T15:30:00Z",
    "processed_at": null,
    "status": "PENDING"
}
```

#### Update Transaction Status
```http
PATCH /api/v1/transactions/{id}/status

Request Body:
{
    "status": "COMPLETED"
}

Response (200 OK):
{
    "message": "Transaction status updated successfully"
}
```

#### List Transactions
```http
GET /api/v1/transactions?limit=10&offset=0

Response (200 OK):
[
    {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "description": "Office Supplies",
        "transaction_date": "2024-01-20T00:00:00Z",
        "amount_usd": "123.45",
        "created_at": "2024-01-20T15:30:00Z",
        "processed_at": null,
        "status": "PENDING"
    },
    // ... more transactions
]
```

#### Convert Transaction Currency
```http
GET /api/v1/transactions/{id}/convert?currency={currencyCode}

Response (200 OK):
{
    "transaction_id": "123e4567-e89b-12d3-a456-426614174000",
    "description": "Office Supplies",
    "transaction_date": "2024-01-20",
    "original_amount_usd": "123.45",
    "exchange_rate": "0.85",
    "converted_amount": "104.93",
    "target_currency": "EUR",
    "exchange_date": "2024-01-19"
}
```

### Transaction States

- `PENDING`: Initial state after creation
- `PROCESSING`: Transaction is being processed
- `COMPLETED`: Transaction has been successfully processed
- `FAILED`: Transaction processing failed

### Validation Rules

1. **Description**
   - Required
   - Maximum 50 characters
   - Cannot be empty

2. **Transaction Date**
   - Required
   - Must be a valid date (YYYY-MM-DD)
   - Cannot be in the future

3. **Amount**
   - Required
   - Must be positive
   - Rounded to 2 decimal places

## Development

### Running Locally

1. Start development services:
```bash
make docker-run
```

2. Run with live reload:
```bash
make watch
```

### Testing

Run unit tests:
```bash
make test
```

Run integration tests:
```bash
make itest
```

### Available Make Commands

- `make build`: Build the application
- `make run`: Run the application
- `make docker-run`: Start all services with Docker
- `make docker-down`: Stop all services
- `make migrate`: Run database migrations
- `make test`: Run unit tests
- `make itest`: Run integration tests
- `make watch`: Run with live reload
- `make clean`: Clean build artifacts

## Monitoring

### Health Check

```http
GET /health

Response (200 OK):
{
    "status": "ok"
}
```

The health check endpoint verifies:
- Database connectivity
- Kafka connectivity

## License

This project is licensed under the Unlicense License - see the [LICENSE](LICENSE) file for details.

## Next Steps

### Continuous Integration and Continuous Deployment (CI/CD)

1. **Set up CI/CD with GitHub Actions:**
   - Create a `.github/workflows` directory in your repository.
   - Add a workflow YAML file for CI/CD pipeline configuration.
   - Configure jobs for building, testing, and deploying the application.

2. **Example CI/CD Workflow:**
```yaml
name: CI/CD Pipeline

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23

      - name: Install dependencies
        run: go mod download

      - name: Build
        run: go build -v ./...

      - name: Run tests
        run: make test

  deploy:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - name: Checkout code
        uses: actions/checkout@v2

      - name: Deploy to Production
        run: |
          echo "Deploying to production..."
```

### API Documentation with Swagger

1. **Install Swagger:**
   - Add Swagger dependencies to your project.
   - Use `swaggo/swag` for generating Swagger documentation.

2. **Generate Swagger Docs:**
   - Annotate your Go code with Swagger comments.
   - Run `swag init` to generate the Swagger documentation.

3. **Serve Swagger UI:**

