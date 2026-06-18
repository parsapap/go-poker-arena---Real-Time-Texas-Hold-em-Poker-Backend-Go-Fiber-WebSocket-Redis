'use client'

import { useEffect, useState, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useRouter } from 'next/navigation'
import Navbar from '@/components/Navbar'
import CreateRoomModal from '@/components/CreateRoomModal'
import RoomTable from '@/components/RoomTable'
import RoomSkeleton from '@/components/RoomSkeleton'
import { Users, TrendingUp, Zap, Coins } from 'lucide-react'

interface User {
  id: number
  username: string
  chips: number
  wins: number
  losses: number
}

interface UserStats {
  user_id: number
  username: string
  chips: number
  wins: number
  losses: number
  total_games: number
  win_rate: number
}

interface Room {
  id: number
  name: string
  small_blind: number
  big_blind: number
  max_players: number
  status: string
  player_count?: number
}

export default function LobbyPage() {
  const router = useRouter()
  const [user, setUser] = useState<User | null>(null)
  const [userStats, setUserStats] = useState<UserStats | null>(null)
  const [rooms, setRooms] = useState<Room[]>([])
  const [loading, setLoading] = useState(true)
  const [activePlayers, setActivePlayers] = useState(0)
  const [ws, setWs] = useState<WebSocket | null>(null)

  // Animated chip counter
  const [displayChips, setDisplayChips] = useState(0)

  // Fetch user stats from API
  const fetchUserStats = async (userId: number) => {
    try {
      const token = localStorage.getItem('token')
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/api/users/${userId}/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      
      if (response.ok) {
        const stats = await response.json()
        setUserStats(stats)
        // Update user chips from fresh stats
        if (stats.chips !== undefined) {
          setUser(prev => prev ? { ...prev, chips: stats.chips, wins: stats.wins, losses: stats.losses } : null)
          setDisplayChips(stats.chips)
          // Update localStorage with fresh data
          const userData = localStorage.getItem('user')
          if (userData) {
            const parsedUser = JSON.parse(userData)
            parsedUser.chips = stats.chips
            parsedUser.wins = stats.wins
            parsedUser.losses = stats.losses
            localStorage.setItem('user', JSON.stringify(parsedUser))
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch user stats:', error)
    }
  }

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }

    const userData = localStorage.getItem('user')
    if (userData) {
      const parsedUser = JSON.parse(userData)
      setUser(parsedUser)
      setDisplayChips(parsedUser.chips)
      // Fetch fresh stats from API
      fetchUserStats(parsedUser.id)
    }

    fetchRooms()
    setupWebSocket()

    // Refresh rooms and stats every 10 seconds
    const interval = setInterval(() => {
      fetchRooms()
      if (user?.id) fetchUserStats(user.id)
    }, 10000)

    return () => {
      clearInterval(interval)
      if (ws) ws.close()
    }
  }, [router])

  // Animate chip counter
  useEffect(() => {
    if (user && displayChips !== user.chips) {
      const diff = user.chips - displayChips
      const steps = 20
      const increment = diff / steps
      let current = displayChips

      const timer = setInterval(() => {
        current += increment
        if ((increment > 0 && current >= user.chips) || (increment < 0 && current <= user.chips)) {
          setDisplayChips(user.chips)
          clearInterval(timer)
        } else {
          setDisplayChips(Math.round(current))
        }
      }, 30)

      return () => clearInterval(timer)
    }
  }, [user?.chips])

  const setupWebSocket = () => {
    try {
      const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'
      const token = localStorage.getItem('token')
      const userData = localStorage.getItem('user')
      
      if (!userData) return

      const user = JSON.parse(userData)
      if (!token) return

      // The backend authenticates the WebSocket via JWT. Identity comes from
      // the verified token (?token=), not user_id/username query params.
      const socket = new WebSocket(
        `${wsUrl}/ws?token=${encodeURIComponent(token)}&room_id=lobby`
      )

      socket.onopen = () => {
        console.log('WebSocket connected')
      }

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'player_count') {
            setActivePlayers(data.count)
          } else if (data.type === 'room_update') {
            fetchRooms()
          }
        } catch (error) {
          console.error('WebSocket message error:', error)
        }
      }

      socket.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      socket.onclose = () => {
        console.log('WebSocket disconnected')
        // Attempt to reconnect after 5 seconds
        setTimeout(setupWebSocket, 5000)
      }

      setWs(socket)
    } catch (error) {
      console.error('Failed to setup WebSocket:', error)
    }
  }

  const fetchRooms = async () => {
    try {
      const token = localStorage.getItem('token')
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/api/rooms`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      
      if (response.ok) {
        const data = await response.json()
        const roomsData = Array.isArray(data) ? data : []
        setRooms(roomsData)
        // Calculate active players from room player counts
        const totalPlayers = roomsData.reduce((sum: number, room: Room) => sum + (room.player_count || 0), 0)
        setActivePlayers(totalPlayers)
      } else {
        setRooms([])
      }
    } catch (error) {
      console.error('Failed to fetch rooms:', error)
      setRooms([])
    } finally {
      setLoading(false)
    }
  }

  const handleCreateRoom = async (roomData: {
    name: string
    small_blind: number
    big_blind: number
    max_players: number
  }) => {
    try {
      const token = localStorage.getItem('token')
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/api/rooms`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(roomData),
      })

      if (response.ok) {
        await fetchRooms()
      } else {
        throw new Error('Failed to create room')
      }
    } catch (error) {
      console.error('Failed to create room:', error)
      throw error
    }
  }

  const handleJoinRoom = async (roomId: number) => {
    try {
      const token = localStorage.getItem('token')
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      
      // Check if room is full before joining
      const room = rooms.find(r => r.id === roomId)
      if (room && room.player_count && room.player_count >= room.max_players) {
        alert('This room is full. Please choose another room.')
        return
      }
      
      // Attempt to join the room
      const response = await fetch(`${apiUrl}/api/rooms/${roomId}/join`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      })
      
      if (!response.ok) {
        const error = await response.json()
        alert(error.error || 'Failed to join room')
        return
      }
      
      router.push(`/game/${roomId}`)
    } catch (error) {
      console.error('Failed to join room:', error)
      alert('Failed to join room. Please try again.')
    }
  }

  const handleQuickPlay = async () => {
    try {
      const token = localStorage.getItem('token')
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/api/matchmaking/join`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          chips: user?.chips || 1000,
          skill_rank: 1500,
        }),
      })

      if (response.ok) {
        // Show matchmaking status
        alert('Joined matchmaking queue! You will be notified when a match is found.')
      }
    } catch (error) {
      console.error('Failed to join matchmaking:', error)
    }
  }

  const calculateWinRate = () => {
    // Use userStats if available (from API), otherwise fall back to user data
    if (userStats) {
      return Math.round(userStats.win_rate)
    }
    if (!user || (user.wins + user.losses) === 0) return 0
    return Math.round((user.wins / (user.wins + user.losses)) * 100)
  }

  const getWins = () => {
    return userStats?.wins ?? user?.wins ?? 0
  }

  if (!user) return null

  return (
    <div className="min-h-screen">
      <Navbar chips={displayChips} />

      <main className="container mx-auto px-4 pt-24 pb-12">
        {/* Header with Glow Effect */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <motion.h1
            className="text-6xl md:text-7xl font-bold mb-4 relative inline-block"
            animate={{
              textShadow: [
                '0 0 20px rgba(255,255,255,0.5)',
                '0 0 40px rgba(255,255,255,0.3)',
                '0 0 20px rgba(255,255,255,0.5)',
              ],
            }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            POKER ARENA
          </motion.h1>
          <p className="text-white/60 text-lg">Welcome back, {user.username}</p>
        </motion.div>

        {/* Animated Chip Balance */}
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          className="max-w-md mx-auto mb-8 glass glow-border rounded-2xl p-6 text-center"
        >
          <div className="flex items-center justify-center gap-3 mb-2">
            <Coins className="w-6 h-6 text-yellow-400 animate-pulse" />
            <span className="text-white/60 text-sm font-medium">Your Balance</span>
          </div>
          <motion.div
            key={displayChips}
            initial={{ scale: 1.2, color: '#fbbf24' }}
            animate={{ scale: 1, color: '#ffffff' }}
            className="text-5xl font-bold"
          >
            {displayChips.toLocaleString()}
          </motion.div>
          <span className="text-white/40 text-sm">chips</span>
        </motion.div>

        {/* Quick Play Button */}
        <motion.button
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          onClick={handleQuickPlay}
          className="w-full max-w-2xl mx-auto mb-8 bg-gradient-to-r from-green-500 to-emerald-600 text-white rounded-xl p-6 flex items-center justify-center gap-3 hover:from-green-600 hover:to-emerald-700 transition-all shadow-lg shadow-green-500/20 relative overflow-hidden group"
        >
          <motion.div
            animate={{ scale: [1, 1.2, 1] }}
            transition={{ duration: 1.5, repeat: Infinity }}
            className="absolute inset-0 bg-white/10"
          />
          <Zap className="w-6 h-6 relative z-10" />
          <span className="text-xl font-bold relative z-10">Quick Play - Auto Matchmaking</span>
        </motion.button>

        {/* Stats */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8"
        >
          <StatCard
            icon={<Users />}
            label="Active Players"
            value={activePlayers || '...'}
            pulse
          />
          <StatCard icon={<TrendingUp />} label="Your Wins" value={getWins()} />
          <StatCard icon={<TrendingUp />} label="Win Rate" value={`${calculateWinRate()}%`} />
        </motion.div>

        {/* Create Room Modal */}
        <div className="mb-8">
          <CreateRoomModal onCreateRoom={handleCreateRoom} />
        </div>

        {/* Rooms Section */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <h2 className="text-2xl font-bold mb-6 flex items-center gap-2">
            <span>Active Tables</span>
            <span className="text-white/40 text-lg">({rooms.length})</span>
          </h2>

          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <RoomSkeleton count={6} />
            </div>
          ) : (
            <RoomTable rooms={rooms} onJoinRoom={handleJoinRoom} />
          )}
        </motion.div>
      </main>
    </div>
  )
}

function StatCard({
  icon,
  label,
  value,
  pulse = false,
}: {
  icon: React.ReactNode
  label: string
  value: string | number
  pulse?: boolean
}) {
  return (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className="glass glow-border rounded-xl p-6"
    >
      <div className="flex items-center gap-3 mb-2">
        <div className={`text-white/60 ${pulse ? 'animate-pulse' : ''}`}>{icon}</div>
        <span className="text-white/60 text-sm">{label}</span>
      </div>
      <motion.div
        key={value}
        initial={{ scale: 1.1 }}
        animate={{ scale: 1 }}
        className="text-3xl font-bold"
      >
        {value}
      </motion.div>
    </motion.div>
  )
}
