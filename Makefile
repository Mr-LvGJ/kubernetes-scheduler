all: local

local: fmt vet
	GOOS=linux GOARCH=amd64 go build  -o=bin/yoda-scheduler ./cmd/scheduler

build:  local
	docker build --no-cache . -t registry.cn-qingdao.aliyuncs.com/k765171999/google-samples:yoda-scheduler-2.34

push:   build
	docker push registry.cn-qingdao.aliyuncs.com/k765171999/google-samples:yoda-scheduler-2.34

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

clean: fmt vet
	sudo rm -f yoda-scheduler
