FROM golang:1.23
WORKDIR /app
COPY . .
RUN go build -o kubenetinsight cmd/kubenetinsight/main.go
CMD ["./kubenetinsight"]