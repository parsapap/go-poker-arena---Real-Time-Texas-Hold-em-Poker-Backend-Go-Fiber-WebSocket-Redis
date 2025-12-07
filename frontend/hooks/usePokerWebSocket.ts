import { useEffect, useRef, useState } from 'react'
import { useGameStore } from '@/store/gameStore'
import { getWebSocketUrl } from '@/lib/websocketUtils'
import { cardsToStrings, type BackendCard } from '@/lib/cardUtils'
import type { WebSocketMessage } from '@/types/websocket'
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
  const [isConnected, setIsConnected] = useState(false)
  const [isReconnecting, setIsReconnecting] = useState(false)
  
  const wsRef = useRef<WebSocket | null>(null)
  const heartbeatRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectCountRef = useRef(0)
  const mountedRef = useRef(true)

  // Track if we've already connected with these params
  const connectedParamsRef = useRef<string>('')

  // Single effect to manage the entire WebSocket lifecycle
  useEffect(() => {
    mountedRef.current = true
    
    // Don't connect if not enabled or missing required data
    if (!enabled || !userId || !username || !roomId) {
      return
    }

    // Create a unique key for current connection params
    const paramsKey = `${userId}-${username}-${roomId}`
    
    // If already connected with same params, don't reconnect
    if (wsRef.current?.readyState === WebSocket.OPEN && connectedParamsRef.current === paramsKey) {
      logger.info('Already connected with same params, skipping')
      return
    }

    let socket: WebSocket | null = null
    let isCleaningUp = false

    const cleanup = () => {
      isCleaningUp = true
      if (heartbeatRef.current) {
        clearInterval(heartbeatRef.current)
        heartbeatRef.current = null
      }
      if (reconnectRef.current) {
        clearTimeout(reconnectRef.current)
        reconnectRef.current = null
      }
      if (socket && socket.readyState === WebSocket.OPEN) {
        try {
          socket.send(JSON.stringify({ type: 'leave', room_id: roomId, user_id: userId }))
        } catch {}
        socket.close(1000)
      }
      socket = null
      wsRef.current = null
    }

    const handleMessage = (data: string) => {
      let msg: WebSocketMessage
      try {
        msg = JSON.parse(data)
      } catch {
        return
      }

      const store = useGameStore.getState()

      switch (msg.type) {
        case 'ping':
          logger.info('Received ping from server')
          if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({ type: 'pong' }))
            logger.info('Sent pong response')
          } else {
            logger.warn('Cannot send pong - socket not open')
          }
          break
        case 'pong':
          logger.info('Received pong from server')
          break
        case 'joined':
          logger.info('Joined room successfully')
          break
        case 'playerJoined':
          store.addChatMessage({ user: 'System', message: `👋 ${msg.username || 'Player'} joined`, timestamp: Date.now() })
          break
        case 'leave':
          if (msg.user_id) {
            store.removePlayer(msg.user_id)
            store.addChatMessage({ user: 'System', message: `${msg.username} left`, timestamp: Date.now() })
          }
          break
        case 'gameState':
          if (msg.payload) {
            const { players, communityCards, pot, currentBet, phase, currentPlayer, pots } = msg.payload
            if (players) store.setPlayers(players)
            if (communityCards) store.setCommunityCards(cardsToStrings(communityCards as BackendCard[]))
            store.setPot(pots?.reduce((s: number, p: any) => s + (p.amount || 0), 0) ?? pot ?? 0)
            store.setCurrentBet(currentBet || 0)
            store.setPhase(phase || 'waiting')
            store.setMyTurn(currentPlayer === userId)
          }
          break
        case 'deal':
          if (msg.payload) {
            const { holeCards, playerId, communityCards } = msg.payload
            if (playerId === userId && holeCards) store.setHoleCards(cardsToStrings(holeCards as BackendCard[]))
            if (communityCards) store.setCommunityCards(cardsToStrings(communityCards as BackendCard[]))
          }
          break
        case 'player_action':
        case 'playerAction':
          if (msg.payload) {
            const { playerId, player_id, action, amount, newPot, pot, game, newBet, current_bet } = msg.payload
            store.updatePlayer({ id: playerId || player_id, lastAction: action, bet: amount })
            const potVal = newPot ?? pot ?? game?.pots?.reduce((s: number, p: any) => s + (p.amount || 0), 0)
            if (potVal !== undefined) store.setPot(potVal)
            const betVal = newBet ?? current_bet ?? game?.current_bet
            if (betVal !== undefined) store.setCurrentBet(betVal)
            store.addChatMessage({ user: msg.username || 'Player', message: `${action}${amount ? ` ${amount}` : ''}`, timestamp: Date.now() })
          }
          break
        case 'phaseChange':
        case 'phase_change':
          if (msg.payload) {
            const { phase, communityCards, community_cards, pot, pots } = msg.payload
            store.setPhase(phase)
            const cards = communityCards || community_cards
            if (cards) store.setCommunityCards(cardsToStrings(cards as BackendCard[]))
            const potVal = pots?.reduce((s: number, p: any) => s + (p.amount || 0), 0) ?? pot
            if (potVal !== undefined) store.setPot(potVal)
          }
          break
        case 'winner':
          if (msg.payload) {
            const { winnerId, winnerName, amount, hand } = msg.payload
            store.setWinner({ id: winnerId, name: winnerName, amount, hand })
            store.addChatMessage({ user: 'System', message: `🏆 ${winnerName} wins ${amount}!`, timestamp: Date.now() })
          }
          break
        case 'turn':
          if (msg.payload) store.setMyTurn(msg.payload.playerId === userId)
          break
        case 'chat':
          if (msg.payload) store.addChatMessage({ user: msg.username || 'Anon', message: msg.payload.message, timestamp: Date.now() })
          break
        case 'error':
          store.addChatMessage({ user: 'System', message: `⚠️ ${msg.payload?.message || 'Error'}`, timestamp: Date.now() })
          break
        case 'gameStarting':
          store.addChatMessage({ user: 'System', message: `🎮 Game starting in ${msg.data?.countdown || msg.payload?.countdown}...`, timestamp: Date.now() })
          break
        case 'showdown':
          if (msg.payload?.players) store.setPlayers(msg.payload.players)
          break
      }
    }

    const connect = () => {
      if (isCleaningUp || !mountedRef.current) return
      if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) return

      const wsUrl = getWebSocketUrl()
      const url = `${wsUrl}/ws?user_id=${userId}&username=${encodeURIComponent(username)}&room_id=${roomId}`
      
      logger.info(`Connecting to ${url}`)
      
      socket = new WebSocket(url)
      wsRef.current = socket

      socket.onopen = () => {
        if (isCleaningUp || !mountedRef.current) {
          socket?.close()
          return
        }
        
        reconnectCountRef.current = 0
        connectedParamsRef.current = `${userId}-${username}-${roomId}`
        setIsConnected(true)
        setIsReconnecting(false)
        logger.ws.connected()
        logger.info('WebSocket onopen fired')
        onConnect?.()

        // Send join
        socket?.send(JSON.stringify({ type: 'join', room_id: roomId, user_id: userId, username }))
        logger.info('Sent join message')

        // Clear any existing heartbeat
        if (heartbeatRef.current) {
          clearInterval(heartbeatRef.current)
        }
        
        // Heartbeat every 25s
        logger.info('Starting heartbeat interval')
        heartbeatRef.current = setInterval(() => {
          if (socket?.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: 'ping' }))
            logger.info('Sent ping')
          } else {
            logger.warn('Socket not open, cannot send ping')
          }
        }, 25000)
      }

      socket.onmessage = (e) => handleMessage(e.data)

      socket.onerror = (e) => {
        logger.ws.error(e)
        onError?.(e)
      }

      socket.onclose = (e) => {
        if (heartbeatRef.current) {
          clearInterval(heartbeatRef.current)
          heartbeatRef.current = null
        }
        
        wsRef.current = null
        socket = null
        
        if (!mountedRef.current || isCleaningUp) return
        
        setIsConnected(false)
        logger.ws.disconnected()
        logger.info(`Close code: ${e.code}`)
        onDisconnect?.()

        // Reconnect unless clean close
        if (e.code !== 1000 && reconnectCountRef.current < 5) {
          reconnectCountRef.current++
          setIsReconnecting(true)
          const delay = Math.min(1000 * Math.pow(2, reconnectCountRef.current - 1), 10000)
          logger.info(`Reconnecting in ${delay}ms`)
          reconnectRef.current = setTimeout(connect, delay)
        }
      }
    }

    connect()

    return () => {
      mountedRef.current = false
      connectedParamsRef.current = ''
      cleanup()
      setIsConnected(false)
      setIsReconnecting(false)
    }
  }, [enabled, userId, username, roomId]) // Only these 4 deps - callbacks are called via closure

  const sendAction = (action: string, amount?: number) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'action',
        room_id: roomId,
        user_id: userId,
        payload: { action, amount: amount || 0 }
      }))
    }
  }

  const sendChat = (message: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'chat',
        room_id: roomId,
        user_id: userId,
        username,
        payload: { message }
      }))
    }
  }

  const disconnect = () => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.close(1000)
    }
  }

  const reconnect = () => {
    disconnect()
    reconnectCountRef.current = 0
    // The useEffect will handle reconnection when socket closes
  }

  return {
    isConnected,
    isReconnecting,
    sendAction,
    sendChat,
    disconnect,
    reconnect
  }
}
