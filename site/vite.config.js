import { defineConfig } from "vite";

// The landing ships in two visual languages:
//
//   night  - the balanov.pro product page. A live instrument panel: dark
//           canvas, amber signal for deviations, cyan for capability.
//   draft  - the paritet.dev page. A technical drawing sheet: pale cool
//           paper, dimension-line annotations, vermilion deviations.
//
// Same markup, same data, two languages. The theme is chosen at build time
// so a page never flashes the wrong one.
const theme = process.env.ASBUILT_THEME === "draft" ? "draft" : "night";
const paritet = theme === "draft";

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
