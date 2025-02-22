.PHONY: build start stop deploy_bin logf all deploy_config
build: ## Build streamer binary
	go build -o deploy/build/pi cmd/PI/*.go

deploy_bin: ## Copy built binaries to /usr/bin/*
	@sudo cp deploy/build/pi /usr/bin/lean

deploy_services: ## Copy systemd service files to /etc/systemd/system/
	@sudo cp -R deploy/systemd/. /etc/systemd/system/

deploy_config: ## Copy systemd service files to /etc/systemd/system/
	@sudo cp -R config/ /opt/

start: ## start
	@bash deploy/scripts/start.sh

stop: ## stop
	@bash deploy/scripts/stop.sh

logf: ##log foll./piow
	@bash deploy/scripts/logf.sh

rebuild: stop all deploy_bin start ## rebuild, deploy all binaries and restart server

all: mk_build_dir build ## Build all binaries to "./deploy/build" dir

mk_build_dir:
	@mkdir -p deploy/build


