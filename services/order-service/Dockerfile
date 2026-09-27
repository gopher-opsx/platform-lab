FROM golang:1.26.6-alpine AS build
WORKDIR /src
COPY services/order-service/go.mod services/order-service/go.sum ./
RUN go mod download
COPY services/order-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/api
FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=build /out/service /service
EXPOSE 8083
ENTRYPOINT ["/service"]
