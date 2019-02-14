FROM golang as builder

RUN mkdir -p /app 
WORKDIR /app 

COPY egos.go /app/ 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o egos egos.go

###############################################################################
FROM scratch as runtime
LABEL maintainer Hugo Josefson <hugo@josefson.org> (https://www.hugojosefson.com/)

COPY --from=builder /app/egos /egos

ENTRYPOINT ["/egos"]

