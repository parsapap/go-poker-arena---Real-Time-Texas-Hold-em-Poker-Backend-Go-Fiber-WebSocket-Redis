import { useEffect, useRef, useState, useCallback } from 'react'
import { useGameStore } from '@/store/gameStore'
import { getWebSocketUrl } from '@/lib/websocketUtils'
import { cardsToStrings, type BackendCard } from '@/lib/cardUtils'
import type { WebSocketMessage, WSMessage } from '@/types/websocket'
import { logger } from '@/lib/logger'

interface UsePokerWebSocketProps {
  roomId: string
  userId: number
  username: string
  enabled?: boolean
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}

export function usePokerWebSocket({
  roomId,
  userId,
  username,
  enabled = true,
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
  const isConnecting = useRef(false) // Prevent duplicate connections

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
    // Prevent duplicate connections
    if (isConnecting.current) {
      if (process.env.NODE_ENV === 'development') {
        logger.debug('Connection already in progress, skipping')
      }
      return
    }
    
    if (ws.current?.readyState === WebSocket.OPEN) {
      if (process.env.NODE_ENV === 'development') {
        logger.debug('WebSocket already connected')
      }
      return
    }
    
    // Don't connect if disabled or user data is not loaded yet
    if (!enabled || !userId || !username) {
      if (process.env.NODE_ENV === 'development') {
        if (!enabled) {
          logger.debug('WebSocket connection disabled')
        } else {
          logger.warn('Cannot connect: user data not loaded')
        }
      }
      return
    }

    try {
      isConnecting.current = true
      setIsReconnecting(reconnectAttempts.current > 0)
      
      // Connect to WebSocket with auto-detected URL
      const wsUrl = getWebSocketUrl()
      const url = `${wsUrl}/ws?user_id=${userId}&username=${encodeURIComponent(username)}&room_id=${roomId}`
      
      if (process.env.NODE_ENV === 'development') {
        logger.info(`Connecting to WebSocket: ${url}`)
      }
      
      ws.current = new WebSocket(url)

      ws.current.onopen = () => {
        isConnecting.current = false
        setIsConnected(true)
        setIsReconnecting(false)
        reconnectAttempts.current = 0
        
        if (process.env.NODE_ENV === 'development') {
          logger.ws.connected()
        }
        
        onConnect?.()

        // Start heartbeat - ping every 20 seconds
        heartbeatInterval.current = setInterval(() => {
          if (ws.current?.readyState === WebSocket.OPEN) {
            try {
              ws.current.send(JSON.stringify({ type: 'ping' }))
            } catch (error) {
              if (process.env.NODE_ENV === 'development') {
                logger.error('Failed to send ping', error)
              }
            }
          }
        }, 20000) // Every 20 seconds

        // Send join message directly (don't use sendMessage to avoid circular dependency)
        try {
          if (ws.current?.readyState === WebSocket.OPEN) {
            ws.current.send(JSON.stringify({
              type: 'join',
              room_id: roomId,
              user_id: userId,
              username
            }))
            if (process.env.NODE_ENV === 'development') {
              logger.debug('Join message sent')
            }
          }
        } catch (error) {
          if (process.env.NODE_ENV === 'development') {
            logger.error('Failed to send join message', error)
          }
        }
      }

      ws.current.onclose = (event) => {
        isConnecting.current = false
        setIsConnected(false)
        clearInterval(heartbeatInterval.current)
        
        if (process.env.NODE_ENV === 'development') {
          logger.ws.disconnected()
          logger.info(`WebSocket closed: code=${event.code}, reason=${event.reason || 'No reason'}`)
          
          // Log common close codes
          const closeReasons: Record<number, string> = {
            1000: 'Normal closure',
            1001: 'Going away',
            1002: 'Protocol error',
            1003: 'Unsupported data',
            1006: 'Abnormal closure (no close frame)',
            1011: 'Server error',
            1012: 'Service restart'
          }
          
          const reasonText = closeReasons[event.code] || 'Unknown'
          logger.warn(`Close code ${event.code}: ${reasonText}`)
        }
        
        onDisconnect?.()

        // Don't reconnect if it was a clean close or if disabled
        if (!enabled || event.code === 1000) {
          if (process.env.NODE_ENV === 'development') {
            logger.info('Clean disconnect, not reconnecting')
          }
          return
        }

        // Attempt reconnection with exponential backoff
        if (reconnectAttempts.current < maxReconnectAttempts) {
          setIsReconnecting(true)
          reconnectAttempts.current++
          
          // Exponential backoff: 1s → 2s → 4s → 8s → 10s (max)
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current - 1), 10000)
          
          if (process.env.NODE_ENV === 'development') {
            logger.ws.reconnecting(reconnectAttempts.current)
            logger.info(`Reconnecting in ${delay}ms (attempt ${reconnectAttempts.current}/${maxReconnectAttempts})`)
          }
          
          reconnectTimeout.current = setTimeout(() => {
            connect()
          }, delay)
        } else {
          if (process.env.NODE_ENV === 'development') {
            logger.error('Max reconnection attempts reached')
          }
          addChatMessage({
            user: 'System',
            message: 'Connection lost. Please refresh the page.',
            timestamp: Date.now()
          })
        }
      }

      ws.current.onerror = (error) => {
        isConnecting.current = false
        setIsConnected(false)
        
        if (process.env.NODE_ENV === 'development') {
          logger.ws.error(error)
        }
        
        onError?.(error)
        
        // Show user-friendly error only once
        if (reconnectAttempts.current === 0) {
          addChatMessage({
            user: 'System',
            message: 'Connection error. Attempting to reconnect...',
            timestamp: Date.now()
          })
        }
      }

      ws.current.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          handleMessage(message)
        } catch (error) {
          if (process.env.NODE_ENV === 'development') {
            logger.error('Error parsing WebSocket message', error)
          }
        }
      }
    } catch (error) {
      isConnecting.current = false
      if (process.env.NODE_ENV === 'development') {
        logger.error('Error connecting to WebSocket', error)
      }
    }
  }, [roomId, userId, username, enabled, onConnect, onDisconnect, onError])

  const handleMessage = useCallback((message: WebSocketMessage) => {
    logger.ws.message(message.type, message.payload)

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
          const { players, communityCards, pot, currentBet, phase, currentPlayer, pots } = message.payload
          setPlayers(players || [])
          
          // Convert backend cards to frontend format
          if (communityCards && Array.isArray(communityCards)) {
            setCommunityCards(cardsToStrings(communityCards as BackendCard[]))
          }
          
          // Calculate total pot from all pots
          const totalPot = pots ? pots.reduce((sum: number, p: any) => sum + (p.amount || 0), 0) : (pot || 0)
          setPot(totalPot)
          setCurrentBet(currentBet || 0)
          setPhase(phase || 'waiting')
          setMyTurn(currentPlayer === userId)
        }
        break

      case 'deal':
        // Cards dealt to players
        if (message.payload) {
          const { holeCards, playerId, communityCards } = message.payload
          if (playerId === userId && holeCards) {
            setHoleCards(cardsToStrings(holeCards as BackendCard[]))
          }
          if (communityCards) {
            setCommunityCards(cardsToStrings(communityCards as BackendCard[]))
          }
        }
        break

      case 'player_action':
      case 'playerAction':
        // Player made an action
        if (message.payload) {
          const { playerId, player_id, action, amount, newPot, newBet, pot, current_bet, game } = message.payload
          const actualPlayerId = playerId || player_id
          
          updatePlayer({ id: actualPlayerId, lastAction: action, bet: amount })
          
          // Update pot from various possible sources
          if (newPot !== undefined) setPot(newPot)
          else if (pot !== undefined) setPot(pot)
          else if (game?.pots) {
            const totalPot = game.pots.reduce((sum: number, p: any) => sum + (p.amount || 0), 0)
            setPot(totalPot)
          }
          
          // Update current bet
          if (newBet !== undefined) setCurrentBet(newBet)
          else if (current_bet !== undefined) setCurrentBet(current_bet)
          else if (game?.current_bet !== undefined) setCurrentBet(game.current_bet)
          
          addChatMessage({
            user: message.username || 'Player',
            message: `${action}${amount ? ` $${amount}` : ''}`,
            timestamp: Date.now()
          })
        }
        break

      case 'phaseChange':
      case 'phase_change':
        // Game phase changed (flop, turn, river)
        if (message.payload) {
          const { phase, communityCards, community_cards, pot, pots } = message.payload
          setPhase(phase)
          
          const cards = communityCards || community_cards
          if (cards && Array.isArray(cards)) {
            setCommunityCards(cardsToStrings(cards as BackendCard[]))
          }
          
          // Update pot on phase change
          if (pots) {
            const totalPot = pots.reduce((sum: number, p: any) => sum + (p.amount || 0), 0)
            setPot(totalPot)
          } else if (pot !== undefined) {
            setPot(pot)
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
        logger.error('Server error', message.payload)
        const errorMsg = message.payload?.message || message.payload?.error || 'Unknown error'
        addChatMessage({
          user: 'System',
          message: `⚠️ Error: ${errorMsg}`,
          timestamp: Date.now()
        })
        
        // Handle specific errors
        if (errorMsg.includes('banned')) {
          setTimeout(() => {
            window.location.href = '/login'
          }, 3000)
        }
        break

      case 'pong':
        // Heartbeat response
        break

      case 'joined':
        // Join confirmation from server
        if (process.env.NODE_ENV === 'development') {
          logger.info('Successfully joined room', message.data || message.payload)
        }
        break

      default:
        logger.warn('Unknown message type', { type: message.type, payload: message.payload })
    }
  }, [userId, setPlayers, setCommunityCards, setHoleCards, setPhase, setPot, setCurrentBet, setMyTurn, addChatMessage, setWinner, updatePlayer, removePlayer])

  const sendMessage = useCallback((message: WebSocketMessage) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message))
      logger.debug('WebSocket message sent', { type: message.type })
    } else {
      logger.warn('WebSocket not connected, cannot send message')
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
    isConnecting.current = false
    
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      try {
        // Send leave message directly
        ws.current.send(JSON.stringify({
          type: 'leave',
          room_id: roomId,
          user_id: userId
        }))
        if (process.env.NODE_ENV === 'development') {
          logger.debug('Leave message sent')
        }
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          logger.error('Failed to send leave message', error)
        }
      }
      
      ws.current.close(1000, 'User disconnected') // Clean close
      ws.current = null
    }
    
    setIsConnected(false)
    setIsReconnecting(false)
    reconnectAttempts.current = 0
  }, [roomId, userId])

  useEffect(() => {
    // Only connect if enabled and user data is available
    if (enabled && userId && username) {
      connect()
    }

    return () => {
      disconnect()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, userId, username, roomId])

  return {
    isConnected,
    isReconnecting,
    sendAction,
    sendChat,
    disconnect,
    reconnect: connect
  }
}
