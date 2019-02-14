FROM golang:alpine
LABEL maintainer Hugo Josefson <hugo@josefson.org> (https://www.hugojosefson.com/)

COPY egos.go /
RUN go build -o /egos /egos.go
RUN rm /egos.go

ENTRYPOINT ["/egos"]

