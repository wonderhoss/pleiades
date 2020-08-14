FROM gargath/combined-builder:latest AS builder
COPY . /go/src/github.com/gargath/pleiades
WORKDIR /go/src/github.com/gargath/pleiades
RUN go mod download
RUN ./configure && make release

FROM alpine:latest
COPY --from=builder /go/src/github.com/gargath/pleiades/pleiades .
ENTRYPOINT ["./pleiades"]
