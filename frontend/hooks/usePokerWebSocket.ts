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
      // Connect to WebSocket with auto-detected URL
      const wsUrl = getWebSocketUrl()
      const url = `${wsUrl}/ws?user_id=${userId}&username=${encodeURIComponent(username)}&room_id=${roomId}`
      
      ws.current = new WebSocket(url)

      ws.current.onopen = () => {
        logger.ws.connected()
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
        logger.ws.disconnected()
        setIsConnected(false)
        clearInterval(heartbeatInterval.current)
        onDisconnect?.()

        // Attempt reconnection
        if (reconnectAttempts.current < maxReconnectAttempts) {
          setIsReconnecting(true)
          reconnectAttempts.current++
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000)
          logger.ws.reconnecting(reconnectAttempts.current)
          
          reconnectTimeout.current = setTimeout(() => {
            connect()
          }, delay)
        }
      }

      ws.current.onerror = (error) => {
        logger.ws.error(error)
        setIsConnected(false)
        onError?.(error)
        
        // Show user-friendly error
        addChatMessage({
          user: 'System',
          message: 'Connection error. Attempting to reconnect...',
          timestamp: Date.now()
        })
      }

      ws.current.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          handleMessage(message)
        } catch (error) {
          logger.error('Error parsing WebSocket message', error)
        }
      }
    } catch (error) {
      logger.error('Error connecting to WebSocket', error)
    }
  }, [roomId, userId, username, onConnect, onDisconnect, onError])

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
