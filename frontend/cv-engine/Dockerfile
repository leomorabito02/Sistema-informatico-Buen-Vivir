# cv-renderer en tres etapas (22/11/2026).
#
# Antes la imagen arrancaba con `tsx`, que transpila TypeScript al vuelo
# con esbuild: el binario de esbuild (compilado en Go) quedaba en la imagen
# y Trivy lo marcaba con vulnerabilidades HIGH de la biblioteca estándar de
# Go. Ahora el servidor se compila en la etapa de build y la imagen final
# solo tiene Node, las dependencias de producción y el bundle.

# 1) Build: todas las dependencias y el bundle del servidor.
FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY cv-engine/package.json ./cv-engine/
RUN pnpm install --filter @buenvivir/cv-engine... --frozen-lockfile
COPY cv-engine ./cv-engine
RUN pnpm --filter @buenvivir/cv-engine run build:server

# 2) Dependencias de producción de cv-engine SOLAMENTE. `pnpm deploy` arma
#    un node_modules aislado; con `install --filter` se colaba todo el
#    workspace del frontend (Angular, Firebase, jsPDF...) y Trivy marcaba sus
#    vulnerabilidades en una imagen que no los usa.
FROM node:22-alpine AS deps
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY cv-engine/package.json ./cv-engine/
RUN pnpm --filter @buenvivir/cv-engine deploy --prod /out

# 3) Imagen final.
FROM node:22-alpine
# La imagen base trae npm, npx, corepack y yarn, que el servidor no usa
# (arranca con `node`). Trivy marcaba vulnerabilidades en paquetes internos
# de npm (brace-expansion, ip-address, pacote, picomatch, sigstore).
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/corepack \
      /usr/local/bin/npm /usr/local/bin/npx /usr/local/bin/corepack \
      /opt/yarn-* /usr/local/bin/yarn /usr/local/bin/yarnpkg
WORKDIR /app
COPY --from=deps /out/node_modules ./cv-engine/node_modules
# package.json: lo usa Node para resolver el import interno #react-pdf-renderer.
COPY cv-engine/package.json ./cv-engine/package.json
# Las fuentes se leen desde <cwd>/fonts (src/fonts.ts).
COPY cv-engine/fonts ./cv-engine/fonts
COPY --from=build /app/cv-engine/dist/server ./cv-engine/dist/server

EXPOSE 5005
ENV PORT=5005
ENV NODE_ENV=production

WORKDIR /app/cv-engine
USER node

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:5005/health || exit 1

CMD ["node", "dist/server/server.mjs"]
