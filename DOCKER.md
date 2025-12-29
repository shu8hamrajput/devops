# Docker Setup Guide

This guide explains how to run the Splitwise application using Docker Compose.

## Prerequisites

- Docker (version 20.10 or higher)
- Docker Compose (version 2.0 or higher)

## Quick Start

1. **Build and start all services:**
   ```bash
   docker-compose up --build
   ```

2. **Access the application:**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/health

3. **Stop all services:**
   ```bash
   docker-compose down
   ```

## Services

### Backend Service
- **Container**: `splitwise-backend`
- **Port**: 8080
- **Health Check**: `/health` endpoint
- **Technology**: Go 1.21

### Frontend Service
- **Container**: `splitwise-frontend`
- **Port**: 3000 (mapped to container port 80)
- **Technology**: React + Vite, served with Nginx

## Development

### Running in Development Mode

For local development, you can run services individually:

**Backend:**
```bash
cd backend
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

### Rebuilding After Changes

If you make changes to the code:

```bash
# Rebuild and restart all services
docker-compose up --build

# Or rebuild a specific service
docker-compose build backend
docker-compose build frontend
```

### Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
```

## Environment Variables

### Backend
- `PORT`: Server port (default: 8080)

### Frontend
- `VITE_API_URL`: Backend API URL (default: http://localhost:8080)

The frontend API URL is set at build time. To change it, modify the `docker-compose.yml` file:

```yaml
frontend:
  build:
    args:
      - VITE_API_URL=http://your-backend-url:8080
```

## Troubleshooting

### CORS Issues
The backend includes CORS middleware that allows all origins. If you need to restrict it, modify `backend/main.go`.

### Port Conflicts
If ports 3000 or 8080 are already in use, modify the port mappings in `docker-compose.yml`:

```yaml
services:
  backend:
    ports:
      - "8081:8080"  # Change host port
  frontend:
    ports:
      - "3001:80"    # Change host port
```

### Container Not Starting
1. Check logs: `docker-compose logs`
2. Verify Docker is running: `docker ps`
3. Check if ports are available: `lsof -i :8080` or `lsof -i :3000`

### Rebuilding from Scratch
```bash
# Remove all containers, networks, and volumes
docker-compose down -v

# Remove images
docker-compose rm -f

# Rebuild everything
docker-compose up --build
```

## Production Considerations

For production deployment, consider:

1. **Environment Variables**: Use `.env` files or secrets management
2. **HTTPS**: Configure reverse proxy (nginx/traefik) with SSL certificates
3. **Database**: Replace in-memory repositories with persistent storage
4. **Security**: 
   - Restrict CORS origins
   - Add authentication/authorization
   - Use environment-specific configurations
5. **Monitoring**: Add health checks and logging
6. **Scaling**: Configure multiple backend instances behind a load balancer

## Docker Commands Reference

```bash
# Start services in background
docker-compose up -d

# Stop services
docker-compose stop

# Stop and remove containers
docker-compose down

# View running containers
docker-compose ps

# Execute command in container
docker-compose exec backend sh
docker-compose exec frontend sh

# View resource usage
docker stats
```

