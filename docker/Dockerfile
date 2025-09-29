FROM ubuntu:25.04

RUN mkdir -p /etc/orders-service
COPY configs/config-docker.yaml /etc/orders-service/config.yaml

RUN useradd -r -u 2000 -s /bin/false orders-service && \
    chown -R orders-service /etc/orders-service

USER orders-service

WORKDIR /app

COPY --chown=orders-service --chmod=500 orders-service ./orders-service

ENTRYPOINT ["./orders-service", "server"]

