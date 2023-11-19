FROM golang:1.19-alpine AS gobuild

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o wallabot .

# Final image
FROM alpine:latest
WORKDIR /app
COPY --from=gobuild /app/wallabot .

ENTRYPOINT ["./wallabot"]
