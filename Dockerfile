FROM golang:1.15-alpine AS builder
RUN apk --update --no-cache add git gcc libc-dev sqlite-dev
WORKDIR /app
COPY . .
RUN go build -o build/mailtrap ./cmd/mailtrap

FROM alpine
WORKDIR /app
COPY --from=builder app/build/mailtrap .
COPY web web
ENTRYPOINT ["/app/mailtrap"]