FROM golang:1.27.4 AS build
COPY . /workspace
WORKDIR /workspace
# No explicit -mod flag on purpose. vendor/ is gitignored, so a bare clone (the
# cluster Job clones the repo and has none) would fail a -mod=vendor build
# outright; with no flag Go uses vendor/ when it is present and the module cache
# otherwise, which keeps the local arm on vendor and lets the cluster arm build.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /main
CMD ["/bin/bash"]

FROM alpine:3.24 AS alpine
ARG BUILD_GIT_VERSION=dev
ARG BUILD_GIT_COMMIT=none
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="Hue"
LABEL org.opencontainers.image.description="Philips Hue desired-state controller — schedules lights via time-of-day / sunrise-sunset triggers across multiple bridges"
LABEL org.opencontainers.image.vendor="Benjamin Borbe"
LABEL org.opencontainers.image.licenses="BSD-2-Clause"
LABEL org.opencontainers.image.source="https://github.com/bborbe/hue"
LABEL org.opencontainers.image.documentation="https://github.com/bborbe/hue"
LABEL org.opencontainers.image.version="${BUILD_GIT_VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${BUILD_GIT_COMMIT}"

RUN apk --no-cache add \
	ca-certificates \
	rsync \
	openssh-client \
	tzdata \
	&& rm -rf /var/cache/apk/*
COPY --from=build /main /main
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
ENV BUILD_GIT_VERSION=${BUILD_GIT_VERSION}
ENV BUILD_GIT_COMMIT=${BUILD_GIT_COMMIT}
ENV BUILD_DATE=${BUILD_DATE}
ENTRYPOINT ["/main"]
