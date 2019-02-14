FROM golang:onbuild
LABEL maintainer Hugo Josefson <hugo@josefson.org> (https://www.hugojosefson.com/)

RUN mkdir /app 
WORKDIR /app 

COPY egos.go /app/ 
RUN go build -o egos egos.go

ENTRYPOINT ["/app/egos"]

