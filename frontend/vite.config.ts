import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import Pages from "vite-plugin-pages";
import "vite-plugin-pages-react";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss(), Pages()],
});
