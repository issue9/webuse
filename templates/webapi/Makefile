.PHONY: gen build install watch test

ROOT = .

BIN = ./webapi

# 生成中间代码
gen:
	go generate $(ROOT)/...

# 编译项目内容
build:
	go build -o=$(BIN) -v $(ROOT)

# 安装依赖
install:
	go mod download

watch:
	web watch -app=-a=serve $(ROOT)

# 执行测试内容
test:
	go test ./... -count=1 -p=1 -parallel=1
