BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
VERSION=$(shell git tag --sort=-v:refname)
GITHASH=$(shell git rev-parse HEAD 2>/dev/null)
BUILDAT=$(shell date +%FT%T%z)
LDFLAGS="-s -w -X github.com/karust/openserp/cmd.GitHash=${GITHASH} -X github.com/karust/openserp/cmd.BuildAt=${BUILDAT} -X github.com/karust/openserp/cmd.Version=${VERSION}"

builddarwin:
	rm -rf release/${BRANCH}/${VERSION}/darwin
	GOOS=darwin GOARCH=amd64 go build -v -o release/${BRANCH}/${VERSION}/darwin/openserp -ldflags ${LDFLAGS}
	cd release/${BRANCH}/${VERSION}/darwin/ && tar -zcvf ../openserp_${VERSION}_darwin_amd64.tar.gz ./* && cd ../../

builddarwinarm:
	rm -rf release/${BRANCH}/${VERSION}/darwinarm
	GOOS=darwin GOARCH=arm64 go build -v -o release/${BRANCH}/${VERSION}/darwinarm/openserp -ldflags ${LDFLAGS}
	cd release/${BRANCH}/${VERSION}/darwinarm/ && tar -zcvf ../openserp_${VERSION}_darwin_arm64.tar.gz ./* && cd ../../

buildlinux:
	rm -rf release/${BRANCH}/${VERSION}/linux
	GOOS=linux GOARCH=amd64 go build -v -o release/${BRANCH}/${VERSION}/linux/openserp -ldflags ${LDFLAGS}
	cd release/${BRANCH}/${VERSION}/linux/ && tar -zcvf ../openserp_${VERSION}_linux_amd64.tar.gz ./* && cd ../../

buildwin:
	rm -rf release/${BRANCH}/${VERSION}/windows
	GOOS=windows GOARCH=amd64 go build -v -o release/${BRANCH}/${VERSION}/windows/openserp.exe -ldflags ${LDFLAGS}
	cd release/${BRANCH}/${VERSION}/windows/ && tar -zcvf ../openserp_${VERSION}_windows_amd64.tar.gz ./* && cd ../../

buildlinuxarm:
	rm -rf release/${BRANCH}/${VERSION}/linux-arm
	GOOS=linux GOARCH=arm64 go build -v -o release/${BRANCH}/${VERSION}/linux-arm/openserp -ldflags ${LDFLAGS}
	cd release/${BRANCH}/${VERSION}/linux-arm/ && tar -zcvf ../openserp_${VERSION}_linux_arm64.tar.gz ./* && cd ../../

buildwinarm:
	rm -rf release/${BRANCH}/${VERSION}/windows-arm
	GOOS=windows GOARCH=arm64 go build -v -o release/${BRANCH}/${VERSION}/windows-arm/openserp.exe -ldflags ${LDFLAGS}
	cd release/${BRANCH}/${VERSION}/windows-arm/ && tar -zcvf ../openserp_${VERSION}_windows_arm64.tar.gz ./* && cd ../../

# 一键编译所有平台
buildall: builddarwin builddarwinarm buildlinux buildwin buildlinuxarm buildwinarm

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
