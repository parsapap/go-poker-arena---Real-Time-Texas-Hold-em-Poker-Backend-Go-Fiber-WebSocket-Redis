'use client'

import { useEffect, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useParams, useRouter } from 'next/navigation'
import Navbar from '@/components/Navbar'
import { ArrowLeft, Send, Trophy, Volume2, VolumeX, MessageCircle, X } from 'lucide-react'
import { usePokerWebSocket } from '@/hooks/usePokerWebSocket'
import { useGameStore } from '@/store/gameStore'
import { soundManager } from '@/lib/sounds'
import { ToastContainer } from '@/components/Toast'
import { ReconnectionOverlay } from '@/components/ReconnectionOverlay'
import { ChipRain } from '@/components/ChipRain'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import type { Toast, ToastType } from '@/types/toast'

function GamePageContent() {
  const params = useParams()
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [raiseAmount, setRaiseAmount] = useState(100)
  const [chatInput, setChatInput] = useState('')
  const [toasts, setToasts] = useState<Toast[]>([])
  const [showChipRain, setShowChipRain] = useState(false)
  const [showChat, setShowChat] = useState(false)

  const {
    players, communityCards, holeCards, phase, pot, currentBet,
    myTurn, winner, chatMessages, soundEnabled, toggleSound
  } = useGameStore()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) { router.push('/login'); return }
    const userData = localStorage.getItem('user')
    if (userData) setUser(JSON.parse(userData))
  }, [router])

  const { isConnected, isReconnecting, sendAction, sendChat, reconnect } = usePokerWebSocket({
    roomId: params.roomId as string,
    userId: user?.id || 0,
    username: user?.username || '',
    enabled: !!user,
    onConnect: () => addToast('success', 'Connected to game'),
    onDisconnect: () => addToast('error', 'Disconnected from game')
  })

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

  const addToast = (type: ToastType, message: string) => {
    setToasts(prev => [...prev, { id: Date.now().toString(), type, message, duration: 3000, timestamp: Date.now() }])
  }

  const handleAction = (action: string, amount?: number) => {
    sendAction(action, amount)
    if (soundEnabled) {
      if (['raise', 'call', 'bet'].includes(action)) soundManager.playChipStack()
      else soundManager.playButtonClick()
    }
  }

  // Authoritative chip/bet values for the current player come from the game
  // store (server-synced), not localStorage (which is stale after a hand).
  const me = players.find((p) => p.id === user?.id)
  const myChips = me?.chips ?? user?.chips ?? 0 // remaining stack
  const myBet = me?.bet ?? 0                     // already committed this round
  const maxRaiseTo = myBet + myChips             // total bet if going all-in
  const minRaiseTo = Math.max((currentBet || 0) * 2, (currentBet || 0) + 20, 20)

  // handleRaise translates the slider's "raise to" total into what the backend
  // expects. The engine treats a raise `amount` as the increment ABOVE the
  // current bet (it needs currentBet - myBet + amount chips). Committing the
  // whole stack must use the dedicated `allin` action, otherwise the required
  // chips exceed the stack and the server rejects it as "insufficient chips".
  const handleRaise = () => {
    const raiseTo = Math.min(raiseAmount, maxRaiseTo)
    if (raiseTo >= maxRaiseTo) {
      handleAction('allin')
      return
    }
    const increment = raiseTo - (currentBet || 0)
    handleAction('raise', increment)
  }

  const handleSendChat = () => {
    if (!chatInput.trim()) return
    sendChat(chatInput)
    setChatInput('')
  }

  if (!user) {
    return (
      <div className="fixed inset-0 bg-gradient-to-br from-gray-900 to-black flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-emerald-500"></div>
      </div>
    )
  }

  // Responsive seat positions - different for mobile vs desktop
  const getPlayerPosition = (idx: number, total: number) => {
    // For 2 players: opponent at top, you at bottom
    if (total <= 2) {
      const positions = [
        { left: '50%', top: '75%' }, // You (bottom)
        { left: '50%', top: '8%' },  // Opponent (top)
      ]
      return positions[idx] || positions[0]
    }
    // For more players, distribute around table
    const angle = (idx / Math.max(total, 6)) * 2 * Math.PI - Math.PI / 2
    const radiusX = 42
    const radiusY = 38
    return {
      left: `${50 + radiusX * Math.cos(angle)}%`,
      top: `${50 + radiusY * Math.sin(angle)}%`
    }
  }

  return (
    <div className="fixed inset-0 bg-gradient-to-br from-gray-900 via-emerald-950 to-gray-900 overflow-hidden">
      <Navbar chips={user?.chips || 5000} />
      
      {/* Top bar with controls */}
      <div className="fixed top-16 left-0 right-0 z-40 px-3 py-2 flex justify-between items-center">
        <motion.button
          whileTap={{ scale: 0.95 }}
          onClick={() => router.push('/')}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white/10 backdrop-blur text-sm"
        >
          <ArrowLeft className="w-4 h-4" />
          <span className="hidden sm:inline">Lobby</span>
        </motion.button>

        <div className="flex items-center gap-2">
          {/* Connection status */}
          <div className={`px-2.5 py-1 rounded-full text-xs flex items-center gap-1.5 ${
            isConnected ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
          }`}>
            <div className={`w-1.5 h-1.5 rounded-full ${isConnected ? 'bg-green-400' : 'bg-red-400 animate-pulse'}`} />
            <span className="hidden sm:inline">{isConnected ? 'Live' : 'Offline'}</span>
          </div>

          {/* Sound toggle */}
          <button onClick={() => { toggleSound(); soundManager.setEnabled(!soundEnabled) }}
            className="p-2 rounded-lg bg-white/10 backdrop-blur">
            {soundEnabled ? <Volume2 className="w-4 h-4" /> : <VolumeX className="w-4 h-4" />}
          </button>

          {/* Chat toggle (mobile) */}
          <button onClick={() => setShowChat(!showChat)}
            className="p-2 rounded-lg bg-white/10 backdrop-blur lg:hidden relative">
            <MessageCircle className="w-4 h-4" />
            {chatMessages.length > 0 && (
              <span className="absolute -top-1 -right-1 w-4 h-4 bg-emerald-500 rounded-full text-[10px] flex items-center justify-center">
                {chatMessages.length > 9 ? '9+' : chatMessages.length}
              </span>
            )}
          </button>
        </div>
      </div>

      <ReconnectionOverlay isReconnecting={isReconnecting} onRetry={reconnect} />
      <ToastContainer toasts={toasts} onClose={(id) => setToasts(prev => prev.filter(t => t.id !== id))} />
      {showChipRain && <ChipRain />}

      {/* Winner Modal */}
      <AnimatePresence>
        {winner && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
            <motion.div initial={{ scale: 0.8, y: 20 }} animate={{ scale: 1, y: 0 }}
              className="bg-gradient-to-br from-gray-800 to-gray-900 p-6 sm:p-8 rounded-2xl text-center max-w-sm w-full border border-amber-500/30">
              <Trophy className="w-16 h-16 mx-auto text-amber-400 mb-4" />
              <h2 className="text-2xl sm:text-3xl font-bold mb-2">
                {winner.id === user?.id ? '🎉 You Won!' : `${winner.name} Wins!`}
              </h2>
              <p className="text-2xl text-emerald-400 font-bold mb-1">${winner.amount}</p>
              <p className="text-gray-400 mb-6">{winner.hand}</p>
              <div className="flex gap-3">
                <button onClick={() => { useGameStore.getState().setWinner(null); useGameStore.getState().setPhase('waiting') }}
                  className="flex-1 py-3 bg-emerald-600 hover:bg-emerald-500 rounded-xl font-semibold transition">
                  Play Again
                </button>
                <button onClick={() => router.push('/')}
                  className="flex-1 py-3 bg-gray-700 hover:bg-gray-600 rounded-xl font-semibold transition">
                  Leave
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Game Area */}
      <div className="absolute inset-0 pt-28 pb-4 px-2 sm:px-4 flex flex-col">
        {/* Poker Table */}
        <div className="flex-1 relative max-w-5xl mx-auto w-full">
          {/* Table background */}
          <div className="absolute inset-4 sm:inset-8 rounded-[40%] sm:rounded-[45%] bg-gradient-to-br from-emerald-800 via-green-800 to-teal-800 shadow-2xl"
            style={{ boxShadow: '0 0 60px rgba(16, 185, 129, 0.2), inset 0 0 60px rgba(0,0,0,0.4)' }} />
          <div className="absolute inset-8 sm:inset-16 rounded-[40%] sm:rounded-[45%] border-2 border-amber-600/20" />

          {/* Community Cards with Pot below */}
          <div className="absolute top-[38%] sm:top-[40%] left-1/2 -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center">
            {/* Cards */}
            <div className="flex gap-1.5 sm:gap-2">
              {communityCards.map((card, i) => (
                <motion.div key={i} initial={{ rotateY: 180, scale: 0 }} animate={{ rotateY: 0, scale: 1 }}
                  transition={{ delay: i * 0.1 }}
                  className={`w-10 h-14 sm:w-14 sm:h-20 md:w-16 md:h-24 bg-white rounded-lg sm:rounded-xl shadow-lg flex items-center justify-center text-lg sm:text-2xl md:text-3xl font-bold border border-gray-200 ${
                    card.includes('♥') || card.includes('♦') ? 'text-red-600' : 'text-gray-900'
                  }`}>
                  {card}
                </motion.div>
              ))}
              {/* Empty card slots */}
              {Array.from({ length: Math.max(0, 5 - communityCards.length) }).map((_, i) => (
                <div key={`empty-${i}`} className="w-10 h-14 sm:w-14 sm:h-20 md:w-16 md:h-24 rounded-lg sm:rounded-xl border-2 border-dashed border-white/10" />
              ))}
            </div>
            
            {/* Pot display - below cards */}
            <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} className="mt-3 sm:mt-4">
              <div className="bg-black/50 backdrop-blur-sm px-4 sm:px-6 py-1.5 sm:py-2 rounded-full border border-amber-500/30">
                <span className="text-[10px] sm:text-xs text-amber-400/70 mr-2">POT</span>
                <span className="text-lg sm:text-xl font-bold text-amber-400">${pot}</span>
              </div>
            </motion.div>
          </div>

          {/* Players */}
          {players.map((player, idx) => {
            const pos = getPlayerPosition(idx, players.length)
            const isYou = player.id === user?.id
            
            return (
              <motion.div key={player.id} initial={{ scale: 0 }} animate={{ scale: 1 }}
                transition={{ delay: idx * 0.1 }}
                className="absolute z-30 -translate-x-1/2 -translate-y-1/2"
                style={{ left: pos.left, top: pos.top }}>
                
                <div className={`relative ${isYou && myTurn ? 'ring-2 ring-emerald-400 ring-offset-2 ring-offset-transparent' : ''} 
                  ${player.isFolded ? 'opacity-40' : ''} rounded-xl`}>
                  
                  {/* Player bet */}
                  {player.bet > 0 && (
                    <div className="absolute -top-6 left-1/2 -translate-x-1/2 bg-amber-500/20 px-2 py-0.5 rounded text-xs text-amber-400 font-bold">
                      ${player.bet}
                    </div>
                  )}

                  {/* Player info card */}
                  <div className={`bg-gray-800/90 backdrop-blur rounded-xl p-2 sm:p-3 min-w-[100px] sm:min-w-[120px] border ${
                    isYou ? 'border-emerald-500/50' : 'border-white/10'
                  }`}>
                    <div className="flex items-center gap-1.5 mb-1">
                      {/* Online indicator - light green when active/online */}
                      <div className={`w-2.5 h-2.5 rounded-full ${
                        player.isActive !== false ? 'bg-lime-400 shadow-sm shadow-lime-400/50' : 'bg-gray-500'
                      }`} />
                      <span className="text-xs sm:text-sm font-medium truncate max-w-[80px]">
                        {isYou ? 'You' : player.username}
                      </span>
                    </div>
                    <div className="text-amber-400 font-bold text-sm sm:text-base">${player.chips}</div>
                  </div>

                  {/* Hole cards for current player */}
                  {isYou && holeCards.length > 0 && (
                    <div className="flex gap-1.5 mt-2 justify-center">
                      {holeCards.map((card, i) => (
                        <motion.div key={i} initial={{ rotateY: 180 }} animate={{ rotateY: 0 }}
                          transition={{ delay: 0.3 + i * 0.1 }}
                          className={`w-12 h-[68px] sm:w-16 sm:h-[88px] md:w-20 md:h-28 bg-white rounded-lg sm:rounded-xl shadow-xl flex items-center justify-center text-xl sm:text-3xl md:text-4xl font-bold border-2 border-gray-200 ${
                            card.includes('♥') || card.includes('♦') ? 'text-red-600' : 'text-gray-900'
                          }`}>
                          {card}
                        </motion.div>
                      ))}
                    </div>
                  )}

                  {/* Face-down cards for opponents */}
                  {!isYou && !player.isFolded && phase !== 'waiting' && (
                    <div className="flex gap-1 mt-2 justify-center">
                      <div className="w-8 h-11 sm:w-10 sm:h-14 bg-gradient-to-br from-blue-800 to-blue-900 rounded-lg border border-white/20" />
                      <div className="w-8 h-11 sm:w-10 sm:h-14 bg-gradient-to-br from-blue-800 to-blue-900 rounded-lg border border-white/20" />
                    </div>
                  )}
                </div>
              </motion.div>
            )
          })}
        </div>

        {/* Action Bar */}
        <AnimatePresence>
          {myTurn && (
            <motion.div initial={{ y: 100, opacity: 0 }} animate={{ y: 0, opacity: 1 }} exit={{ y: 100, opacity: 0 }}
              className="mt-auto bg-gray-900/95 backdrop-blur rounded-t-2xl sm:rounded-2xl p-3 sm:p-4 mx-auto w-full max-w-lg border-t sm:border border-white/10">
              
              {/* Raise slider */}
              <div className="mb-3">
                <div className="flex justify-between text-xs sm:text-sm mb-1.5">
                  <span className="text-gray-400">Raise to</span>
                  <span className="text-amber-400 font-bold">
                    ${Math.min(raiseAmount, maxRaiseTo)}{Math.min(raiseAmount, maxRaiseTo) >= maxRaiseTo ? ' (All-in)' : ''}
                  </span>
                </div>
                <input type="range" min={Math.min(minRaiseTo, maxRaiseTo)} max={maxRaiseTo} step="10"
                  value={Math.min(raiseAmount, maxRaiseTo)} onChange={(e) => setRaiseAmount(Number(e.target.value))}
                  className="w-full h-2 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-emerald-500" />
                <div className="flex justify-between text-[10px] sm:text-xs text-gray-500 mt-1">
                  <button onClick={() => setRaiseAmount(Math.min(minRaiseTo, maxRaiseTo))} className="hover:text-white">Min</button>
                  <button onClick={() => setRaiseAmount(Math.min(Math.max(pot || 50, minRaiseTo), maxRaiseTo))} className="hover:text-white">Pot</button>
                  <button onClick={() => setRaiseAmount(maxRaiseTo)} className="hover:text-white">All-in</button>
                </div>
              </div>

              {/* Action buttons */}
              <div className="grid grid-cols-3 gap-2">
                <motion.button whileTap={{ scale: 0.95 }} onClick={() => handleAction('fold')}
                  className="py-3 sm:py-4 rounded-xl font-bold text-sm sm:text-base bg-red-500/20 border border-red-500/50 text-red-400 active:bg-red-500/30">
                  Fold
                </motion.button>
                <motion.button whileTap={{ scale: 0.95 }} onClick={() => handleAction(currentBet > 0 ? 'call' : 'check')}
                  className="py-3 sm:py-4 rounded-xl font-bold text-sm sm:text-base bg-gray-700 border border-gray-600 active:bg-gray-600">
                  {currentBet > 0 ? `Call $${currentBet}` : 'Check'}
                </motion.button>
                <motion.button whileTap={{ scale: 0.95 }} onClick={handleRaise}
                  className="py-3 sm:py-4 rounded-xl font-bold text-sm sm:text-base bg-emerald-600 active:bg-emerald-500 shadow-lg shadow-emerald-500/30">
                  {Math.min(raiseAmount, maxRaiseTo) >= maxRaiseTo ? 'All-in' : 'Raise'}
                </motion.button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Chat Panel - Desktop sidebar / Mobile overlay */}
      <AnimatePresence>
        {(showChat || typeof window !== 'undefined' && window.innerWidth >= 1024) && (
          <motion.div initial={{ x: 300, opacity: 0 }} animate={{ x: 0, opacity: 1 }} exit={{ x: 300, opacity: 0 }}
            className={`fixed z-50 bg-gray-900/95 backdrop-blur border-l border-white/10 flex flex-col
              ${showChat ? 'inset-0 pt-16' : 'hidden lg:flex right-0 top-28 bottom-4 w-72 rounded-l-2xl'}`}>
            
            {/* Chat header */}
            <div className="flex items-center justify-between p-3 border-b border-white/10">
              <h3 className="font-semibold text-sm">Chat</h3>
              <button onClick={() => setShowChat(false)} className="lg:hidden p-1">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-3 space-y-2">
              {chatMessages.map((msg, i) => (
                <div key={i} className="text-sm">
                  <span className="text-emerald-400 font-medium">{msg.user}:</span>
                  <span className="text-gray-300 ml-1.5">{msg.message}</span>
                </div>
              ))}
              {chatMessages.length === 0 && (
                <p className="text-gray-500 text-sm text-center py-4">No messages yet</p>
              )}
            </div>

            {/* Input */}
            <div className="p-3 border-t border-white/10 flex gap-2">
              <input type="text" value={chatInput} onChange={(e) => setChatInput(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSendChat()}
                placeholder="Message..." className="flex-1 bg-gray-800 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-emerald-500" />
              <button onClick={handleSendChat} className="p-2 bg-emerald-600 rounded-lg hover:bg-emerald-500">
                <Send className="w-4 h-4" />
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

export default function GamePage() {
  return <ErrorBoundary><GamePageContent /></ErrorBoundary>
}
