import { defineConfig } from "vitest/config";
import path from "node:path";

export default defineConfig({
  resolve: {
    alias: [
      { find: "#react-pdf-renderer", replacement: "@react-pdf/renderer" },
      { find: /^@reactive-resume\/schema\/(.*)$/, replacement: path.resolve(__dirname, "./vendor/schema/$1") },
      { find: "@reactive-resume/schema", replacement: path.resolve(__dirname, "./vendor/schema/index.ts") },
      { find: /^@reactive-resume\/utils\/(.*)$/, replacement: path.resolve(__dirname, "./vendor/utils/$1") },
      { find: "@reactive-resume/utils", replacement: path.resolve(__dirname, "./vendor/utils/index.ts") },
      { find: /^@reactive-resume\/resume\/(.*)$/, replacement: path.resolve(__dirname, "./vendor/resume/$1") },
      { find: "@reactive-resume/resume", replacement: path.resolve(__dirname, "./vendor/resume/index.ts") },
      { find: /^@reactive-resume\/pdf\/(.*)$/, replacement: path.resolve(__dirname, "./vendor/pdf/$1") },
      { find: "@reactive-resume/pdf", replacement: path.resolve(__dirname, "./vendor/pdf/index.ts") },
      { find: "@reactive-resume/fonts", replacement: path.resolve(__dirname, "./src/fonts.ts") },
    ],
  },
});
