# Stage 1: Build the React App
FROM oven/bun:1 AS builder
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install
COPY . .
# This builds the static HTML/JS files using the .env.production variable
RUN bun run build

# Stage 2: Serve with NGINX
FROM nginx:alpine
# Copy build output to NGINX html folder
COPY --from=builder /app/dist /usr/share/nginx/html
# Default NGINX config is fine for React Router if we add a try_files fallback
RUN echo 'server { listen 80; location / { root /usr/share/nginx/html; index index.html index.htm; try_files $uri $uri/ /index.html; } }' > /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]