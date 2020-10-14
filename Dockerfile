FROM alpine
WORKDIR /app

COPY ./dist/server /app/server
COPY ./dist/client /app/client

CMD ["/app/server"]
