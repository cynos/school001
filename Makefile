.PHONY: build

export CGO_ENABLED=0

build:
	@rm -rf dist || true
	@mkdir dist
	@mkdir -p dist/log
	@cp -r pages/ ./dist/
	@cp -r static/ ./dist/
	@cp -r views/ ./dist/
	@cp config.json ./dist/
	@go build -a -o ./dist/main .
	@rm -rf dist.tar.gz || true
	@tar -zcvf dist.tar.gz ./dist

push:
	@scp -P 148 dist.tar.gz developer@103.139.245.180:~/budi

exec:
	@ssh -p 148 developer@103.139.245.180 " \
		cd ~/budi && \
		(rm -rf dist || true) && \
		tar -zxvf dist.tar.gz \
	" \