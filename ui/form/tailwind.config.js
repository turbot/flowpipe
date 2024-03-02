/** @type {import('tailwindcss').Config} */
export default {
  content: ["./src/**/*.{js,jsx,ts,tsx}", "./index.html"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "var(--color-background)",
        foreground: "var(--color-foreground)",
        "foreground-light": "var(--color-foreground-light)",
        "flowpipe-blue": "rgb(var(--color-flowpipe-blue) / <alpha-value>)",
        "flowpipe-blue-dark":
          "rgb(var(--color-flowpipe-blue-dark) / <alpha-value>)",
        modal: "var(--color-modal)",
        "modal-divide": "var(--color-modal-divide)",
        alert: "rgb(var(--color-alert) / <alpha-value>)",
        ok: "rgb(var(--color-ok) / <alpha-value>)",
        info: "rgb(var(--color-info) / <alpha-value>)",
      },
    },
  },
  plugins: [],
};
