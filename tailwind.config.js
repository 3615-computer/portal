// npx tailwindcss -i ./assets/src/css/main.css -o ./public/css/main.css --watch
/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./views/**/*.html", "./views/*.html", "./assets/src/css/**.css"],
    theme: {
        extend: {
            colors: {
                amethyst: {
                    50: "#fcf5fe",
                    100: "#f9eafd",
                    200: "#f2d5f9",
                    300: "#eab4f3",
                    400: "#de87eb",
                    500: "#cc58dd",
                    600: "#bc4bca",
                    700: "#952c9f",
                    800: "#7b2682",
                    900: "#67246b",
                    950: "#430c46",
                },
                blaze: {
                    50: "#fff8ec",
                    100: "#fff0d3",
                    200: "#ffdda5",
                    300: "#ffc36d",
                    400: "#ff9e32",
                    500: "#ff800a",
                    600: "#ff6700",
                    700: "#cc4902",
                    800: "#a1390b",
                    900: "#82310c",
                    950: "#461604",
                },
            },
        },
    },
    plugins: [require("@tailwindcss/forms")],
};
