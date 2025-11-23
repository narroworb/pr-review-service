FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /root/

COPY --from=build /app/app .
COPY --from=build /app/migrations ./migrations
COPY --from=build /app/test_data ./test_data

ENV POSTGRES_DSN=${POSTGRES_DSN}

EXPOSE 8080
CMD ["./app"]
