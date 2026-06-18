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
  
  // Use ref for userId to avoid stale closure issues
  const userIdRef = useRef(userId)
  userIdRef.current = userId

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
          // Game state can be in msg.payload, msg.game, or msg.data
          const gameData = msg.payload || (msg as any).game || msg.data
          if (gameData) {
            const { players, communityCards, community_cards, pot, currentBet, current_bet, phase, currentPlayer, current_position, pots } = gameData
            logger.info(`gameState: phase=${phase}, current_position=${current_position}, players=${players?.length}`)
            if (players) {
              store.setPlayers(players)
              logger.info(`gameState players: ${players.map((p: any) => `${p.id}:${p.username}`).join(', ')}`)
            }
            const cards = communityCards || community_cards
            if (cards) store.setCommunityCards(cardsToStrings(cards as BackendCard[]))
            store.setPot(pots?.reduce((s: number, p: any) => s + (p.amount || 0), 0) ?? pot ?? 0)
            store.setCurrentBet(currentBet || current_bet || 0)
            store.setPhase(phase || 'waiting')
            // current_position is index, need to check if it matches user
            const currentPos = currentPlayer ?? current_position
            if (players && currentPos !== undefined) {
              const currentPlayerObj = players[currentPos]
              const currentUserId = userIdRef.current
              logger.info(`gameState turn: currentPos=${currentPos}, currentPlayer=${currentPlayerObj?.id}:${currentPlayerObj?.username}, userId=${currentUserId}, isMyTurn=${currentPlayerObj?.id === currentUserId}`)
              store.setMyTurn(currentPlayerObj?.id === currentUserId)
            }
          }
          break
        case 'deal':
          console.log('[DEAL] Raw message:', JSON.stringify(msg))
          // Deal data can be in payload or at root level
          const dealData = msg.payload || msg.data || msg
          console.log('[DEAL] dealData:', JSON.stringify(dealData))
          if (dealData) {
            const { holeCards, hole_cards, playerId, player_id, communityCards, community_cards, phase } = dealData as any
            const dealPlayerId = playerId || player_id
            const dealHoleCards = holeCards || hole_cards
            const currentUserId = userIdRef.current
            console.log(`[DEAL] player_id=${dealPlayerId}, userId=${currentUserId}, match=${dealPlayerId == currentUserId}`)
            console.log(`[DEAL] holeCards:`, dealHoleCards)
            // Use == for loose comparison in case of type mismatch
            if (dealPlayerId == currentUserId && dealHoleCards && dealHoleCards.length > 0) {
              const convertedCards = cardsToStrings(dealHoleCards as BackendCard[])
              console.log(`[DEAL] Setting hole cards:`, convertedCards)
              store.setHoleCards(convertedCards)
            }
            const cards = communityCards || community_cards
            if (cards) store.setCommunityCards(cardsToStrings(cards as BackendCard[]))
            if (phase) store.setPhase(phase)
          }
          break
        case 'player_action':
        case 'playerAction':
          // Data can be in payload or at root level
          const actionData = msg.payload || msg
          const actionPlayerId = (actionData as any).playerId || (actionData as any).player_id || (msg as any).player_id
          const actionType = (actionData as any).action || (msg as any).action
          const actionAmount = (actionData as any).amount || (msg as any).amount || 0
          const actionGame = (actionData as any).game || (msg as any).game
          
          if (actionPlayerId) {
            store.updatePlayer({ id: actionPlayerId, lastAction: actionType, bet: actionAmount })
          }
          
          // Get username from game players or message
          let actionUsername = msg.username
          if (!actionUsername && actionGame?.players) {
            const player = actionGame.players.find((p: any) => p.id === actionPlayerId)
            actionUsername = player?.username
          }
          
          // Update game state from the game object
          if (actionGame) {
            const { pots, current_bet, current_position, players, community_cards, phase } = actionGame
            if (pots) store.setPot(pots.reduce((s: number, p: any) => s + (p.amount || 0), 0))
            if (current_bet !== undefined) store.setCurrentBet(current_bet)
            // Update community cards if phase changed
            if (community_cards && community_cards.length > 0) {
              store.setCommunityCards(cardsToStrings(community_cards as BackendCard[]))
            }
            if (phase) store.setPhase(phase)
            // Update whose turn it is
            if (players && current_position !== undefined) {
              const currentPlayer = players[current_position]
              store.setMyTurn(currentPlayer?.id === userIdRef.current)
            }
          }
          
          if (actionType) {
            store.addChatMessage({ user: actionUsername || 'Player', message: `${actionType}${actionAmount ? ` $${actionAmount}` : ''}`, timestamp: Date.now() })
          }
          break
        case 'phaseChange':
        case 'phase_change':
          // Phase data can be in payload or at root level
          const phaseData = msg.payload || msg.data || msg
          if (phaseData) {
            const phase = (phaseData as any).phase || (msg as any).phase
            if (phase) store.setPhase(phase)
            const { communityCards, community_cards, pot, pots } = phaseData as any
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
          if (msg.payload) store.setMyTurn(msg.payload.playerId === userIdRef.current)
          break
        case 'chat':
          if (msg.payload) store.addChatMessage({ user: msg.username || 'Anon', message: msg.payload.message, timestamp: Date.now() })
          break
        case 'error':
          const errorMsg = msg.payload?.message || msg.data?.message || (msg as any).message || 'Error'
          store.addChatMessage({ user: 'System', message: `⚠️ ${errorMsg}`, timestamp: Date.now() })
          break
        case 'gameStarting':
          // countdown can be at root level or in data/payload
          logger.info('gameStarting message:', JSON.stringify(msg))
          const countdown = (msg as any).countdown || msg.data?.countdown || msg.payload?.countdown
          store.addChatMessage({ user: 'System', message: `🎮 Game starting in ${countdown}...`, timestamp: Date.now() })
          break
        case 'potUpdate':
          const potUpdateData = msg.payload || msg.data || msg
          if (potUpdateData) {
            const { pot, current_bet, currentBet } = potUpdateData as any
            if (pot !== undefined) store.setPot(pot)
            if (current_bet !== undefined || currentBet !== undefined) store.setCurrentBet(current_bet || currentBet)
          }
          break
        case 'playerTurn':
          const turnData = msg.payload || msg.data || msg
          if (turnData) {
            const turnPlayerId = (turnData as any).player_id || (turnData as any).playerId
            const currentUserId = userIdRef.current
            logger.info(`playerTurn: player_id=${turnPlayerId}, userId=${currentUserId}, isMyTurn=${turnPlayerId === currentUserId}`)
            store.setMyTurn(turnPlayerId === currentUserId)
          }
          break
        case 'showdown':
          // Showdown data can be at root level or in payload
          const showdownData = msg.payload || msg.data || msg
          console.log('[SHOWDOWN] Raw message:', JSON.stringify(msg))
          if (showdownData) {
            const players = (showdownData as any).players
            const communityCards = (showdownData as any).community_cards
            if (players) {
              store.setPlayers(players)
              console.log('[SHOWDOWN] Updated players with hole cards')
            }
            if (communityCards) {
              store.setCommunityCards(cardsToStrings(communityCards as BackendCard[]))
            }
            store.setPhase('showdown')
          }
          break
        case 'game_end':
          // Handle game end with winners array
          const gameEndData = msg.payload || msg.data || msg
          console.log('[GAME_END] Raw message:', JSON.stringify(msg))
          const winners = (gameEndData as any).winners
          console.log('[GAME_END] Winners:', winners)
          if (winners && winners.length > 0) {
            // Find the actual winner (player with best hand)
            const winner = winners[0]
            const winAmount = winner.amount || store.pot
            console.log('[GAME_END] Winner:', winner.username, 'Hand:', winner.hand, 'Amount:', winAmount)
            store.setWinner({ 
              id: winner.player_id, 
              name: winner.username, 
              amount: winAmount,
              hand: winner.hand 
            })
            store.addChatMessage({ 
              user: 'System', 
              message: `🏆 ${winner.username} wins $${winAmount} with ${winner.hand}!`, 
              timestamp: Date.now() 
            })
            store.setPhase('finished')
          }
          break
      }
    }

    const connect = () => {
      if (isCleaningUp || !mountedRef.current) return
      if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) return

      const wsUrl = getWebSocketUrl()
      // The backend authenticates the WebSocket via JWT (?token=). Identity is
      // derived from the verified token, not user_id/username query params.
      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null
      const url = `${wsUrl}/ws?token=${encodeURIComponent(token || '')}&room_id=${roomId}`

      logger.info(`Connecting to ${wsUrl}/ws (room ${roomId})`)

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
