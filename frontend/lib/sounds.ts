// Sound effects manager using Web Audio API
// No external dependencies needed

class SoundManager {
  private audioContext: AudioContext | null = null
  private sounds: Map<string, AudioBuffer> = new Map()
  private enabled: boolean = true

  constructor() {
    if (typeof window !== 'undefined') {
      this.audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
    }
  }

  setEnabled(enabled: boolean) {
    this.enabled = enabled
  }

  // Generate card deal sound (short click)
  private generateCardDeal(): AudioBuffer {
    if (!this.audioContext) return null as any
    
    const sampleRate = this.audioContext.sampleRate
    const duration = 0.05 // 50ms
    const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate)
    const data = buffer.getChannelData(0)

    for (let i = 0; i < buffer.length; i++) {
      const t = i / sampleRate
      // Short click sound
      data[i] = Math.sin(2 * Math.PI * 800 * t) * Math.exp(-t * 50)
    }

    return buffer
  }

  // Generate chip stack sound (multiple clicks)
  private generateChipStack(): AudioBuffer {
    if (!this.audioContext) return null as any
    
    const sampleRate = this.audioContext.sampleRate
    const duration = 0.3
    const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate)
    const data = buffer.getChannelData(0)

    for (let i = 0; i < buffer.length; i++) {
      const t = i / sampleRate
      // Multiple clicks with decay
      const clicks = Math.sin(2 * Math.PI * 600 * t) * Math.exp(-t * 8)
      const rattle = Math.sin(2 * Math.PI * 400 * t * (1 + Math.sin(t * 30))) * Math.exp(-t * 10)
      data[i] = (clicks + rattle * 0.3) * 0.5
    }

    return buffer
  }

  // Generate win fanfare (ascending notes)
  private generateWinFanfare(): AudioBuffer {
    if (!this.audioContext) return null as any
    
    const sampleRate = this.audioContext.sampleRate
    const duration = 1.5
    const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate)
    const data = buffer.getChannelData(0)

    const notes = [523.25, 659.25, 783.99, 1046.50] // C5, E5, G5, C6

    for (let i = 0; i < buffer.length; i++) {
      const t = i / sampleRate
      let sample = 0

      notes.forEach((freq, idx) => {
        const noteStart = idx * 0.25
        const noteEnd = noteStart + 0.4
        if (t >= noteStart && t < noteEnd) {
          const noteT = t - noteStart
          sample += Math.sin(2 * Math.PI * freq * noteT) * Math.exp(-noteT * 3)
        }
      })

      data[i] = sample * 0.3
    }

    return buffer
  }

  // Generate button click sound
  private generateButtonClick(): AudioBuffer {
    if (!this.audioContext) return null as any
    
    const sampleRate = this.audioContext.sampleRate
    const duration = 0.08
    const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate)
    const data = buffer.getChannelData(0)

    for (let i = 0; i < buffer.length; i++) {
      const t = i / sampleRate
      data[i] = Math.sin(2 * Math.PI * 1000 * t) * Math.exp(-t * 40)
    }

    return buffer
  }

  // Initialize all sounds
  init() {
    if (!this.audioContext) return

    this.sounds.set('cardDeal', this.generateCardDeal())
    this.sounds.set('chipStack', this.generateChipStack())
    this.sounds.set('winFanfare', this.generateWinFanfare())
    this.sounds.set('buttonClick', this.generateButtonClick())
  }

  // Play a sound
  play(soundName: string, volume: number = 1.0) {
    if (!this.enabled || !this.audioContext) return

    const buffer = this.sounds.get(soundName)
    if (!buffer) {
      console.warn(`Sound ${soundName} not found`)
      return
    }

    const source = this.audioContext.createBufferSource()
    const gainNode = this.audioContext.createGain()
    
    source.buffer = buffer
    gainNode.gain.value = volume
    
    source.connect(gainNode)
    gainNode.connect(this.audioContext.destination)
    
    source.start(0)
  }

  // Play card deal sound
  playCardDeal() {
    this.play('cardDeal', 0.3)
  }

  // Play chip stack sound
  playChipStack() {
    this.play('chipStack', 0.4)
  }

  // Play win fanfare
  playWinFanfare() {
    this.play('winFanfare', 0.5)
  }

  // Play button click
  playButtonClick() {
    this.play('buttonClick', 0.2)
  }
}

// Singleton instance
export const soundManager = new SoundManager()

// Initialize sounds on first user interaction
if (typeof window !== 'undefined') {
  const initSounds = () => {
    soundManager.init()
    document.removeEventListener('click', initSounds)
    document.removeEventListener('touchstart', initSounds)
  }
  
  document.addEventListener('click', initSounds, { once: true })
  document.addEventListener('touchstart', initSounds, { once: true })
}
