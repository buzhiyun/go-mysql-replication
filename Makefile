#NAME = registry.cn-hangzhou.aliyuncs.com/septnet/gitlab-webhook
#VERSION = myoa-0.0.2
GOEXEC = go
CGO = CGO_ENABLED=0
GOARCH = GOARCH=amd64
LINUX_GOOS = GOOS=linux
WINDOWS_GOOS = GOOS=windows
LD_FLAGS = '-s -w'

.PHONY: build start push test

build-web:
	cd web && yarn run build


gen-bindata: build-web
	rm -rf ./assets
	cp -r web/dist ./assets
	go-bindata ./assets/...

build-linux:
	GOPROXY=https://goproxy.cn,direct GO111MODULE=on ${GOEXEC} mod vendor
	${CGO} ${GOARCH} ${LINUX_GOOS} ${GOEXEC} build -ldflags ${LD_FLAGS} -o bin/go-mysql-replication

build-windows:
	GOPROXY=https://goproxy.cn,direct GO111MODULE=on ${GOEXEC} mod vendor
	${CGO} ${GOARCH} ${WINDOWS_GOOS} ${GOEXEC} build -o bin/go-mysql-replication.exe


test:
#	rm -rf ./assets
#	cp -r web/dist ./assets
#	go-bindata ./assets/...
	${GOEXEC} run . -c config.yml


