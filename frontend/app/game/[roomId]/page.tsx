'use client'

import { useEffect, useState, useRef } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useParams, useRouter } from 'next/navigation'
import Navbar from '@/components/Navbar'
import { ArrowLeft, Send, Trophy, Sparkles } from 'lucide-react'

interface Player {
  id: string
  username: string
  chips: number
  bet: number
  cards: string[]
  isFolded: boolean
  isActive: boolean
  position: number
}

interface GameState {
  pot: number
  communityCards: string[]
  currentPlayer: string
  players: Player[]
  phase: 'preflop' | 'flop' | 'turn' | 'river' | 'showdown'
  winner?: string
}

export default function GamePage() {
  const params = useParams()
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [gameState, setGameState] = useState<GameState>({
    pot: 2500,
    communityCards: ['🂠', '🂠', '🂠', '🂠', '🂠'],
    currentPlayer: '',
    players: [],
    phase: 'preflop'
  })
  const [raiseAmount, setRaiseAmount] = useState(100)
  const [timeLeft, setTimeLeft] = useState(10)
  const [isMyTurn, setIsMyTurn] = useState(true)
  const [chatMessages, setChatMessages] = useState<Array<{user: string, msg: string}>>([])
  const [chatInput, setChatInput] = useState('')
  const [showWinner, setShowWinner] = useState(false)

  // Mock players for demo
  const mockPlayers: Player[] = [
    { id: '1', username: 'You', chips: 5000, bet: 100, cards: ['🂡', '🂱'], isFolded: false, isActive: true, position: 0 },
    { id: '2', username: 'Player2', chips: 3200, bet: 100, cards: ['🂠', '🂠'], isFolded: false, isActive: false, position: 1 },
    { id: '3', username: 'Player3', chips: 0, bet: 0, cards: [], isFolded: false, isActive: false, position: 2 },
    { id: '4', username: 'Player4', chips: 4500, bet: 100, cards: ['🂠', '🂠'], isFolded: false, isActive: false, position: 3 },
    { id: '5', username: 'Player5', chips: 0, bet: 0, cards: [], isFolded: false, isActive: false, position: 4 },
    { id: '6', username: 'Player6', chips: 2800, bet: 100, cards: ['🂠', '🂠'], isFolded: true, isActive: false, position: 5 },
    { id: '7', username: 'Player7', chips: 0, bet: 0, cards: [], isFolded: false, isActive: false, position: 6 },
    { id: '8', username: 'Player8', chips: 6100, bet: 100, cards: ['🂠', '🂠'], isFolded: false, isActive: false, position: 7 },
  ]

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

    setGameState(prev => ({ ...prev, players: mockPlayers }))

    // Timer countdown
    const timer = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 0) return 10
        return prev - 1
      })
    }, 1000)

    return () => clearInterval(timer)
  }, [params.roomId, router])

  const handleAction = (action: string, amount?: number) => {
    console.log(`Action: ${action}`, amount)
    // TODO: Send action to WebSocket
  }

  const sendChat = () => {
    if (!chatInput.trim()) return
    setChatMessages(prev => [...prev, { user: user?.username || 'You', msg: chatInput }])
    setChatInput('')
  }

  if (!user) return null

  const seatPositions = [
    { x: '50%', y: '85%', transform: 'translate(-50%, -50%)' }, // Bottom (You)
    { x: '15%', y: '70%', transform: 'translate(-50%, -50%)' }, // Bottom-left
    { x: '5%', y: '40%', transform: 'translate(-50%, -50%)' },  // Left
    { x: '15%', y: '15%', transform: 'translate(-50%, -50%)' }, // Top-left
    { x: '40%', y: '5%', transform: 'translate(-50%, -50%)' },  // Top-left-center
    { x: '60%', y: '5%', transform: 'translate(-50%, -50%)' },  // Top-right-center
    { x: '85%', y: '15%', transform: 'translate(-50%, -50%)' }, // Top-right
    { x: '95%', y: '40%', transform: 'translate(-50%, -50%)' }, // Right
  ]

  return (
    <div className="fixed inset-0 bg-black overflow-hidden">
      {/* Elegant felt background */}
      <div className="absolute inset-0 bg-gradient-to-br from-emerald-950 via-green-950 to-teal-950" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-emerald-900/20 via-transparent to-transparent" />
      
      {/* Subtle pattern overlay */}
      <div className="absolute inset-0 opacity-5" style={{
        backgroundImage: 'repeating-linear-gradient(45deg, transparent, transparent 10px, rgba(255,255,255,.03) 10px, rgba(255,255,255,.03) 20px)'
      }} />

      <Navbar chips={user?.chips || 5000} />

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

      {/* Main poker table */}
      <div className="absolute inset-0 flex items-center justify-center p-4 pt-20">
        <div className="relative w-full max-w-7xl aspect-[16/10]">
          
          {/* Table felt with premium border */}
          <motion.div
            initial={{ scale: 0.9, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="absolute inset-0 rounded-[50%] bg-gradient-to-br from-emerald-900 via-green-900 to-teal-900 shadow-2xl"
            style={{
              boxShadow: '0 0 100px rgba(16, 185, 129, 0.3), inset 0 0 100px rgba(0,0,0,0.5)'
            }}
          />
          
          {/* Inner felt border */}
          <div className="absolute inset-8 rounded-[50%] border-4 border-amber-600/30 shadow-inner" />
          <div className="absolute inset-12 rounded-[50%] border-2 border-amber-500/20" />

          {/* Center pot with glow */}
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
                  key={gameState.pot}
                  initial={{ scale: 1.5, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  className="text-3xl font-bold text-amber-400 text-center"
                >
                  ${gameState.pot.toLocaleString()}
                </motion.div>
              </div>
            </motion.div>
          </motion.div>

          {/* Community cards */}
          <div className="absolute top-[55%] left-1/2 -translate-x-1/2 -translate-y-1/2 flex gap-3 z-10">
            {gameState.communityCards.map((card, i) => (
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
                  {card}
                </div>
              </motion.div>
            ))}
          </div>

          {/* Player seats */}
          {gameState.players.map((player, idx) => {
            const pos = seatPositions[idx]
            const isEmpty = player.chips === 0
            const isYou = idx === 0
            
            return (
              <motion.div
                key={player.id}
                initial={{ scale: 0, opacity: 0 }}
                animate={{ scale: 1, opacity: isEmpty ? 0.3 : 1 }}
                transition={{ delay: idx * 0.1 }}
                className="absolute z-30"
                style={{ left: pos.x, top: pos.y, transform: pos.transform }}
              >
                {/* Timer ring */}
                {isMyTurn && isYou && (
                  <motion.div
                    className="absolute -inset-4 rounded-full"
                    style={{
                      background: `conic-gradient(from 0deg, #10b981 ${(timeLeft / 10) * 360}deg, transparent ${(timeLeft / 10) * 360}deg)`
                    }}
                  />
                )}

                {/* Player card */}
                <motion.div
                  animate={isYou && isMyTurn ? { 
                    boxShadow: ['0 0 20px rgba(16, 185, 129, 0.5)', '0 0 40px rgba(16, 185, 129, 0.8)', '0 0 20px rgba(16, 185, 129, 0.5)']
                  } : {}}
                  transition={{ duration: 1.5, repeat: Infinity }}
                  className={`relative glass rounded-2xl p-3 min-w-[140px] ${
                    player.isFolded ? 'opacity-40' : ''
                  } ${isEmpty ? 'border-dashed border-white/20' : 'border-2 border-white/30'}`}
                >
                  {/* Winner glow */}
                  {gameState.winner === player.id && (
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
                        
                        {/* Bet amount */}
                        {player.bet > 0 && (
                          <motion.div
                            initial={{ scale: 0 }}
                            animate={{ scale: 1 }}
                            className="absolute -top-8 left-1/2 -translate-x-1/2 glass px-3 py-1 rounded-lg text-xs font-bold text-amber-400"
                          >
                            ${player.bet}
                          </motion.div>
                        )}

                        {/* Player cards */}
                        {isYou && player.cards.length > 0 && (
                          <div className="flex gap-1 mt-2 justify-center">
                            {player.cards.map((card, i) => (
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
                      </>
                    )}
                  </div>
                </motion.div>
              </motion.div>
            )
          })}
        </div>
      </div>

      {/* Action bar (bottom) */}
      <AnimatePresence>
        {isMyTurn && (
          <motion.div
            initial={{ y: 100, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: 100, opacity: 0 }}
            className="fixed bottom-0 left-0 right-0 z-40 pb-4 px-4"
          >
            <div className="max-w-4xl mx-auto glass rounded-2xl p-4 border-2 border-emerald-500/30">
              {/* Raise slider */}
              <div className="mb-4">
                <div className="flex justify-between text-sm mb-2">
                  <span className="text-white/60">Raise Amount</span>
                  <span className="text-amber-400 font-bold">${raiseAmount}</span>
                </div>
                <input
                  type="range"
                  min="100"
                  max="5000"
                  step="50"
                  value={raiseAmount}
                  onChange={(e) => setRaiseAmount(Number(e.target.value))}
                  className="w-full h-2 bg-white/10 rounded-lg appearance-none cursor-pointer slider"
                />
                <div className="flex justify-between text-xs text-white/40 mt-1">
                  <button onClick={() => setRaiseAmount(200)} className="hover:text-white">2x</button>
                  <button onClick={() => setRaiseAmount(300)} className="hover:text-white">3x</button>
                  <button onClick={() => setRaiseAmount(gameState.pot)} className="hover:text-white">Pot</button>
                  <button onClick={() => setRaiseAmount(5000)} className="hover:text-white">All-in</button>
                </div>
              </div>

              {/* Action buttons */}
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
                  onClick={() => handleAction('call')}
                  className="py-4 rounded-xl font-bold glass hover:bg-white/10 border-2 border-white/30"
                >
                  Call $100
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
              <span className="text-white/80 ml-2">{msg.msg}</span>
            </div>
          ))}
        </div>
        <div className="flex gap-2">
          <input
            type="text"
            value={chatInput}
            onChange={(e) => setChatInput(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && sendChat()}
            placeholder="Type message..."
            className="flex-1 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-emerald-500/50"
          />
          <button
            onClick={sendChat}
            className="glass p-2 rounded-lg hover:bg-white/10"
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
      </motion.div>

      {/* Winner modal */}
      <AnimatePresence>
        {showWinner && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center"
            onClick={() => setShowWinner(false)}
          >
            <motion.div
              initial={{ scale: 0, rotateY: 180 }}
              animate={{ scale: 1, rotateY: 0 }}
              exit={{ scale: 0, rotateY: 180 }}
              className="relative"
            >
              {/* Confetti effect */}
              {[...Array(50)].map((_, i) => (
                <motion.div
                  key={i}
                  initial={{ y: 0, opacity: 1 }}
                  animate={{
                    y: [0, -500],
                    x: [0, (Math.random() - 0.5) * 400],
                    opacity: [1, 0],
                    rotate: [0, Math.random() * 360]
                  }}
                  transition={{ duration: 2, delay: Math.random() * 0.5 }}
                  className="absolute top-0 left-1/2 w-3 h-3 rounded-full"
                  style={{ backgroundColor: ['#fbbf24', '#f59e0b', '#10b981', '#3b82f6'][Math.floor(Math.random() * 4)] }}
                />
              ))}

              <div className="glass rounded-3xl p-12 text-center border-4 border-amber-500/50">
                <motion.div
                  animate={{ rotate: [0, 10, -10, 0] }}
                  transition={{ duration: 0.5, repeat: Infinity, repeatDelay: 1 }}
                >
                  <Trophy className="w-24 h-24 mx-auto mb-4 text-amber-400" />
                </motion.div>
                <h2 className="text-4xl font-bold mb-2 text-amber-400">You Win!</h2>
                <p className="text-2xl text-white/80 mb-4">$2,500</p>
                <motion.div
                  animate={{ scale: [1, 1.1, 1] }}
                  transition={{ duration: 1, repeat: Infinity }}
                  className="flex items-center justify-center gap-2 text-emerald-400"
                >
                  <Sparkles className="w-5 h-5" />
                  <span className="font-bold">Royal Flush!</span>
                  <Sparkles className="w-5 h-5" />
                </motion.div>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

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
