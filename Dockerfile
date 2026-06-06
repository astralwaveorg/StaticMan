# 多阶段构建：前端编译 → Go编译 → 精简运行镜像

# 阶段1: 编译前端
FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

# 阶段2: 编译Go后端
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
# 嵌入前端静态资源
COPY --from=web-builder /app/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -tags withweb -o /staticman ./cmd/server

# 阶段3: 运行镜像
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /staticman /app/staticman
EXPOSE 8080
VOLUME /app/data
ENTRYPOINT ["/app/staticman"]