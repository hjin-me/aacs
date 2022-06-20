FROM golang:1.17 as goBuilder
WORKDIR /go_build
COPY . /go_build
ARG VERSION_TAG
ARG COMMIT_ID
ENV VERSION_TAG ${VERSION_TAG:-0.0.0}
ENV GOPROXY=https://goproxy.cn,direct
RUN go build -ldflags "-X main.Version=$VERSION_TAG" -o /go/bin/ ./...
# debian release as the same as golang image
# set TimeZone as Asia/Shanghai
# set Local as zh-hans
FROM debian:bullseye
RUN set -ex; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
	    tzdata \
	    locales \
	    ca-certificates;
RUN locale-gen zh_CN.UTF-8; \
    update-locale zh_CN.UTF-8;
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime;
ENV TZ Asia/Shanghai
ENV LANG zh_US.utf8

ARG VERSION_TAG
ARG COMMIT_ID
ENV VERSION_TAG ${VERSION_TAG:-0.0.0}
EXPOSE 8080
COPY --from=goBuilder /go/bin/* /usr/local/bin/

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf
ENTRYPOINT ["aacs"]
CMD ["-conf", "/data/conf/config.yaml"]
