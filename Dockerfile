FROM alpine
WORKDIR /app

COPY ./dist/xlan /app/xlan

CMD ["/app/xlan", "server"]
