# 使用场景示例

本文档提供了不同场景下使用 Ora2Pg-Admin 的具体示例和最佳实践。

## 场景1：开发环境迁移

### 背景
- 小型开发数据库
- 数据量较小（< 1GB）
- 结构相对简单
- 允许停机时间较长

### 操作步骤

1. **创建项目**
```bash
ora2pg-admin 初始化 --template=basic --description="开发环境迁移" dev-migration
cd dev-migration
```

2. **配置数据库连接**
```bash
# 设置环境变量
export ORACLE_PASSWORD="dev_password"
export PG_PASSWORD="dev_password"

# 配置连接
ora2pg-admin 配置 数据库
```

3. **检查环境**
```bash
ora2pg-admin 检查 环境
ora2pg-admin 检查 连接
```

4. **执行迁移**
```bash
# 一次性完整迁移
ora2pg-admin 迁移 全部
```

### 配置要点
```yaml
migration:
  types:
    - "TABLE"
    - "VIEW"
    - "SEQUENCE"
    - "COPY"
  parallel_jobs: 2
  batch_size: 1000
```

## 场景2：生产环境迁移

### 背景
- 大型生产数据库
- 数据量巨大（> 100GB）
- 复杂的业务逻辑
- 停机时间要求极短

### 操作步骤

1. **创建项目**
```bash
ora2pg-admin 初始化 --template=advanced --description="生产环境迁移" prod-migration
cd prod-migration
```

2. **分阶段配置**
```bash
# 配置数据库连接
ora2pg-admin 配置 数据库

# 配置迁移选项
ora2pg-admin 配置 选项
```

3. **分阶段执行**
```bash
# 第一阶段：结构迁移（可在业务低峰期执行）
ora2pg-admin 迁移 结构

# 第二阶段：数据迁移（需要停机）
ora2pg-admin 迁移 数据 --parallel=8 --timeout=6h
```

### 配置要点
```yaml
migration:
  parallel_jobs: 8
  batch_size: 10000
  
advanced:
  data_export:
    commit_count: 50000
    disable_triggers: true
    disable_fkey: true
```

## 场景3：部分数据迁移

### 背景
- 只迁移特定的表或模式
- 需要数据过滤
- 排除敏感数据

### 操作步骤

1. **创建项目**
```bash
ora2pg-admin 初始化 --template=custom --description="部分数据迁移" partial-migration
cd partial-migration
```

2. **自定义配置**
编辑 `.ora2pg-admin/config.yaml`：
```yaml
advanced:
  filters:
    include_tables: "^(USERS|ORDERS|PRODUCTS).*"
    exclude_tables: "^(TEMP_|LOG_|AUDIT_).*"
    exclude_columns: "^(PASSWORD|SSN|CREDIT_CARD).*"
```

3. **执行迁移**
```bash
ora2pg-admin 迁移 全部
```

## 场景4：测试和验证

### 背景
- 迁移前的测试验证
- 数据一致性检查
- 性能对比测试

### 操作步骤

1. **预览模式**
```bash
# 使用预览模式查看迁移计划
ora2pg-admin 迁移 全部 --dry-run
```

2. **小批量测试**
```bash
# 先迁移少量数据进行测试
ora2pg-admin 配置 选项
# 设置 batch_size: 100, parallel_jobs: 1
ora2pg-admin 迁移 结构
```

3. **验证结果**
```bash
# 启用详细验证
ora2pg-admin 迁移 数据 --validate --verbose
```

## 场景5：增量迁移

### 背景
- 需要多次迁移
- 保持数据同步
- 最小化停机时间

### 操作步骤

1. **初始迁移**
```bash
# 第一次完整迁移
ora2pg-admin 迁移 全部
```

2. **增量迁移**
```bash
# 配置增量迁移
# 编辑配置文件，设置时间戳过滤
ora2pg-admin 迁移 数据 --resume
```

### 配置要点
```yaml
advanced:
  filters:
    # 使用时间戳过滤增量数据
    include_tables: "WHERE LAST_MODIFIED > '2024-01-01'"
```

## 场景6：跨网络迁移

### 背景
- 源数据库和目标数据库在不同网络
- 需要通过跳板机或VPN
- 网络带宽限制

### 操作步骤

1. **网络配置**
```bash
# 设置代理或隧道
ssh -L 1521:oracle-server:1521 jumphost
ssh -L 5432:postgres-server:5432 jumphost
```

2. **调整配置**
```yaml
oracle:
  host: "localhost"  # 通过隧道连接
  port: 1521

postgresql:
  host: "localhost"
  port: 5432

migration:
  parallel_jobs: 2   # 降低并发度
  batch_size: 500    # 减小批处理大小
```

3. **执行迁移**
```bash
ora2pg-admin 迁移 全部 --timeout=12h
```

## 场景7：容器化部署

### 背景
- 在Docker容器中运行迁移
- 需要持久化配置和数据
- 自动化部署

### Dockerfile示例
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ora2pg-admin

FROM alpine:latest
RUN apk --no-cache add ca-certificates perl
WORKDIR /root/
COPY --from=builder /app/ora2pg-admin .
COPY --from=builder /app/templates ./templates
CMD ["./ora2pg-admin"]
```

### Docker Compose示例
```yaml
version: '3.8'
services:
  ora2pg-admin:
    build: .
    volumes:
      - ./projects:/projects
      - ./logs:/logs
    environment:
      - ORACLE_PASSWORD=${ORACLE_PASSWORD}
      - PG_PASSWORD=${PG_PASSWORD}
    working_dir: /projects
```

### 操作步骤
```bash
# 构建镜像
docker build -t ora2pg-admin .

# 运行迁移
docker run -it --rm \
  -v $(pwd)/projects:/projects \
  -v $(pwd)/logs:/logs \
  -e ORACLE_PASSWORD=secret \
  -e PG_PASSWORD=secret \
  ora2pg-admin 迁移 全部
```

## 场景8：自动化CI/CD集成

### 背景
- 集成到CI/CD流水线
- 自动化测试和部署
- 版本控制和回滚

### GitLab CI示例
```yaml
stages:
  - validate
  - migrate
  - verify

variables:
  PROJECT_NAME: "ci-migration"

validate:
  stage: validate
  script:
    - ora2pg-admin 检查 环境
    - ora2pg-admin 检查 连接
  only:
    - main

migrate:
  stage: migrate
  script:
    - ora2pg-admin 初始化 --template=basic $PROJECT_NAME
    - cd $PROJECT_NAME
    - ora2pg-admin 迁移 全部 --timeout=2h
  artifacts:
    paths:
      - $PROJECT_NAME/output/
      - $PROJECT_NAME/logs/
    expire_in: 1 week
  only:
    - main

verify:
  stage: verify
  script:
    - cd $PROJECT_NAME
    - ora2pg-admin 检查 连接
    - ./scripts/verify-data.sh
  only:
    - main
```

## 最佳实践总结

### 1. 迁移前准备
- 充分了解源数据库结构
- 评估数据量和复杂度
- 制定详细的迁移计划
- 准备回滚方案

### 2. 配置优化
- 根据硬件资源调整并行度
- 合理设置批处理大小
- 启用适当的日志级别
- 配置必要的过滤规则

### 3. 执行策略
- 优先迁移结构，再迁移数据
- 大表可以考虑分批迁移
- 在业务低峰期执行
- 实时监控迁移进度

### 4. 验证和测试
- 验证数据完整性
- 检查业务逻辑正确性
- 进行性能测试
- 准备应急预案

### 5. 生产部署
- 制定详细的切换计划
- 准备数据回滚方案
- 监控系统性能
- 及时处理问题反馈
