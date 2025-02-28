import { build } from "bun";
import { rm } from "node:fs/promises";
import copyContents from "./copier";

await rm("dist", { recursive: true, force: true });

await Promise.all([
  build({
    entrypoints: ["./src/landing/index.html"],
    outdir: "dist/landing",
    target: "browser",
    minify: true,
    sourcemap: "external",
  }),

  build({
    entrypoints: ["./src/popup/index.html"],
    outdir: "dist/popup",
    target: "browser",
    minify: true,
    sourcemap: "external",
  }),

  build({
    entrypoints: ["./src/service-worker.ts"],
    outdir: "dist",
    target: "browser",
    minify: true,
    sourcemap: "external",
  }),
]);

await copyContents("./public/", "./dist");

// eslint-disable-next-line no-console
console.log("✅ Build completed successfully");
