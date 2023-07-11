// npx tailwindcss -i ./assets/src/css/main.css -o ./public/css/main.css --watch
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/layouts/*.html",
    "./views/partials/*.html",
    "./views/*.html",
    "./assets/src/css/**.css",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
}