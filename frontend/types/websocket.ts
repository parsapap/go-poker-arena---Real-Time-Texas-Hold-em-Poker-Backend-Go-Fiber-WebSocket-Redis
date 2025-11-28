// Type definitions for WebSocket messages

import { BackendCard } from '@/lib/cardUtils'

export interface WebSocketMessage {
  type: string
  room_id?: string
  user_id?: number
  username?: string
  payload?: any
  data?: Record<string, any>
}

export interface JoinMessage extends WebSocketMessage {
  type: 'join'
  payload: {
    id: number
    username: string
    chips: number
    position: number
  }
}

export interface LeaveMessage extends WebSocketMessage {
  type: 'leave'
  user_id: number
  username: string
}

export interface GameStateMessage extends WebSocketMessage {
  type: 'gameState'
  payload: {
    players: Player[]
    communityCards: BackendCard[]
    pot: number
    pots: Pot[]
    currentBet: number
    phase: GamePhase
    currentPlayer: number
  }
}

export interface DealMessage extends WebSocketMessage {
  type: 'deal'
  payload: {
    holeCards?: BackendCard[]
    playerId?: number
    communityCards?: BackendCard[]
  }
}

export interface PlayerActionMessage extends WebSocketMessage {
  type: 'player_action' | 'playerAction'
  payload: {
    playerId?: number
    player_id?: number
    action: string
    amount?: number
    newPot?: number
    newBet?: number
    pot?: number
    current_bet?: number
    game?: {
      pots: Pot[]
      current_bet: number
    }
  }
}

export interface PhaseChangeMessage extends WebSocketMessage {
  type: 'phaseChange' | 'phase_change'
  payload: {
    phase: GamePhase
    communityCards?: BackendCard[]
    community_cards?: BackendCard[]
    pot?: number
    pots?: Pot[]
  }
}

export interface ShowdownMessage extends WebSocketMessage {
  type: 'showdown'
  payload: {
    players: Player[]
  }
}

export interface WinnerMessage extends WebSocketMessage {
  type: 'winner'
  payload: {
    winnerId: number
    winnerName: string
    amount: number
    hand: string
  }
}

export interface TurnMessage extends WebSocketMessage {
  type: 'turn'
  payload: {
    playerId: number
    timeLeft?: number
  }
}

export interface ChatMessage extends WebSocketMessage {
  type: 'chat'
  payload: {
    message: string
  }
}

export interface ErrorMessage extends WebSocketMessage {
  type: 'error'
  payload: {
    message: string
  }
}

// Game types
export type GamePhase = 'waiting' | 'preflop' | 'flop' | 'turn' | 'river' | 'showdown' | 'finished'

export interface Player {
  id: number
  username: string
  chips: number
  bet: number
  position: number
  cards?: string[]
  hole_cards?: BackendCard[]
  isFolded?: boolean
  folded?: boolean
  isActive?: boolean
  active?: boolean
  lastAction?: string
}

export interface Pot {
  amount: number
  players: number[]
}

// Union type of all possible messages
export type WSMessage =
  | JoinMessage
  | LeaveMessage
  | GameStateMessage
  | DealMessage
  | PlayerActionMessage
  | PhaseChangeMessage
  | ShowdownMessage
  | WinnerMessage
  | TurnMessage
  | ChatMessage
  | ErrorMessage
  | WebSocketMessage
