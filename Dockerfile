FROM golang:1.23.3 AS build
COPY . /workspace
WORKDIR /workspace
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags "-s" -a -installsuffix cgo -o /main
CMD ["/bin/bash"]

FROM alpine:3.20 AS alpine
RUN apk --no-cache add \
	ca-certificates \
	rsync \
	openssh-client \
	tzdata \
	&& rm -rf /var/cache/apk/*
COPY --from=build /main /main
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
ENTRYPOINT ["/main"]
