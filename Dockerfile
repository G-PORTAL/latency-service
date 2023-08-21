FROM harbor.g-portal.se/alpine/golang:latest

ARG RUNTIME_VERSION
ARG COMMIT_SHA1

# Import application source
COPY ./ /opt/app-root/src

# Change working directory
WORKDIR /opt/app-root/src

# Build binary for Latency Service
RUN go build -v -o "${APP_ROOT}/latency-service" cmd/run.go && \
    setcap cap_net_bind_service+ep "${APP_ROOT}/latency-service"

# Finally delete application source
RUN rm -rf /opt/app-root/src/*

VOLUME /data

EXPOSE 8080
EXPOSE 8443

RUN /usr/bin/fix-permissions ${APP_ROOT} && \
    /usr/bin/fix-permissions /data/

CMD ["/opt/app-root/latency-service"]