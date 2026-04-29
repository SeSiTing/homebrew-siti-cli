生成日期：2026-04-20（优化版 2026-04-22）

本文描述迁移大方向、每阶段 Apollo 路由配置和验收要点。代码改造、路由模型、工具入参见同目录 `技术改动设计-2.md`。

<callout emoji="💡" background-color="light-blue">
## 速览

**两阶段**：先把 v5 验证到位，再开始大租户拆分。大租户拆分不依赖 v4。

```
阶段一：公共索引改造（分片 20 → 200）
  新建 v5 → 公共双写 → v4→v5 回填 → 公共切读 → 验证通过 → 公共写收敛 → v4 下线
                                                    ↑
                                            v5 完整验证门禁

阶段二：大租户拆分（可按 org 独立执行，长期常态化）
  建 org 独立索引 → v5→org 回填 → 观察 → 稳定后 v5 冗余清理
```
</callout>

**数据分布演进**（左读右写，加粗=该步关键切换）：

| 步骤 | 读 | 写 |
|------|------|------|
| 初始 | v4 | v4 |
| ①-a 新建 v5 | v4 | v4 |
| ①-b 公共双写 + 回填 | v4 | v4 + v5 |
| ①-c 公共切读（验证 v5） | **v5** | v4 + v5 |
| ①-d 公共写收敛，v4 下线 | v5 | **v5** |
| ②-a 大租户 MIRROR + 回填 | v5 | v5 + org_i |
| ②-b 大租户 CUTOVER | **org_i** / v5 | v5 + org_i |
| ②-c 大租户 DEDICATED | org_i / v5\* | **org_i** / v5 |

> `org_i` = 命中阈值的大租户独立索引，其他 org 走公共；v5\* = filtered 排除 dedicated orgs。

**为什么先 v5 再 org**：

- **v5 正确性先验证再拆 org**：阶段二启动的唯一前置是「v5 已切读稳定」，大租户问题与 v5 问题彻底解耦，排查面更小。
- **v5 作为完整回滚兜底**：阶段二任意 org 出问题，可单步回退到公共 v5（DEDICATED 未清理前 v5 仍持有该 org 数据；CUTOVER 阶段 v5 持续双写）。
- **对齐长期常态**：阶段二是独立 runbook，后续新出现的大租户可随时按此流程单独拆分，不依赖任何公共索引切换。
- **代价**：v5 会写入未来被拆走的大 org 数据（最大的 20495300 占 26%）；接受这份短期冗余换可验证性和可回滚性。每个 org DEDICATED 稳定 ≥ 14 天后必须 `delete_by_query` 清理 v5 中该 org 数据，回收存储。

**核心约束**：

- 公共 v4 → v5：主分片 20 → 200；routing 规则 `orgId-objectCode` 不变。
- 大租户阈值：文档量 > 1 亿独立拆为 `objectplatform_org_{orgId}_v1`（1 分片）。
- 路由统一走 Apollo 单字段 `op.es.route.config`，读用 alias，写直配物理索引。
- 每个 org 在 `MIRROR → CUTOVER → DEDICATED` 三态间均可单步回滚。

**最终态索引清单**（全部阶段完成后，`objectplatform_v4` 已下线）：

**物理索引**：

| 类型 | 索引 | 分片 | 用途 / 承载数据 |
|------|------|------|------|
| 公共 | `objectplatform_v5` | 200/1 | 非独立拆分租户（全量 − 已 DEDICATED 的 org） |
| 大租户 | `objectplatform_org_20495300_v1` | 1/1 | 英华利 demo（首批 #1） |
| 大租户 | `objectplatform_org_20552848_v1` | 1/1 | 苏州袁记（首批 #2） |
| 大租户 | `objectplatform_org_20530526_v1` | 1/1 | 广州昆仰（首批 #3） |
| 大租户 | `objectplatform_org_20789270_v1` | 1/1 | 元宏（首批 #4） |
| 大租户 | `objectplatform_org_20917632_v1` | 1/1 | 南海大欣（首批 #5） |
| 大租户 | `objectplatform_org_20592730_v1` | 1/1 | 艾泰克（首批 #6） |
| 大租户 | `objectplatform_org_20891325_v1` | 1/1 | 芜湖悠派（首批 #7） |
| 大租户 | `objectplatform_org_{新 orgId}_v1` | 1/1 | 后续新增大租户，按需增长 |

---

## 1. 当前基线

| 项 | 值 |
|------|------|
| 读 alias | `objectplatform_read -> objectplatform_v4` |
| 写 / 物理索引 | `objectplatform_v4`（20 主分片） |
| routing | `orgId + "-" + objectCode` |

