# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

RUN npm run build

# Stage 2: Runtime
FROM nginx:alpine

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf
COPY proxy_params /etc/nginx/proxy_params

# Copy built assets from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Create directory for health check
RUN mkdir -p /usr/share/nginx/html

# Health check

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
