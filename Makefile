.PHONY: build run

REPO  ?= registry.cn-hangzhou.aliyuncs.com/vprix/centos_desktop
TAG   ?= xfce_7.6


build:
	go build -o ./bin/vprix
	docker build -f ./docker/Dockerfile -t $(REPO):$(TAG) .

run:
	docker run -it --rm \
	-p 8080:8080  \
	--name centos_xfce_desktop_test \
	-v /home/lzm/GolandProjects/single-agent/bin/vprix:/bin/vprix \
	-v /home/lzm/GolandProjects/bitcloud-dockerfile/home/:/home/vprix \
	-v /home/lzm/GolandProjects/bitcloud-dockerfile/src/common/xfce:/desktop/ \
	-e "HOME=/home/vprix" \
	$(REPO):$(TAG)


exec:
	docker exec -it centos_xfce_desktop_test sh