FROM golang:1.17 as goBuilder
WORKDIR /go_build
COPY _ci /go_build
ARG VERSION_TAG
ARG COMMIT_ID
ENV VERSION_TAG ${VERSION_TAG:-0.0.0}
ENV GOPROXY=https://goproxy.cn,direct
RUN go build -mod=vendor -ldflags "-X main.Version=$VERSION_TAG" -o /go/bin/ ./...
# debian release as the same as golang image
# set TimeZone as Asia/Shanghai
# set Local as zh-hans
FROM debian:bullseye

ARG VERSION_TAG
ARG COMMIT_ID
ENV VERSION_TAG ${VERSION_TAG:-0.0.0}
EXPOSE 8080
COPY --from=goBuilder /go/bin/* /usr/local/bin/

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf
ENTRYPOINT ["aacs"]
CMD ["-conf", "/data/conf"]
