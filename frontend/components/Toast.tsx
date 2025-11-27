import { motion, AnimatePresence } from 'framer-motion'
import { X, Trophy, AlertCircle, Info } from 'lucide-react'
import { useEffect } from 'react'

export interface ToastProps {
  id: string
  type: 'success' | 'error' | 'info' | 'win'
  message: string
  duration?: number
  onClose: (id: string) => void
}

export function Toast({ id, type, message, duration = 5000, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose(id)
    }, duration)

    return () => clearTimeout(timer)
  }, [id, duration, onClose])

  const icons = {
    success: <Trophy className="w-5 h-5" />,
    error: <AlertCircle className="w-5 h-5" />,
    info: <Info className="w-5 h-5" />,
    win: <Trophy className="w-6 h-6" />
  }

  const colors = {
    success: 'from-emerald-500 to-teal-500',
    error: 'from-red-500 to-rose-500',
    info: 'from-blue-500 to-cyan-500',
    win: 'from-amber-400 to-yellow-500'
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: -50, scale: 0.3 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, scale: 0.5, transition: { duration: 0.2 } }}
      className={`glass rounded-xl p-4 shadow-2xl border-2 ${
        type === 'win' ? 'border-amber-500/50' : 'border-white/20'
      } min-w-[300px] max-w-md`}
    >
      <div className="flex items-start gap-3">
        <motion.div
          animate={type === 'win' ? { rotate: [0, 10, -10, 0] } : {}}
          transition={{ duration: 0.5, repeat: type === 'win' ? 3 : 0 }}
          className={`p-2 rounded-lg bg-gradient-to-br ${colors[type]}`}
        >
          {icons[type]}
        </motion.div>
        
        <div className="flex-1">
          <p className="text-white font-medium">{message}</p>
        </div>

        <button
          onClick={() => onClose(id)}
          className="text-white/60 hover:text-white transition-colors"
        >
          <X className="w-4 h-4" />
        </button>
      </div>

      {/* Progress bar */}
      <motion.div
        initial={{ scaleX: 1 }}
        animate={{ scaleX: 0 }}
        transition={{ duration: duration / 1000, ease: 'linear' }}
        className={`h-1 bg-gradient-to-r ${colors[type]} rounded-full mt-3 origin-left`}
      />
    </motion.div>
  )
}

interface ToastContainerProps {
  toasts: Array<{
    id: string
    type: 'success' | 'error' | 'info' | 'win'
    message: string
    duration?: number
    timestamp?: number
  }>
  onClose: (id: string) => void
}

export function ToastContainer({ toasts, onClose }: ToastContainerProps) {
  return (
    <div className="fixed top-24 right-4 z-50 space-y-3">
      <AnimatePresence>
        {toasts.map((toast) => (
          <Toast key={toast.id} {...toast} onClose={onClose} />
        ))}
      </AnimatePresence>
    </div>
  )
}
