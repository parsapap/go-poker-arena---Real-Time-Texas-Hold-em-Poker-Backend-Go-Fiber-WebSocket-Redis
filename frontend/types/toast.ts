// Toast notification types

export type ToastType = 'success' | 'error' | 'info' | 'warning' | 'win'

export interface Toast {
  id: string
  type: ToastType
  message: string
  duration?: number
  timestamp: number
}

export interface ToastOptions {
  duration?: number
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center'
}
