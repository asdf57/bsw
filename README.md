# BSW
BSW is a small shared-expense tracker.

It lets you:
- create users
- record payments (in a variety of currencies)
- track net debts between users
- use a simple web UI, REST API, and Discord bot

## What’s In The Repo
- `cmd/api`: Go API server
- `frontend`: lightweight PHP UI
- `cmd/bot`: Discord bot
- `internal`: app logic, models, handlers, debt logic, currency logic
- `docs`: generated Swagger docs

## Stack
- Go
- Gin
- GORM
- Postgres
- PHP frontend
- DiscordGo
- Docker Compose

## How to run it
The simplest path:

```bash
make build
```

That:
- generates Swagger docs
- builds containers
- starts Postgres, the API, the frontend, and the Discord bot

Useful URLs:
- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Frontend UI: `http://localhost:8081`

Useful commands:
```bash
make connect-db # debug the DB
make clean      # clean up the compose cluster
```

## How To Use It
Typical flow:
1. Create a few users.
2. Add a payment with a payer, amount, description, currency, and debtors.
3. View current payments.
4. View the current debt ledger.

The app stores debts as net balances between user pairs, so the debt list stays simplified instead of tracking both directions separately.

## Main Features

### Users
- create users
- list users
- delete users when they are no longer referenced by active data

### Payments
- create payments
- list all payments
- fetch a single payment
- delete payments

When a payment is created, the app also updates the debt ledger.

### Debts
- list all current debts
- list grouped debt views by user

### Exchange Rates
- fetch exchange rates
- cache them in the database
- reuse cached daily rates for later payments

### Admin
- database health check
- create backups
- download backups
- inspect cached exchange rates

## API Routes
Main routes:
- `GET /api/v1/user`
- `POST /api/v1/user`
- `DELETE /api/v1/user/:id`
- `GET /api/v1/payment/:id`
- `GET /api/v1/payment/all`
- `POST /api/v1/payment`
- `DELETE /api/v1/payment/:id`
- `GET /api/v1/debts`
- `GET /api/v1/debts/debts/users`
- `GET /api/v1/health`

Admin routes:
- `POST /api/v1/admin/backup`
- `GET /api/v1/admin/backup/:filename`
- `GET /api/v1/admin/exchange-rate`
- `GET /api/v1/admin/exchange-rates`

For request and response details, please view the Swagger docs.

## Discord Bot
The bot talks to the same API and supports slash commands for common actions such as:

- creating payments
- listing payments
- listing debts
- deleting payments
- adding users

The bot expects:
- `DISCORD_BOT_TOKEN`
- `API_URL`

## Notes
- `make build` starts the Discord bot too (for now)
- Swagger docs are pieces of "living code" that will change as the API evolves
