# Real-Time Payment Processing System

A high-performance payment processing system built with Go, demonstrating advanced concurrency patterns, financial transaction handling, and production-ready architecture.

## Features

- **Real-time Transaction Processing** - Deposit, withdrawal, and transfer operations
- **Concurrency Control** - Goroutines, mutexes, and deadlock prevention
- **Idempotency** - Prevents duplicate transactions
- **Immutable Ledger** - Complete audit trail with append-only design
- **JWT Authentication** - Secure user authentication
- **ACID Transactions** - Database-level transaction guarantees
- **Production-Ready** - Structured logging, error handling, graceful shutdown

## Tech Stack

- **Language**: Go 1.22
- **Web Framework**: Gin
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Message Queue**: RabbitMQ 3
- **Auth**: JWT
- **Logging**: Zerolog

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│   API Gateway   │  (JWT, Rate Limit, Validation)
└──────┬──────────┘
       │
       ▼
┌──────────────────────────────┐
│      Core Services           │
│  ┌────────┐  ┌────────────┐ │
│  │Account │  │Transaction │ │
│  │Service │  │  Service   │ │
│  └────────┘  └────────────┘ │
│  ┌────────┐  ┌────────────┐ │
│  │Ledger  │  │    Auth    │ │
│  │Service │  │  Service   │ │
│  └────────┘  └────────────┘ │
└──────────────┬───────────────┘
               │
       ┌───────┴────────┐
       ▼                ▼
┌─────────────┐  ┌─────────────┐
│  PostgreSQL │  │    Redis    │
└─────────────┘  └─────────────┘
```

## Project Structure

```
real-time-payments/
├── cmd/
│   └── api-server/          # Main application entry point
├── internal/                # Private application code
│   ├── account/            # Account domain
│   ├── transaction/        # Transaction domain with concurrency
│   ├── ledger/             # Immutable ledger
│   ├── auth/               # JWT authentication
│   └── user/               # User management
├── pkg/                    # Public reusable packages
│   ├── database/           # Database connection
│   ├── middleware/         # HTTP middlewares
│   ├── logger/             # Structured logging
│   └── errors/             # Error definitions
├── api/                    # HTTP layer
│   ├── handlers/           # HTTP handlers
│   └── router.go           # Route definitions
├── config/                 # Configuration management
├── scripts/                # Database migrations
└── docker-compose.yml      # Infrastructure setup
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional)

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd Real-Time-Payment-Processing-System
```

### 2. Setup Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Start Infrastructure

```bash
docker-compose up -d
```

This starts:
- PostgreSQL on port 5432
- Redis on port 6379
- RabbitMQ on port 5672 (management UI on 15672)

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run the Application

```bash
go run cmd/api-server/main.go
```

Or using Make:

```bash
make dev
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Authentication

#### Register
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe",
  "cpf": "123.456.789-00",
  "phone": "+5511999999999"
}
```

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe"
  }
}
```

### Accounts

#### Get My Account
```bash
GET /api/v1/accounts/me
Authorization: Bearer <token>
```

#### Get Balance
```bash
GET /api/v1/accounts/{id}/balance
Authorization: Bearer <token>
```

### Transactions

#### Deposit
```bash
POST /api/v1/transactions/deposit
Authorization: Bearer <token>
Content-Type: application/json

{
  "idempotency_key": "unique-key-123",
  "account_id": "uuid",
  "amount": 100.00,
  "description": "Salary deposit"
}
```

#### Withdrawal
```bash
POST /api/v1/transactions/withdrawal
Authorization: Bearer <token>
Content-Type: application/json

{
  "idempotency_key": "unique-key-456",
  "account_id": "uuid",
  "amount": 50.00,
  "description": "ATM withdrawal"
}
```

#### Transfer
```bash
POST /api/v1/transactions/transfer
Authorization: Bearer <token>
Content-Type: application/json

{
  "idempotency_key": "unique-key-789",
  "from_account_id": "uuid",
  "to_account_id": "uuid",
  "amount": 200.00,
  "description": "Payment to friend"
}
```

