# BENZHI_README

基于 Go 实现的heritage-tree-relocation-clearance HTTP API 项目，一款后端服务，面向古树保护技术人员的移栽前审查 HTTP 服务，完整实现树体建档、根系勘查、保护方案、风险阻断整改、独立复核、现场冻结、不可变放行凭据及双向审计查询。

## 项目说明
- 项目：benzhi-project-852a5c05-898b-478d-897c-a9b36626a532
- 项目用途：面向古树保护技术人员的移栽前审查 HTTP 服务，完整实现树体建档、根系勘查、保护方案、风险阻断整改、独立复核、现场冻结、不可变放行凭据及双向审计查询。
- Go 工具链：`golang:1.22`
- 前端工具链：无

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/server -addr=127.0.0.1:19081 -selfcheck
cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-project-852a5c05-898b-478d-897c-a9b36626a532-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-project-852a5c05-898b-478d-897c-a9b36626a532-arm64 linux/arm64
docker run -it benzhi-project-852a5c05-898b-478d-897c-a9b36626a532-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/server -addr=127.0.0.1:19081 -selfcheck`
