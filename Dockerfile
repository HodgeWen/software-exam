# ---- 前端构建:产出纯静态 dist ----
FROM node:24-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- 后端构建:纯 Go SQLite 驱动,无需 CGO ----
FROM golang:1.27-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /exam-server ./cmd/server

# ---- 运行时:单进程同时提供 REST API 与前端页面 ----
FROM alpine:3.22
COPY --from=backend /exam-server /usr/local/bin/exam-server
COPY --from=frontend /app/dist /static
ENV EXAM_PORT=8080 \
    EXAM_DB_PATH=/data/exam.db \
    EXAM_STATIC_DIR=/static
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["exam-server"]
