/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  async rewrites() {
    // Use backend container name in Docker, localhost otherwise
    const apiUrl = process.env.API_URL || 'http://backend:8080'
    return [
      {
        source: '/api/:path*',
        destination: `${apiUrl}/api/:path*`,
      },
      {
        source: '/auth/:path*',
        destination: `${apiUrl}/auth/:path*`,
      },
      {
        source: '/ws',
        destination: `${apiUrl}/ws`,
      },
    ]
  },
}

module.exports = nextConfig
