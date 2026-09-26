FROM node:24.19.0-bookworm-slim AS builder

WORKDIR /build

COPY ui/package.json ui/package-lock.json ./

RUN npm ci

COPY ./ui .

RUN npm run build

FROM nginx:1.30.5-alpine3.24

COPY --from=builder /build/dist /usr/share/nginx/html

COPY ./docker/ui/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
