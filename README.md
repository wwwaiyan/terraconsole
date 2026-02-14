# TerraConsole 🏗️

A self-hosted Terraform Cloud UI for managing infrastructure as code. Built with Go (backend), React + TypeScript (frontend), PostgreSQL, and Docker.

## Features

- **Authentication** — Login, Signup, TOTP MFA (Google Authenticator / Authy)
- **Organizations** — Team management with role-based access (Owner, Admin, Member, Viewer)
- **Projects** — Group workspaces by project
- **Workspaces** — Full workspace management with variables, state, and runs
- **Terraform Operations** — Plan, Apply, Destroy, Refresh with approval workflows
- **Multiple Terraform Versions** — Each workspace can use a different Terraform version
- **Variables & Secrets** — Terraform and environment variables with AES-256-GCM encryption
- **State Management** — Terraform HTTP backend for remote state, version history, outputs
- **Docker Ready** — Full Docker Compose setup for one-command deployment

## Quick Start

### Prerequisites

- Docker & Docker Compose
- (Optional) Go 1.22+ and Node.js 20+ for local development

### Run with Docker

```bash
# Clone the repository
git clone https://github.com/your-org/terraconsole.git
cd terraconsole

# Start all services
docker compose up -d

# Access the UI
open http://localhost
```

### Environment Variables

Create a `.env` file in the project root or set these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://terraconsole:...@db:5432/terraconsole?sslmode=disable` | PostgreSQL connection string |
| `REDIS_URL` | `redis://redis:6379` | Redis connection string |
| `JWT_SECRET` | `change-me-in-production-...` | JWT signing secret (change in production!) |
| `ENCRYPTION_KEY` | `0123456789abcdef...` | AES-256 hex key for encrypting secrets |
| `TERRAFORM_DIR` | `/opt/terraform` | Directory for Terraform binaries |
| `WORKING_DIR` | `/opt/terraconsole/workspaces` | Working directory for workspace files |
| `ALLOWED_ORIGINS` | `http://localhost,http://localhost:3000` | CORS allowed origins |

### Local Development

```bash
# Backend
cd api
go mod tidy
go run cmd/server/main.go

# Frontend (in another terminal)
cd frontend
npm install
npm run dev
```

## Architecture

```
terraconsole/
├── api/                        # Go backend
│   ├── cmd/server/main.go      # Entry point
│   └── internal/
│       ├── config/             # Configuration loader
│       ├── database/           # GORM PostgreSQL connection
│       ├── handlers/           # HTTP route handlers
│       ├── middleware/         # JWT auth middleware
│       ├── models/            # GORM data models
│       └── services/          # Encryption service
├── frontend/                   # React + TypeScript + Vite
│   └── src/
│       ├── api/               # API client
│       ├── components/        # Layout components
│       ├── context/           # Auth context provider
│       ├── pages/             # 9 page components
│       └── types/             # TypeScript interfaces
├── nginx/                      # Reverse proxy config
├── docker-compose.yml          # Full stack orchestration
├── Dockerfile.api              # Go backend container
└── Dockerfile.frontend         # React frontend container
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/signup` | Create account |
| POST | `/api/auth/login` | Login (with optional TOTP) |
| GET | `/api/auth/me` | Get current user |
| POST | `/api/auth/mfa/setup` | Generate MFA QR code |
| POST | `/api/auth/mfa/verify` | Verify & enable MFA |
| GET | `/api/organizations` | List organizations |
| POST | `/api/organizations` | Create organization |
| GET | `/api/organizations/{id}/projects` | List projects |
| GET | `/api/projects/{id}/workspaces` | List workspaces |
| GET | `/api/workspaces/{id}` | Get workspace details |
| POST | `/api/workspaces/{id}/runs` | Create a run |
| GET | `/api/runs/{id}` | Get run details |
| POST | `/api/runs/{id}/approve` | Approve a planned run |
| GET | `/api/workspaces/{id}/variables` | List variables |
| GET | `/api/workspaces/{id}/state` | Get current state |
| GET | `/api/terraform/versions` | List available TF versions |
| POST | `/api/terraform/versions/{v}/install` | Install a TF version |

## Terraform Remote State

Configure your Terraform backend to use TerraConsole:

```hcl
terraform {
  backend "http" {
    address        = "http://localhost/api/state/WORKSPACE_ID"
    lock_address   = "http://localhost/api/state/WORKSPACE_ID/lock"
    unlock_address = "http://localhost/api/state/WORKSPACE_ID/unlock"
  }
}
```

## Tech Stack

- **Backend**: Go 1.22, Chi router, GORM, JWT, TOTP, AES-256-GCM
- **Frontend**: React 18, TypeScript, Vite, React Router
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Proxy**: Nginx
- **Container**: Docker + Docker Compose

## License

MIT
