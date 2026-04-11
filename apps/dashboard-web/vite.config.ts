import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import { visualizer } from 'rollup-plugin-visualizer';
import { resolve } from 'path';

// Security headers plugin for Vite
const securityHeadersPlugin = () => ({
  name: 'security-headers',
  transformIndexHtml(html: string) {
    // Determine if we're in production or development
    const isProduction = process.env.NODE_ENV === 'production';

    // Build CSP based on environment
    // Production: No unsafe-inline for scripts (XSS protection)
    // Development: Allow unsafe-inline for HMR and React DevTools
    // frame-ancestors and X-Frame-Options can only be set via HTTP headers, not <meta>
    // They are set via server.headers below
    let csp = [
      "default-src 'self'",
      // Scripts: production forbids inline scripts, dev allows for HMR
      isProduction
        ? "script-src 'self' https://us.i.posthog.com"
        : "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://us.i.posthog.com",
      // Styles: keep unsafe-inline for React inline styles (acceptable risk)
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
      // Fonts
      "font-src 'self' https://fonts.gstatic.com",
      // Images and media (includes tile servers)
      "img-src 'self' data: https: http://localhost:3000",
      // Connections (API, WebSocket, analytics)
      "connect-src 'self' http://localhost:8084 https://us.i.posthog.com https://*.trycloudflare.com wss://*.trycloudflare.com",
    ].join('; ');

    const tags = [
      `<meta http-equiv="Content-Security-Policy" content="${csp}">`,
      `<meta http-equiv="X-Content-Type-Options" content="nosniff">`,
      `<meta http-equiv="X-XSS-Protection" content="1; mode=block">`,
      `<meta name="referrer" content="strict-origin-when-cross-origin">`
    ].join('\n      ');

    return html.replace('</head>', `      ${tags}\n    </head>`);
  },
});

export default defineConfig({
  plugins: [react(), securityHeadersPlugin()],
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-redux',
      '@reduxjs/toolkit',
      '@tanstack/react-query',
      'lucide-react',
      '@heroicons/react',
      'date-fns',
    ],
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@flyfleet/config': resolve(__dirname, '../../packages/config/src'),
    },
  },
  server: {
    port: 3000,
    allowedHosts: true,
    headers: {
      'X-Frame-Options': 'DENY',
      'Content-Security-Policy': "frame-ancestors 'none'",
    },
    proxy: {
      '/api/v1/ws': {
        target: 'ws://localhost:8084',
        ws: true,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/v1\/ws/, '/ws/telemetry'),
      },
      '/api': {
        target: 'http://localhost:8084',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8084',
        ws: true,
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8084',
        changeOrigin: true,
      },
      '/carto-proxy': {
        target: 'https://basemaps.cartocdn.com',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/carto-proxy/, ''),
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    chunkSizeWarningLimit: 2000,
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // Vendor chunk - React ecosystem
          if (id.includes('react') || id.includes('react-dom') || id.includes('react-redux') || id.includes('react-router-dom')) {
            return 'vendor';
          }
          // Charts & Maps - heavy visualization libraries
          if (id.includes('recharts') || id.includes('leaflet') || id.includes('maplibre-gl') || id.includes('supercluster')) {
            return 'charts-maps';
          }
          // Icons
          if (id.includes('lucide-react') || id.includes('@heroicons/react') || id.includes('material-symbols')) {
            return 'icons';
          }
          // Animation
          if (id.includes('framer-motion')) {
            return 'animations';
          }
          // Date/utility
          if (id.includes('date-fns')) {
            return 'date-utils';
          }
        },
      },
    },
  },
  define: {
    'process.env.REACT_APP_API_URL': JSON.stringify(process.env.REACT_APP_API_URL || '/api/v1'),
    'process.env.REACT_APP_PUBLIC_POSTHOG_KEY': JSON.stringify(process.env.REACT_APP_PUBLIC_POSTHOG_KEY || ''),
    'process.env.REACT_APP_PUBLIC_POSTHOG_HOST': JSON.stringify(process.env.REACT_APP_PUBLIC_POSTHOG_HOST || 'https://us.i.posthog.com'),
  },
  test: {
    globals: true,
    environment: 'jsdom',
    environmentOptions: {
      url: 'http://localhost/',
    },
    setupFiles: './src/setupTests.ts',
    css: true,
    include: ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    exclude: ['node_modules', 'dist', 'tests'],
  },
});
