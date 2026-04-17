/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    "./public/index.html"
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "primary": "#137fec",
        "primary-dark": "#0f65bd",
        "background-light": "#f6f7f8",
        "background-dark": "#0b1219",
        "surface-dark": "#16202a",
        "surface-darker": "#111a22",
        "surface-light": "#ffffff",
        "border-color": "#233648",
        "text-muted": "#92adc9",
        
        // Obsidian Fleet Design Tokens
        "surface-container": "#16202a",
        "surface-container-low": "#111a22",
        "surface-container-high": "#1d2936",
        "surface-container-highest": "#233648",
        "surface-container-lowest": "#0b1219",
        "surface-dim": "#0b1219",
        "surface-bright": "#16202a",
        "on-surface": "#ffffff",
        "on-surface-variant": "#92adc9",
        "outline": "#233648",
        "outline-variant": "#334d66",
        "secondary-container": "#233648",
        "inverse-surface": "#f9f9ff",
        "inverse-on-surface": "#181c22",
        "tertiary": "#964400",
        "tertiary-container": "#bc5700",
        "on-tertiary": "#ffffff",
        
        // Monitoring Module Specific
        "monitor-primary": "#ec5b13",
        "monitor-primary-dark": "#d44908",
        "monitor-bg-light": "#f6f7f8",
        "monitor-bg-dark": "#0b1219",
        "monitor-surface": "#16202a",
        "monitor-surface-darker": "#111a22",
        "monitor-border": "#233648",
        "monitor-text-muted": "#92adc9",
        "monitor-success": "#10b981",
        "monitor-warning": "#f59e0b",
        "monitor-danger": "#ef4444",
      },
      fontFamily: {
        "display": ["Inter", "sans-serif"],
        "body": ["Inter", "sans-serif"],
        "public-sans": ["Public Sans", "sans-serif"],
        "mono": ["JetBrains Mono", "monospace"]
      },
      borderRadius: {
        "lg": "0.75rem",
        "xl": "1rem",
        "2xl": "1.5rem",
      },
      animation: {
        'ping-slow': 'ping 3s cubic-bezier(0, 0, 0.2, 1) infinite',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      }
    },
  },
  plugins: [],
}
