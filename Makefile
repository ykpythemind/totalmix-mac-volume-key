BINARY_NAME = totalmix-mac-volume-key
INSTALL_DIR = /usr/local/bin
PLIST_NAME = com.ykpythemind.totalmix-mac-volume-key.plist
PLIST_DIR = $(HOME)/Library/LaunchAgents

.PHONY: build install uninstall start stop restart

build:
	go build -o $(BINARY_NAME) .

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	mkdir -p $(PLIST_DIR)
	sed 's|__BINARY_PATH__|$(INSTALL_DIR)/$(BINARY_NAME)|g' $(PLIST_NAME) > $(PLIST_DIR)/$(PLIST_NAME)
	launchctl load $(PLIST_DIR)/$(PLIST_NAME)
	@echo "Installed and started. Check: launchctl list | grep totalmix"

uninstall: stop
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -f $(PLIST_DIR)/$(PLIST_NAME)
	@echo "Uninstalled."

start:
	launchctl load $(PLIST_DIR)/$(PLIST_NAME)

stop:
	-launchctl unload $(PLIST_DIR)/$(PLIST_NAME)

restart: stop start
