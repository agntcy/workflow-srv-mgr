# Copyright AGNTCY Contributors (https://github.com/agntcy)
# SPDX-License-Identifier: Apache-2.0

include ../scripts/main.mk


export TEST_PKGS=

GEN_DIR = manifests

OPENAPI_DESCRIPTOR = spec/manifest.yaml

OPENAPI_GENERATOR_VERSION = v7.0.1
OPENAPI_GENERATOR = openapitools/openapi-generator-cli:v7.0.1

OPENAPI_GENERATOR_REPOSITORY=openapitools/openapi-generator-cli
#OPENAPI_GENERATOR_REPOSITORY=harbor.eticloud.io/gbear-dev/openapi-generator-cli
OPENAPI_GENERATOR_IMAGE=${OPENAPI_GENERATOR_REPOSITORY}:${OPENAPI_GENERATOR_VERSION}

GORELEASER_VERSION = 1.10.3


.PHONY: gen-dir
gen-dir:
	mkdir -p ${GEN_DIR}

# https://openapi-generator.tech/docs/generators/go-server/
.PHONY: models
models: gen-dir ## generate server stubs from openapi
	if [ -d ${GEN_DIR} ]; then rm -rf ${GEN_DIR}; fi
	@docker run --rm -v $${PWD}:/temp ${OPENAPI_GENERATOR} \
		generate \
		-g go \
		-i /temp/${OPENAPI_DESCRIPTOR} \
		-o /temp/${GEN_DIR} \
		--additional-properties=packageName=manifests \
		--additional-properties=sourceFolder=go \
		--global-property=models \
		--global-property=modelTests="false" \
		--global-property=modelDocs="false" \
		--global-property=withGoMod="false" \
		--global-property=noservice="true" \
		--global-property=supportingFiles=utils.go:client.go:error.go:configuration.go
	go fmt ./${GEN_DIR}


check-arg: ARG=argument
check-arg:
ifndef ${ARG}
	$(error ${ARG} is undefined)
endif

bin/goreleaser: bin/goreleaser-${GORELEASER_VERSION}
	@ln -sf goreleaser-${GORELEASER_VERSION} bin/goreleaser

# Note: removing the last line of the script because we install and run
# goreleaser in 2 separate steps, the last line is the execution line.
bin/goreleaser-${GORELEASER_VERSION}:
	@mkdir -p bin
	curl -sfL https://git.io/goreleaser | sed '$$ d' | TMPDIR=./bin VERSION=v${GORELEASER_VERSION} bash
	@mv bin/goreleaser $@

.PHONY: release
release: bin/goreleaser # Publish a release
	GORELEASER_LDFLAGS="$(LDFLAGS)" bin/goreleaser release ${GORELEASERFLAGS}

# release-%: TAG_PREFIX = v
release-%:
ifneq (${DRY}, 1)
#	@sed -e "s/^## \[Unreleased\]$$/## [Unreleased]\\"$$'\n'"\\"$$'\n'"\\"$$'\n'"## [$*] - $$(date +%Y-%m-%d)/g; s|^\[Unreleased\]: \(.*\/compare\/\)\(.*\)...HEAD$$|[Unreleased]: \1${TAG_PREFIX}$*...HEAD\\"$$'\n'"[$*]: \1\2...${TAG_PREFIX}$*|g" CHANGELOG.md > CHANGELOG.md.new
#	@mv CHANGELOG.md.new CHANGELOG.md

ifeq (${TAG}, 1)
#	git add CHANGELOG.md
#	git commit -m 'Prepare release $*'
	git tag -m 'Release $*' ${TAG_PREFIX}$*
ifeq (${PUSH}, 1)
	git push; git push origin ${TAG_PREFIX}$*
endif
endif
endif

	@echo "Version updated to $*!"
ifneq (${PUSH}, 1)
	@echo
	@echo "Review the changes made by this script then execute the following:"
ifneq (${TAG}, 1)
	@echo
#	@echo "git add CHANGELOG.md && git commit -m 'Prepare release $*' && git tag -m 'Release $*' ${TAG_PREFIX}$*"
	@echo "git tag -m 'Release $*' ${TAG_PREFIX}$*"
	@echo
	@echo "Finally, push the changes:"
endif
	@echo
	@echo "git push; git push origin ${TAG_PREFIX}$*"
endif

.PHONY: release-dev
release-dev: bin/goreleaser # Publish an experimental release
	GORELEASER_LDFLAGS="$(LDFLAGS)" bin/goreleaser release -f .goreleaser.dev.yml ${GORELEASERFLAGS}

.PHONY: build-chart-asset
build-chart-asset:
	tar -czvf assets/agent-chart.tar.gz charts

.DEFAULT_GOAL := help

# enable CGO for the project (instrument the build in the included makefile)
build: export CGO_ENABLED=1


