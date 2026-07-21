FROM library/debian:trixie-slim AS slim
COPY north_outage /usr/bin/north_outage
ENTRYPOINT ["/usr/bin/north_outage" ]
CMD ["--config","/config.yaml"]
