# 🐳 Docker Setup for Blockhead Consulting

This directory contains Docker configuration for development, testing, and deployment of the Blockhead Consulting website.

## 📋 Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) (20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (1.29+)

### Development Environment

```bash
# Set up development environment (one-time setup)
make docker-setup

# Start development environment
make docker-dev-bg

# View running services
make docker-status

# View logs
make docker-logs

# Stop services
make docker-stop
```

## 🚀 Available Services

| Service | Port | Description | Web Interface |
|---------|------|-------------|---------------|
| **App** | 8085 | Main application | http://localhost:8085 |
| **MailHog** | 8025 | Email testing | http://localhost:8025 |
| **Gitea** | 3000 | Git server | http://localhost:3000 |
| **Redis** | 6379 | Caching (future use) | - |

## 🧪 Testing

### Run Integration Tests
```bash
# Full integration test suite
make docker-test

# Manual testing
curl http://localhost:8085/health
curl -X POST http://localhost:8085/contact \
  -d "name=Test&email=test@example.com&message=Hello"
```

### Test Email Functionality
1. Submit contact form: http://localhost:8085
2. Check emails in MailHog: http://localhost:8025

## 🏗️ Docker Architecture

### Development Stack (`docker-compose.yml`)
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    App      │    │   MailHog   │    │   Gitea     │
│   :8085     │◄──►│   :8025     │    │   :3000     │
│             │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
       │
       ▼
┌─────────────┐
│   Redis     │
│   :6379     │
└─────────────┘
```

### Test Stack (`docker-compose.test.yml`)
- Isolated test environment
- Automated integration tests
- Coverage reporting

## 📝 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GIT_ENCRYPTION_KEY` | dev_key_... | 32-byte encryption key |
| `SMTP_HOST` | mailhog | SMTP server hostname |
| `SMTP_PORT` | 1025 | SMTP server port |
| `CALENDAR_ENABLED` | false | Enable/disable calendar |
| `BLOG_ENABLED` | true | Enable/disable blog |

### Volume Mounts
- `./data:/app/data` - Persistent data
- `git_storage:/app/data/messages` - Git repository
- `redis_data:/data` - Redis persistence

## 🛠️ Development Workflow

### 1. Initial Setup
```bash
git clone <repository>
cd blockhead-consulting
make docker-setup
```

### 2. Development Cycle
```bash
# Start services
make docker-dev-bg

# Make code changes
# (files are mounted, changes reflect immediately)

# Run tests
make docker-test

# Check logs
make docker-logs

# Stop when done
make docker-stop
```

### 3. Testing New Features
```bash
# Start clean environment
make docker-clean
make docker-setup

# Test specific functionality
curl -X POST http://localhost:8085/contact \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "name=Test&email=test@example.com&message=Testing"

# Check email was sent
open http://localhost:8025
```

## 🚨 Troubleshooting

### Common Issues

**Port Already in Use**
```bash
# Kill processes on ports
sudo lsof -ti :8085 | xargs kill -9
sudo lsof -ti :8025 | xargs kill -9
sudo lsof -ti :3000 | xargs kill -9

# Or use make target
make kill-port
```

**Services Won't Start**
```bash
# Check Docker status
docker-compose ps

# View detailed logs
docker-compose logs app

# Restart specific service
docker-compose restart app
```

**Email Not Sending**
```bash
# Check MailHog logs
docker-compose logs mailhog

# Verify SMTP settings
curl http://localhost:8085/health
```

**Git Storage Issues**
```bash
# Check encryption key
echo $GIT_ENCRYPTION_KEY

# View git storage logs
docker-compose logs app | grep GIT
```

### Reset Everything
```bash
# Nuclear option - clean everything
make docker-clean
docker system prune -a -f
make docker-setup
```

## 📊 Monitoring & Health Checks

### Health Endpoint
```bash
curl http://localhost:8085/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "services": {
    "blog": "ok (3 posts)",
    "contact": "ok",
    "email": "ok",
    "storage": "ok",
    "templates": "ok"
  }
}
```

### Service Status
```bash
# Docker service status
make docker-status

# Application logs
make docker-logs

# MailHog interface
open http://localhost:8025

# Gitea interface
open http://localhost:3000
```

## 🚀 Production Deployment

### Build Production Image
```bash
# Build optimized image
docker build -t blockhead-consulting:latest .

# Run production container
docker run -p 8085:8085 \
  -e ENVIRONMENT=production \
  -e GIT_ENCRYPTION_KEY=your_production_key \
  blockhead-consulting:latest
```

### Environment-Specific Configs
- Development: `docker-compose.yml`
- Testing: `docker-compose.test.yml`  
- Production: Use environment variables or Docker secrets

## 📚 Additional Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [MailHog Documentation](https://github.com/mailhog/MailHog)
- [Gitea Documentation](https://docs.gitea.io/)

## 💡 Pro Tips

1. **Use Background Mode**: `make docker-dev-bg` for development
2. **Watch Logs**: `make docker-logs` in separate terminal
3. **Clean Regularly**: `make docker-clean` to avoid disk usage
4. **Test Often**: `make docker-test` before committing
5. **Health Checks**: Monitor `/health` endpoint for issues