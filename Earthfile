FROM golang:1.15-alpine
RUN apk --update --no-cache add git
WORKDIR /mailtrap

all:
  BUILD +docker

build:
  RUN apk --update --no-cache add gcc libc-dev sqlite-dev
  COPY . .
  RUN go build -o build/mailtrap ./cmd/mailtrap
  SAVE ARTIFACT build/mailtrap AS LOCAL build/mailtrap

docker:
  FROM alpine
  WORKDIR /mailtrap
  COPY +build/mailtrap .
  COPY web web
  ENTRYPOINT ["/mailtrap/mailtrap"]
  SAVE IMAGE mailtrap:latest