FROM golang:1.16 AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY main.go .

RUN go build -o tiny-echo .

FROM scratch

COPY --from=builder /build/tiny-echo .

ENTRYPOINT [ "/tiny-echo" ]