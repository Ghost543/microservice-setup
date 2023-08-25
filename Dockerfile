FROM golang:latest
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
COPY *.go .
RUN go mod download
RUN go build -o orders-ms

EXPOSE 8001
CMD ["/orders-ms"]