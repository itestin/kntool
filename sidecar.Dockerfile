FROM golang:1.15.4

WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 go build -mod vendor -o kntool-sidecar cmd/sidecar/main.go


FROM alpine:3.9

COPY --from=0 /workspace/kntool-sidecar /bin/kntool-sidecar

RUN echo "http://mirrors.aliyun.com/alpine/v3.9/main/" > /etc/apk/repositories
RUN apk update
RUN apk add --no-cache ca-certificates tzdata curl bash iproute2

CMD ["kntool-sidecar"]
