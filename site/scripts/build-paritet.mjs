// Builds the paritet.dev variant of the landing page.
//
// The variant uses the "draft" visual language and carries the paritet
// branding. Keeping the switch in one portable script means the same
// command works on Windows, macOS and Linux.
//
//   node scripts/build-paritet.mjs
process.env.ASBUILT_THEME = "draft";

const { build } = await import("vite");

await build({
  build: {
    outDir: "dist-paritet",
    emptyOutDir: true,
  },
});
