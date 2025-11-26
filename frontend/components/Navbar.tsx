'use client'

import { motion } from 'framer-motion'
import { LogOut, Coins } from 'lucide-react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'

export default function Navbar({ chips = 1000 }: { chips?: number }) {
  const router = useRouter()

  const handleLogout = () => {
    localStorage.removeItem('token')
    router.push('/login')
  }

  return (
    <motion.nav
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      className="fixed top-0 left-0 right-0 z-50 glass border-b border-white/10"
    >
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2 group">
          <motion.div
            whileHover={{ rotate: 360 }}
            transition={{ duration: 0.5 }}
            className="text-2xl"
          >
            🃏
          </motion.div>
          <span className="text-xl font-bold tracking-tight group-hover:text-white/80 transition-colors">
            POKER ARENA
          </span>
        </Link>

        <div className="flex items-center gap-4">
          <motion.div
            whileHover={{ scale: 1.05 }}
            className="glass px-4 py-2 rounded-lg flex items-center gap-2"
          >
            <Coins className="w-5 h-5 text-yellow-400 animate-pulse" />
            <motion.span
              key={chips}
              initial={{ scale: 1.2, color: '#fbbf24' }}
              animate={{ scale: 1, color: '#ffffff' }}
              className="font-bold"
            >
              {chips.toLocaleString()}
            </motion.span>
          </motion.div>

          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={handleLogout}
            className="glass glass-hover px-4 py-2 rounded-lg flex items-center gap-2"
          >
            <LogOut className="w-4 h-4" />
            <span className="hidden sm:inline">Logout</span>
          </motion.button>
        </div>
      </div>
    </motion.nav>
  )
}
