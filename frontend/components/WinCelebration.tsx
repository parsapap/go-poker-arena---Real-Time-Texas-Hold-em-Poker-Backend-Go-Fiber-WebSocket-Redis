'use client'

import { useEffect, useState } from 'react'
import Confetti from 'react-confetti'
import { motion, AnimatePresence } from 'framer-motion'

interface WinCelebrationProps {
  show: boolean
  winnerName: string
  amount: number
  onComplete?: () => void
}

export default function WinCelebration({ show, winnerName, amount, onComplete }: WinCelebrationProps) {
  const [windowSize, setWindowSize] = useState({ width: 0, height: 0 })

  useEffect(() => {
    if (typeof window !== 'undefined') {
      setWindowSize({ width: window.innerWidth, height: window.innerHeight })
      
      const handleResize = () => {
        setWindowSize({ width: window.innerWidth, height: window.innerHeight })
      }
      
      window.addEventListener('resize', handleResize)
      return () => window.removeEventListener('resize', handleResize)
    }
  }, [])

  useEffect(() => {
    if (show && onComplete) {
      const timer = setTimeout(onComplete, 5000)
      return () => clearTimeout(timer)
    }
  }, [show, onComplete])

  return (
    <AnimatePresence>
      {show && (
        <>
          <Confetti
            width={windowSize.width}
            height={windowSize.height}
            recycle={false}
            numberOfPieces={500}
            gravity={0.3}
          />
          
          <motion.div
            initial={{ opacity: 0, scale: 0.5, y: 50 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.5, y: -50 }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            className="fixed inset-0 flex items-center justify-center z-50 pointer-events-none"
          >
            <div className="bg-gradient-to-br from-yellow-400 via-yellow-500 to-yellow-600 text-white px-12 py-8 rounded-2xl shadow-2xl text-center">
              <motion.div
                animate={{ rotate: [0, 10, -10, 10, 0] }}
                transition={{ duration: 0.5, repeat: 2 }}
              >
                <h2 className="text-5xl font-bold mb-4">🎉 {winnerName} Wins! 🎉</h2>
                <motion.p
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ delay: 0.3, type: 'spring', stiffness: 200 }}
                  className="text-3xl font-bold"
                >
                  💰 ${amount.toLocaleString()}
                </motion.p>
              </motion.div>
            </div>
          </motion.div>

          {/* Chip rain effect */}
          <div className="fixed inset-0 pointer-events-none z-40">
            {[...Array(20)].map((_, i) => (
              <motion.div
                key={i}
                initial={{ 
                  x: Math.random() * windowSize.width, 
                  y: -50,
                  rotate: 0 
                }}
                animate={{ 
                  y: windowSize.height + 50,
                  rotate: 360 * (Math.random() > 0.5 ? 1 : -1)
                }}
                transition={{ 
                  duration: 2 + Math.random() * 2,
                  delay: Math.random() * 0.5,
                  ease: 'linear'
                }}
                className="absolute text-4xl"
              >
                🪙
              </motion.div>
            ))}
          </div>
        </>
      )}
    </AnimatePresence>
  )
}
