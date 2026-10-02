FROM node:24.19.0-bookworm-slim AS builder

WORKDIR /build

COPY ui/package.json ui/package-lock.json ./

RUN npm ci

COPY ./ui .

RUN npm run build

FROM nginxinc/nginx-unprivileged:1.30-alpine3.24-slim

COPY --from=builder /build/dist /usr/share/nginx/html

COPY docker/ui/nginx/default.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 7006

CMD ["nginx", "-g", "daemon off;"]
