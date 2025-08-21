# 多阶段构建 Dockerfile for Ora2Pg-Admin

# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -w -s" \
    -a -installsuffix cgo \
    -o ora2pg-admin .

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    perl \
    perl-dbi \
    perl-dbd-oracle \
    perl-dbd-pg \
    curl \
    bash

# 创建非root用户
RUN addgroup -g 1001 -S ora2pg && \
    adduser -u 1001 -S ora2pg -G ora2pg

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/ora2pg-admin /usr/local/bin/ora2pg-admin

# 复制模板文件（如果存在）
COPY --from=builder /app/templates ./templates

# 复制文档
COPY --from=builder /app/docs ./docs

# 创建数据目录
RUN mkdir -p /data/projects /data/logs /data/output && \
    chown -R ora2pg:ora2pg /data /app

# 设置环境变量
ENV PATH="/usr/local/bin:${PATH}"
ENV ORA2PG_ADMIN_DATA_DIR="/data"

# 切换到非root用户
USER ora2pg

# 设置工作目录为数据目录
WORKDIR /data

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ora2pg-admin --version || exit 1

# 暴露端口（如果有Web界面的话）
# EXPOSE 8080

# 设置入口点
ENTRYPOINT ["ora2pg-admin"]

# 默认命令
CMD ["--help"]

# 标签
LABEL maintainer="ora2pg-admin team" \
      version="${VERSION}" \
      description="Oracle to PostgreSQL migration tool with Chinese interface" \
      org.opencontainers.image.title="Ora2Pg-Admin" \
      org.opencontainers.image.description="Oracle to PostgreSQL migration tool with Chinese interface" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.licenses="MIT"
