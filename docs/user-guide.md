# Ora2Pg-Admin 用户使用指南

## 概述

Ora2Pg-Admin 是一个基于 Go 语言开发的中文命令行工具，旨在简化 Oracle 到 PostgreSQL 数据库迁移的运维操作。它提供了友好的中文界面、自动配置生成、Oracle 客户端管理、交互式向导等核心功能。

## 系统要求

### 基础环境
- 操作系统：Windows 10+、Linux、macOS
- Go 版本：1.19+ （仅开发时需要）

### 依赖工具
- **Oracle 客户端**：Oracle Instant Client 或完整 Oracle 客户端
- **ora2pg**：Perl 工具，用于实际的数据库迁移
- **PostgreSQL 客户端**：psql 工具（可选，用于连接测试）

## 安装指南

### 1. 下载可执行文件
从项目发布页面下载适合您操作系统的可执行文件：
- Windows: `ora2pg-admin.exe`
- Linux: `ora2pg-admin`
- macOS: `ora2pg-admin`

### 2. 安装 Oracle 客户端
根据您的操作系统安装 Oracle Instant Client：

**Windows:**
1. 下载 Oracle Instant Client
2. 解压到目录（如 `C:\instantclient`）
3. 将目录添加到 PATH 环境变量
4. 设置 ORACLE_HOME 环境变量

**Linux:**
```bash
# 下载并安装 RPM 包
sudo yum install oracle-instantclient-basic
# 或解压 ZIP 包到 /opt/instantclient
export ORACLE_HOME=/opt/instantclient
export PATH=$ORACLE_HOME:$PATH
export LD_LIBRARY_PATH=$ORACLE_HOME:$LD_LIBRARY_PATH
```

**macOS:**
```bash
# 解压到 /opt/instantclient
export ORACLE_HOME=/opt/instantclient
export PATH=$ORACLE_HOME:$PATH
export DYLD_LIBRARY_PATH=$ORACLE_HOME:$DYLD_LIBRARY_PATH
```

### 3. 安装 ora2pg
```bash
# 使用 CPAN 安装
cpan Ora2Pg

# 或从源码安装
wget https://github.com/darold/ora2pg/archive/v24.3.tar.gz
tar -xzf v24.3.tar.gz
cd ora2pg-24.3
perl Makefile.PL
make && sudo make install
```

## 快速开始

### 1. 创建新项目
```bash
# 创建基础项目
ora2pg-admin 初始化 我的迁移项目

# 使用高级模板
ora2pg-admin 初始化 --template=advanced --description="生产环境迁移" 生产迁移
```

### 2. 配置数据库连接
```bash
# 进入项目目录
cd 我的迁移项目

# 配置数据库连接
ora2pg-admin 配置 数据库
```

### 3. 检查环境
```bash
# 检查 Oracle 客户端和环境配置
ora2pg-admin 检查 环境

# 测试数据库连接
ora2pg-admin 检查 连接
```

### 4. 执行迁移
```bash
# 迁移数据库结构
ora2pg-admin 迁移 结构

# 迁移数据内容
ora2pg-admin 迁移 数据

# 完整迁移（推荐）
ora2pg-admin 迁移 全部
```

## 命令详解

### 初始化命令
创建新的迁移项目，生成项目结构和配置文件。

```bash
ora2pg-admin 初始化 [项目名称] [选项]
```

**选项：**
- `--template, -t`：项目模板（basic、advanced、custom）
- `--description, -d`：项目描述
- `--force, -f`：强制覆盖已存在的项目

**示例：**
```bash
# 交互式创建项目
ora2pg-admin 初始化

# 使用参数创建项目
ora2pg-admin 初始化 --template=basic --description="测试迁移" 测试项目
```

### 配置命令
配置数据库连接和迁移选项。

```bash
ora2pg-admin 配置 [子命令] [选项]
```

**子命令：**
- `数据库`：配置 Oracle 和 PostgreSQL 连接
- `选项`：配置迁移类型和性能参数

**选项：**
- `--file, -f`：指定配置文件路径
- `--backup`：配置前创建备份（默认启用）
- `--force`：强制覆盖现有配置

### 检查命令
检查环境配置和数据库连接状态。

```bash
ora2pg-admin 检查 [子命令] [选项]
```

