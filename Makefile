SHELL := /bin/bash

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
UMASK := 022

VERSION := 0.8.13
ITERATION := 1
REVISION := $(shell cd '$(ROOT)' && git rev-parse --short HEAD)
ENVIRONMENT ?= production
GO111MODULE := on

ARCHITECTURE := $(shell go env GOARCH)

export GO111MODULE

CGO_LDFLAGS=
GO_BUILD_TAGS=
export CGO_LDFLAGS

FPM = \
	fpm -s dir \
		--name varnishmon \
		--version '$(VERSION)' \
		--architecture '$(ARCHITECTURE)' \
		--description 'varnishmon' \
		--maintainer 'Allenta Consulting S.L. <info@allenta.com>' \
		--vendor 'Allenta Consulting S.L. <info@allenta.com>' \
		--url 'https://github.com/allenta/varnishmon' \
		--license 'BSD 2-Clause License' \
		--config-files /etc/varnish/varnishmon.yml

.PHONY: build
build: mrproper
	@( \
		set -e; \
		\
		export LD_FLAGS="\
			-X github.com/allenta/varnishmon/pkg/config.version=$(VERSION) \
			-X github.com/allenta/varnishmon/pkg/config.revision=$(REVISION) \
			-X github.com/allenta/varnishmon/pkg/config.environment=$(ENVIRONMENT) \
			-s -w"; \
		export CGO_ENABLED=1; \
		\
		echo '> Building...'; \
		for CMD in '$(ROOT)/cmd/'*; do \
			echo "- $$CMD (linux $(ARCHITECTURE))"; \
			GOOS=linux GOARCH=$(ARCHITECTURE) go build \
				-tags=$(GO_BUILD_TAGS) \
				-trimpath \
				-ldflags "$$LD_FLAGS" \
				-o build/bin/$${CMD##*/} \
				./cmd/$${CMD##*/}; \
		done; \
	)

.PHONY: fmt
fmt:
	@( \
		set -e; \
		\
		echo '> Running goimports (includes gofmt)...'; \
		FILES=$$(find '$(ROOT)/cmd' '$(ROOT)/pkg' -name '*.go'); \
		for FILE in $$(go tool goimports -l $$FILES); do \
			echo "- $$FILE"; \
			go tool goimports -w "$$FILE"; \
		done; \
	)

.PHONY: lint
lint:
	@( \
		set -e; \
		\
		if [ ! -f "$$HOME/go/bin/golangci-lint" ]; then \
			echo '> Installing golangci-lint...'; \
			curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.12.2; \
		fi; \
		\
		echo '> Running golangci-lint...'; \
		~/go/bin/golangci-lint cache clean; \
		~/go/bin/golangci-lint run --timeout 5m '$(ROOT)/cmd/...' '$(ROOT)/pkg/...'; \
	)

.PHONY: vet
vet:
	@( \
		set -e; \
		\
		echo '> Running go vet...'; \
		go vet ./...; \
	)

# See: https://github.com/golang/go/issues/73279.
.PHONY: modernize
modernize:
	@( \
		set -e; \
		\
		echo '> Running modernize...'; \
		go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@v0.47.0 -fix ./...; \
	)

TEST_PACKAGES ?= '$(ROOT)/...'
TEST_PATTERN ?= .

.PHONY: test
test:
	@( \
		set -e; \
		\
		echo '> Running tests...'; \
		go test \
			$(TEST_PACKAGES) \
			-failfast \
			-race \
			-coverprofile=$(ROOT)/coverage.txt \
			-covermode=atomic \
			-run '$(TEST_PATTERN)' \
			-tags=$(GO_BUILD_TAGS) \
			-timeout=2m; \
	)

.PHONY: mod
mod:
	@( \
		set -e; \
		\
		echo '> Adding missing and removing unused modules...'; \
		go mod tidy -compat=1.26; \
		\
		echo '> Printing module requirement graph...'; \
		go mod graph; \
	)

.PHONY: mocks
mocks:
	@( \
		set -e; \
		\
		echo '> Removing previous mocks from Git...'; \
		find '$(ROOT)/pkg' -name 'mock_*.go' -type f -exec git rm -f {} \;; \
		\
		echo '> Running mockery...'; \
		go tool mockery; \
		\
		echo '> Add new mocks to Git...'; \
		find '$(ROOT)/pkg' -name 'mock_*.go' -type f -exec git add -v {} \;; \
	)

.PHONY: dist
dist: build
	@( \
		set -e; \
		umask $(UMASK); \
		\
		[ '$(PLATFORM)' = 'resolute' -o '$(PLATFORM)' = 'noble' -o \
		    '$(PLATFORM)' = 'jammy' -o '$(PLATFORM)' = 'trixie' -o \
			'$(PLATFORM)' = 'bookworm' -o '$(PLATFORM)' = 'rhel10' -o \
			'$(PLATFORM)' = 'rhel9' ] || \
				{ echo >&2 'Invalid platform ($(PLATFORM))'; exit 1; }; \
		\
		[ '$(ENVIRONMENT)' = 'production' ] || \
				{ echo >&2 'Invalid environment ($(ENVIRONMENT))'; exit 1; }; \
		\
		echo '> Building distribution...'; \
		\
		mkdir -p \
		    '$(ROOT)/build/dist/usr/bin' \
		    '$(ROOT)/build/dist/etc/varnish' \
		    '$(ROOT)/build/dist/var/log/varnishmon' \
		    '$(ROOT)/build/dist/var/lib/varnishmon' \
		    '$(ROOT)/build/dist/usr/share/doc/varnishmon' \
		    '$(ROOT)/build/dist/usr/share/man/man1'; \
		\
		find '$(ROOT)/build/bin/' -type f ! -name helper \
			-exec cp -t '$(ROOT)/build/dist/usr/bin/' {} +; \
		\
		cp '$(ROOT)/extras/packaging/varnishmon.yml' '$(ROOT)/build/dist/etc/varnish/'; \
		\
		cp \
			'$(ROOT)/README.md' \
   		   	'$(ROOT)/LICENSE.txt' \
   		   	'$(ROOT)/CHANGELOG.md' \
			'$(ROOT)/build/dist/usr/share/doc/varnishmon/'; \
		\
		'$(ROOT)/build/bin/helper' man 1 '$(ROOT)/build/dist/usr/share/man/man1/'; \
		gzip '$(ROOT)/build/dist/usr/share/man/man1/'*.1; \
	)

ifeq ($(PLATFORM),$(filter $(PLATFORM),resolute noble jammy trixie bookworm))
	@( \
		set -e; \
		umask $(UMASK); \
		\
		mkdir -p \
		    '$(ROOT)/build/dist/etc/default' \
		    '$(ROOT)/build/dist/etc/logrotate.d' \
            '$(ROOT)/build/dist/lib/systemd/system'; \
		\
		cp '$(ROOT)/extras/packaging/varnishmon.params' '$(ROOT)/build/dist/etc/default/varnishmon'; \
		cp '$(ROOT)/extras/packaging/debian/varnishmon.logrotate' '$(ROOT)/build/dist/etc/logrotate.d/varnishmon'; \
		cp '$(ROOT)/extras/packaging/debian/varnishmon.service' '$(ROOT)/build/dist/lib/systemd/system/'; \
	)
else ifeq ($(PLATFORM),$(filter $(PLATFORM),rhel10 rhel9))
	@( \
		set -e; \
		umask $(UMASK); \
		\
		mkdir -p \
		    '$(ROOT)/build/dist/etc/sysconfig' \
		    '$(ROOT)/build/dist/etc/logrotate.d' \
            '$(ROOT)/build/dist/lib/systemd/system'; \
		\
		cp '$(ROOT)/extras/packaging/varnishmon.params' '$(ROOT)/build/dist/etc/sysconfig/varnishmon'; \
		cp '$(ROOT)/extras/packaging/redhat/varnishmon.logrotate' '$(ROOT)/build/dist/etc/logrotate.d/varnishmon'; \
		cp '$(ROOT)/extras/packaging/redhat/varnishmon.service' '$(ROOT)/build/dist/lib/systemd/system/'; \
	)
endif

	@( \
		set -e; \
		umask $(UMASK); \
		\
		NAME='varnishmon-$(VERSION)-$(ITERATION)-$(REVISION)-$(PLATFORM)-$(ARCHITECTURE)'; \
		cd '$(ROOT)/build'; \
		tar cfz "$$NAME.tgz" \
			--transform "s,^.*/,$$NAME/," \
			dist/usr/bin/* \
			dist/usr/share/doc/varnishmon/*; \
	)

.PHONY: package
package: dist
ifeq ($(PLATFORM),$(filter $(PLATFORM),resolute noble jammy trixie bookworm))
	@( \
		set -e; \
		umask $(UMASK); \
		\
		echo '> Building package...'; \
		\
		cd '$(ROOT)/build'; \
		$(FPM) -t deb \
			--iteration '$(ITERATION)+$(PLATFORM)' \
			--after-install '$(ROOT)/extras/packaging/debian/varnishmon.postinst' \
			--before-remove '$(ROOT)/extras/packaging/debian/varnishmon.prerm' \
			--depends logrotate \
			--config-files /etc/default/varnishmon \
			--config-files /etc/logrotate.d/varnishmon \
			-C '$(ROOT)/build/dist' .; \
	)
else ifeq ($(PLATFORM),$(filter $(PLATFORM),rhel10 rhel9))
	@( \
		set -e; \
		umask $(UMASK); \
		\
		echo '> Building package...'; \
		\
		cd '$(ROOT)/build'; \
		$(FPM) -t rpm \
			--iteration '$(ITERATION).$(PLATFORM)' \
			--after-install '$(ROOT)/extras/packaging/redhat/varnishmon.postinst' \
			--before-remove '$(ROOT)/extras/packaging/redhat/varnishmon.prerm' \
			--depends logrotate \
			--config-files /etc/sysconfig/varnishmon \
			--config-files /etc/logrotate.d/varnishmon \
			-C '$(ROOT)/build/dist' .; \
	)
endif

define WEB_NPM_TASK
    @( \
        set -e; \
        cd '$(ROOT)/assets/web'; \
        npm install; \
        npm run $(1); \
    )
endef

.PHONY: web-serve
web-serve:
	$(call WEB_NPM_TASK,serve)

.PHONY: web-watch
web-watch:
	$(call WEB_NPM_TASK,watch)

.PHONY: web-build
web-build:
	$(call WEB_NPM_TASK,build)

.PHONY: web-lint
web-lint:
	$(call WEB_NPM_TASK,lint)

.PHONY: web-prettier
web-prettier:
	$(call WEB_NPM_TASK,prettier)

.PHONY: mrproper
mrproper:
	@( \
		echo '> Cleaning up...'; \
		rm -rf '$(ROOT)/build'; \
		git clean -f -x -d -e .env -e assets/web/node_modules -e varnishmon.db $(ROOT); \
	)
