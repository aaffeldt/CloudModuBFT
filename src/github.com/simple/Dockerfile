FROM golang:1.24.3 AS builder


WORKDIR /app 

COPY github.com/zistvan/ /app/github.com/zistvan

WORKDIR /app/github.com/zistvan

RUN go build -o cli main.go 

ENTRYPOINT [ "./cli" ]

CMD [ "--id", "0", "--nodes", "10,2,3,4", "--roles", "2,1,1,0", "--clusters", "0,1,0,0" ]