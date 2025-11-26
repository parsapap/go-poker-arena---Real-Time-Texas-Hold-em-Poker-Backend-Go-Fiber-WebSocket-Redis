'use client'

import { motion } from 'framer-motion'
import { Users, Play } from 'lucide-react'

interface Room {
  id: number
  name: string
  small_blind: number
  big_blind: number
  max_players: number
  status: string
  player_count?: number
}

interface RoomTableProps {
  rooms: Room[]
  onJoinRoom: (roomId: number) => void
}

export default function RoomTable({ rooms, onJoinRoom }: RoomTableProps) {
  if (rooms.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="col-span-full text-center py-12 glass glow-border rounded-xl"
      >
        <div className="text-6xl mb-4">🃏</div>
        <h3 className="text-xl font-bold mb-2">No Active Rooms</h3>
        <p className="text-white/60">Create a room to start playing!</p>
      </motion.div>
    )
  }

  return (
    <>
      {/* Desktop Table View */}
      <div className="hidden md:block glass glow-border rounded-xl overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-white/10">
              <th className="text-left p-4 font-bold text-white/80">Room Name</th>
              <th className="text-left p-4 font-bold text-white/80">Blinds</th>
              <th className="text-left p-4 font-bold text-white/80">Players</th>
              <th className="text-left p-4 font-bold text-white/80">Status</th>
              <th className="text-right p-4 font-bold text-white/80">Action</th>
            </tr>
          </thead>
          <tbody>
            {rooms.map((room, index) => (
              <motion.tr
                key={room.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.05 }}
                whileHover={{ backgroundColor: 'rgba(255, 255, 255, 0.05)' }}
                className="border-b border-white/5 last:border-0 cursor-pointer"
                onClick={() => onJoinRoom(room.id)}
              >
                <td className="p-4 font-bold">{room.name}</td>
                <td className="p-4 text-white/60">
                  {room.small_blind}/{room.big_blind}
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-2">
                    <Users className="w-4 h-4 text-white/60" />
                    <span className="text-white/60">
                      {room.player_count || 0}/{room.max_players}
                    </span>
                  </div>
                </td>
                <td className="p-4">
                  <span
                    className={`px-3 py-1 rounded-full text-xs font-bold ${
                      room.status === 'waiting'
                        ? 'bg-green-500/20 text-green-400'
                        : 'bg-yellow-500/20 text-yellow-400'
                    }`}
                  >
                    {room.status === 'waiting' ? 'Waiting' : 'Playing'}
                  </span>
                </td>
                <td className="p-4 text-right">
                  <motion.button
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    onClick={(e) => {
                      e.stopPropagation()
                      onJoinRoom(room.id)
                    }}
                    className="bg-white text-black px-4 py-2 rounded-lg font-bold hover:bg-white/90 transition-all inline-flex items-center gap-2"
                  >
                    <Play className="w-4 h-4" />
                    Join
                  </motion.button>
                </td>
              </motion.tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Mobile Card View */}
      <div className="md:hidden space-y-4">
        {rooms.map((room, index) => (
          <motion.div
            key={room.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
            whileHover={{ scale: 1.02 }}
            className="glass glow-border rounded-xl p-6"
          >
            <div className="flex items-start justify-between mb-4">
              <h3 className="text-xl font-bold">{room.name}</h3>
              <span
                className={`px-3 py-1 rounded-full text-xs font-bold ${
                  room.status === 'waiting'
                    ? 'bg-green-500/20 text-green-400'
                    : 'bg-yellow-500/20 text-yellow-400'
                }`}
              >
                {room.status === 'waiting' ? 'Waiting' : 'Playing'}
              </span>
            </div>
            <div className="space-y-2 text-sm text-white/60 mb-4">
              <div className="flex justify-between">
                <span>Blinds:</span>
                <span className="text-white font-bold">
                  {room.small_blind}/{room.big_blind}
                </span>
              </div>
              <div className="flex justify-between">
                <span>Players:</span>
                <span className="text-white font-bold flex items-center gap-2">
                  <Users className="w-4 h-4" />
                  {room.player_count || 0}/{room.max_players}
                </span>
              </div>
            </div>
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              onClick={() => onJoinRoom(room.id)}
              className="w-full bg-white text-black py-3 rounded-lg font-bold hover:bg-white/90 transition-all flex items-center justify-center gap-2"
            >
              <Play className="w-5 h-5" />
              Join Table
            </motion.button>
          </motion.div>
        ))}
      </div>
    </>
  )
}
