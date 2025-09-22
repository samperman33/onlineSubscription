FROM golang:1.24
COPY . /app
WORKDIR /app
RUN go build cmd/onlineSubscription/main.go
EXPOSE 8080
ENTRYPOINT ["/app/main"]

