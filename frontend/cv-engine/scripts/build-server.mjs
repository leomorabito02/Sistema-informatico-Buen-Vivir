// Compila el servidor de cv-renderer en un solo archivo con TODAS sus
// dependencias adentro (22/11/2026). Así la imagen Docker final no necesita
// node_modules: con pnpm 9, `pnpm deploy` copiaba el almacén del workspace
// entero (Angular, Firebase, jsPDF...) y Trivy marcaba sus vulnerabilidades
// en una imagen que no los usa.
//
// Algunos paquetes (css-tree) leen JSON en tiempo de ejecución con
// createRequire(import.meta.url), que esbuild no puede empaquetar. El plugin
// cambia esos require("...json") por imports estáticos para que los datos
// queden dentro del bundle.
//
// Uso: pnpm run build:server   (salida: dist/server/server.mjs)
import { build } from "esbuild";
import { readFile } from "node:fs/promises";

const outfile = process.argv[2] || "dist/server/server.mjs";

const jsonEnTiempoDeEjecucion = {
  name: "json-en-tiempo-de-ejecucion",
  setup(b) {
    // Solo los paquetes que lo necesitan: abrir todo node_modules agota los
    // descriptores de archivo (miles de íconos).
    b.onLoad({ filter: /[\\/]node_modules[\\/](css-tree|pdfkit)[\\/].*\.m?js$/ }, async (args) => {
      const src = await readFile(args.path, "utf8");
      // require("x.json") (css-tree) y require$1("#standard-fonts/Helvetica")
      // (pdfkit: import interno resuelto por el mapa "imports" del paquete).
      const patron = /\brequire(?:\$\d+)?\(\s*['"]((?:[^'"]+\.json)|(?:#[^'"]+))['"]\s*\)/g;
      if (!src.includes("createRequire") || !patron.test(src)) {
        return undefined;
      }
      patron.lastIndex = 0;
      const imports = [];
      const contents = src.replace(patron, (_, ruta) => {
        const id = `__json${imports.length}`;
        imports.push(`import ${id} from ${JSON.stringify(ruta)};`);
        return id;
      });
      return { contents: imports.join("\n") + "\n" + contents, loader: "js" };
    });
  },
};

await build({
  entryPoints: ["src/renderer/server.ts"],
  bundle: true,
  platform: "node",
  format: "esm",
  target: "node22",
  outfile,
  logLevel: "warning",
  plugins: [jsonEnTiempoDeEjecucion],
  // Dependencias CommonJS que usan require() de módulos de Node.
  banner: {
    js: "import { createRequire as __bvCreateRequire } from 'node:module'; const require = __bvCreateRequire(import.meta.url);",
  },
});
