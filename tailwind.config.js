/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/*.html"],
  theme: {
    container: {
      center: true
    },
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        light: {
          ...require("daisyui/src/colors/themes")["[data-theme=garden]"],
          primary: "#f28c18",
          secondary: "rgba(70,70,70,0.7)",
        },
        dark: {
          ...require("daisyui/src/colors/themes")["[data-theme=halloween]"],
          secondary: "#a0a0a0",
        },
      },
    ],
    darkTheme: "dark",
  },
};
