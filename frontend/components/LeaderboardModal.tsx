'use client'

import { motion, AnimatePresence } from 'framer-motion'
import { useEffect, useState } from 'react'

interface LeaderboardEntry {
  rank: number
  username: string
  chips: number
  wins: number
}

interface LeaderboardModalProps {
  isOpen: boolean
  onClose: () => void
}

export default function LeaderboardModal({ isOpen, onClose }: LeaderboardModalProps) {
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isOpen) {
      fetchLeaderboard()
    }
  }, [isOpen])

  const fetchLeaderboard = async () => {
    setLoading(true)
    try {
      const token = localStorage.getItem('token')
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      
      const response = await fetch(`${apiUrl}/api/leaderboard`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        setLeaderboard(data.slice(0, 10))
      }
    } catch (error) {
      console.error('Failed to fetch leaderboard:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50"
          />

          {/* Modal */}
          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 20 }}
            className="fixed inset-0 flex items-center justify-center z-50 p-4"
          >
            <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[80vh] overflow-hidden border border-yellow-500/20">
              {/* Header */}
              <div className="bg-gradient-to-r from-yellow-600 to-yellow-500 px-6 py-4 flex items-center justify-between">
                <h2 className="text-2xl font-bold text-white flex items-center gap-2">
                  🏆 Leaderboard
                </h2>
                <button
                  onClick={onClose}
                  className="text-white hover:text-gray-200 transition-colors"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              {/* Content */}
              <div className="p-6 overflow-y-auto max-h-[calc(80vh-80px)]">
                {loading ? (
                  <div className="flex items-center justify-center py-12">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-yellow-500"></div>
                  </div>
                ) : leaderboard.length === 0 ? (
                  <div className="text-center py-12 text-gray-400">
                    <p className="text-xl">No players yet</p>
                    <p className="text-sm mt-2">Be the first to join!</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {leaderboard.map((entry, index) => (
                      <motion.div
                        key={entry.username}
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.05 }}
                        className={`
                          flex items-center justify-between p-4 rounded-lg
                          ${index === 0 ? 'bg-gradient-to-r from-yellow-600/20 to-yellow-500/20 border border-yellow-500/30' : ''}
                          ${index === 1 ? 'bg-gradient-to-r from-gray-400/20 to-gray-300/20 border border-gray-400/30' : ''}
                          ${index === 2 ? 'bg-gradient-to-r from-orange-600/20 to-orange-500/20 border border-orange-500/30' : ''}
                          ${index > 2 ? 'bg-gray-800/50 border border-gray-700/30' : ''}
                          hover:scale-105 transition-transform
                        `}
                      >
                        <div className="flex items-center gap-4">
                          <div className={`
                            text-2xl font-bold w-10 h-10 flex items-center justify-center rounded-full
                            ${index === 0 ? 'bg-yellow-500 text-white' : ''}
                            ${index === 1 ? 'bg-gray-400 text-white' : ''}
                            ${index === 2 ? 'bg-orange-500 text-white' : ''}
                            ${index > 2 ? 'bg-gray-700 text-gray-300' : ''}
                          `}>
                            {index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : entry.rank}
                          </div>
                          <div>
                            <p className="font-semibold text-white">{entry.username}</p>
                            <p className="text-sm text-gray-400">{entry.wins} wins</p>
                          </div>
                        </div>
                        <div className="text-right">
                          <p className="text-xl font-bold text-yellow-500">
                            💰 {entry.chips.toLocaleString()}
                          </p>
                          <p className="text-xs text-gray-400">chips</p>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}
