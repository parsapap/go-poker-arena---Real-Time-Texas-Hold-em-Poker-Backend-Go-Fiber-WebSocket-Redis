# Project Structure

## Clean Project Layout

```
go-poker-arena/
├── README.md                    # Main project documentation
├── PROJECT_STRUCTURE.md         # This file
├── LICENSE                      # MIT License
├── docker-compose.yml           # Docker services configuration
├── docker-compose.local.yml     # Local development override
├── .gitignore                   # Git ignore rules
├── .dockerignore               # Docker ignore rules
├── .env.example                # Environment variables template
│
├── backend/                    # Go backend application
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        # Application entry point
│   ├── internal/              # Internal packages
│   │   ├── auth/             # Authentication & authorization
│   │   ├── poker/            # Poker game logic
│   │   ├── rooms/            # Room management
│   │   ├── websocket/        # WebSocket handling
│   │   ├── database/         # Database connection & migrations
│   │   ├── models/           # Data models
│   │   ├── middleware/       # HTTP middleware
│   │   ├── logger/           # Logging utilities
│   │   ├── metrics/          # Prometheus metrics
│   │   ├── anticheat/        # Anti-cheat validation
│   │   ├── matchmaking/      # Matchmaking queue
│   │   ├── leaderboard/      # Leaderboard management
│   │   └── history/          # Game history
│   ├── test/                 # Test files
│   ├── go.mod                # Go dependencies
│   ├── go.sum                # Go dependency checksums
│   ├── .env                  # Environment variables (gitignored)
│   ├── Dockerfile            # Backend Docker image
│   └── Makefile              # Build commands
│
├── frontend/                  # Next.js frontend application
│   ├── app/                  # Next.js App Router pages
│   │   ├── page.tsx         # Lobby page
│   │   ├── layout.tsx       # Root layout
│   │   ├── globals.css      # Global styles
│   │   ├── login/           # Login page
│   │   ├── register/        # Registration page
│   │   └── game/            # Game pages
│   │       └── [roomId]/    # Dynamic room page
│   ├── components/           # React components
│   │   ├── ui/              # UI components
│   │   ├── Navbar.tsx       # Navigation bar
│   │   ├── RoomTable.tsx    # Room list
│   │   ├── Toast.tsx        # Toast notifications
│   │   ├── ErrorBoundary.tsx # Error handling
│   │   └── ...
│   ├── hooks/                # Custom React hooks
│   │   └── usePokerWebSocket.ts # WebSocket hook
│   ├── lib/                  # Utility functions
│   │   ├── cardUtils.ts     # Card utilities
│   │   ├── sounds.ts        # Sound effects
│   │   ├── logger.ts        # Logging
│   │   └── websocketUtils.ts # WebSocket utilities
│   ├── store/                # State management
│   │   └── gameStore.ts     # Zustand game store
│   ├── types/                # TypeScript type definitions
│   │   ├── toast.ts         # Toast types
│   │   └── websocket.ts     # WebSocket types
│   ├── public/               # Static assets
│   ├── .env.local           # Environment variables (gitignored)
│   ├── .env.local.example   # Environment template
│   ├── package.json         # Node dependencies
│   ├── tsconfig.json        # TypeScript configuration
│   ├── tailwind.config.ts   # Tailwind CSS configuration
│   ├── next.config.js       # Next.js configuration
│   └── Dockerfile           # Frontend Docker image
│
├── docs/                     # Documentation
│   ├── SETUP_GUIDE.md       # Setup instructions
│   ├── TROUBLESHOOTING.md   # Common issues & solutions
│   ├── API.md               # API documentation
│   ├── FEATURES.md          # Feature list
│   ├── CODE_REVIEW.md       # Code analysis
│   ├── CONTRIBUTING.md      # Contribution guidelines
│   ├── SECURITY.md          # Security practices
│   ├── DEPLOYMENT.md        # Deployment guide
│   ├── DOCKER_SETUP.md      # Docker setup
│   ├── CHANGELOG.md         # Version history
│   └── archive/             # Old documentation (for reference)
│
├── test-auth.sh             # Authentication testing script
└── test-websocket.html      # WebSocket testing tool
```

## Key Directories

### Backend (`/backend`)
- **cmd/server/** - Application entry point
- **internal/** - Private application code
  - Well-organized by feature (auth, poker, rooms, etc.)
  - Each package has clear responsibility
  - Models, middleware, and utilities separated

### Frontend (`/frontend`)
- **app/** - Next.js 14 App Router pages
- **components/** - Reusable React components
- **hooks/** - Custom React hooks (WebSocket, etc.)
- **lib/** - Utility functions
- **store/** - Zustand state management
- **types/** - TypeScript type definitions

### Documentation (`/docs`)
- **SETUP_GUIDE.md** - How to set up the project
- **TROUBLESHOOTING.md** - Solutions to common problems
- **API.md** - API endpoint documentation
- **archive/** - Old docs kept for reference

## Important Files

### Root Level
- **README.md** - Main project documentation
- **docker-compose.yml** - Docker services (PostgreSQL, Redis)
- **.env.example** - Environment variables template
- **test-auth.sh** - Quick authentication testing
- **test-websocket.html** - WebSocket connection testing

### Backend
- **cmd/server/main.go** - Application entry point
- **internal/poker/game.go** - Core poker game logic
- **internal/websocket/hub.go** - WebSocket hub
- **go.mod** - Go dependencies

### Frontend
- **app/page.tsx** - Lobby page
- **app/game/[roomId]/page.tsx** - Game room page
- **hooks/usePokerWebSocket.ts** - WebSocket connection
- **store/gameStore.ts** - Game state management
- **package.json** - Node dependencies

## File Count Summary

- **Backend:** ~50 Go files
- **Frontend:** ~30 TypeScript/React files
- **Documentation:** 10 main docs + archive
- **Configuration:** 10+ config files
- **Tests:** Test files in backend/test/

## Clean vs Archive

### Active Files (Root)
- README.md
- docker-compose.yml
- test-auth.sh
- test-websocket.html

### Documentation (docs/)
- 7 main documentation files
- 1 archive folder with old docs

### Removed/Archived
- 47+ redundant documentation files
- 10+ duplicate shell scripts
- Multiple fix/debug documents

All archived files are in `docs/archive/` for reference if needed.

## Development Workflow

1. **Start services:** `docker-compose up -d`
2. **Start backend:** `cd backend && go run cmd/server/main.go`
3. **Start frontend:** `cd frontend && npm run dev`
4. **Test:** Use test-auth.sh and test-websocket.html
5. **Debug:** Check docs/TROUBLESHOOTING.md

## Production Deployment

See `docs/DEPLOYMENT.md` for production deployment instructions.

## Notes

- All sensitive files (.env, .env.local) are gitignored
- Test tools are kept in root for easy access
- Documentation is organized in docs/ folder
- Old/redundant docs archived in docs/archive/
- Clean, maintainable structure