> 公共 20 → 200 主分片降低整体承载和 routing 碰撞；大租户独立索引做租户隔离，不承诺解决单 `orgId+objectCode` 巨热点，若仍有热点再单独设计 bucket routing。

## 2. 索引模板

mapping / dynamic templates / analyzer 与 `objectplatform_v4` 完全一致，仅区分匹配范围和分片数。

| 模板 | index_patterns | 主分片 | 副本 | 用途 |
|------|------|------|------|------|
| `objectplatform_common` | `objectplatform_v*` | 200 | 1 | 公共大索引 |
| `objectplatform_org` | `objectplatform_org_*_v*` | 1 | 1 | 大租户独立索引 |

## 3. Apollo 路由模型

单一 Apollo key：`op.es.route.config`。

```json
{
  "enabled": true,
  "common": {
    "read": "objectplatform_read",
    "write": ["objectplatform_v4"]
  },
  "orgs": {
    "{orgId}": {
      "enabled": true,
      "mode": "MIRROR",
      "read": "objectplatform_org_{orgId}_read",
      "write": ["objectplatform_org_{orgId}_v1"]
    }
  }
}
```

**mode 语义**：

| mode | 普通读 | 写目标 | 适用阶段 |
|------|------|------|------|
| `COMMON` | `common.read` | `common.write` | 初始 / 预登记 |
| `MIRROR` | `common.read` | `common.write ∪ org.write` | 镜像写，读仍走公共 |
| `CUTOVER` | `org.read` | `common.write ∪ org.write` | 大租户切读，保留公共回滚 |
| `DEDICATED` | `org.read` | `org.write` | 最终态 |

**兼容**：`enabled=true` 以本字段为准；`enabled=false` 或缺失则回退到旧的 `op.es.index.read/write`。迁移完成后删除旧配置。

## 4. 双写一致性口径

- **时间点**：Apollo 双写发布时记录 `T0`（公共）或 `T_org`（大租户）。
- **顺序**：**先开双写并记录时间点 → 再全量回填 → 最后按时间点做增量追平**，严格按序。
- **追平窗口**：`updatedAt >= T - 60s`，避免时钟抖动和 scroll 快照边界丢数。
- **版本保护**：同步工具须按 `updatedAt` 或 `_version` 避免旧覆盖新；否则必须先全量后增量、不允许并发。
- **MIRROR 发布前窗口**：独立索引建好到 Apollo 发布之间的增量仅写公共，回填 + `T_org` 追平会覆盖，无需特殊处理。

## 5. `objectplatform_all` 汇总 alias

仅供运维/统计，不进业务代码。**每个业务文档只能被其覆盖一次**。

| 阶段 | 指向 |
|------|------|
| 初始 | `objectplatform_v4` |
| 公共切读后 | `objectplatform_v5` |
| 出现 `DEDICATED` org | `objectplatform_v5`（filtered 排除所有 dedicated orgs）∪ 各 dedicated 独立索引 |

多 dedicated org 每次都要重建 filtered alias，建议维护脚本自动生成。

---

## 6. 阶段一：公共索引改造

**目标**：v4 → v5，分片 20 → 200，完成公共索引切读、写收敛、v4 下线。本阶段不涉及任何租户独立索引。

### 6.1 建 v5

1. 创建模板 `objectplatform_common`、`objectplatform_org`。
2. 创建 `objectplatform_v5`（200 主分片 / 1 副本，mapping 从 v4 复制）。
3. 补齐 alias `objectplatform_all -> objectplatform_v4`。
4. Apollo 显式发布 `op.es.route.config`（默认值接管路由）。

### 6.2 公共双写 + 回填

1. Apollo `common.write = ["v4", "v5"]`，记录 `T0`。
2. 同步工具：**源 v4 → 目标 v5**，无过滤全量。
3. 全量结束后追平：同向，`updatedAt >= T0 - 60s`。

### 6.3 公共切读（v5 正确性门禁）

1. 低峰切 alias：`objectplatform_read`、`objectplatform_all` 从 v4 → v5。
2. Apollo 保持公共双写，观察不少于一个完整业务高峰周期。
3. **此步通过是阶段二启动的唯一前置**，未通过不进入阶段二。

### 6.4 公共写收敛 + v4 下线

1. v5 充分稳定后 Apollo `common.write = ["v5"]`。
2. 观察稳定后，v4 停写 → 保留只读一段时间 → 最终下线。

### 阶段一验收

