FROM alpine:3.11.0
RUN adduser -D -u 1000 chart-registry
COPY bin/linux/amd64/chart-registry /chart-registry
RUN mkdir /charts && chown -R 1000 /charts && chmod a+x /chart-registry
USER 1000
ENTRYPOINT ["/chart-registry"]
