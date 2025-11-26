import { motion } from 'framer-motion'
import { useEffect, useState } from 'react'

interface Chip {
  id: number
  x: number
  delay: number
  duration: number
  rotation: number
  color: string
}

export function ChipRain() {
  const [chips, setChips] = useState<Chip[]>([])

  useEffect(() => {
    const colors = ['#fbbf24', '#f59e0b', '#d97706', '#b45309']
    const newChips: Chip[] = []

    for (let i = 0; i < 50; i++) {
      newChips.push({
        id: i,
        x: Math.random() * 100,
        delay: Math.random() * 2,
        duration: 2 + Math.random() * 2,
        rotation: Math.random() * 720 - 360,
        color: colors[Math.floor(Math.random() * colors.length)]
      })
    }

    setChips(newChips)
  }, [])

  return (
    <div className="fixed inset-0 pointer-events-none z-40 overflow-hidden">
      {chips.map((chip) => (
        <motion.div
          key={chip.id}
          initial={{ y: -100, x: `${chip.x}vw`, opacity: 1, rotate: 0 }}
          animate={{
            y: '110vh',
            rotate: chip.rotation,
            opacity: [1, 1, 0.5, 0]
          }}
          transition={{
            duration: chip.duration,
            delay: chip.delay,
            ease: 'easeIn'
          }}
          className="absolute"
        >
          <div
            className="w-8 h-8 rounded-full shadow-lg"
            style={{
              background: `radial-gradient(circle at 30% 30%, ${chip.color}, ${chip.color}dd)`,
              boxShadow: `0 4px 8px rgba(0,0,0,0.3), inset 0 2px 4px rgba(255,255,255,0.3)`
            }}
          >
            <div className="w-full h-full rounded-full border-4 border-white/20 flex items-center justify-center">
              <span className="text-xs font-bold text-white/80">$</span>
            </div>
          </div>
        </motion.div>
      ))}
    </div>
  )
}
