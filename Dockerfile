FROM golang:latest AS builder
ARG RELEASE
RUN set -x \
    && mkdir /tmp/dist \
    && GOOS="$(go env GOOS)" \
    && GOARCH="$(go env GOARCH)" \
    && curl -s -f -L -o /tmp/dist/simple-command-output-filter.tar.gz "https://github.com/joeycumines/simple-command-output-filter/releases/download/${RELEASE}/simple-command-output-filter-${RELEASE}-${GOOS}-${GOARCH}.tar.gz" \
    && curl -s -f -L -o /tmp/dist/simple-command-output-filter.tar.gz.sha256 "https://github.com/joeycumines/simple-command-output-filter/releases/download/${RELEASE}/simple-command-output-filter-${RELEASE}-${GOOS}-${GOARCH}.tar.gz.sha256" \
    && expected_sha256="$(cat /tmp/dist/simple-command-output-filter.tar.gz.sha256)" \
    && echo "$expected_sha256  /tmp/dist/simple-command-output-filter.tar.gz" | sha256sum -c - \
    && tar -C /tmp/dist -xzf /tmp/dist/simple-command-output-filter.tar.gz simple-command-output-filter \
    && echo 'nobody:x:65534:65534:Nobody:/:' >/tmp/dist/etc-passwd

FROM scratch
ARG RELEASE
LABEL org.opencontainers.image.version=${RELEASE}
COPY --from=builder --chown=0:0 --chmod=0755 /tmp/dist/simple-command-output-filter /usr/local/bin/simple-command-output-filter
COPY --from=builder --chown=0:0 --chmod=0644 /tmp/dist/etc-passwd /etc/passwd
USER 65534:65534
ENTRYPOINT ["simple-command-output-filter"]
