#NAME = registry.cn-hangzhou.aliyuncs.com/septnet/gitlab-webhook
#VERSION = myoa-0.0.2
GOEXEC = go
CGO = CGO_ENABLED=0
GOARCH = GOARCH=amd64
LINUX_GOOS = GOOS=linux

.PHONY: build start push test
#
#build: build-version
#
build-web:
	cd web && yarn run build


gen-bindata: build-web
	rm -rf ./assets
	cp -r web/dist ./assets
	go-bindata ./assets/...

build-linux: gen-bindata
	${CGO} ${GOARCH} ${LINUX_GOOS} ${GOEXEC} build -o bin/go-mysql-replication
	mv  bin/go-mysql-replication bin/go-mysql-transfer
	md5 bin/*


test:
#	rm -rf ./assets
#	cp -r web/dist ./assets
#	go-bindata ./assets/...
	${GOEXEC} run . -c config.yml


