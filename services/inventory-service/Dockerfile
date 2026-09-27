FROM golang:1.26.6-alpine AS build
WORKDIR /src
COPY services/inventory-service/go.mod services/inventory-service/go.sum ./
RUN go mod download
COPY services/inventory-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/worker
FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=build /out/service /service
EXPOSE 8084
ENTRYPOINT ["/service"]
