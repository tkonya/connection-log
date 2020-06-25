FROM golang:1.14.4-alpine3.11 as builder
MAINTAINER Trevor Konya <trevor@konya.io>

#ENV EMAIL_USER youremail@gmail.com
#ENV EMAIL_PASS yourgmailpassword
#ENV EMAIL_TO emailtosendto@anything.com
#ENV PING_INTERVAL 30s
#ENV SPEED_INTERVAL 20m
#ENV EMAIL_INTERVAL 168h

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN go mod vendor
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" ./cmd/connection-logger/...

FROM alpine:3.11
COPY --from=builder /build/connection-logger /app/
COPY --from=builder /build/scripts/speedtest /app/
WORKDIR /app

CMD ["/app/connection-logger"]