#### Get Transaction
```bash
GET /api/v1/transactions/{id}
Authorization: Bearer <token>
```

#### List Transactions
```bash
GET /api/v1/transactions?account_id=uuid&limit=50&offset=0
Authorization: Bearer <token>
```

### Ledger

#### Get Ledger History
```bash
GET /api/v1/ledger/{account_id}?limit=50&offset=0
Authorization: Bearer <token>
```

## Key Features Explained

### 1. Concurrency Control

The transaction service uses **mutex locks** and **goroutines** to safely handle concurrent operations:

```go
// Lock accounts in sorted order to prevent deadlock
accountIDs := []string{fromAccountID, toAccountID}
sort.Strings(accountIDs)

for _, id := range accountIDs {
    mu := s.getAccountLock(id)
    mu.Lock()
    defer mu.Unlock()
}
```

### 2. Idempotency

Every transaction requires a unique `idempotency_key`. If the same key is used twice, the system returns the original transaction:

```go
existing, err := s.txRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
if existing != nil {
    return existing // Return cached result
}
```

### 3. Immutable Ledger

The ledger table is **append-only** and uses PostgreSQL rules to prevent updates/deletes:

```sql
CREATE RULE ledger_no_update AS ON UPDATE TO ledger DO INSTEAD NOTHING;
CREATE RULE ledger_no_delete AS ON DELETE TO ledger DO INSTEAD NOTHING;
```

### 4. ACID Transactions

All financial operations use database transactions with `Serializable` isolation:

```go
tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
})
defer tx.Rollback()

// ... perform operations ...

tx.Commit()
```

## Testing

### Run Unit Tests
```bash
make test
```

### Run Tests with Coverage
```bash
make test-coverage
```

### Example Test
```go
func TestTransfer_Success(t *testing.T) {
    // Setup
    service := setupTestService(t)

    // Create accounts with balance
    acc1 := createAccount(t, 1000.00)
    acc2 := createAccount(t, 500.00)

    // Transfer
    req := &TransferRequest{
        IdempotencyKey: "test-key",
        FromAccountID:  acc1.ID,
        ToAccountID:    acc2.ID,
        Amount:         200.00,
    }

    result, err := service.Transfer(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, StatusCompleted, result.Status)
    assert.Equal(t, 800.00, getBalance(acc1.ID))
    assert.Equal(t, 700.00, getBalance(acc2.ID))
}
```

## Database Schema

### Key Tables

- **users** - User authentication and profile
- **accounts** - Account information with balance
- **transactions** - Transaction records with status
- **ledger** - Immutable audit trail (append-only)

See `scripts/init.sql` for complete schema.

## Configuration

Edit `.env` file:

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/payments?sslmode=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=3600

# Logging
LOG_LEVEL=info
```

## Development Commands

```bash
# Start infrastructure
make docker-up

# Stop infrastructure
make docker-down

# Run application
make run

# Build binary
make build

# Run tests
make test

# Clean build artifacts
make clean

# View logs
make docker-logs
```

## Production Considerations

### What's Included
✅ Structured logging with zerolog
✅ Graceful shutdown
✅ Database connection pooling
✅ Error handling and custom errors
✅ JWT authentication
✅ CORS middleware
✅ Request logging
✅ Panic recovery

### What to Add for Production
- [ ] Rate limiting (Redis-based)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Metrics (Prometheus)
- [ ] Load testing (k6)
- [ ] CI/CD pipeline
- [ ] Kubernetes deployment
- [ ] Fraud detection rules
- [ ] Webhook notifications

## Performance

Expected performance on standard hardware:
- **Throughput**: 1000+ req/s
- **Latency (p50)**: < 50ms
- **Latency (p99)**: < 200ms

## Security Features

- Password hashing with bcrypt
- JWT token-based authentication
- SQL injection prevention (parameterized queries)
- Input validation
- CORS protection

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License


## Acknowledgments

Built to demonstrate:
- Advanced Go concurrency patterns
- Financial transaction processing
- Production-ready architecture
- Clean code principles

---

**Author**: Your Name
**Contact**: your.email@example.com
**GitHub**: https://github.com/yourusername

Made with Go
