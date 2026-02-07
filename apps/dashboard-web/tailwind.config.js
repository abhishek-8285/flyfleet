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
        "background-dark": "#101922",
        "surface-dark": "#16202a",
        "surface-darker": "#111a22",
        "surface-light": "#ffffff",
        "border-color": "#233648",
        "text-muted": "#92adc9",
        "border-dark": "#233648", // keeping as alias
        "text-secondary": "#92adc9", // keeping as alias
      },
      fontFamily: {
        "display": ["Inter", "sans-serif"],
        "body": ["Inter", "sans-serif"]
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
