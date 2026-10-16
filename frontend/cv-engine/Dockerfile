FROM node:22-alpine

WORKDIR /app

# Enable pnpm
RUN corepack enable && corepack prepare pnpm@9.15.0 --activate

# Copy root workspace configs
COPY pnpm-workspace.yaml package.json pnpm-lock.yaml* ./
COPY cv-engine/package.json ./cv-engine/

# Install dependencies for cv-engine
RUN pnpm install --filter @buenvivir/cv-engine...

# Copy cv-engine code
COPY cv-engine ./cv-engine

EXPOSE 5005
ENV PORT=5005
ENV NODE_ENV=production

WORKDIR /app/cv-engine

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:5005/health || exit 1

CMD ["pnpm", "exec", "tsx", "src/renderer/server.ts"]
