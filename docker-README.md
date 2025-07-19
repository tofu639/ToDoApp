# Docker Deployment Guide

This guide explains how to run the Todo API Backend using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+

## Quick Start

1. **Clone the repository and navigate to the project directory**

2. **Copy environment configuration**
   ```bash
   cp .env.example .env
   ```
   Update the `.env` file with your configuration values.

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Check service health**
   ```bash
   docker-compose ps
   curl http://localhost:8080/health
   ```

## Environment Files

- `.env.example` - Template with all available configuration options
- `.env.docker` - Docker-specific configuration
- `.env.production` - Production environment template
- `.env` - Your local configuration (not tracked in git)

## Services

### API Service
- **Port**: 8080
- **Health Check**: `GET /health`
- **Environment**: Configurable via environment variables

### PostgreSQL Database
- **Port**: 5432
- **Database**: todoapi
- **User**: user
- **Password**: password (change in production)

## Docker Commands

### Development
```bash
# Start services in development mode
docker-compose up

# Start services in background
docker-compose up -d

# View logs
docker-compose logs -f api
docker-compose logs -f postgres

# Stop services
docker-compose down

# Rebuild and start
docker-compose up --build
```

### Production
```bash
# Use production environment file
docker-compose --env-file .env.production up -d

# Or set environment inline
ENVIRONMENT=production docker-compose up -d
```

### Maintenance
```bash
# Remove containers and volumes (WARNING: This deletes data)
docker-compose down -v

# Remove containers, volumes, and images
docker-compose down -v --rmi all

# View container status
docker-compose ps

# Execute commands in running container
docker-compose exec api sh
docker-compose exec postgres psql -U user -d todoapi
```

## Troubleshooting

### Database Connection Issues
1. Ensure PostgreSQL container is healthy:
   ```bash
   docker-compose ps postgres
   ```

2. Check database logs:
   ```bash
   docker-compose logs postgres
   ```

3. Test database connection:
   ```bash
   docker-compose exec postgres psql -U user -d todoapi -c "SELECT 1;"
   ```

### API Issues
1. Check API logs:
   ```bash
   docker-compose logs api
   ```

2. Verify environment variables:
   ```bash
   docker-compose exec api env | grep -E "(PORT|DATABASE_URL|JWT_SECRET)"
   ```

3. Test API health:
   ```bash
   curl -v http://localhost:8080/health
   ```

### Performance Tuning
- Adjust PostgreSQL memory settings in docker-compose.yml
- Configure connection pooling in the application
- Use Docker resource limits for production deployments

## Security Notes

1. **Change default passwords** in production
2. **Use strong JWT secrets** (minimum 32 characters)
3. **Configure CORS properly** for production
4. **Use SSL/TLS** in production environments
5. **Regularly update base images** for security patches