# Splitwise Microservice - Ports and Adapters Pattern

This is a Go microservice implementing the **Ports and Adapters** (Hexagonal Architecture) pattern for a Splitwise-like expense sharing application.

## Architecture Overview

The project follows the Ports and Adapters pattern, which separates the business logic from external dependencies:

```
┌─────────────────────────────────────────────────────────┐
│                    Inbound Adapters                      │
│              (HTTP Handlers - Driving Side)              │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                  Inbound Ports (Interfaces)              │
│              (UserService, ExpenseService, etc.)         │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              Application Services (Use Cases)            │
│         (Implements inbound ports, uses outbound)        │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                 Outbound Ports (Interfaces)              │
│        (UserRepository, ExpenseRepository, etc.)         │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                  Outbound Adapters                       │
│         (Repositories - Driven Side)                     │
└─────────────────────────────────────────────────────────┘
```

## Project Structure

```
splitwise/
├── backend/                   # Go backend microservice
│   ├── domain/                # Core business logic
│   │   ├── entity/            # Domain entities
│   │   │   ├── user.go
│   │   │   ├── expense.go
│   │   │   └── group.go
│   │   └── port/              # Ports (interfaces)
│   │       ├── inbound/       # Driving ports (what the app needs)
│   │       │   ├── user_service.go
│   │       │   ├── expense_service.go
│   │       │   └── group_service.go
│   │       └── outbound/      # Driven ports (what the app provides)
│   │           ├── user_repository.go
│   │           ├── expense_repository.go
│   │           └── group_repository.go
│   ├── application/           # Application layer
│   │   └── service/           # Use case implementations
│   │       ├── user_service.go
│   │       ├── expense_service.go
│   │       └── group_service.go
│   ├── adapter/               # Adapters (implementations)
│   │   ├── inbound/           # Driving adapters
│   │   │   └── http/          # HTTP handlers
│   │   │       ├── user_handler.go
│   │   │       ├── expense_handler.go
│   │   │       └── group_handler.go
│   │   └── outbound/          # Driven adapters
│   │       └── repository/    # Repository implementations
│   │           ├── memory_user_repository.go
│   │           ├── memory_expense_repository.go
│   │           └── memory_group_repository.go
│   ├── main.go                # Dependency injection & wiring
│   ├── Dockerfile             # Backend Docker image
│   ├── go.mod
│   └── go.sum
├── frontend/                  # React frontend application
│   ├── src/                   # Source code
│   ├── Dockerfile             # Frontend Docker image
│   └── package.json
├── docker-compose.yml         # Docker Compose configuration
└── README.md
```

## Key Concepts

### Ports (Interfaces)
- **Inbound Ports**: Define what the application can do (e.g., `UserService`, `ExpenseService`)
- **Outbound Ports**: Define what the application needs from external systems (e.g., `UserRepository`, `ExpenseRepository`)

### Adapters (Implementations)
- **Inbound Adapters**: Implement how external systems interact with the application (e.g., HTTP handlers, gRPC handlers, CLI)
- **Outbound Adapters**: Implement how the application interacts with external systems (e.g., database repositories, external API clients)

### Benefits
1. **Testability**: Easy to mock dependencies
2. **Flexibility**: Swap implementations without changing business logic
3. **Independence**: Business logic doesn't depend on frameworks or databases
4. **Maintainability**: Clear separation of concerns

## Running the Service

### Option 1: Using Docker Compose (Recommended)

Run both backend and frontend together:

```bash
docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Jaeger UI: http://localhost:16686

See [DOCKER.md](./DOCKER.md) for detailed Docker setup instructions.
See [backend/TRACING.md](./backend/TRACING.md) for Jaeger tracing documentation.

### Option 2: Local Development

**Backend:**
```bash
cd backend

# Install dependencies
go mod download

# Run the service
go run main.go
```

The service will start on `http://localhost:8080`

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:3000`

## API Endpoints

### Users
- `POST /users` - Create a new user
- `GET /users` - Get all users
- `GET /users/{id}` - Get user by ID

### Expenses
- `POST /expenses` - Create a new expense
- `GET /expenses/{id}` - Get expense by ID
- `GET /groups/{group_id}/expenses` - Get all expenses for a group

### Groups
- `POST /groups` - Create a new group
- `GET /groups/{id}` - Get group by ID
- `POST /groups/{id}/users` - Add user to group

### Health
- `GET /health` - Health check endpoint

## Example Usage

### Create a User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'
```

### Create a Group
```bash
curl -X POST http://localhost:8080/groups \
  -H "Content-Type: application/json" \
  -d '{"name": "Trip to Paris", "user_ids": ["<user-id-1>", "<user-id-2>"]}'
```

### Create an Expense
```bash
curl -X POST http://localhost:8080/expenses \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Dinner",
    "amount": 100.50,
    "paid_by": "<user-id>",
    "group_id": "<group-id>"
  }'
```

## Extending the Architecture

To add a new database adapter, simply:
1. Create a new repository implementation in `backend/adapter/outbound/repository/`
2. Implement the outbound port interface
3. Update `backend/main.go` to use the new repository

The business logic remains unchanged!

