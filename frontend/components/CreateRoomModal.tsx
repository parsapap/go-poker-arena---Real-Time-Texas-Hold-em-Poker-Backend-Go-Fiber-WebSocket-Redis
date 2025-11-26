'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Plus } from 'lucide-react'

interface CreateRoomModalProps {
  onCreateRoom: (roomData: {
    name: string
    small_blind: number
    big_blind: number
    max_players: number
  }) => Promise<void>
}

export default function CreateRoomModal({ onCreateRoom }: CreateRoomModalProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [formData, setFormData] = useState({
    name: '',
    small_blind: 10,
    big_blind: 20,
    max_players: 6,
  })

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      await onCreateRoom(formData)
      setOpen(false)
      setFormData({ name: '', small_blind: 10, big_blind: 20, max_players: 6 })
    } catch (error) {
      console.error('Failed to create room:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <motion.button
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          className="w-full glass glow-border rounded-xl p-6 flex items-center justify-center gap-3 hover:bg-white/10 transition-all"
        >
          <Plus className="w-6 h-6" />
          <span className="text-xl font-bold">Create New Room</span>
        </motion.button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Poker Room</DialogTitle>
          <DialogDescription>
            Set up your custom poker table with your preferred blinds and player limit
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Room Name</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-lg focus:outline-none glow-border-focus"
              placeholder="High Stakes Table"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">Small Blind</label>
              <input
                type="number"
                value={formData.small_blind}
                onChange={(e) => setFormData({ ...formData, small_blind: parseInt(e.target.value) })}
                className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-lg focus:outline-none glow-border-focus"
                min="1"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Big Blind</label>
              <input
                type="number"
                value={formData.big_blind}
                onChange={(e) => setFormData({ ...formData, big_blind: parseInt(e.target.value) })}
                className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-lg focus:outline-none glow-border-focus"
                min="1"
                required
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-2">Max Players (2-8)</label>
            <input
              type="number"
              value={formData.max_players}
              onChange={(e) => setFormData({ ...formData, max_players: parseInt(e.target.value) })}
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-lg focus:outline-none glow-border-focus"
              min="2"
              max="8"
              required
            />
          </div>

          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            type="submit"
            disabled={loading}
            className="w-full bg-white text-black py-3 rounded-lg font-bold hover:bg-white/90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Creating...' : 'Create Room'}
          </motion.button>
        </form>
      </DialogContent>
    </Dialog>
  )
}
