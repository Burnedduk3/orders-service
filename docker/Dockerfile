FROM ubuntu:25.04

RUN mkdir -p /etc/orders-service
COPY configs/config-docker.yaml /etc/orders-service/config.yaml

RUN adduser --system --uid 2000 --no-create-home orders-service && \
    chown -R orders-service /etc/orders-service

USER orders-service

WORKDIR /app

COPY --chown=orders-service --chmod=500 orders-service ./orders-service

ENTRYPOINT ["./orders-service", "server"]

