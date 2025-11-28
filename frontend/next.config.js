/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  
  async rewrites() {
    // Use backend container name in Docker, localhost otherwise
    const apiUrl = process.env.API_URL || 'http://localhost:8080'
    
    return [
      // WebSocket proxy - MUST be first for proper upgrade
      {
        source: '/ws',
        destination: `${apiUrl}/ws`,
      },
      // API routes
      {
        source: '/api/:path*',
        destination: `${apiUrl}/api/:path*`,
      },
      // Auth routes
      {
        source: '/auth/:path*',
        destination: `${apiUrl}/auth/:path*`,
      },
    ]
  },

  // Webpack config for better error logging
  webpack: (config, { dev, isServer }) => {
    // Reduce noise in development
    if (dev && !isServer) {
      config.infrastructureLogging = { level: 'error' }
    }
    return config
  },

  // Allow WebSocket connections
  experimental: {
    serverActions: true,
  },
}

module.exports = nextConfig
