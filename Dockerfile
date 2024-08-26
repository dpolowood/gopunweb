FROM alpine:latest

COPY ./gopunweb /app/gopunweb

COPY ./static/ /app/static/

WORKDIR /app

RUN apk add libc6-compat

CMD ["./gopunweb"]


