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
      // Auth, Setup, Users, Tenants, Config → auth-service
      '/api/v1/auth': { target: 'http://localhost:8001', changeOrigin: true },
      '/api/v1/setup': { target: 'http://localhost:8001', changeOrigin: true },
      '/api/v1/users': { target: 'http://localhost:8001', changeOrigin: true },
      '/api/v1/tenants': { target: 'http://localhost:8001', changeOrigin: true },
      '/api/v1/config': { target: 'http://localhost:8001', changeOrigin: true },
      '/api/v1/contacts': { target: 'http://localhost:8003', changeOrigin: true },
      '/api/v1/invitations': { target: 'http://localhost:8001', changeOrigin: true },
      // Inventory → inventory-service
      '/api/v1/equipment': { target: 'http://localhost:8002', changeOrigin: true },
      '/api/v1/categories': { target: 'http://localhost:8002', changeOrigin: true },
      '/api/v1/flightcases': { target: 'http://localhost:8002', changeOrigin: true },
      // Projects → project-service
      '/api/v1/projects': { target: 'http://localhost:8003', changeOrigin: true },
      '/api/v1/packlists': { target: 'http://localhost:8003', changeOrigin: true },
      '/api/v1/reservations': { target: 'http://localhost:8003', changeOrigin: true },
      '/api/v1/customers': { target: 'http://localhost:8003', changeOrigin: true },
      // Scanner → scanner-service
      '/api/v1/scan': { target: 'http://localhost:8004', changeOrigin: true },
      '/api/v1/scanner': { target: 'http://localhost:8004', changeOrigin: true },
      // Warehouse → warehouse-service
      '/api/v1/locations': { target: 'http://localhost:8005', changeOrigin: true },
      '/api/v1/movements': { target: 'http://localhost:8005', changeOrigin: true },
      '/api/v1/inventory': { target: 'http://localhost:8005', changeOrigin: true },
      '/api/v1/warehouses': { target: 'http://localhost:8005', changeOrigin: true },
      // Invoices → invoice-service
      '/api/v1/invoices': { target: 'http://localhost:8006', changeOrigin: true },
      '/api/v1/quotes': { target: 'http://localhost:8006', changeOrigin: true },
      '/api/v1/credit-notes': { target: 'http://localhost:8006', changeOrigin: true },
      '/api/v1/dunning': { target: 'http://localhost:8006', changeOrigin: true },
      '/api/v1/export': { target: 'http://localhost:8006', changeOrigin: true },
      // Documents → document-service
      '/api/v1/documents': { target: 'http://localhost:8007', changeOrigin: true },
      '/api/v1/templates': { target: 'http://localhost:8007', changeOrigin: true },
      // Crew → crew-service
      '/api/v1/crew': { target: 'http://localhost:8008', changeOrigin: true },
      '/api/v1/time-entries': { target: 'http://localhost:8008', changeOrigin: true },
      // Federation → federation-service
      '/api/v1/federation': { target: 'http://localhost:8009', changeOrigin: true },
      // Maintenance → maintenance-service
      '/api/v1/maintenance': { target: 'http://localhost:8010', changeOrigin: true },
      '/api/v1/dguv': { target: 'http://localhost:8010', changeOrigin: true },
      // Transport → transport-service
      '/api/v1/transport': { target: 'http://localhost:8011', changeOrigin: true },
      // Insurance → insurance-service
      '/api/v1/policies': { target: 'http://localhost:8012', changeOrigin: true },
      '/api/v1/claims': { target: 'http://localhost:8012', changeOrigin: true },
      // Workflows → workflow-service
      '/api/v1/workflows': { target: 'http://localhost:8013', changeOrigin: true },
      // AI → ai-service
      '/api/v1/ai': { target: 'http://localhost:8014', changeOrigin: true },
      // Notifications & Mail → notification-service
      '/api/v1/notifications': { target: 'http://localhost:8015', changeOrigin: true },
      '/api/v1/mail': { target: 'http://localhost:8015', changeOrigin: true },
      // Reporting → reporting-service
      '/api/v1/reports': { target: 'http://localhost:8016', changeOrigin: true },
      // Audit → audit-service
      '/api/v1/audit': { target: 'http://localhost:8017', changeOrigin: true },
      // Expenses → expense-service
      '/api/v1/expenses': { target: 'http://localhost:8018', changeOrigin: true },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser',
  },
})
