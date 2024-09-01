FROM alpine:latest

ENV PORT=443

COPY ./gopunweb /app/gopunweb

COPY ./static/ /app/static/

COPY ./secrets/ /app/secrets/

WORKDIR /app

RUN apk add libc6-compat

CMD ["./gopunweb"]


