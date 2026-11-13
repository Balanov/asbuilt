import { defineConfig } from "vite";

// The landing ships in two visual languages:
//
//   night  - a live instrument panel: dark canvas, amber signal for
//           deviations, cyan for capability.
//   draft  - a technical drawing sheet: pale cool paper, dimension-line
//           annotations, vermilion deviations.
//
// The page opens in the light language and the reader can switch. Same
// markup, same data, two languages. The theme is chosen at build time so a
// page never flashes the wrong one.
const theme = process.env.ASBUILT_THEME === "night" ? "night" : "draft";
const paritet = process.env.ASBUILT_PARITET === "1";

function asbuiltTheme() {
  return {
    name: "asbuilt-theme",
    transformIndexHtml(html) {
      return html
        .replaceAll("__THEME__", theme)
        .replaceAll("__PARITET__", paritet ? "paritet" : "");
    },
  };
}

export default defineConfig({
  base: "./",
  plugins: [asbuiltTheme()],
});
