FROM golang:1.11-alpine3.9 AS builder

WORKDIR /go/src/github.com/jakoblorz/netpkg

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o netpkg .

FROM scratch

WORKDIR /bin

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/jakoblorz/netpkg/netpkg .

ENTRYPOINT ["netpkg"]

EXPOSE 8080

CMD ["netpkg"]