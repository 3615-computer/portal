// npx tailwindcss -i ./assets/src/css/main.css -o ./public/css/main.css --watch
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/views/**/*.html", "./app/views/*.html", "./app/assets/src/css/**.css"],
  plugins: [require("@tailwindcss/forms")],
};
