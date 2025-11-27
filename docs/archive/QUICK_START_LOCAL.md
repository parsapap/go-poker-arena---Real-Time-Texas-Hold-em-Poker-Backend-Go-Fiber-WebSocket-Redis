# 🚀 Quick Start - Local Development

## ✅ Current Status

- ✅ Docker installed and running
- ✅ PostgreSQL running on port 5432
- ✅ Redis running on port 6379
- ✅ Go 1.16+ installed
- ✅ Node.js 20+ installed

## 🎯 Start Everything in 3 Steps

### Step 1: Verify Docker Services (Already Running!)

```bash
sudo docker ps
```

You should see `poker-postgres` and `poker-redis` running.

### Step 2: Start Backend

Open a terminal and run:

```bash
./run-backend.sh
```

Wait for: `Server starting on port 8080`

### Step 3: Start Frontend

Open a **new terminal** and run:

```bash
./run-frontend.sh
```

Wait for: `Ready on http://localhost:3000`

### Step 4: Open Browser

Go to: **http://localhost:3000**

## 🎮 That's It!

You're ready to play poker! 🎰

---

## 📝 Manual Commands (if scripts don't work)

### Backend:
```bash
cd backend
go run cmd/server/main.go
```

### Frontend:
```bash
cd frontend
npm install  # first time only
npm run dev
```

---

## 🛑 Stop Everything

### Stop Backend/Frontend:
Press `Ctrl+C` in each terminal

### Stop Docker:
```bash
sudo docker-compose -f docker-compose.local.yml down
```

### Start Docker Again:
```bash
sudo docker-compose -f docker-compose.local.yml up -d
```
