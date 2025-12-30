# Splitwise Application

A full-stack expense splitting application built with Go (backend) and React (frontend).

## Quick Start

### Using Docker Compose

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database (port 5432)
- Redis cache (port 6379)
- Jaeger tracing (port 16686)
- Backend API (port 8080)
- Frontend (port 80)

### Populating Mock Data

To populate the database with mock data for testing:

```bash
# Install Python dependencies (if not already installed)
pip3 install psycopg2-binary

# Run the population script
python3 populate_mock_data.py
```

Or use the shell script:
```bash
./populate_mock_data.sh
```

The script will create:
- 10 users
- 5 groups with 3-5 members each
- Multiple expenses with automatic equal splits
- User-to-user expenses

## Features

- User authentication (JWT)
- Group management
- Expense splitting (Equal, Percentage, Exact Amount, Shares)
- Balance calculation
- Real-time balance visualization

## Development

See individual README files in `backend/` and `frontend/` directories for development setup.

