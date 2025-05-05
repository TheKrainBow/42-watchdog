# Makefile at the root of your project

# Binaries
SERVER_BIN := watchdog-server
CLIENT_BIN := watchdog-client

GOBIN ?= $(shell go env GOPATH)/bin

# Service
SYSTEMD_FILE := /etc/systemd/system/watchdog.service
USER := $(shell whoami)
WORKDIR := $(shell pwd)
INSTALL_PATH := /usr/local/bin
CONFIG_DST := /etc/watchdog/config.yml
LOGFILE := /var/log/watchdog.log

# Paths
SERVER_DIR := ./cmd/server
CLIENT_DIR := ./cmd/client

help:
	@echo "📦  Watchdog Makefile commands:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: $(SERVER_BIN) $(CLIENT_BIN) ## Build server and client binaries localy

re: clean all ## Clean and rebuild server/client binaries localy

$(SERVER_BIN): ## Build server binary localy
	go build -o $(SERVER_BIN) $(SERVER_DIR)

$(CLIENT_BIN): ## Build client binary localy
	go build -o $(CLIENT_BIN) $(CLIENT_DIR)

clean: ## Clean local binaries
	rm -f $(SERVER_BIN) $(CLIENT_BIN)

server-install: $(SERVER_BIN) ## Install watchdog service
	sudo mkdir -p /etc/watchdog
	sudo cp config.yml $(CONFIG_DST)
	sudo touch $(LOGFILE)
	sudo chown $(USER) $(LOGFILE)
	sudo cp $(SERVER_BIN) $(INSTALL_PATH)/$(SERVER_BIN)

	@echo "Creating systemd service file $(SYSTEMD_FILE) ..."
	@echo
	@echo "[Unit]" | sudo tee $(SYSTEMD_FILE) > /dev/null
	@echo "Description=Watchdog Server Service" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "After=network.target" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "[Service]" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "Type=simple" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "ExecStart=$(INSTALL_PATH)/$(SERVER_BIN) $(CONFIG_DST) $(LOGFILE)" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "Restart=on-failure" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "User=$(USER)" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "WorkingDirectory=$(WORKDIR)" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "[Install]" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@echo "WantedBy=multi-user.target" | sudo tee -a $(SYSTEMD_FILE) > /dev/null
	@cat $(SYSTEMD_FILE)
	@echo

	sudo systemctl daemon-reload
	@echo "✅ Server installed. Use 'make server-start' to start it."

server-start: ## Start and enable watchdog service
	sudo systemctl start watchdog.service
	sudo systemctl enable watchdog.service

server-stop: ## Stop watchdog service
	sudo systemctl stop watchdog.service
	sudo systemctl disable watchdog.service

server-restart: ## Restart watchdog service
	sudo systemctl restart watchdog.service

server-reload: ## Reload server using watchdog.service file, without stopping the service
	sudo systemctl daemon-reload

server-status: ## Display server service status
	systemctl status watchdog.service

server-logs: ## Display server logs in real time
	journalctl -u watchdog.service -f

server-delete: ## Delete watchdog.service, service file and config. (Keep logs)
	@echo "Stopping and disabling watchdog service..."
	-sudo systemctl stop watchdog.service
	-sudo systemctl disable watchdog.service
	-sudo rm -f $(SYSTEMD_FILE)
	-sudo rm -f $(INSTALL_PATH)/$(SERVER_BIN)
	-sudo rm -f $(CONFIG_DST)
	sudo systemctl daemon-reload
	@echo "✅ Watchdog server uninstalled (logs kept at $(LOGFILE))"

cron-setup:  ## Install crontab to start/notify/stop server daily
	@echo "Installing cron jobs for watchdog-client..."
	@crontab -l 2>/dev/null | grep -v 'watchdog-client' > cron.tmp || true
	@echo "30 7  * * * /usr/local/bin/watchdog-client start" >> cron.tmp
	@echo "30 19 * * * /usr/local/bin/watchdog-client notify" >> cron.tmp
	@echo "30 20 * * * /usr/local/bin/watchdog-client stop --post-attendance" >> cron.tmp
	@crontab cron.tmp
	@rm cron.tmp
	@echo "✅ Cron jobs installed:"
	@crontab -l | grep watchdog-client

cron-remove:  ## Remove all watchdog-client-related cron jobs
	@echo "Removing watchdog-client cron jobs..."
	@crontab -l 2>/dev/null | grep -v 'watchdog-client' > cron.tmp || true 
	@crontab cron.tmp
	@rm cron.tmp
	@echo "✅ Cron jobs cleaned"
