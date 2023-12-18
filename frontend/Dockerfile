# frontend/Dockerfile

# Build Stage
FROM node:16.20.2-alpine AS builder

WORKDIR /app
COPY . .
RUN npm install
RUN npm run build

# Final Stage
FROM nginx:stable-alpine-slim

COPY --from=builder /app/build /usr/share/nginx/html
