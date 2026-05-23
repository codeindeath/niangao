# ============================================
# 年糕 App — Docker & CI/CD
# ============================================

# ---- Go Backend ----
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20 AS backend
RUN apk add --no-cache ca-certificates
COPY --from=backend-builder /server /server
EXPOSE 8080
CMD ["/server"]

# ---- Python AI Service ----
FROM python:3.11-slim AS ai-service
WORKDIR /app
COPY ai-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY ai-service/ .
EXPOSE 8000
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
