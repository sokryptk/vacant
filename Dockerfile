FROM node:22-alpine AS web
WORKDIR /src
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.26-alpine AS go
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY *.go ./
COPY --from=web /src/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o vacant

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go /src/vacant .
EXPOSE 8080
ENTRYPOINT ["./vacant", "serve", "-addr", ":8080"]
