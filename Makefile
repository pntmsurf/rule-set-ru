MIHOMO_VERSION := v1.19.29
SINGBOX_VERSION := v1.9.3

ROOT_DIR := $(abspath .)
WORK_DIR := $(ROOT_DIR)/.build
RELEASE_DIR := $(ROOT_DIR)/release

MIHOMO_BIN := $(WORK_DIR)/mihomo
SINGBOX_DIR := $(WORK_DIR)/sing-box-$(patsubst v%,%,$(SINGBOX_VERSION))-linux-amd64
SINGBOX_BIN := $(SINGBOX_DIR)/sing-box

.PHONY: all prepare build mihomo singbox xray clean

all: build

prepare: $(MIHOMO_BIN) $(SINGBOX_BIN)

$(MIHOMO_BIN):
	mkdir -p "$(WORK_DIR)"
	wget -q -O "$(WORK_DIR)/mihomo.gz" \
	"https://github.com/MetaCubeX/mihomo/releases/download/$(MIHOMO_VERSION)/mihomo-linux-amd64-$(MIHOMO_VERSION).gz"
	gunzip -f "$(WORK_DIR)/mihomo.gz"
	 chmod +x "$@"

$(SINGBOX_BIN):
	mkdir -p "$(WORK_DIR)"
	wget -q -O "$(WORK_DIR)/sing-box.tar.gz" \
	"https://github.com/SagerNet/sing-box/releases/download/$(SINGBOX_VERSION)/sing-box-$(patsubst v%,%,$(SINGBOX_VERSION))-linux-amd64.tar.gz"
	tar -xzf "$(WORK_DIR)/sing-box.tar.gz" -C "$(WORK_DIR)"

mihomo: prepare
	mkdir -p "$(RELEASE_DIR)"
	cd "$(ROOT_DIR)" && go run ./cmd/mihomo \
	-root="$(ROOT_DIR)" \
	-mihomo-bin="$(MIHOMO_BIN)"

singbox: prepare
	mkdir -p "$(RELEASE_DIR)"
	cd "$(ROOT_DIR)" && go run ./cmd/singbox \
	-root="$(ROOT_DIR)" \
	-singbox-bin="$(SINGBOX_BIN)"

xray:
	mkdir -p "$(RELEASE_DIR)"
	cd "$(ROOT_DIR)" && go run ./cmd/xraycore \
	-root="$(ROOT_DIR)"

build: mihomo singbox xray

clean:
	rm -rf "$(WORK_DIR)"
	rm -f "$(RELEASE_DIR)"/*.mrs
	rm -f "$(RELEASE_DIR)"/*.srs
	rm -f "$(RELEASE_DIR)"/*.dat