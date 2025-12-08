# 🃏 Go Poker Arena

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-14-black?style=flat-square&logo=next.js)](https://nextjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.3-blue?style=flat-square&logo=typescript)](https://typescriptlang.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat-square&logo=redis&logoColor=white)](https://redis.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat-square&logo=postgresql&logoColor=white)](https://postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)

**Real-time multiplayer Texas Hold'em Poker**

</div>

---

## 📸 Screenshots

<div align="center">

| Login | GamePlay |
|:---:|:---:|
| ![Login](screenShots/game2.png) | ![Sign Up](screenShots/signup.png) |

| Winner | Lobby |
|:---:|:---:|
| ![Lobby](screenShots/lobby.png) | ![Game](screenShots/game1.png) |

| Gameplay + Chat |
|:---:|
| ![Gameplay](screenShots/login.png) |

</div>

---

## 🎬 Demo

<div align="center">

![Demo](screenShots/demo.gif)

</div>

---

## ⚡ Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/parsapap/go-poker-arena.git
cd go-poker-arena

# Start all services (backend, frontend, PostgreSQL, Redis)
docker-compose up -d

# View logs
docker-compose logs -f
```

**Access:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

**Stop:**
```bash
docker-compose down
```

---

## 🎮 Features

| Feature | Description |
|---------|-------------|
| 🃏 Full Poker Rules | Pre-flop, Flop, Turn, River, Showdown |
| 🎯 All Actions | Fold, Check, Call, Raise, All-in |
| 🏆 Hand Rankings | Royal Flush to High Card |
| 🔄 Real-time | WebSocket with auto-reconnect |
| 👥 Multiplayer | Up to 9 players per table |
| 🔒 Secure | JWT auth, rate limiting, anti-cheat |
| 📱 Responsive | Mobile and desktop support |
| 📊 Stats | Leaderboards, history, win rate |

---

## 🏗️ Tech Stack

```
Frontend          Backend           Database
─────────         ───────           ────────
Next.js 14        Go + Fiber        PostgreSQL
TypeScript        WebSocket         Redis
Tailwind CSS      GORM              
Framer Motion     JWT Auth          
```

---

## 📁 Structure

```
go-poker-arena/
├── backend/
│   ├── cmd/server/      # Entry point
│   └── internal/        # Poker engine, WebSocket, Auth
├── frontend/
│   ├── app/             # Pages (lobby, game, auth)
│   ├── components/      # UI components
│   └── hooks/           # WebSocket hooks
├── docs/                # Documentation
└── docker-compose.yml
```

---

## 🚀 Development (Without Docker)

**Backend:**
```bash
cd backend
cp .env.example .env    # Configure database & Redis
go mod download
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

---

## 🧪 Testing

```bash
# Run all backend tests
cd backend
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/poker/... -v
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

<div align="center">

**Built with Go, Fiber, Next.js & TypeScript**

</div>
