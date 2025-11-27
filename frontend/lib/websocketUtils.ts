/**
 * Get WebSocket URL based on environment
 * Handles both development and production scenarios
 */
export function getWebSocketUrl(): string {
  // In browser
  if (typeof window !== 'undefined') {
    // Use environment variable if set
    if (process.env.NEXT_PUBLIC_WS_URL) {
      return process.env.NEXT_PUBLIC_WS_URL
    }

    // Auto-detect based on current location
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    
    // In Docker, backend is on port 8080
    // In development, backend is on port 8080
    const port = process.env.NODE_ENV === 'production' ? window.location.port : '8080'
    
    return `${protocol}//${host}:${port}`
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
