FROM alpine

RUN apk add --no-cache busybox-extras
EXPOSE 80

CMD ["sh", "-c", "while true; do echo -e 'HTTP/1.1 200 OK\\r\\nContent-Length: 13\\r\\n\\r\\nHello, world!' | nc -l -p 80; done"]
