# Go Poker Arena - Frontend

Next.js frontend for the Go Poker Arena real-time poker game.

## Tech Stack

- **Next.js 14** - React framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **Socket.IO** - WebSocket client
- **Zustand** - State management
- **Axios** - HTTP client

## Quick Start

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Environment Variables

Create a `.env.local` file:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## Project Structure

```
frontend/
├── app/
│   ├── layout.tsx       # Root layout
│   ├── page.tsx         # Home page
│   ├── play/            # Game page
│   ├── rooms/           # Rooms list
│   └── globals.css      # Global styles
├── components/
│   ├── PokerTable.tsx   # Poker table component
│   ├── Card.tsx         # Playing card
│   └── PlayerSeat.tsx   # Player seat
├── lib/
│   ├── api.ts           # API client
│   ├── websocket.ts     # WebSocket client
│   └── store.ts         # Zustand store
└── public/              # Static assets
```

## Features

- ✅ Real-time poker gameplay
- ✅ Responsive design
- ✅ WebSocket connection
- ✅ Authentication
- ✅ Leaderboards
- ✅ Game history
- ✅ Player statistics

## Development

```bash
# Type checking
npm run type-check

# Linting
npm run lint

# Format code
npm run format
```

## Deployment

Deploy to Vercel:

```bash
vercel deploy
```

Or build Docker image:

```bash
docker build -t poker-frontend .
docker run -p 3000:3000 poker-frontend
```
