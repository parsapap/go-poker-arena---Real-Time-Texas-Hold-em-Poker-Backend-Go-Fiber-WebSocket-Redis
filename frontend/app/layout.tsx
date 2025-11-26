import type { Metadata } from 'next'
import { IBM_Plex_Mono } from 'next/font/google'
import './globals.css'

const ibmPlexMono = IBM_Plex_Mono({ 
  weight: ['400', '500', '600', '700'],
  subsets: ['latin'],
  variable: '--font-geist-mono',
})

export const metadata: Metadata = {
  title: 'Poker Arena - Real-Time Texas Hold\'em',
  description: 'Play Texas Hold\'em poker online with real-time multiplayer gameplay',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <body className={`${ibmPlexMono.variable} font-mono bg-black text-white antialiased`}>
        {children}
      </body>
    </html>
  )
}
