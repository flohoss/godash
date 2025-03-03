const { addDynamicIconSelectors } = require('@iconify/tailwind');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./views/**/*.templ'],
  theme: {
    container: {
      center: true
    },
  },
  plugins: [addDynamicIconSelectors()],
};