| 维度 | 标准 |
|------|------|
| 数据一致性 | v4/v5 总 count 一致；Top100 objectCode、Top20 org、Top20 routing 聚合一致；抽样 `_id` 同 routing 下 `_source` 一致 |
| 切读稳定 | 切读后错误率不升高，p95/p99 不明显劣化 |
| 写入健康 | 双写期间无持续 bulk reject、无显著写延迟上升 |
| 回滚演练 | 切读后能通过 alias 一键切回 v4 |

### 阶段一回滚

| 时机 | 回滚方式 |
|------|------|
| 双写回填期 | `common.write = ["v4"]`，必要时清空 v5 重跑 |
| 切读观察期 | alias 切回 v4（双写期间 v5 是超集，直接切即可） |
| 写收敛后 | 先把 v5 → v4 增量回补，再改回双写 → 切回读 |

## 7. 阶段二：大租户拆分

**前置**：阶段一完成，v4 已下线，v5 作为唯一公共索引稳定运行。

**性质**：独立 runbook，可按 org 逐个执行，长期常态化（未来新出现的大租户同样走这套）。

**候选规则**（阈值：文档量 > 1 亿，约 43 GB）：

| 条件 | 计划 |
|------|------|
| 文档量 > 1 亿 | **拆分到 `objectplatform_org_{orgId}_v1`** |
| ≤ 1 亿 | 暂留公共 v5，稳定后再评估 |
| 已停用租户 | 先评估归档清理，避免无效搬运 |
| 业务方明确保障租户 | 叠加到候选集 |

**Top20 数据量与拆分决策**（2026-04-22 快照，底层 `objectplatform_v4`，全量 39.44 亿；执行前需重新 `_count` 复核避免漂移）：

| 排名 | orgId | 租户名称 | 文档量 | 占比 | 预估主存储 | 拆分建议 |
|------|------|------|------|------|------|------|
| 1 | 20495300 | 深圳英华利demo（赣州英华利3.0） | 10.3 亿 | 26.1% | ~449 GB | ✅ **单独拆（超大，专项窗口）** |
| 2 | 20552848 | 苏州袁记 | 2.45 亿 | 6.2% | ~107 GB | ✅ **单独拆（超大，专项窗口）** |
| 3 | 20530526 | 广州昆仰电子有限公司 | 1.62 亿 | 4.1% | ~71 GB | ✅ 单独拆 |
| 4 | 20789270 | 元宏（袁记子） | 1.47 亿 | 3.7% | ~64 GB | ✅ 单独拆 |
| 5 | 20917632 | 佛山市南海大欣针织 | 1.41 亿 | 3.6% | ~61 GB | ✅ 单独拆 |
| 6 | 20592730 | 艾泰克-高新园 | 1.20 亿 | 3.0% | ~52 GB | ✅ 单独拆 |
| 7 | 20891325 | 芜湖悠派护理用品 | 1.12 亿 | 2.9% | ~49 GB | ✅ 单独拆 |
| 8 | 20907573 | 深圳英华利 | 0.83 亿 | 2.1% | ~36 GB | 暂留 v5（建议与 20495300 同批规划） |
| 9 | 20005612 | 江苏大艺科技 | 0.82 亿 | 2.1% | ~36 GB | 暂留 v5 |
| 10 | 20453333 | 浙江德宝通讯科技 | 0.77 亿 | 1.9% | ~34 GB | 暂留 v5 |
| 11 | 20002493 | 深圳市常兴技术 | 0.77 亿 | 1.9% | ~34 GB | ⚠ 已停用，优先评估归档 |
| 12 | 20152162 | 库博标准汽车配件（苏州） | 0.60 亿 | 1.5% | ~26 GB | 暂留 v5 |
| 13 | 20437262 | 汉嫂（袁记子） | 0.51 亿 | 1.3% | ~22 GB | 暂留 v5 |
| 14 | 20278958 | 艾欧史密斯 | 0.49 亿 | 1.2% | ~21 GB | 暂留 v5 |
| 15 | 20296156 | 浙江银湖箱包 | 0.46 亿 | 1.2% | ~20 GB | ⚠ 已停用，优先评估归档 |
| 16 | 20826326 | 特充（上海）新能源 | 0.44 亿 | 1.1% | ~19 GB | 暂留 v5 |
| 17 | 20969742 | 麦德利食品（袁记子） | 0.41 亿 | 1.0% | ~18 GB | 暂留 v5 |
| 18 | 20219347 | Mega Management Services | 0.39 亿 | 1.0% | ~17 GB | 暂留 v5 |
| 19 | 20990892 | 美琪（九江）婴儿用品 | 0.38 亿 | 1.0% | ~17 GB | 暂留 v5 |
| 20 | 20317466 | 蜡笔小新（福建）食品工业 | 0.33 亿 | 0.8% | ~14 GB | 暂留 v5 |
| — | 其他 | 所有剩余租户合计 | 12.7 亿 | 32.1% | ~554 GB | — |

