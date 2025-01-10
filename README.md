# Project vr-software-challenge

This project is an HTTP REST application built in Go that stores purchase transactions. It uses PostgreSQL for data storage and Kafka for asynchronous processing.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Docker
- Docker Compose

### Installation

1. Clone the repository:

```bash
git clone https://github.com/Athla/vr-software-challenge.git
cd vr-software-challenge
```

2. Create a `.env` file based on the `.env.example` file and update the environment variables as needed.

3. Build and run the Docker containers:

```bash
make docker-run
```

4. Run database migrations:

```bash
make migrate
```

### Running the Application

To run the application:

```bash
make run
```

### Running Tests

To run the test suite:

```bash
make test
```

To run integration tests:

```bash
make itest
```

### Live Reload

To enable live reload during development:

```bash
make watch
```

### Shutting Down

To shut down the Docker containers:

```bash
make docker-down
```

## API Endpoints

All the endpoints are under the `/api/v1/`:

- `POST /transactions`: Create a new transaction.
- `GET /transactions/:id`: Get a transaction by ID.
- `PATCH /transactions/:id/status`: Update the status of a transaction.
- `GET /transactions`: List transactions.

## License

This project is licensed under the Unlicense License - see the [LICENSE](LICENSE) file for details.
