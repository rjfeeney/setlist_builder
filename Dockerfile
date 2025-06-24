FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o setlist_builder .

FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY --from=builder /app/setlist_builder .
CMD ["./setlist_builder"]