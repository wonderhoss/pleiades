FROM golang:1.14.6 AS builder
COPY . /go/src/github.com/gargath/pleiades
WORKDIR /go/src/github.com/gargath/pleiades
RUN go mod download
RUN ./configure --without-ginkgo --without-linter && make pleiades

FROM alpine:latest
COPY --from=builder /go/src/github.com/gargath/pleiades/pleiades .
ENTRYPOINT ["./pleiades"]
