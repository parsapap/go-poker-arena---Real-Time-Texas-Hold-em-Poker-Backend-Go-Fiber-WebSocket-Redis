import { create } from 'zustand'

export interface Player {
  id: number
  username: string
  chips: number
  bet: number
  cards?: string[]
  isFolded: boolean
  isActive: boolean
  position: number
  lastAction?: string
}

export interface ChatMessage {
  user: string
  message: string
  timestamp: number
}

export interface Winner {
  id: number
  name: string
  amount: number
  hand: string
}

interface GameState {
  // Game state
  roomId: string
  players: Player[]
  communityCards: string[]
  holeCards: string[]
  phase: 'waiting' | 'preflop' | 'flop' | 'turn' | 'river' | 'showdown'
  pot: number
  currentBet: number
  myTurn: boolean
  winner: Winner | null
  
  // Chat
  chatMessages: ChatMessage[]
  typingUsers: string[]
  
  // UI state
  soundEnabled: boolean
  
  // Actions
  setRoomId: (roomId: string) => void
  setPlayers: (players: Player[]) => void
  updatePlayer: (player: Partial<Player> & { id: number }) => void
  removePlayer: (playerId: number) => void
  setCommunityCards: (cards: string[]) => void
  setHoleCards: (cards: string[]) => void
  setPhase: (phase: GameState['phase']) => void
  setPot: (pot: number) => void
  setCurrentBet: (bet: number) => void
  setMyTurn: (isMyTurn: boolean) => void
  setWinner: (winner: Winner | null) => void
  addChatMessage: (message: ChatMessage) => void
  clearChat: () => void
  setTypingUser: (username: string, isTyping: boolean) => void
  toggleSound: () => void
  resetGame: () => void
}

export const useGameStore = create<GameState>((set) => ({
  // Initial state
  roomId: '',
  players: [],
  communityCards: [],
  holeCards: [],
  phase: 'waiting',
  pot: 0,
  currentBet: 0,
  myTurn: false,
  winner: null,
  chatMessages: [],
  typingUsers: [],
  soundEnabled: true,

  // Actions
  setRoomId: (roomId) => set({ roomId }),

  setPlayers: (players) => set({ players }),

  updatePlayer: (updatedPlayer) =>
    set((state) => ({
      players: state.players.map((player) =>
        player.id === updatedPlayer.id
          ? { ...player, ...updatedPlayer }
          : player
      )
    })),

  removePlayer: (playerId) =>
    set((state) => ({
      players: state.players.filter((player) => player.id !== playerId)
    })),

  setCommunityCards: (cards) => set({ communityCards: cards }),

  setHoleCards: (cards) => set({ holeCards: cards }),

  setPhase: (phase) => set({ phase }),

  setPot: (pot) => set({ pot }),

  setCurrentBet: (bet) => set({ currentBet: bet }),

  setMyTurn: (isMyTurn) => set({ myTurn: isMyTurn }),

  setWinner: (winner) => set({ winner }),

  addChatMessage: (message) =>
    set((state) => ({
      chatMessages: [...state.chatMessages, message].slice(-100) // Keep last 100 messages
    })),

  clearChat: () => set({ chatMessages: [] }),

  setTypingUser: (username, isTyping) =>
    set((state) => ({
      typingUsers: isTyping
        ? [...state.typingUsers, username]
        : state.typingUsers.filter((u) => u !== username)
    })),

  toggleSound: () => set((state) => ({ soundEnabled: !state.soundEnabled })),

  resetGame: () =>
    set({
      communityCards: [],
      holeCards: [],
      phase: 'waiting',
      pot: 0,
      currentBet: 0,
      myTurn: false,
      winner: null
    })
}))
