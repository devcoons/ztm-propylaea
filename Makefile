cnf ?= Makefile.env
include $(cnf)

ifeq ($(OS), Windows_NT)
	WHERE_WHICH ?= where
else
	WHERE_WHICH ?= which
endif

ifeq (, $(shell $(WHERE_WHICH) podman))
	DOCKER := $(shell $(WHERE_WHICH) docker)
	ifeq (, $(DOCKER))
		$(error "Neither Docker nor Podman is installed. Please install one of them.")
	endif
	CONTAINER_ENGINE := docker
else
	CONTAINER_ENGINE := podman
endif

all: build run

build:
	-$(CONTAINER_ENGINE) build -f Dockerfile -t $(IMG_NAME):$(IMG_TAG) .	

build-debug: 
	-$(CONTAINER_ENGINE) build -f Dockerfile.debug -t $(IMG_NAME).debug:$(IMG_TAG) .	

run:
	-$(CONTAINER_ENGINE) run -d -v"$(CURDIR)/$(APP_CFG_FILE):/tmp/config.cfg" -e IMSCFGFILE=/tmp/config.cfg -p $(APP_PORT):8080 --name $(APP_NAME) $(IMG_NAME):$(IMG_TAG)

run-debug:
	-$(CONTAINER_ENGINE) run -d -v"$(CURDIR)/$(APP_CFG_FILE):/tmp/config.cfg" -e IMSCFGFILE=/tmp/config.cfg -p $(APP_PORT):8080 -v"$(CURDIR)/.app:/app" --name $(APP_NAME) $(IMG_NAME).debug:$(IMG_TAG)

attach-net:
	-$(CONTAINER_ENGINE) network connect $(net) $(APP_NAME) 

attach-default-net:
	-$(CONTAINER_ENGINE) network create ztm-net 
	-$(CONTAINER_ENGINE) network create ztm-net-db 
	-$(CONTAINER_ENGINE) network connect ztm-net $(APP_NAME)

debug: build-debug run-debug attach-default-net