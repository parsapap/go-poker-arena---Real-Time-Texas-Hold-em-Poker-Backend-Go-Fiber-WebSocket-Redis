import { useEffect, useRef, useState, useCallback } from 'react'
import { useGameStore } from '@/store/gameStore'

interface WebSocketMessage {
  type: string
  room_id?: string
  user_id?: number
  username?: string
  payload?: any
}

interface UsePokerWebSocketProps {
  roomId: string
  userId: number
  username: string
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}

export function usePokerWebSocket({
  roomId,
  userId,
  username,
  onConnect,
  onDisconnect,
  onError
}: UsePokerWebSocketProps) {
  const ws = useRef<WebSocket | null>(null)
  const reconnectTimeout = useRef<NodeJS.Timeout>()
  const heartbeatInterval = useRef<NodeJS.Timeout>()
  const [isConnected, setIsConnected] = useState(false)
  const [isReconnecting, setIsReconnecting] = useState(false)
  const reconnectAttempts = useRef(0)
  const maxReconnectAttempts = 5

  const {
    setPlayers,
    setCommunityCards,
    setHoleCards,
    setPhase,
    setPot,
    setCurrentBet,
    setMyTurn,
    addChatMessage,
    setWinner,
    updatePlayer,
    removePlayer
  } = useGameStore()

  const connect = useCallback(() => {
    if (ws.current?.readyState === WebSocket.OPEN) return

    try {
      // Connect to WebSocket (proxied through Next.js)
      const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'
      const url = `${wsUrl}/ws?user_id=${userId}&username=${encodeURIComponent(username)}&room_id=${roomId}`
      
      ws.current = new WebSocket(url)

      ws.current.onopen = () => {
        console.log('✅ WebSocket connected')
        setIsConnected(true)
        setIsReconnecting(false)
        reconnectAttempts.current = 0
        onConnect?.()

        // Start heartbeat
        heartbeatInterval.current = setInterval(() => {
          if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({ type: 'ping' }))
          }
        }, 30000) // Every 30 seconds

        // Send join message
        sendMessage({
          type: 'join',
          room_id: roomId,
          user_id: userId,
          username
        })
      }

      ws.current.onclose = () => {
        console.log('❌ WebSocket disconnected')
        setIsConnected(false)
        clearInterval(heartbeatInterval.current)
        onDisconnect?.()

        // Attempt reconnection
        if (reconnectAttempts.current < maxReconnectAttempts) {
          setIsReconnecting(true)
          reconnectAttempts.current++
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000)
          console.log(`🔄 Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current})`)
          
          reconnectTimeout.current = setTimeout(() => {
            connect()
          }, delay)
        }
      }

      ws.current.onerror = (error) => {
        console.error('❌ WebSocket error:', error)
        onError?.(error)
      }

      ws.current.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          handleMessage(message)
        } catch (error) {
          console.error('Error parsing message:', error)
        }
      }
    } catch (error) {
      console.error('Error connecting to WebSocket:', error)
    }
  }, [roomId, userId, username, onConnect, onDisconnect, onError])

  const handleMessage = useCallback((message: WebSocketMessage) => {
    console.log('📨 Received:', message.type, message.payload)

    switch (message.type) {
      case 'join':
        if (message.payload) {
          updatePlayer(message.payload)
          addChatMessage({
            user: 'System',
            message: `${message.username} joined the table`,
            timestamp: Date.now()
          })
        }
        break

      case 'leave':
        if (message.user_id) {
          removePlayer(message.user_id)
          addChatMessage({
            user: 'System',
            message: `${message.username} left the table`,
            timestamp: Date.now()
          })
        }
        break

      case 'gameState':
        // Full game state update
        if (message.payload) {
          const { players, communityCards, pot, currentBet, phase, currentPlayer } = message.payload
          setPlayers(players || [])
          setCommunityCards(communityCards || [])
          setPot(pot || 0)
          setCurrentBet(currentBet || 0)
          setPhase(phase || 'waiting')
          setMyTurn(currentPlayer === userId)
        }
        break

      case 'deal':
        // Cards dealt to players
        if (message.payload) {
          const { holeCards, playerId } = message.payload
          if (playerId === userId) {
            setHoleCards(holeCards)
          }
        }
        break

      case 'playerAction':
        // Player made an action
        if (message.payload) {
          const { playerId, action, amount, newPot, newBet } = message.payload
          updatePlayer({ id: playerId, lastAction: action, bet: amount })
          if (newPot !== undefined) setPot(newPot)
          if (newBet !== undefined) setCurrentBet(newBet)
          
          addChatMessage({
            user: message.username || 'Player',
            message: `${action}${amount ? ` $${amount}` : ''}`,
            timestamp: Date.now()
          })
        }
        break

      case 'phaseChange':
        // Game phase changed (flop, turn, river)
        if (message.payload) {
          const { phase, communityCards } = message.payload
          setPhase(phase)
          if (communityCards) {
            setCommunityCards(communityCards)
          }
        }
        break

      case 'showdown':
        // Show all players' cards
        if (message.payload) {
          const { players } = message.payload
          setPlayers(players)
        }
        break

      case 'winner':
        // Game winner announced
        if (message.payload) {
          const { winnerId, winnerName, amount, hand } = message.payload
          setWinner({
            id: winnerId,
            name: winnerName,
            amount,
            hand
          })
          
          addChatMessage({
            user: 'System',
            message: `🏆 ${winnerName} wins $${amount} with ${hand}!`,
            timestamp: Date.now()
          })
        }
        break

      case 'turn':
        // It's someone's turn
        if (message.payload) {
          const { playerId, timeLeft } = message.payload
          setMyTurn(playerId === userId)
        }
        break

      case 'chat':
        // Chat message
        if (message.payload) {
          addChatMessage({
            user: message.username || 'Anonymous',
            message: message.payload.message,
            timestamp: Date.now()
          })
        }
        break

      case 'error':
        // Error message
        console.error('Server error:', message.payload)
        addChatMessage({
          user: 'System',
          message: `Error: ${message.payload?.message || 'Unknown error'}`,
          timestamp: Date.now()
        })
        break

      case 'pong':
        // Heartbeat response
        break

      default:
        console.log('Unknown message type:', message.type)
    }
  }, [userId, setPlayers, setCommunityCards, setHoleCards, setPhase, setPot, setCurrentBet, setMyTurn, addChatMessage, setWinner, updatePlayer, removePlayer])

  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket not connected, cannot send message')
    }
  }, [])

  const sendAction = useCallback((action: string, amount?: number) => {
    sendMessage({
      type: 'action',
      room_id: roomId,
      user_id: userId,
      payload: {
        action,
        amount: amount || 0
      }
    })
  }, [roomId, userId, sendMessage])

  const sendChat = useCallback((message: string) => {
    sendMessage({
      type: 'chat',
      room_id: roomId,
      user_id: userId,
      username,
      payload: {
        message
      }
    })
  }, [roomId, userId, username, sendMessage])

  const disconnect = useCallback(() => {
    clearTimeout(reconnectTimeout.current)
    clearInterval(heartbeatInterval.current)
    
    if (ws.current) {
      // Send leave message
      sendMessage({
        type: 'leave',
        room_id: roomId,
        user_id: userId
      })
      
      ws.current.close()
      ws.current = null
    }
    
    setIsConnected(false)
    setIsReconnecting(false)
  }, [roomId, userId, sendMessage])

  useEffect(() => {
    connect()

    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  return {
    isConnected,
    isReconnecting,
    sendAction,
    sendChat,
    disconnect,
    reconnect: connect
  }
}
