# RentFlow Quick Start Guide

## 30-Second Setup

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with your settings (passwords, domain, etc.)

# 2. Start all services
docker-compose up -d

# 3. Wait for services to be healthy
docker-compose ps

# 4. Access the application
# Frontend:  http://localhost:3000 (dev) or https://rentflow.example.com (prod)
# Dashboard: http://localhost:8080 (Traefik)
# Metrics:   http://localhost:9090 (Prometheus)
```

## Development Setup (Local Services)

```bash
# Terminal 1: Start infrastructure
docker-compose up postgres redis kurrentdb traefik prometheus

# Terminal 2: Start frontend
cd frontend
npm install
npm run dev
# Runs at http://localhost:3000

# Terminal 3+: Run microservices locally
cd services/auth-service
go run ./cmd/server
```

## Common Commands

### Manage Services
```bash
# View running services
docker-compose ps

# View logs
docker-compose logs -f                    # All services
docker-compose logs -f auth-service       # Specific service

# Restart a service
docker-compose restart auth-service

# Scale a service
docker-compose up -d --scale auth-service=3

# Stop all services
docker-compose down

# Remove all volumes (clean slate)
docker-compose down -v
```

### Build & Deploy
```bash
# Rebuild all services
docker-compose build

# Rebuild specific service
docker-compose build auth-service

# Push to registry (configure REGISTRY env var first)
docker tag rentflow-auth-service:latest $REGISTRY/auth-service:latest
docker push $REGISTRY/auth-service:latest
```

### Database Operations
```bash
# Access PostgreSQL
docker-compose exec postgres psql -U rentflow -d rentflow

# Backup database
docker-compose exec postgres pg_dump -U rentflow rentflow > backup.sql

# Restore database
docker-compose exec -T postgres psql -U rentflow < backup.sql

# View database list
docker-compose exec postgres psql -U rentflow -l
```

### Monitoring
```bash
# View metrics in Prometheus
# http://localhost:9090

# Query specific metric
# Go to http://localhost:9090/graph
# Query: http_requests_total
# Query: http_request_duration_seconds

# View service health
curl http://localhost:8001/health    # Auth service
curl http://localhost:8002/health    # Inventory service
# etc.
```

### Frontend Development
```bash
cd frontend

# Development server with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type checking
npm run type-check

# Linting
npm run lint

# All dependencies installed? Check:
npm list react react-router-dom zustand @tanstack/react-query axios
```

## Environment Variables

### Critical for Production
```bash
# Database security
POSTGRES_USER=rentflow
POSTGRES_PASSWORD=<very-secure-password>

# JWT security
JWT_SECRET=<base64-32+ characters>

# Let's Encrypt (auto HTTPS)
LETSENCRYPT_EMAIL=admin@rentflow.example.com
DOMAIN=rentflow.example.com

# Redis security
REDIS_PASSWORD=<very-secure-password>
```

### Optional but Useful
```bash
# AI/LLM Integration
OLLAMA_ENDPOINT=http://localhost:11434

# Email notifications
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=<app-password>

# Monitoring
PROMETHEUS_RETENTION=30d
LOG_LEVEL=info
```

## Troubleshooting

### Services won't start
```bash
# Check logs
docker-compose logs <service>

# Check dependencies are healthy
docker-compose ps

# Rebuild
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### Port already in use
```bash
# Find what's using the port
lsof -i :8001

# Change the port in docker-compose.yml
# ports:
#   - "8002:8001"  # Use 8002 instead
```

### Can't connect to frontend
```bash
# Frontend logs
docker-compose logs frontend

# Check if Nginx is serving correctly
curl -v http://localhost:3000

# Verify API proxy
curl -v http://localhost:3000/api/v1/auth/me
```

### Database connection errors
```bash
# Test PostgreSQL connection
docker-compose exec postgres psql -U rentflow -d rentflow -c "SELECT 1;"

# Check environment variables in service
docker-compose exec auth-service env | grep DB_

# View PostgreSQL logs
docker-compose logs postgres
```

## Architecture at a Glance

