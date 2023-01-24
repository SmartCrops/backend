FROM golang:alpine
WORKDIR /build
COPY . .
RUN go build -o main
EXPOSE 8080
CMD ["./main"]