FROM alpine:3.22

RUN mkdir -p /etc/orders-service
COPY configs/config-docker.yaml /etc/orders-service/config.yaml

RUN adduser -u 2000 -S orders-service && \
    chown -R orders-service /etc/orders-service

USER orders-service

WORKDIR /app

COPY --chown=orders-service --chmod=500 orders-service ./orders-service

ENTRYPOINT ["./orders-service", "server"]

