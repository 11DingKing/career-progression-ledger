# Foundation 规格冻结表

| 项目 | 冻结内容 |
|---|---|
| 档位与容量 | `compact_10`；非测试生产 Go 至少 2000 行、20 文件、10 package；测试至少 1500 行；同一基线最多 10 道题 |
| 业务边界 | 职业本科毕业生离校、授权、首次就业、岗位/薪酬/能力事件、资格培训、转岗晋升、失联补录、回访、申诉更正、统计冻结、恢复审计 |
| 角色 | graduate、employer、counselor、major_admin；服务端会话可撤销、过期和退出撤销 |
| 持久化 | SQLite 真实关系数据库，版本化 migration；graduates、employers、employment_records、career_events、skills、training_records、consents、appeals、audit_events、idempotency_keys、jobs、statistic_snapshots 等关联表 |
| 事务 | 事件写入、经历合并、申诉处理、统计冻结在跨实体事务中提交；失败回滚；唯一约束和版本号保护并发 |
| 状态机 | consent active/revoked；employment draft/active/ended/disputed；appeal open/in_review/resolved/rejected；job pending/running/succeeded/failed |
| 并发 | employment version 乐观锁、SQLite busy timeout、条件更新；worker claim 使用状态条件；并发测试固定屏障，不依赖随机 sleep |
| context/错误 | HTTP request id 传递 context；service/repository 保留 `%w` 错误链；超时/取消可中止事务与 worker |
| HTTP | `/healthz`、`/readyz`、`/v1/auth/*`、`/v1/graduates`、`/v1/employment`、`/v1/events`、`/v1/appeals`、`/v1/statistics`；统一 JSON 错误和请求 ID |
| Worker | 回访任务轮询、重试退避、永久失败记录、优雅停止和恢复未完成任务 |
| 审计/隐私 | 操作者、对象、动作、结果、request id；按角色分层字段，撤回授权后限制读取；申诉保留历史事实 |
| Docker | 多阶段 Go 1.26、`WORKDIR /src`、`go build ./cmd/server`、非 root 入口；健康检查 `/healthz`、就绪检查 `/readyz` |
| 测试 | 领域状态、service 事务/幂等/并发、真实 SQLite migration/重启、HTTP 契约、worker 取消重试、分页过滤和时间边界 |
| 禁止题材 | 已核对排除清单；本项目不是电商、问卷、报表、库存、CRM、工单、博客、预约或桌面工具 |
| 后续边界容量 | 10 个独立运行时边界：授权撤回、经历合并冲突、版本竞争、事件时间冲突、审计失败回滚、幂等重放、context 取消、worker 重试、统计冻结、隐私分层 |

该表在任何源码创建前冻结，基础阶段不创建题目分支、私测、题面或答案材料。
