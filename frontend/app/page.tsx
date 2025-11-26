'use client'

import { useState } from 'react'
import Link from 'next/link'

export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-poker-green to-green-900">
      <div className="container mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h1 className="text-6xl font-bold text-white mb-4">
            🃏 Go Poker Arena
          </h1>
          <p className="text-2xl text-green-100 mb-8">
            Real-Time Texas Hold'em Poker
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              href="/play"
              className="bg-poker-gold hover:bg-yellow-600 text-white font-bold py-4 px-8 rounded-lg text-xl transition"
            >
              Play Now
            </Link>
            <Link
              href="/rooms"
              className="bg-white hover:bg-gray-100 text-poker-green font-bold py-4 px-8 rounded-lg text-xl transition"
            >
              Browse Rooms
            </Link>
          </div>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          <FeatureCard
            icon="⚡"
            title="Real-Time Gameplay"
            description="Play with up to 9 players in real-time with WebSocket technology"
          />
          <FeatureCard
            icon="🔒"
            title="Secure & Fair"
            description="Enterprise-grade security with anti-cheat validation"
          />
          <FeatureCard
            icon="🏆"
            title="Leaderboards"
            description="Compete for the top spot on global leaderboards"
          />
        </div>

        <div className="mt-16 text-center">
          <h2 className="text-3xl font-bold text-white mb-8">
            Game Features
          </h2>
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4 max-w-4xl mx-auto">
            <Feature text="Complete Texas Hold'em Rules" />
            <Feature text="Auto-Matchmaking" />
            <Feature text="Game History" />
            <Feature text="Player Statistics" />
            <Feature text="Multiple Tables" />
            <Feature text="Chat System" />
            <Feature text="Mobile Responsive" />
            <Feature text="Free to Play" />
          </div>
        </div>
      </div>
    </main>
  )
}

function FeatureCard({ icon, title, description }: { icon: string; title: string; description: string }) {
  return (
    <div className="bg-white/10 backdrop-blur-sm rounded-lg p-6 text-center">
      <div className="text-5xl mb-4">{icon}</div>
      <h3 className="text-xl font-bold text-white mb-2">{title}</h3>
      <p className="text-green-100">{description}</p>
    </div>
  )
}

function Feature({ text }: { text: string }) {
  return (
    <div className="bg-white/10 backdrop-blur-sm rounded-lg p-4 text-white">
      ✓ {text}
    </div>
  )
}
