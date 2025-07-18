FROM alpine:latest

RUN apk add --no-cache build-base

COPY bomb.c /bomb.c

RUN gcc -o /mem-bomb /bomb.c

CMD ["/mem-bomb"]
