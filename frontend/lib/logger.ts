/**
 * Frontend logger utility
 * Provides structured logging with different levels
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

class Logger {
  private isDevelopment = process.env.NODE_ENV === 'development'
  private minLevel: LogLevel = this.isDevelopment ? 'debug' : 'info'

  private levels: Record<LogLevel, number> = {
    debug: 0,
    info: 1,
    warn: 2,
    error: 3,
  }

  private shouldLog(level: LogLevel): boolean {
    return this.levels[level] >= this.levels[this.minLevel]
  }

  private formatMessage(level: LogLevel, message: string, data?: any): string {
    const timestamp = new Date().toISOString()
    const prefix = `[${timestamp}] [${level.toUpperCase()}]`
    
    if (data) {
      return `${prefix} ${message} ${JSON.stringify(data)}`
    }
    return `${prefix} ${message}`
  }

  debug(message: string, data?: any) {
    if (this.shouldLog('debug')) {
      console.debug(this.formatMessage('debug', message, data))
    }
  }

  info(message: string, data?: any) {
    if (this.shouldLog('info')) {
      console.info(this.formatMessage('info', message, data))
    }
  }

  warn(message: string, data?: any) {
    if (this.shouldLog('warn')) {
      console.warn(this.formatMessage('warn', message, data))
    }
  }

  error(message: string, error?: any) {
    if (this.shouldLog('error')) {
      console.error(this.formatMessage('error', message, error))
      
      // In production, you could send errors to a service like Sentry
      if (!this.isDevelopment && typeof window !== 'undefined') {
        // Example: window.Sentry?.captureException(error)
      }
    }
  }

  // WebSocket specific logging
  ws = {
    connected: () => this.info('WebSocket connected'),
    disconnected: () => this.warn('WebSocket disconnected'),
    reconnecting: (attempt: number) => this.info(`WebSocket reconnecting (attempt ${attempt})`),
    error: (error: any) => this.error('WebSocket error', error),
    message: (type: string, payload?: any) => this.debug(`WebSocket message: ${type}`, payload),
  }

  // API specific logging
  api = {
    request: (method: string, url: string) => this.debug(`API ${method} ${url}`),
    response: (method: string, url: string, status: number) => 
      this.debug(`API ${method} ${url} - ${status}`),
    error: (method: string, url: string, error: any) => 
      this.error(`API ${method} ${url} failed`, error),
  }
}

export const logger = new Logger()
