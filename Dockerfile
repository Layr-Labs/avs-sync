FROM golang:1.22-alpine AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod tidy && go mod verify

COPY . .

RUN go build -v -o avs-sync

FROM alpine:3.18
COPY --from=build /usr/src/app/avs-sync /usr/local/bin/avs-sync
ENTRYPOINT [ "avs-sync" ]