**首批候选清单（7 个，合计约 19.57 亿文档，占总量 49.6%）**：

```
20495300 (英华利demo) ─ 超大，独立批次
20552848 (苏州袁记)   ─ 超大，独立批次
20530526 (广州昆仰)
20789270 (元宏)
20917632 (南海大欣)
20592730 (艾泰克)
20891325 (芜湖悠派)
```

**附加规划**：

- **袁记集团**（20552848 / 20789270 / 20437262 / 20969742，合计 4.84 亿）：建议按集团统一时间窗口，便于数据一致性校验。
- **英华利集团**（20495300 / 20907573，合计 11.13 亿）：20907573 虽 < 1 亿，建议与 20495300 同批拆分，避免跨窗口数据漂移。
- **已停用租户**（20002493 / 20296156，合计 1.23 亿）：迁移前优先评估归档清理，归档后直接剔除出候选集。

### 7.1 单个 org 执行步骤

| # | 动作 | 读 | 写 |
|---|------|------|------|
| 1 | 建独立索引 + 读 alias `objectplatform_org_{orgId}_read` | v5 | v5 |
| 2 | Apollo 切 `MIRROR`，记录 `T_org` | v5 | v5 + org_i |
| 3 | 全量回填：**源 v5（`orgId` 过滤）→ 目标 org 独立索引** | 同上 | 同上 |
| 4 | 增量追平：同向，`orgId + updatedAt >= T_org - 60s` | 同上 | 同上 |
| 5 | 验数通过 | 同上 | 同上 |
| 6 | 切 `CUTOVER` | org_i | v5 + org_i |
| 7 | 观察一个完整业务高峰周期 | org_i | v5 + org_i |
| 8 | 切 `DEDICATED`，同步更新 `objectplatform_all`（见 §5） | org_i | org_i |

### 7.2 阶段二验收（单 org 维度）

1. v5 中该 org 与独立索引 count 一致。
2. 该 org 的 `objectCode` 聚合、最近增量一致。
3. 新增/更新/删除在公共写和独立写均生效。
4. `CUTOVER` 后该 org 读写错误率不升高，p95/p99 不明显劣化。

### 7.3 阶段二回滚（单 org 维度，按反序）

| 状态 | 回滚方式 |
|------|------|
| `DEDICATED` | 独立索引 → v5 回补缺失增量 → 改回 `CUTOVER` |
| `CUTOVER` | 改回 `MIRROR`，读立即回到 v5 |
| `MIRROR` | 移除该 org 配置，回到公共，必要时清空独立索引重跑 |

## 8. 阶段三：v5 冗余清理

**背景**：阶段一全量回填时 v5 包含所有 org 的数据（含后续被拆到独立索引的大 org）；DEDICATED 之后这部分数据是冗余的，必须清理以回收存储和降低公共索引压力。正确性由 `objectplatform_all` filtered alias 保证，清理前后读写均不受影响。

**触发**：每个 dedicated org 稳定运行 ≥ 14 天、确认无回滚需求后，按 org 独立触发，不阻塞其他 org 的拆分进度。

**步骤**：

1. 确认该 org 已进入 `DEDICATED` 且观察期达标。
2. 分批 `delete_by_query { term: { orgId } }`，按分片并发和批量控制限速，调度到低峰。
3. 清理完成后 `_forcemerge` 回收磁盘。
4. 更新 `objectplatform_all` 口径检查（见 §5），确保无重复和漏数。

**风险控制**：`delete_by_query` 代价高，必须限速；不在关键路径上，允许多次分批执行。

## 9. 待办

| P | 事项 |
|---|------|
| **阶段一** | |
| P0 | 创建模板 `objectplatform_common`、`objectplatform_org` |
| P0 | 创建 v5，补齐 `objectplatform_all` |
| P0 | 公共双写 + v4→v5 全量回填 + 增量追平 |
| P0 | 公共切读 + 稳定观察 + 回滚演练 |
| P0 | 公共写收敛 + v4 下线 |
| **阶段二** | |
| P1 | Top20 复核 + 首批大租户名单 |
| P1 | 单 org 拆分 runbook（MIRROR → 回填 → 追平 → CUTOVER → DEDICATED） |
| P1 | 逐个大租户按 runbook 执行 |
| P2 | `objectplatform_all` 自动维护脚本 |
| **阶段三** | |
| P1 | dedicated org 稳定后 `delete_by_query` 清理 v5 冗余 |
