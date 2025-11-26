'use client'

import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useParams, useRouter } from 'next/navigation'
import Navbar from '@/components/Navbar'
import { ArrowLeft, Send, Trophy, Sparkles, Volume2, VolumeX } from 'lucide-react'
import { usePokerWebSocket } from '@/hooks/usePokerWebSocket'
import { useGameStore } from '@/store/gameStore'
import { soundManager } from '@/lib/sounds'
import { ToastContainer } from '@/components/Toast'
import { ReconnectionOverlay, ConnectionStatus } from '@/components/ReconnectionOverlay'
import { ChipRain } from '@/components/ChipRain'

export default function GamePage() {
  const params = useParams()
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [raiseAmount, setRaiseAmount] = useState(100)
  const [chatInput, setChatInput] = useState('')
  const [toasts, setToasts] = useState<any[]>([])
  const [showChipRain, setShowChipRain] = useState(false)

  const {
    players,
    communityCards,
    holeCards,
    phase,
    pot,
    currentBet,
    myTurn,
    winner,
    chatMessages,
    soundEnabled,
    toggleSound
  } = useGameStore()

  const { isConnected, isReconnecting, sendAction, sendChat, reconnect } = usePokerWebSocket({
    roomId: params.roomId as string,
    userId: user?.id || 0,
    username: user?.username || '',
    onConnect: () => {
      addToast('success', 'Connected to game')
    },
    onDisconnect: () => {
      addToast('error', 'Disconnected from game')
    }
  })

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
  }, [router])

  // Watch for winner and trigger celebrations
  useEffect(() => {
    if (winner) {
      if (winner.id === user?.id) {
        addToast('win', `🎉 You won $${winner.amount} with ${winner.hand}!`)
        setShowChipRain(true)
        if (soundEnabled) soundManager.playWinFanfare()
        
        setTimeout(() => setShowChipRain(false), 4000)
      } else {
        addToast('info', `${winner.name} won $${winner.amount}`)
      }
    }
  }, [winner, user?.id, soundEnabled])

  // Play sounds for game events
  useEffect(() => {
    if (communityCards.length > 0 && soundEnabled) {
      soundManager.playCardDeal()
    }
  }, [communityCards.length, soundEnabled])

  const addToast = (type: string, message: string) => {
    const id = Date.now().toString()
    setToasts((prev: any) => [...prev, { id, type, message }])
  }

  const removeToast = (id: string) => {
    setToasts((prev: any) => prev.filter((t: any) => t.id !== id))
  }

  const handleAction = (action: string, amount?: number) => {
    sendAction(action, amount)
    if (soundEnabled) {
      if (action === 'raise' || action === 'call' || action === 'bet') {
        soundManager.playChipStack()
      } else {
        soundManager.playButtonClick()
      }
    }
  }

  const handleSendChat = () => {
    if (!chatInput.trim()) return
    sendChat(chatInput)
    setChatInput('')
  }

  if (!user) return null

  const seatPositions = [
    { x: '50%', y: '85%', transform: 'translate(-50%, -50%)' },
    { x: '15%', y: '70%', transform: 'translate(-50%, -50%)' },
    { x: '5%', y: '40%', transform: 'translate(-50%, -50%)' },
    { x: '15%', y: '15%', transform: 'translate(-50%, -50%)' },
    { x: '40%', y: '5%', transform: 'translate(-50%, -50%)' },
    { x: '60%', y: '5%', transform: 'translate(-50%, -50%)' },
    { x: '85%', y: '15%', transform: 'translate(-50%, -50%)' },
    { x: '95%', y: '40%', transform: 'translate(-50%, -50%)' },
  ]

  return (
    <div className="fixed inset-0 bg-black overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 bg-gradient-to-br from-emerald-950 via-green-950 to-teal-950" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-emerald-900/20 via-transparent to-transparent" />
      
      <Navbar chips={user?.chips || 5000} />

      {/* Connection status */}
      {!isConnected && <ConnectionStatus isConnected={isConnected} />}

      {/* Reconnection overlay */}
      <ReconnectionOverlay isReconnecting={isReconnecting} onRetry={reconnect} />

      {/* Toast notifications */}
      <ToastContainer toasts={toasts} onClose={removeToast} />

      {/* Chip rain for wins */}
      {showChipRain && <ChipRain />}

      {/* Back button */}
      <motion.button
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        whileHover={{ scale: 1.05 }}
        onClick={() => router.push('/')}
        className="fixed top-20 left-4 z-50 glass px-4 py-2 rounded-lg flex items-center gap-2 hover:bg-white/10"
      >
        <ArrowLeft className="w-4 h-4" />
        <span className="hidden sm:inline">Lobby</span>
      </motion.button>

      {/* Sound toggle */}
      <motion.button
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        whileHover={{ scale: 1.05 }}
        onClick={() => {
          toggleSound()
          soundManager.setEnabled(!soundEnabled)
        }}
        className="fixed top-20 right-4 z-50 glass p-3 rounded-lg hover:bg-white/10"
      >
        {soundEnabled ? <Volume2 className="w-5 h-5" /> : <VolumeX className="w-5 h-5" />}
      </motion.button>

      {/* Main poker table */}
      <div className="absolute inset-0 flex items-center justify-center p-4 pt-20">
        <div className="relative w-full max-w-7xl aspect-[16/10]">
          
          {/* Table felt */}
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="absolute inset-0 rounded-[50%] bg-gradient-to-br from-emerald-900 via-green-900 to-teal-900 shadow-2xl"
            style={{
              boxShadow: '0 0 100px rgba(16, 185, 129, 0.3), inset 0 0 100px rgba(0,0,0,0.5)'
            }}
          />
          
          <div className="absolute inset-8 rounded-[50%] border-4 border-amber-600/30 shadow-inner" />
          <div className="absolute inset-12 rounded-[50%] border-2 border-amber-500/20" />

          {/* Center pot */}
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            className="absolute top-[45%] left-1/2 -translate-x-1/2 -translate-y-1/2 z-20"
          >
            <motion.div
              animate={{ scale: [1, 1.05, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
              className="relative"
            >
              <div className="absolute inset-0 bg-amber-500/20 blur-xl rounded-full" />
              <div className="relative glass px-8 py-4 rounded-2xl border-2 border-amber-500/30">
                <div className="text-xs text-amber-400/80 mb-1 text-center font-semibold tracking-wider">POT</div>
                <motion.div
                  key={pot}
                  initial={{ scale: 1.5, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  className="text-3xl font-bold text-amber-400 text-center"
                >
                  ${pot.toLocaleString()}
                </motion.div>
              </div>
            </motion.div>
          </motion.div>

          {/* Community cards */}
          <div className="absolute top-[55%] left-1/2 -translate-x-1/2 -translate-y-1/2 flex gap-3 z-10">
            {communityCards.map((card, i) => (
              <motion.div
                key={i}
                initial={{ rotateY: 180, y: -100, opacity: 0 }}
                animate={{ rotateY: 0, y: 0, opacity: 1 }}
                transition={{ delay: i * 0.15, type: 'spring', stiffness: 200 }}
                whileHover={{ y: -10, scale: 1.05 }}
                className="relative group"
              >
                <div className="absolute inset-0 bg-gradient-to-br from-amber-400/20 to-orange-500/20 blur-lg opacity-0 group-hover:opacity-100 transition-opacity rounded-xl" />
                <div className="relative w-16 h-24 sm:w-20 sm:h-28 bg-white rounded-xl shadow-2xl flex items-center justify-center text-4xl sm:text-5xl border-2 border-white/20">
                  {card || '🂠'}
                </div>
              </motion.div>
            ))}
          </div>

          {/* Player seats */}
          {players.map((player, idx) => {
            const pos = seatPositions[idx]
            const isEmpty = player.chips === 0
            const isYou = player.id === user?.id
            
            return (
              <motion.div
                key={player.id}
                initial={{ scale: 0, opacity: 0 }}
                animate={{ scale: 1, opacity: isEmpty ? 0.3 : 1 }}
                transition={{ delay: idx * 0.1 }}
                className="absolute z-30"
                style={{ left: pos.x, top: pos.y, transform: pos.transform }}
              >
                <motion.div
                  animate={isYou && myTurn ? { 
                    boxShadow: ['0 0 20px rgba(16, 185, 129, 0.5)', '0 0 40px rgba(16, 185, 129, 0.8)', '0 0 20px rgba(16, 185, 129, 0.5)']
                  } : {}}
                  transition={{ duration: 1.5, repeat: Infinity }}
                  className={`relative glass rounded-2xl p-3 min-w-[140px] ${
                    player.isFolded ? 'opacity-40' : ''
                  } ${isEmpty ? 'border-dashed border-white/20' : 'border-2 border-white/30'}`}
                >
                  {winner?.id === player.id && (
                    <motion.div
                      animate={{ scale: [1, 1.2, 1], opacity: [0.5, 1, 0.5] }}
                      transition={{ duration: 1, repeat: Infinity }}
                      className="absolute -inset-1 bg-gradient-to-r from-amber-400 to-yellow-500 rounded-2xl blur-xl"
                    />
                  )}

                  <div className="relative">
                    {isEmpty ? (
                      <div className="text-center text-white/40 text-sm py-2">Empty Seat</div>
                    ) : (
                      <>
                        <div className="flex items-center gap-2 mb-2">
                          <div className={`w-2 h-2 rounded-full ${player.isActive ? 'bg-green-400 animate-pulse' : 'bg-gray-500'}`} />
                          <div className="font-bold text-sm truncate">{player.username}</div>
                        </div>
                        <div className="text-amber-400 font-bold text-lg">${player.chips.toLocaleString()}</div>
                        
                        {player.bet > 0 && (
                          <motion.div
                            initial={{ scale: 0 }}
                            animate={{ scale: 1 }}
                            className="absolute -top-8 left-1/2 -translate-x-1/2 glass px-3 py-1 rounded-lg text-xs font-bold text-amber-400"
                          >
                            ${player.bet}
                          </motion.div>
                        )}

                        {/* Player cards - only show to owner */}
                        {isYou && holeCards.length > 0 && (
                          <div className="flex gap-1 mt-2 justify-center">
                            {holeCards.map((card, i) => (
                              <motion.div
                                key={i}
                                initial={{ rotateY: 180, scale: 0 }}
                                animate={{ rotateY: 0, scale: 1 }}
                                transition={{ delay: 0.5 + i * 0.1, type: 'spring' }}
                                whileHover={{ y: -5, scale: 1.1 }}
                                className="w-10 h-14 bg-white rounded-lg shadow-lg flex items-center justify-center text-2xl cursor-pointer"
                              >
                                {card}
                              </motion.div>
                            ))}
                          </div>
                        )}
                        {!isYou && player.cards && player.cards.length > 0 && (
                          <div className="flex gap-1 mt-2 justify-center">
                            {player.cards.map((card, i) => (
                              <div
                                key={i}
                                className="w-10 h-14 bg-gradient-to-br from-blue-900 to-blue-950 rounded-lg shadow-lg border-2 border-white/20"
                              />
                            ))}
                          </div>
                        )}
                      </>
                    )}
                  </div>
                </motion.div>
              </motion.div>
            )
          })}
        </div>
      </div>

      {/* Action bar */}
      <AnimatePresence>
        {myTurn && (
          <motion.div
            initial={{ y: 100, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: 100, opacity: 0 }}
            className="fixed bottom-0 left-0 right-0 z-40 pb-4 px-4"
          >
            <div className="max-w-4xl mx-auto glass rounded-2xl p-4 border-2 border-emerald-500/30">
              <div className="mb-4">
                <div className="flex justify-between text-sm mb-2">
                  <span className="text-white/60">Raise Amount</span>
                  <span className="text-amber-400 font-bold">${raiseAmount}</span>
                </div>
                <input
                  type="range"
                  min={currentBet * 2}
                  max={user?.chips || 5000}
                  step="50"
                  value={raiseAmount}
                  onChange={(e) => setRaiseAmount(Number(e.target.value))}
                  className="w-full h-2 bg-white/10 rounded-lg appearance-none cursor-pointer slider"
                />
                <div className="flex justify-between text-xs text-white/40 mt-1">
                  <button onClick={() => setRaiseAmount(currentBet * 2)} className="hover:text-white">2x</button>
                  <button onClick={() => setRaiseAmount(currentBet * 3)} className="hover:text-white">3x</button>
                  <button onClick={() => setRaiseAmount(pot)} className="hover:text-white">Pot</button>
                  <button onClick={() => setRaiseAmount(user?.chips || 5000)} className="hover:text-white">All-in</button>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3">
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => handleAction('fold')}
                  className="py-4 rounded-xl font-bold bg-red-500/20 hover:bg-red-500/30 border-2 border-red-500/50 text-red-400"
                >
                  Fold
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => handleAction(currentBet > 0 ? 'call' : 'check')}
                  className="py-4 rounded-xl font-bold glass hover:bg-white/10 border-2 border-white/30"
                >
                  {currentBet > 0 ? `Call $${currentBet}` : 'Check'}
                </motion.button>
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => handleAction('raise', raiseAmount)}
                  className="py-4 rounded-xl font-bold bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-400 hover:to-teal-400 shadow-lg shadow-emerald-500/50"
                >
                  Raise ${raiseAmount}
                </motion.button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Chat sidebar */}
      <motion.div
        initial={{ x: 300, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        className="fixed right-4 top-24 bottom-24 w-72 glass rounded-2xl p-4 hidden lg:flex flex-col z-40"
      >
        <h3 className="font-bold mb-3 text-sm text-white/60">CHAT</h3>
        <div className="flex-1 overflow-y-auto space-y-2 mb-3">
          {chatMessages.map((msg, i) => (
            <div key={i} className="text-sm">
              <span className="text-emerald-400 font-semibold">{msg.user}:</span>
              <span className="text-white/80 ml-2">{msg.message}</span>
            </div>
          ))}
        </div>
        <div className="flex gap-2">
          <input
            type="text"
            value={chatInput}
            onChange={(e) => setChatInput(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSendChat()}
            placeholder="Type message..."
            className="flex-1 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500/50"
          />
          <button
            onClick={handleSendChat}
            className="glass p-2 rounded-lg hover:bg-white/10"
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
      </motion.div>

      <style jsx>{`
        .slider::-webkit-slider-thumb {
          appearance: none;
          width: 20px;
          height: 20px;
          border-radius: 50%;
          background: linear-gradient(135deg, #10b981, #14b8a6);
          cursor: pointer;
          box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
        }
        .slider::-moz-range-thumb {
          width: 20px;
          height: 20px;
          border-radius: 50%;
          background: linear-gradient(135deg, #10b981, #14b8a6);
          cursor: pointer;
          box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
          border: none;
        }
      `}</style>
    </div>
  )
}
