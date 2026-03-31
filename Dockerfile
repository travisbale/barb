# Build frontend
FROM node:22-alpine AS frontend

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Build backend
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache make

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/cmd/barb/dist ./cmd/barb/dist

ARG VERSION=dev
RUN go build -ldflags "-X main.Version=${VERSION}" -o /barb ./cmd/barb

# Runtime
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /barb /usr/local/bin/barb

VOLUME ["/data"]

EXPOSE 443

ENTRYPOINT ["/usr/local/bin/barb", "serve", "--db", "/data/barb.db", "--addr", ":443"]
