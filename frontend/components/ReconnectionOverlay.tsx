import { motion } from 'framer-motion'
import { Wifi, WifiOff, Loader2 } from 'lucide-react'

interface ReconnectionOverlayProps {
  isReconnecting: boolean
  onRetry?: () => void
}

export function ReconnectionOverlay({ isReconnecting, onRetry }: ReconnectionOverlayProps) {
  if (!isReconnecting) return null

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center"
    >
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        className="glass rounded-2xl p-8 max-w-md mx-4 text-center"
      >
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
          className="w-16 h-16 mx-auto mb-4"
        >
          <WifiOff className="w-full h-full text-amber-400" />
        </motion.div>

        <h3 className="text-2xl font-bold mb-2">Connection Lost</h3>
        <p className="text-white/60 mb-6">
          Attempting to reconnect to the game...
        </p>

        <div className="flex items-center justify-center gap-2 mb-6">
          <motion.div
            animate={{ scale: [1, 1.2, 1] }}
            transition={{ duration: 1, repeat: Infinity }}
            className="w-2 h-2 bg-amber-400 rounded-full"
          />
          <motion.div
            animate={{ scale: [1, 1.2, 1] }}
            transition={{ duration: 1, repeat: Infinity, delay: 0.2 }}
            className="w-2 h-2 bg-amber-400 rounded-full"
          />
          <motion.div
            animate={{ scale: [1, 1.2, 1] }}
            transition={{ duration: 1, repeat: Infinity, delay: 0.4 }}
            className="w-2 h-2 bg-amber-400 rounded-full"
          />
        </div>

        {onRetry && (
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={onRetry}
            className="px-6 py-3 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-lg font-bold hover:from-emerald-400 hover:to-teal-400 transition-all"
          >
            Retry Connection
          </motion.button>
        )}
      </motion.div>
    </motion.div>
  )
}

interface ConnectionStatusProps {
  isConnected: boolean
}

export function ConnectionStatus({ isConnected }: ConnectionStatusProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      className="fixed top-20 left-1/2 -translate-x-1/2 z-40"
    >
      <div
        className={`glass px-4 py-2 rounded-full flex items-center gap-2 ${
          isConnected ? 'border-emerald-500/50' : 'border-red-500/50'
        } border-2`}
      >
        {isConnected ? (
          <>
            <motion.div
              animate={{ scale: [1, 1.2, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
            >
              <Wifi className="w-4 h-4 text-emerald-400" />
            </motion.div>
            <span className="text-sm text-emerald-400 font-medium">Connected</span>
          </>
        ) : (
          <>
            <WifiOff className="w-4 h-4 text-red-400" />
            <span className="text-sm text-red-400 font-medium">Disconnected</span>
          </>
        )}
      </div>
    </motion.div>
  )
}
