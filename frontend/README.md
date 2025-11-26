# Poker Arena - Frontend

Stunning minimal black & white Next.js 14 frontend with poker aesthetic.

## 🎨 Design Features

- **Pure Black & White**: Minimal color palette (#000 background, white text)
- **Glassmorphism**: Frosted glass effects with backdrop blur
- **Glowing Borders**: Animated glow effects on focus
- **Spring Animations**: Smooth Framer Motion transitions
- **IBM Plex Mono**: Monospace font for poker vibe
- **Mobile Responsive**: Works perfectly on all devices

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Open http://localhost:3000
```

## 📁 Structure

```
frontend/
├── app/
│   ├── page.tsx           # Lobby (protected)
│   ├── login/page.tsx     # Login page
│   ├── register/page.tsx  # Register page
│   ├── game/[roomId]/     # Game room
│   ├── layout.tsx         # Root layout
│   └── globals.css        # Global styles
├── components/
│   └── Navbar.tsx         # Navigation bar
├── next.config.js         # API proxy config
└── tailwind.config.ts     # Tailwind config
```

## 🎯 Pages

### Login (`/login`)
- Glassmorphism form
- Glowing input focus
- Spring hover animations
- Password visibility toggle
- Error handling

### Register (`/register`)
- Similar to login
- Email validation
- Password strength
- Animated background particles

### Lobby (`/`)
- Protected route
- Room list
- Create room button
- Player statistics
- Animated chip counter

### Game (`/game/[roomId]`)
- Poker table view
- Community cards
- Player seats
- Action buttons
- Real-time updates (WebSocket)

## 🔧 Configuration

### API Proxy

`next.config.js` proxies requests:
- `/api/*` → `http://localhost:8080/*`
- `/ws` → `http://localhost:8080/ws`

### Environment Variables

Create `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

## 🎨 Styling

### Tailwind Classes

- `glass` - Glassmorphism effect
- `glass-hover` - Hover state
- `glow-border` - Glowing border
- `glow-border-focus` - Focus glow

### Animations

- `animate-glow` - Pulsing glow
- `animate-float` - Floating effect
- Framer Motion spring animations

## 📱 Mobile Responsive

- Breakpoints: `sm`, `md`, `lg`, `xl`
- Touch-friendly buttons
- Responsive grid layouts
- Mobile navigation

## 🚀 Deployment

### Vercel (Recommended)

```bash
vercel deploy
```

### Docker

```bash
docker build -t poker-frontend .
docker run -p 3000:3000 poker-frontend
```

## 🔗 API Integration

### Authentication

```typescript
// Login
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username, password }),
})

// Store token
localStorage.setItem('token', data.token)
```

### Protected Routes

```typescript
useEffect(() => {
  const token = localStorage.getItem('token')
  if (!token) {
    router.push('/login')
  }
}, [])
```

## 🎮 Features

- ✅ JWT Authentication
- ✅ Protected Routes
- ✅ Real-time Updates
- ✅ Responsive Design
- ✅ Glassmorphism UI
- ✅ Animated Transitions
- ✅ Error Handling
- ✅ Loading States

## 🛠️ Tech Stack

- **Next.js 14** - React framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **Framer Motion** - Animations
- **Lucide React** - Icons
- **Radix UI** - Headless components

## 📝 TODO

- [ ] WebSocket integration
- [ ] Game state management
- [ ] Chat system
- [ ] Sound effects
- [ ] Leaderboard page
- [ ] Profile page
- [ ] Settings page

---

**Built with ❤️ using Next.js 14 and Tailwind CSS**
