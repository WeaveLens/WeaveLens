FROM node:20-alpine AS builder

WORKDIR /src

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web .
RUN npm run build

FROM nginx:1.27-alpine AS runtime

COPY --from=builder /src/dist /usr/share/nginx/html
COPY docker/frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Allow the non-root nginx user to read static files and write run/cache/log dirs.
RUN mkdir -p /var/cache/nginx /var/log/nginx /run \
    && chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx /run

USER 101:101

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["wget", "-q", "-O", "-", "http://localhost:8080/"]

ENTRYPOINT ["nginx", "-g", "daemon off;"]
