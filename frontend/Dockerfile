FROM node:22-alpine AS build

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm exec ng build --configuration production

FROM nginxinc/nginx-unprivileged:alpine

COPY --from=build /app/dist/frontend-buen-vivir/browser /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

USER nginx
CMD ["nginx", "-g", "daemon off;"]
