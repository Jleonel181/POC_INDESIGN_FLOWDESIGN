# Stage 1: Compilar idmlgen (Go)
FROM golang:1.21-alpine AS go-builder
WORKDIR /src
COPY idmllib/go.mod idmllib/go.sum ./
RUN go mod download
COPY idmllib/ ./
RUN CGO_ENABLED=0 go build -o /idmlgen ./cmd/idmlgen

# Stage 2: Compilar backend (TypeScript)
FROM node:22-alpine AS ts-builder
WORKDIR /app
COPY backend/package*.json ./
RUN npm ci
COPY backend/ ./
RUN npx tsc --outDir dist

# Stage 3: Imagen final
FROM node:22-alpine
WORKDIR /app
COPY --from=go-builder /idmlgen /usr/local/bin/idmlgen
COPY backend/package*.json ./
RUN npm ci --omit=dev
COPY --from=ts-builder /app/dist ./dist
EXPOSE 3001
CMD ["node", "dist/main.js"]