**子命令：**
- `环境`：检查 Oracle 客户端、ora2pg 工具等环境配置
- `连接`：测试 Oracle 和 PostgreSQL 数据库连接

**选项：**
- `--verbose, -v`：显示详细检查信息
- `--config, -c`：指定配置文件路径

### 迁移命令
执行数据库迁移操作。

```bash
ora2pg-admin 迁移 [子命令] [选项]
```

**子命令：**
- `结构`：迁移数据库结构（表、视图、序列等）
- `数据`：迁移数据内容
- `全部`：执行完整迁移流程

**选项：**
- `--timeout`：迁移超时时间（默认2小时）
- `--parallel`：并行作业数（0表示使用配置文件设置）
- `--resume`：恢复中断的迁移
- `--validate`：迁移后验证结果（默认启用）
- `--backup`：迁移前创建备份（默认启用）

## 配置文件说明

项目配置文件位于 `.ora2pg-admin/config.yaml`，包含以下主要部分：

### 项目信息
```yaml
project:
  name: "项目名称"
  version: "1.0.0"
  description: "项目描述"
  created: "2024-01-01T00:00:00Z"
  updated: "2024-01-01T00:00:00Z"
```

### Oracle 数据库配置
```yaml
oracle:
  host: "localhost"
  port: 1521
  sid: "ORCL"              # 或使用 service
  service: ""              # Service Name
  username: "system"
  password: "${ORACLE_PASSWORD}"  # 支持环境变量
  schema: ""               # 可选，指定模式
```

### PostgreSQL 数据库配置
```yaml
postgresql:
  host: "localhost"
  port: 5432
  database: "postgres"
  username: "postgres"
  password: "${PG_PASSWORD}"
  schema: "public"
```

### 迁移配置
```yaml
migration:
  types:                   # 迁移对象类型
    - "TABLE"
    - "VIEW"
    - "SEQUENCE"
    - "INDEX"
  parallel_jobs: 4         # 并行作业数
  batch_size: 1000         # 批处理大小
  output_dir: "output"     # 输出目录
  log_level: "INFO"        # 日志级别
```

## 最佳实践

### 1. 迁移前准备
- 备份源数据库和目标数据库
- 确保网络连接稳定
- 验证用户权限充足
- 预估迁移时间和资源需求

### 2. 分阶段迁移
建议按以下顺序执行迁移：
1. **结构迁移**：先迁移表结构、视图、序列
2. **数据迁移**：迁移表数据
3. **索引创建**：创建索引和约束
4. **程序对象**：迁移触发器、函数、存储过程
5. **权限设置**：设置用户权限

### 3. 性能优化
- 根据服务器配置调整并行作业数
- 大表可以考虑分批迁移
- 在迁移期间暂时禁用不必要的索引
- 监控系统资源使用情况

### 4. 错误处理
- 仔细查看错误日志
- 对于失败的对象，可以单独重试
- 保留迁移日志用于问题排查
- 必要时联系技术支持

## 故障排除

### 常见问题

**Q: 提示"未找到 Oracle 客户端"**
A: 请确认已正确安装 Oracle Instant Client 并设置环境变量。

**Q: 连接 Oracle 数据库失败**
A: 检查网络连接、用户名密码、SID/Service Name 是否正确。

**Q: ora2pg 命令不存在**
A: 请安装 ora2pg 工具并确保在 PATH 中可以找到。

**Q: 迁移过程中断**
A: 可以使用 `--resume` 参数恢复中断的迁移。

### 日志查看
- 应用日志：`logs/` 目录
- ora2pg 日志：`logs/ora2pg-*.log`
- 迁移输出：`output/` 目录

### 获取帮助
```bash
# 查看命令帮助
ora2pg-admin --help
ora2pg-admin 初始化 --help
ora2pg-admin 配置 --help

# 查看版本信息
ora2pg-admin --version
```

## 高级功能

### 环境变量支持
配置文件支持环境变量替换：
```yaml
oracle:
  password: "${ORACLE_PASSWORD}"
postgresql:
  password: "${PG_PASSWORD}"
```

### 自定义模板
可以创建自定义的项目模板和配置模板。

### 批量操作
支持批量处理多个数据库或模式的迁移。

---

更多详细信息请参考项目文档或联系技术支持。
