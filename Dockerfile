FROM golang:1.12-alpine3.9 AS builder
WORKDIR /build/

RUN apk --no-cache add git build-base
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN $(cd cmd/softcopy-server && go build)
RUN go test ./...

FROM alpine:3.9
COPY --from=builder /build/cmd/softcopy-server/softcopy-server /usr/bin/
COPY ./configs/docker-default.yml /etc/softcopy/config.yml
CMD ["softcopy-server"]