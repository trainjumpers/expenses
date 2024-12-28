import typography from "@tailwindcss/typography";
import daisyui from "daisyui";

const config = {
  content: ["./src/**/*.{vue,js,ts}"],
  plugins: [daisyui, typography],
};
export default config;
module.exports = {
  ...config,
  daisyui: {
    themes: [
      "light",
      "dark",
      "cupcake",
      "bumblebee",
      "emerald",
      "corporate",
      "synthwave",
      "retro",
      "cyberpunk",
      "valentine",
      "halloween",
      "garden",
      "forest",
      "aqua",
      "lofi",
      "pastel",
      "fantasy",
      "wireframe",
      "black",
      "luxury",
      "dracula",
      "cmyk",
      "autumn",
      "business",
      "acid",
      "lemonade",
      "night",
      "coffee",
      "winter",
      "procyon",
      "solar",
      "joker",
    ],
  },
};
