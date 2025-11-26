'use client'

import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { useParams, useRouter } from 'next/navigation'
import Navbar from '@/components/Navbar'
import { ArrowLeft } from 'lucide-react'

export default function GamePage() {
  const params = useParams()
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [gameState, setGameState] = useState<any>(null)

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

    // TODO: Connect to WebSocket
    // const ws = new WebSocket(`ws://localhost:8080/ws?user_id=${user.id}&username=${user.username}&room_id=${params.roomId}`)
  }, [params.roomId, router])

  if (!user) return null

  return (
    <div className="min-h-screen">
      <Navbar chips={user?.chips || 1000} />

      <main className="container mx-auto px-4 pt-24 pb-12">
        <motion.button
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          whileHover={{ scale: 1.05 }}
          onClick={() => router.push('/')}
          className="glass px-4 py-2 rounded-lg flex items-center gap-2 mb-8"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Lobby
        </motion.button>

        {/* Poker Table */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="glass glow-border rounded-3xl p-8 aspect-video max-w-6xl mx-auto relative overflow-hidden"
        >
          {/* Table felt */}
          <div className="absolute inset-8 rounded-3xl border-4 border-white/20 bg-gradient-to-br from-green-900/20 to-green-950/20" />

          {/* Center pot */}
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 z-10">
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              className="glass px-6 py-3 rounded-lg"
            >
              <div className="text-sm text-white/60 mb-1">Pot</div>
              <div className="text-2xl font-bold">$0</div>
            </motion.div>
          </div>

          {/* Community cards */}
          <div className="absolute top-1/3 left-1/2 -translate-x-1/2 flex gap-2 z-10">
            {[...Array(5)].map((_, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.1 }}
                className="w-16 h-24 bg-white/10 rounded-lg border border-white/20"
              />
            ))}
          </div>

          {/* Player seats */}
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="text-white/40 text-center">
              <div className="text-6xl mb-4">🃏</div>
              <div className="text-xl">Waiting for players...</div>
            </div>
          </div>
        </motion.div>

        {/* Action buttons */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="flex gap-4 justify-center mt-8"
        >
          <ActionButton label="Fold" />
          <ActionButton label="Check" />
          <ActionButton label="Call" />
          <ActionButton label="Raise" primary />
        </motion.div>
      </main>
    </div>
  )
}

function ActionButton({ label, primary }: { label: string; primary?: boolean }) {
  return (
    <motion.button
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      className={`px-8 py-3 rounded-lg font-bold transition-all ${
        primary
          ? 'bg-white text-black hover:bg-white/90'
          : 'glass glow-border hover:bg-white/10'
      }`}
    >
      {label}
    </motion.button>
  )
}
