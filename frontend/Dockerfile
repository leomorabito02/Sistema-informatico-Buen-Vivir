FROM node:20-alpine AS build

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm run build -- --configuration production

FROM nginxinc/nginx-unprivileged:alpine

COPY --from=build /app/dist/frontend-buen-vivir/browser /usr/share/nginx/html

EXPOSE 8080

USER nginx
CMD ["nginx", "-g", "daemon off;"]
