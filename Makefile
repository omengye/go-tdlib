TAG := dd1b761fda7e47f4e0275c4d319f80a04db1997f
COMMIT := dd1b761

schema-update:
	curl https://raw.githubusercontent.com/tdlib/td/${TAG}/td/generate/scheme/td_api.tl 2>/dev/null > ./data/td_api.tl

generate-json:
	go run ./cmd/generateJson/main.go \
		-version "${TAG}" \
		-output "./data/td_api.json"

generate-code:
	go run ./cmd/generateCode/main.go \
		-version "${TAG}" \
		-outputDir "./client" \
		-package client \
		-functionFile function_generated.go \
		-typeFile type_generated.go \
		-unmarshalerFile unmarshaler_generated.go \
		-versionFile version_generated.go
	go fmt ./...

git-clone-tdlib:
	git clone https://github.com/tdlib/td.git ./docker/td
	cd docker/td &&	git checkout ${COMMIT}

build-tdlib-image:
	docker build -f docker/Dockerfile -t omengye/go-tdlib:${COMMIT} ./docker