FROM golang:1.13 as builder
WORKDIR /workspace
ARG go_proxy=https://goproxy.cn
ENV GO111MODULE=on \
    GOPROXY=${go_proxy} \
    CGO_ENABLED=0
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/
RUN go build -ldflags '-w -s' -a -installsuffix cgo -o chart-registry main.go



FROM alpine:3.11.0
RUN adduser -D -u 1000 chart-registry
COPY --from=builder /workspace/chart-registry /chart-registry
RUN mkdir /charts && chown -R 1000 /charts && chmod a+x /chart-registry
USER 1000
ENTRYPOINT ["/chart-registry"]