```
Internet
   ↓
Traefik (80/443) — Load balancer & HTTPS
   ↓
   ├─→ React Frontend (SPA)
   │
   ├─→ Auth Service (JWT, user management)
   │   ├─→ PostgreSQL (auth_service db)
   │   ├─→ Redis (sessions, cache)
   │   └─→ KurrentDB (audit events)
   │
   ├─→ Inventory Service (equipment management)
   │   ├─→ PostgreSQL (inventory_service db)
   │   └─→ Redis (cache)
   │
   ├─→ [16 other services]...
   │
   └─→ Prometheus (metrics collection)
```

## File Structure

```
rentflow/
├── docker-compose.yml          ← Start here!
├── .env                        ← Your configuration
│
├── frontend/                   ← React app (npm run dev)
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── services/           ← API calls
│   │   └── stores/             ← State management
│   └── Dockerfile
│
├── services/                   ← 18 microservices
│   ├── auth-service/           ← go run ./cmd/server
│   ├── inventory-service/
│   └── ...
│
└── infra/                      ← Infrastructure config
    ├── traefik/                ← API Gateway
    ├── postgres/               ← Database init
    └── prometheus/             ← Metrics
```

## Common Workflows

### Add a New Feature
1. Create frontend page: `frontend/src/pages/NewFeature.tsx`
2. Add route: Edit `frontend/src/App.tsx`
3. Add navigation: Edit `frontend/src/components/Layout/Sidebar.tsx`
4. Add API endpoint: Edit `frontend/src/services/api.ts`
5. Implement backend: Add to microservice
6. Test: `npm run dev` in frontend, run microservice locally

### Deploy to Production
1. Configure `.env` with production values
2. Set `DOMAIN=rentflow.example.com`
3. Set `LETSENCRYPT_EMAIL=admin@rentflow.example.com`
4. Update DNS to point to server IP
5. Run `docker-compose up -d`
6. Traefik will auto-provision HTTPS certificate

### Run All Services Locally
1. Start infrastructure: `docker-compose up postgres redis kurrentdb`
2. In separate terminals, run each service:
   ```bash
   cd services/auth-service && go run ./cmd/server
   cd services/inventory-service && go run ./cmd/server
   # etc.
   ```
3. Start frontend: `cd frontend && npm run dev`
4. Access at http://localhost:3000

### Debug a Service
```bash
# View service logs
docker-compose logs -f auth-service

# Scale to single instance
docker-compose up -d --scale auth-service=1

# Exec into container
docker-compose exec auth-service sh

# Run with debug logging
docker-compose up auth-service --env LOG_LEVEL=debug
```

## Performance Tips

- Use `.gitkeep` files instead of empty directories (faster git operations)
- Enable gzip in Nginx (already configured)
- Cache static assets for 1 year (already configured)
- Use Redis for session storage (already configured)
- Use TanStack Query caching (frontend already configured)
- Enable HMR in development (Vite, already configured)

## Security Checklist

- [ ] Changed all default passwords in `.env`
- [ ] Generated strong JWT_SECRET (32+ random characters)
- [ ] Set ALLOWED_ORIGINS to your domain only
- [ ] Enabled TLS/HTTPS (automatic with Let's Encrypt)
- [ ] Set up firewall rules (only ports 80, 443 public)
- [ ] Enabled authentication for Traefik dashboard
- [ ] Configured rate limiting (Traefik middleware)
- [ ] Enabled audit logging
- [ ] Set up backups

## Next Steps

1. **Read the full guides**
   - `DOCKER_SETUP.md` - Detailed Docker configuration
   - `frontend/README.md` - Frontend development guide

2. **Customize the setup**
   - Edit `.env` for your environment
   - Modify frontend styling in `frontend/src/styles/variables.scss`
   - Add your business logic to microservices

3. **Deploy**
   - Configure domain/DNS
   - Set production environment variables
   - Run `docker-compose up -d`

4. **Monitor**
   - Check Prometheus at http://localhost:9090
   - View logs: `docker-compose logs -f`
   - Monitor via Traefik dashboard at http://localhost:8080

## Support

- Detailed Docker guide: See `DOCKER_SETUP.md`
- Frontend development: See `frontend/README.md`
- Microservice structure: Check individual service Dockerfiles
- API routing: Check `infra/traefik/dynamic/config.yml`

Happy developing! 🚀
