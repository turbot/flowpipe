/** @type {import('tailwindcss').Config} */
export default {
  content: ["./src/**/*.{js,jsx,ts,tsx}", "./index.html"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "var(--color-background)",
        foreground: "rgb(var(--color-foreground) / <alpha-value>)",
        "foreground-light": "var(--color-foreground-light)",
        "flowpipe-blue": "rgb(var(--color-flowpipe-blue) / <alpha-value>)",
        "flowpipe-blue-dark":
          "rgb(var(--color-flowpipe-blue-dark) / <alpha-value>)",
        modal: "var(--color-modal)",
        "modal-divide": "var(--color-modal-divide)",
        alert: "var(--color-alert)",
        ok: "var(--color-ok)",
        info: "rgb(var(--color-info) / <alpha-value>)",
      },
    },
  },
  plugins: [],
};
