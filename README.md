# 古树移栽前审查与放行服务

本项目面向古树保护技术员、根系勘查人员和移栽安全复核负责人，提供版本化 JSON HTTP API。服务将树体建档、四方位根系勘查、保护方案修订、风险阻断整改、独立复核、现场条件核验和不可变放行凭据串成一条可追溯业务链路。

所有写请求都携带 `expectedVersion` 和 `idempotencyKey`。事实以长度前缀 JSON 事件帧保存到本地，事件帧包含递增序号、前序摘要与校验和；服务启动时会完整校验并重放日志，检测到截断或篡改会拒绝启动。查询使用重放所得投影，已签发的凭据只追加、不更新。

## 构建与测试

```bash
go build ./...
go test ./...
```

## 运行服务

默认仅监听高位回环地址 `127.0.0.1:19081`，事件数据写入 `./data`：

```bash
go run ./cmd/server
```

可以显式指定回环地址和数据目录：

```bash
go run ./cmd/server -addr=127.0.0.1:19443 -data-dir=./runtime-data
```

未提供 `-addr` 时，也可通过 `PORT` 指定端口，服务会绑定 `127.0.0.1:<PORT>`；`-addr` 的优先级高于 `PORT`。服务拒绝未显式授权的非回环地址、低于 `1024` 的端口和非法端口。

## 完整自检

以下命令会创建隔离的临时事件存储，在真实回环监听上通过 HTTP 完成一条含多个阻断项、整改复核、现场冻结和凭据查询的流程，验证审计轨迹后主动优雅退出：

```bash
go run ./cmd/server -addr=127.0.0.1:19081 -selfcheck
```

主要路由以 `/api/v1` 开头，包含案卷、`root-surveys`、`protection-plans`、`risk-reviews`、发现项整改与复核、`site-verifications`、`credentials` 和 `audit`。完整审计轨迹既可按案卷编号查询，也可通过 `/api/v1/clearance-credentials/{credentialID}/audit` 按凭据编号查询。请求必须使用 `Content-Type: application/json`；未知字段及超过 1 MiB 的请求体会被拒绝。
