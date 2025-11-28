/**
 * Get WebSocket URL based on environment
 * Handles both development and production scenarios
 */
export function getWebSocketUrl(): string {
  // In browser
  if (typeof window !== 'undefined') {
    // 1. Use environment variable if explicitly set
    if (process.env.NEXT_PUBLIC_WS_URL) {
      return process.env.NEXT_PUBLIC_WS_URL
    }

    // 2. Auto-detect based on current location (production)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    
    // In production, use same host as frontend (no port, goes through proxy)
    // In development, connect directly to backend on port 8080
    if (process.env.NODE_ENV === 'production') {
      // Production: use wss://current-domain/ws (no port)
      return `${protocol}//${host}`
    } else {
      // Development: connect directly to backend
      return `${protocol}//${host}:8080`
    }
  }

  // Fallback for SSR
  return 'ws://localhost:8080'
}

/**
 * Get API URL based on environment
 */
export function getApiUrl(): string {
  if (typeof window !== 'undefined') {
    // Use proxy in browser (Next.js rewrites)
    return ''
  }
  
  // Server-side: use environment variable
  return process.env.API_URL || 'http://localhost:8080'
}
