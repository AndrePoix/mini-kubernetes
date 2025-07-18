# cpu-bomb/Dockerfile
FROM alpine:latest

RUN apk add --no-cache build-base

COPY bomb.c /bomb.c

RUN gcc -O2 -o /cpu-bomb /bomb.c

CMD ["/cpu-bomb"]
