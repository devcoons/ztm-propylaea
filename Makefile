cnf ?= Makefile.env
include $(cnf)

build:
	docker build -f Dockerfile -t $(IMG_NAME):$(IMG_TAG) .	

build-debug: 
	docker build -f Dockerfile.debug -t $(IMG_NAME).debug:$(IMG_TAG) .	

run:
	docker run -d -v"$(CURDIR)/$(APP_CFG_FILE):/tmp/config.cfg" -e IMSCFGFILE=/tmp/config.cfg -p $(APP_PORT):8080 --name $(APP_NAME) $(IMG_NAME):$(IMG_TAG)

run-debug:
	docker run -d -v"$(CURDIR)/$(APP_CFG_FILE):/tmp/config.cfg" -e IMSCFGFILE=/tmp/config.cfg -p $(APP_PORT):8080 -v"$(CURDIR)/.app:/app" --name $(APP_NAME) $(IMG_NAME).debug:$(IMG_TAG)

attach-net:
	docker network connect $(net) $(APP_NAME) 

all: build run
