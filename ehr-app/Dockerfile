FROM golang:1.22-alpine AS builder


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ehr-app .

FROM gcr.io/distroless/base-debian11

COPY --from=builder /app/ehr-app /ehr-app

EXPOSE 8081

ENTRYPOINT ["/ehr-app"]
