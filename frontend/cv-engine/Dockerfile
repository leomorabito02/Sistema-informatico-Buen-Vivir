# cv-renderer en dos etapas (22/11/2026).
#
# 1) Build: instala dependencias y compila el servidor en UN solo archivo
#    (scripts/build-server.mjs) con todas sus dependencias adentro.
# 2) Imagen final: solo Node, el bundle y las fuentes. Sin node_modules, sin
#    npm/npx/corepack/yarn y sin esbuild.
#
# Por qué: Trivy marcaba vulnerabilidades HIGH/CRITICAL de cosas que el
# servidor no usa: el binario de esbuild (tsx al arrancar), el workspace
# entero del frontend (Angular, Firebase, jsPDF...) que se colaba con
# `pnpm install --filter` y `pnpm deploy`, y paquetes internos del npm de
# la imagen base.

# 1) Build.
FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY cv-engine/package.json ./cv-engine/
RUN pnpm install --filter @buenvivir/cv-engine... --frozen-lockfile
COPY cv-engine ./cv-engine
RUN pnpm --filter @buenvivir/cv-engine run build:server

# 2) Imagen final.
FROM node:22-alpine
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/corepack \
      /usr/local/bin/npm /usr/local/bin/npx /usr/local/bin/corepack \
      /opt/yarn-* /usr/local/bin/yarn /usr/local/bin/yarnpkg
WORKDIR /app/cv-engine
# Las fuentes se leen desde <cwd>/fonts (src/fonts.ts).
COPY cv-engine/fonts ./fonts
COPY --from=build /app/cv-engine/dist/server ./dist/server

EXPOSE 5005
ENV PORT=5005
ENV NODE_ENV=production

USER node

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:5005/health || exit 1

CMD ["node", "dist/server/server.mjs"]
