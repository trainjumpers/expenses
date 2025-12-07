# --- Builder Stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

# Cache Go modules: only re-download if these change
COPY go.mod go.sum ./
RUN go mod download

# Copy only Go source files you need for the build
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/neurospend/

# --- Runtime Stage ---
FROM alpine:3.19
WORKDIR /app

# Create a non-root user and group
RUN addgroup -S appuser && adduser -S appuser -G appuser
COPY --from=builder /app/server .

EXPOSE 8080
USER appuser

CMD ["./server"]
