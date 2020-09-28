build-mac:
	go build -v --ldflags="-w" \
    		-o bin/darwin/amd64/chart-registry main.go # mac osx



build-linux: export GOOS=linux
build-linux: export GOARCH=amd64
build-linux: export CGO_ENABLED=0
build-linux:
	go build -v --ldflags="-w" \
		-o bin/linux/amd64/chart-registry main.go  # linux


build: build-linux

image: build-linux
	@docker build -t harbor-b.alauda.cn/3rdparty/chart-registry:v1.0 .
	@docker push harbor-b.alauda.cn/3rdparty/chart-registry:v1.0

build-arm64: export GOOS=linux
build-arm64: export GOARCH=arm64
build-arm64: export CGO_ENABLED=0
build-arm64:
	go build -v --ldflags="-w " \
		-o bin/linux/arm64/chart-registry main.go  # linux

container-arm64: build-arm64
container-arm64:
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker build . -t armharbor.alauda.cn/3rdparty/chart-registry:v1.0 -f Dockerfile.arm
	docker push armharbor.alauda.cn/3rdparty/chart-registry:v1.0


test:
	echo "No test now"

