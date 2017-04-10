# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Usage:
# 	[PREFIX=gsosa/php-fpm_exporter] [ARCH=amd64] [TAG=1.1.0] make (server|container|push)


all: push

TAG=${TRAVIS_TAG}
PREFIX?=gsosa/php-fpm_exporter
ARCH?=amd64
GOLANG_VERSION=1.8
TEMP_DIR:=$(shell mktemp -d)

server: main.go metrics.go
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) GOARM=6 go build -a -installsuffix cgo -ldflags '-w -s' -o php-fpm_exporter

container:
	# Compile the binary inside a container for reliable builds
	docker pull golang:$(GOLANG_VERSION)
	docker run --rm -it -v $(PWD):/go/src/php-fpm_exporter golang:$(GOLANG_VERSION) /bin/bash -c "make -C /go/src/php-fpm_exporter server ARCH=$(ARCH)"

build:
ifneq ($(TAG),)
	docker build --pull -t $(PREFIX):$(TAG) -t $(PREFIX):latest .
else
	docker build --pull -t $(PREFIX):master .
endif

push: server build
	docker login -u="$(DOCKER_USERNAME)" -p="$(DOCKER_PASSWORD)"
ifneq ($(TAG),)
	docker push $(PREFIX):$(TAG)
	docker push $(PREFIX):latest
else
	docker push $(PREFIX):master
endif

clean:
	rm -f php-fpm_exporter
