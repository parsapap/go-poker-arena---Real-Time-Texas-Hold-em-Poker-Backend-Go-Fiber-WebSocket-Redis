// Card utility functions for converting between backend and frontend formats

export interface BackendCard {
  suit: number | string
  rank: number | string
}

const suitMap: Record<number, string> = {
  0: '♥️',
  1: '♦️',
  2: '♣️',
  3: '♠️',
}

const suitNameMap: Record<string, string> = {
  'hearts': '♥️',
  'diamonds': '♦️',
  'clubs': '♣️',
  'spades': '♠️',
}

const rankMap: Record<number, string> = {
  0: '2',
  1: '3',
  2: '4',
  3: '5',
  4: '6',
  5: '7',
  6: '8',
  7: '9',
  8: '10',
  9: 'J',
  10: 'Q',
  11: 'K',
  12: 'A',
}

const rankNameMap: Record<string, string> = {
  '2': '2', '3': '3', '4': '4', '5': '5', '6': '6',
  '7': '7', '8': '8', '9': '9', '10': '10',
  'J': 'J', 'Q': 'Q', 'K': 'K', 'A': 'A',
}

/**
 * Convert backend card format to frontend display string
 * Backend: { suit: 0, rank: 12 } or { suit: "hearts", rank: "A" }
 * Frontend: "A♥️"
 */
export function cardToString(card: BackendCard): string {
  let suit: string
  let rank: string

  // Handle suit
  if (typeof card.suit === 'number') {
    suit = suitMap[card.suit] || '?'
  } else {
    suit = suitNameMap[card.suit.toLowerCase()] || '?'
  }

  // Handle rank
  if (typeof card.rank === 'number') {
    rank = rankMap[card.rank] || '?'
  } else {
    rank = rankNameMap[card.rank] || card.rank
  }

  return `${rank}${suit}`
}

/**
 * Convert array of backend cards to frontend strings
 */
export function cardsToStrings(cards: BackendCard[]): string[] {
  return cards.map(cardToString)
}

/**
 * Get card color for styling
 */
export function getCardColor(card: string): 'red' | 'black' {
  if (card.includes('♥️') || card.includes('♦️')) {
    return 'red'
  }
  return 'black'
}
