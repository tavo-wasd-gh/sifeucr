FROM golang:1.23.4

RUN apt-get update && apt-get install -y hugo && rm -rf /var/lib/apt/lists/*

WORKDIR /sifeucr

COPY . .

RUN go mod tidy

RUN hugo && go build -o sifeucr

EXPOSE 8080

CMD ["./sifeucr"]
