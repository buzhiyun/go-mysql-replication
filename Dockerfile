FROM node:14.16.1-slim as web-build
# 打包 前端
ADD . /app
WORKDIR /app

# 生成到 /app/web/dist
RUN  yarn config set registry https://registry.npm.taobao.org && \
  cd web &&  yarn install && yarn run build


# 生成最终运行程序
FROM golang:1.14-alpine3.13 as compiler
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0
ADD . /app
COPY --from=web-build /app/web/dist /app/assets
WORKDIR /app
# 这里我比较懒，还是直接把前端的包bindata到程序里吧
RUN ls -al /app/assets && go get -u github.com/go-bindata/go-bindata/... && \
  go-bindata ./assets/... && go build -ldflags '-s -w' -o bin/go-mysql-replication && cp config.yml bin/


# 最终镜像
FROM alpine:3.13
WORKDIR /app
COPY --from=compiler /app/bin /app
CMD ["/app/go-mysql-replication","-c","config.yml"]
