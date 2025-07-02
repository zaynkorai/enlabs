# Enlabs
### Balance Processing API

#### Project Structure
The project adheres to a clean architecture pattern within the standard Go project layout, promoting modularity, testability, and maintainability.
    
    ├── cmd/                 # Application entry points
    │   └── server/
    │       └── main.go
    ├── internal/            # Private application code (not importable by other modules)
    │   ├── app/             # Application-specific logic (server setup, services)
    │   │   ├── server/
    │   │   └── services/    # Core business logic / use cases
    │   ├── domain/          # Business entities, interfaces (contracts for services/repositories)
    │   │   ├── <entity1>/
    │   │   └── <entity2>/
    │   ├── platform/        # Implementations of domain interfaces (persistence, external clients)
    │   │   ├── persistence/
    │   └── transport/       # For external communication (HTTP handlers, CLI)
    │       ├── http/        # HTTP handlers/controllers
    │           ├── <entity1>/
    │           └── <entity2>/
    ├── pkg/                 # Reusable, generic packages (errors, logging, utils)
    ├── go.mod
    ├── go.sum
    ├── Dockerfile
    ├── Makefile
    └── README.md

## Prerequisites

Before running the application, ensure you have Docker Desktop installed:

  * **[Docker Desktop](https://www.docker.com/products/docker-desktop/)** (or Docker Engine and Docker Compose standalone)

## Getting Started

Follow these steps to get the Balance Processing API up and running quickly using Docker Compose.

### 1\. Clone the Repository

First, clone the project repository to your local machine:

```bash
git clone https://github.com/zaynkorai/enlabs.git
cd enlabs
```

### 2\. Configure Environment Variables

The application relies on environment variables for configuration, especially for database connection details.

Create a `.env` file in the root directory of the project by copying the example file:

```bash
cp .env.example .env
```

The `.env` file should contain:

```ini
# .env
DB_HOST=db
DB_PORT=5432
DB_USER=user
DB_PASSWORD=St62ew&uyh
DB_NAME=enlabs_db
APP_PORT=8089
```

### 3\. Run with Docker Compose

This command will:

  * Build the Docker image for the Go application.
  * Pull the PostgreSQL Docker image.
  * Create and start both the `db` (PostgreSQL) and `app` (Go application) containers.
  * Wait for the PostgreSQL database to become healthy.
  * Run GORM's auto-migrations to set up the database schema and insert predefined users (ID 1, 2, 3).


```bash
docker compose up --build -d
```

  * `--build`: Ensures the Docker image for the Go application is (re)built.

### 4\. Verify Running Services

You can check the status of your Docker containers to ensure everything is running as expected:

```bash
docker compose ps
```

You should see output similar to this, indicating both services are `running` and `healthy`:

```
NAME                COMMAND                  SERVICE             STATUS              PORTS
enlabs-app-1   "/bin/sh -c './balan…"   app                 running             0.0.0.0:8089->8089/tcp
enlabs-db-1    "docker-entrypoint.s…"   db                  running (healthy)   0.0.0.0:5432->5432/tcp
```

The API should now be accessible at `http://localhost:8089`.

Upon the first successful startup using `docker compose up`, the database will be initialized, and the users with id(1, 2 and 3) automatically created with an initial balance of `0.00`:

## How to Test

You can use `curl` from your terminal, Postman, or any other HTTP client to interact with the API.

Assuming the application is running on `http://localhost:8089`.

### Example Tests with `curl`

**1. Get initial balance for user 1:**

```bash
curl -v http://localhost:8089/user/1/balance
```

*Expected Output (initial balance):*

```json
{
  "userId": 1,
  "balance": "0.00"
}
```

**2. Deposit (win) 10.50 for user 1:**

```bash
curl -v -X POST \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{"state": "win", "amount": "10.50", "transactionId": "txn-user1-win-1"}' \
  http://localhost:8089/user/1/transaction
```

*Expected Output (success):* `HTTP/1.1 200 OK` (with an empty JSON response `{}`)

**3. Get updated balance for user 1:**

```bash
curl -v http://localhost:8089/user/1/balance
```

*Expected Output:*

```json
{
  "userId": 1,
  "balance": "10.50"
}
```

**4. Withdraw (lose) 2.25 for user 1:**

```bash
curl -v -X POST \
  -H "Source-Type: payment" \
  -H "Content-Type: application/json" \
  -d '{"state": "lose", "amount": "2.25", "transactionId": "txn-user1-lose-1"}' \
  http://localhost:8089/user/1/transaction
```

*Expected Output (success):* `HTTP/1.1 200 OK`


## Shutting Down

To stop and remove the running Docker containers and Docker volumes:

```bash
docker compose down -v
```

-----