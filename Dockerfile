FROM alpine
COPY article_notifier .
RUN apk update && apk add --no-cache ca-certificates
ENTRYPOINT ["./article_notifier"]
