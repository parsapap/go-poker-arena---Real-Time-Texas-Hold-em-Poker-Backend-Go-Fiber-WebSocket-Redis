'use client'

import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { useRouter } from 'next/navigation'
import Navbar from '@/components/Navbar'
import { Plus, Users, TrendingUp } from 'lucide-react'

export default function LobbyPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [rooms, setRooms] = useState([])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }

    const userData = localStorage.getItem('user')
    if (userData) {
      setUser(JSON.parse(userData))
    }

    // Fetch rooms
    fetchRooms()
  }, [router])

  const fetchRooms = async () => {
    try {
      const token = localStorage.getItem('token')
      const response = await fetch('/api/rooms', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })
      const data = await response.json()
      setRooms(data)
    } catch (error) {
      console.error('Failed to fetch rooms:', error)
    }
  }

  if (!user) return null

  return (
    <div className="min-h-screen">
      <Navbar chips={user?.chips || 1000} />

      <main className="container mx-auto px-4 pt-24 pb-12">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <h1 className="text-5xl font-bold mb-4">Welcome, {user?.username}</h1>
          <p className="text-white/60 text-lg">Choose a table or create your own</p>
        </motion.div>

        {/* Stats */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-12"
        >
          <StatCard icon={<Users />} label="Active Players" value="1,234" />
          <StatCard icon={<TrendingUp />} label="Your Wins" value={user?.wins || 0} />
          <StatCard icon={<TrendingUp />} label="Win Rate" value="65%" />
        </motion.div>

        {/* Create Room Button */}
        <motion.button
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          className="w-full glass glow-border rounded-xl p-6 mb-8 flex items-center justify-center gap-3 hover:bg-white/10 transition-all"
        >
          <Plus className="w-6 h-6" />
          <span className="text-xl font-bold">Create New Room</span>
        </motion.button>

        {/* Rooms List */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {rooms.length === 0 ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="col-span-full text-center py-12 text-white/40"
            >
              No active rooms. Create one to start playing!
            </motion.div>
          ) : (
            rooms.map((room: any, index) => (
              <RoomCard key={room.id} room={room} index={index} />
            ))
          )}
        </div>
      </main>
    </div>
  )
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: string | number }) {
  return (
    <div className="glass glow-border rounded-xl p-6">
      <div className="flex items-center gap-3 mb-2">
        <div className="text-white/60">{icon}</div>
        <span className="text-white/60 text-sm">{label}</span>
      </div>
      <div className="text-3xl font-bold">{value}</div>
    </div>
  )
}

function RoomCard({ room, index }: { room: any; index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05 }}
      whileHover={{ scale: 1.02 }}
      className="glass glow-border rounded-xl p-6 cursor-pointer hover:bg-white/10 transition-all"
    >
      <h3 className="text-xl font-bold mb-2">{room.name}</h3>
      <div className="space-y-2 text-sm text-white/60">
        <div className="flex justify-between">
          <span>Blinds:</span>
          <span className="text-white">{room.small_blind}/{room.big_blind}</span>
        </div>
        <div className="flex justify-between">
          <span>Players:</span>
          <span className="text-white">0/{room.max_players}</span>
        </div>
        <div className="flex justify-between">
          <span>Status:</span>
          <span className="text-green-400">{room.status}</span>
        </div>
      </div>
      <motion.button
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
        className="w-full mt-4 bg-white text-black py-2 rounded-lg font-bold hover:bg-white/90 transition-all"
      >
        Join Table
      </motion.button>
    </motion.div>
  )
}
