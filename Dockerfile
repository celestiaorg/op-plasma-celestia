FROM golang:1.22.2-alpine3.18 as builder

WORKDIR /
COPY . op-plasma-celestia
RUN apk add --no-cache make
WORKDIR /op-plasma-celestia
RUN make da-server

FROM alpine:3.18

COPY --from=builder /op-plasma-celestia/bin/da-server /usr/local/bin/da-server

EXPOSE 3100
ENTRYPOINT ["/usr/local/bin/da-server"]
