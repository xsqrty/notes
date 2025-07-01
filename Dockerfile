FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o ./bin/notes -ldflags "-X '{{.PACKAGE}}/internal/config.Version=$Version' -X '{{.PACKAGE}}/internal/config.AppName=$AppName'" ./cmd/notes.go

CMD ["notes"]

FROM scratch
COPY --from=builder /app/bin/notes .

EXPOSE 8080
EXPOSE 1323
EXPOSE 9090

ENV HOST=0.0.0.0
ENV SWAG_HOST=0.0.0.0
ENV METRICS_HOST=0.0.0.0

ENTRYPOINT ["/notes"]
