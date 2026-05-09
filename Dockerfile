# # Use phusion/baseimage as base image. To make your builds reproducible, make
# # sure you lock down to a specific version, not to `latest`!
# # See https://github.com/phusion/baseimage-docker/blob/master/Changelog.md for
# # a list of version numbers.
# FROM phusion/baseimage:<VERSION>

# # Use baseimage-docker's init system.
# CMD ["/sbin/my_init"]

# # ...put your own build instructions here...

# # Clean up APT when done.
# RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

FROM golang:alpine AS builder
WORKDIR /app
COPY . .
ARG GITHUB_SHA
ARG VERSION

# 镜像描述标签
LABEL org.opencontainers.image.description="高性能[测活、测速、媒体检测]代理检测筛选工具，支持100-1000高并发低占用运行，大幅减少数倍检测时间。"
LABEL org.opencontainers.image.keywords="subs-check,测活,测速,媒体检测,节点管理,流媒体检测,测速节点,自动化,GoReleaser,Docker,best-sub,proxy,proxies,mihomo,v2ay,clash"
LABEL org.opencontainers.image.url="https://github.com/sinspired/subs-check-pro"
LABEL org.opencontainers.image.documentation="https://github.com/sinspired/subs-check-pro/wiki"
LABEL org.opencontainers.image.source="https://github.com/sinspired/subs-check-pro"

RUN echo "Building commit: ${GITHUB_SHA:0:7}" && \
    go mod tidy && \
    go build -ldflags="-s -w -X main.Version=${VERSION} -X main.CurrentCommit=${GITHUB_SHA:0:7}" -trimpath -o subs-check .

FROM alpine
WORKDIR /app
ENV TZ=Asia/Shanghai
# 👇 显式标记运行在 Docker 中
ENV RUNNING_IN_DOCKER=true
RUN apk add --no-cache alpine-conf ca-certificates &&\
    /usr/sbin/setup-timezone -z Asia/Shanghai && \
    apk del alpine-conf && \
    rm -rf /var/cache/apk/*

COPY --from=builder /app/subs-check /app/subs-check
# shell form -> exec form
CMD ["/app/subs-check"]
EXPOSE 8199
EXPOSE 8299
