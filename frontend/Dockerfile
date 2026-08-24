FROM node:20-alpine AS build

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .
RUN npm run build -- --configuration production

FROM nginxinc/nginx-unprivileged:alpine

COPY --from=build /app/dist/frontend-buen-vivir/browser /usr/share/nginx/html

EXPOSE 8080

USER nginx
CMD ["nginx", "-g", "daemon off;"]
