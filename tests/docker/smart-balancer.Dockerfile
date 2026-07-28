FROM alpine:3.22

RUN apk add --no-cache socat

EXPOSE 9000

CMD ["socat", "TCP-LISTEN:9000,fork,reuseaddr", "TCP:post_service:8082"]
