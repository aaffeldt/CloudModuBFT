
FROM golang:1.25 AS builder



WORKDIR /app 

COPY ./cli/ /app

RUN mkdir /app/out

ENTRYPOINT ["bash" ,"-c"]
