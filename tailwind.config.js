const { addDynamicIconSelectors } = require('@iconify/tailwind');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./views/**/*.templ'],
  theme: {
    container: {
      center: true
    },
  },
  plugins: [require('daisyui'), addDynamicIconSelectors()],
  daisyui: {
    themes: [
      {
        light: {
          ...require("daisyui/src/theming/themes")["garden"],
          primary: "#f28c18",
          secondary: "rgba(70,70,70,0.7)",
        },
        dark: {
          ...require("daisyui/src/theming/themes")["halloween"],
          secondary: "#a0a0a0",
        },
      },
    ],
    darkTheme: "dark",
  },
};
