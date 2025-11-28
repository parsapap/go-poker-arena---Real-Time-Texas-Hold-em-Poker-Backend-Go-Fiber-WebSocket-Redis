'use client'

import { motion } from 'framer-motion'
import Link from 'next/link'
import { useEffect, useState } from 'react'

export default function NotFound() {
  const [cards, setCards] = useState<{ id: number; suit: string; value: string }[]>([])

  useEffect(() => {
    const suits = ['♠️', '♥️', '♦️', '♣️']
    const values = ['A', 'K', 'Q', 'J', '10', '9', '8', '7', '6', '5', '4', '3', '2']
    
    const randomCards = Array.from({ length: 5 }, (_, i) => ({
      id: i,
      suit: suits[Math.floor(Math.random() * suits.length)],
      value: values[Math.floor(Math.random() * values.length)]
    }))
    
    setCards(randomCards)
  }, [])

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-green-900 to-gray-900 flex items-center justify-center p-4">
      <div className="text-center">
        {/* Animated 404 */}
        <motion.div
          initial={{ scale: 0, rotate: -180 }}
          animate={{ scale: 1, rotate: 0 }}
          transition={{ duration: 0.8, type: 'spring' }}
          className="mb-8"
        >
          <h1 className="text-9xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-yellow-400 to-red-600">
            404
          </h1>
        </motion.div>

        {/* Poker Table */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="mb-8"
        >
          <div className="relative w-80 h-48 mx-auto bg-green-800 rounded-full border-8 border-yellow-700 shadow-2xl flex items-center justify-center">
            <div className="absolute inset-0 rounded-full bg-gradient-to-br from-green-700 to-green-900 opacity-50"></div>
            <p className="relative text-2xl font-bold text-yellow-400 text-center px-8">
              Table Not Found
            </p>
          </div>
        </motion.div>

        {/* Floating Cards */}
        <div className="flex justify-center gap-2 mb-8">
          {cards.map((card, index) => (
            <motion.div
              key={card.id}
              initial={{ opacity: 0, y: -50, rotate: -180 }}
              animate={{ 
                opacity: 1, 
                y: 0, 
                rotate: 0,
              }}
              transition={{ 
                delay: 0.5 + index * 0.1,
                type: 'spring',
                stiffness: 200
              }}
              whileHover={{ 
                y: -10,
                rotate: 5,
                transition: { duration: 0.2 }
              }}
              className="w-16 h-24 bg-white rounded-lg shadow-xl flex flex-col items-center justify-center border-2 border-gray-300"
            >
              <span className={`text-3xl ${card.suit === '♥️' || card.suit === '♦️' ? 'text-red-600' : 'text-black'}`}>
                {card.suit}
              </span>
              <span className={`text-xl font-bold ${card.suit === '♥️' || card.suit === '♦️' ? 'text-red-600' : 'text-black'}`}>
                {card.value}
              </span>
            </motion.div>
          ))}
        </div>

        {/* Message */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1 }}
          className="mb-8"
        >
          <p className="text-xl text-gray-300 mb-2">
            Looks like this table doesn't exist!
          </p>
          <p className="text-gray-400">
            The cards have been shuffled, but we can't find what you're looking for.
          </p>
        </motion.div>

        {/* Action Buttons */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.2 }}
          className="flex gap-4 justify-center"
        >
          <Link href="/">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="px-8 py-3 bg-gradient-to-r from-yellow-500 to-yellow-600 text-white font-bold rounded-lg shadow-lg hover:shadow-xl transition-shadow"
            >
              🏠 Back to Lobby
            </motion.button>
          </Link>
          
          <Link href="/rooms">
            <motion.button
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              className="px-8 py-3 bg-gradient-to-r from-green-600 to-green-700 text-white font-bold rounded-lg shadow-lg hover:shadow-xl transition-shadow"
            >
              🎰 Find a Table
            </motion.button>
          </Link>
        </motion.div>

        {/* Decorative Chips */}
        <div className="absolute inset-0 pointer-events-none overflow-hidden">
          {[...Array(10)].map((_, i) => (
            <motion.div
              key={i}
              initial={{ 
                x: Math.random() * window.innerWidth,
                y: -50,
                rotate: 0
              }}
              animate={{ 
                y: window.innerHeight + 50,
                rotate: 360
              }}
              transition={{ 
                duration: 5 + Math.random() * 5,
                delay: Math.random() * 2,
                repeat: Infinity,
                ease: 'linear'
              }}
              className="absolute text-4xl opacity-20"
            >
              🪙
            </motion.div>
          ))}
        </div>
      </div>
    </div>
  )
}
