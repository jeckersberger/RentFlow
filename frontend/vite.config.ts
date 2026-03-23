import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      strategies: 'generateSW',
      registerType: 'autoUpdate',
      includeAssets: [
        'favicon.ico',
        'apple-touch-icon.png',
        'masked-icon.svg',
        'icons/*.png',
      ],
      manifest: {
        name: 'RentFlow \u2014 Vermietungssoftware',
        short_name: 'RentFlow',
        description: 'Professionelle Lagerverwaltung und Rechnungssoftware f\u00fcr Veranstaltungstechnik',
        theme_color: '#0a0f1a',
        background_color: '#0a0f1a',
        display: 'standalone',
        orientation: 'any',
        scope: '/',
        start_url: '/',
        categories: ['business', 'productivity'],
        icons: [
          {
            src: '/icons/icon-72.png',
            sizes: '72x72',
            type: 'image/png',
          },
          {
            src: '/icons/icon-96.png',
            sizes: '96x96',
            type: 'image/png',
          },
          {
            src: '/icons/icon-128.png',
            sizes: '128x128',
            type: 'image/png',
          },
          {
            src: '/icons/icon-144.png',
            sizes: '144x144',
            type: 'image/png',
          },
          {
            src: '/icons/icon-152.png',
            sizes: '152x152',
            type: 'image/png',
          },
          {
            src: '/icons/icon-192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: '/icons/icon-384.png',
            sizes: '384x384',
            type: 'image/png',
          },
          {
            src: '/icons/icon-512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: '/icons/icon-512-maskable.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
        shortcuts: [
          {
            name: 'Scanner',
            short_name: 'Scan',
            url: '/scanner',
            icons: [{ src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' }],
          },
          {
            name: 'Neues Projekt',
            short_name: 'Projekt',
            url: '/projects/new',
            icons: [{ src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' }],
          },
          {
            name: 'Equipment',
            url: '/equipment',
            icons: [{ src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' }],
          },
        ],
        screenshots: [
          {
            src: 'screenshot-540x720.png',
            sizes: '540x720',
            type: 'image/png',
          },
          {
            src: 'screenshot-1280x720.png',
            sizes: '1280x720',
            type: 'image/png',
          },
        ],
      },
      workbox: {
        maximumFileSizeToCacheInBytes: 5 * 1024 * 1024, // 5 MiB
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webp,woff2}'],
        globIgnores: ['**/node_modules/**/*'],
        runtimeCaching: [
          {
            // Equipment & Projects list API — cache-first for fast offline access
            urlPattern: /\/api\/(equipment|projects)(\?.*)?$/i,
            handler: 'StaleWhileRevalidate',
            options: {
              cacheName: 'api-lists-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 60 * 24, // 24 hours
              },
              cacheableResponse: {
                statuses: [0, 200],
              },
            },
          },
          {
            // All other API calls — network-first with fallback
            urlPattern: /\/api\/.*/i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              networkTimeoutSeconds: 3,
              expiration: {
                maxEntries: 200,
                maxAgeSeconds: 60 * 60 * 24, // 24 hours
              },
              cacheableResponse: {
                statuses: [0, 200],
              },
            },
          },
          {
            // Google Fonts or CDN assets
            urlPattern: /^https:\/\/fonts\.(googleapis|gstatic)\.com\/.*/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'google-fonts-cache',
              expiration: {
                maxEntries: 30,
                maxAgeSeconds: 60 * 60 * 24 * 365, // 1 year
              },
              cacheableResponse: {
                statuses: [0, 200],
              },
            },
          },
        ],
        navigateFallback: '/index.html',
        navigateFallbackDenylist: [/^\/api\//],
        cleanupOutdatedCaches: true,
      },
      devOptions: {
        enabled: false,
        navigateFallbackAllowlist: [/^\/(?!.*\.).*/],
      },
    }),
  ],
  server: {
    port: 3000,
    strictPort: false,
    proxy: {
      '/api': {
        target: 'http://localhost:80',
        changeOrigin: true,
        rewrite: (path) => path,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser',
  },
})
