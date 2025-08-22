# Ora2Pg-Admin

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/badge/Release-v1.0.0-orange.svg)](https://github.com/zaops/ora2pg-admin/releases)
[![Build Status](https://github.com/zaops/ora2pg-admin/workflows/Build%20Linux%20x64/badge.svg)](https://github.com/zaops/ora2pg-admin/actions)
[![codecov](https://codecov.io/gh/zaops/ora2pg-admin/branch/main/graph/badge.svg)](https://codecov.io/gh/zaops/ora2pg-admin)

一个基于 Go 语言开发的中文命令行工具，旨在简化 Oracle 到 PostgreSQL 数据库迁移的运维操作。

## ✨ 特性

- 🇨🇳 **完全中文界面** - 提供友好的中文命令和提示信息
- 🚀 **一键初始化** - 快速创建迁移项目和配置文件
- 🔧 **交互式配置** - 智能配置向导，简化复杂参数设置
- 🔍 **环境检测** - 自动检测 Oracle 客户端和依赖工具
- 📊 **实时进度** - 可视化进度条和详细状态信息
- 🛡️ **安全可靠** - 支持环境变量、配置验证和错误恢复
- 📈 **高性能** - 支持并行迁移和批处理优化
- 📝 **详细日志** - 完整的操作日志和问题诊断

## 🎯 项目简介

ora2pg 是一个强大的 Oracle 到 PostgreSQL 数据库迁移工具，但其命令行界面学习成本较高。Ora2Pg-Admin 为 ora2pg 提供了友好的中文命令行界面，让运维人员能够轻松完成数据库迁移任务。

### 为什么选择 Ora2Pg-Admin？

- **降低学习成本** - 中文界面和交互式向导，无需深入学习 ora2pg 复杂参数
- **提高工作效率** - 自动化配置生成和环境检测，减少手动配置错误
- **增强可靠性** - 完善的错误处理和恢复机制，确保迁移过程稳定
- **改善用户体验** - 实时进度显示和详细日志，让迁移过程可视化

## 🛠️ 技术栈

- **Go 1.24** - 主要开发语言，提供高性能和跨平台支持
- **Cobra** - CLI 框架，支持中文命令和子命令
- **Viper** - 配置管理，支持多种配置格式
- **promptui** - 交互式用户界面，提供友好的配置体验
- **logrus** - 结构化日志，支持多种输出格式

## 🚀 快速开始

### 系统要求

#### 基础环境
- 操作系统：Windows 10+、Linux、macOS
- 内存：建议 4GB 以上
- 磁盘空间：根据数据库大小预留足够空间

#### 依赖工具
- **Oracle 客户端**：Oracle Instant Client 11g+ 或完整 Oracle 客户端
- **ora2pg**：版本 20.0+ （Perl 工具）
- **PostgreSQL 客户端**：psql 工具（可选）

### 安装方式

#### 方式一：下载预编译版本（推荐）
```bash
# 从 GitHub Releases 下载对应平台的可执行文件
wget https://github.com/zaops/ora2pg-admin/releases/latest/download/ora2pg-admin-linux-amd64
chmod +x ora2pg-admin-linux-amd64
sudo mv ora2pg-admin-linux-amd64 /usr/local/bin/ora2pg-admin
```

#### 方式二：从源码构建
```bash
# 克隆项目
git clone https://github.com/zaops/ora2pg-admin.git
cd ora2pg-admin

# 构建
go build -o ora2pg-admin

# 安装到系统路径（可选）
sudo mv ora2pg-admin /usr/local/bin/
```

#### 方式三：使用 Docker
```bash
# 拉取镜像
docker pull zaops/ora2pg-admin:latest

# 运行容器
docker run -it --rm \
  -v $(pwd)/projects:/data/projects \
  -v $(pwd)/logs:/data/logs \
  zaops/ora2pg-admin:latest
```

### 基本使用

```bash
# 1. 创建新的迁移项目
ora2pg-admin 初始化 我的迁移项目
cd 我的迁移项目

# 2. 检查环境配置
ora2pg-admin 检查 环境

# 3. 配置数据库连接
ora2pg-admin 配置 数据库

# 4. 测试数据库连接
ora2pg-admin 检查 连接

# 5. 配置迁移选项
ora2pg-admin 配置 选项

# 6. 执行迁移
ora2pg-admin 迁移 全部
```

## 📋 命令参考

### 初始化命令
```bash
ora2pg-admin 初始化 [项目名称] [选项]

选项:
  --template, -t    项目模板 (basic, advanced, custom)
  --description, -d 项目描述
  --force, -f       强制覆盖已存在的项目
```

### 配置命令
```bash
ora2pg-admin 配置 [子命令] [选项]

子命令:
  数据库    配置 Oracle 和 PostgreSQL 连接
  选项      配置迁移类型和性能参数

选项:
  --file, -f    指定配置文件路径
  --backup      配置前创建备份
  --force       强制覆盖现有配置
```

### 检查命令
```bash
ora2pg-admin 检查 [子命令] [选项]

子命令:
  环境      检查 Oracle 客户端、ora2pg 工具等环境配置
  连接      测试 Oracle 和 PostgreSQL 数据库连接

选项:
  --verbose, -v 显示详细检查信息
  --config, -c  指定配置文件路径
```

### 迁移命令
```bash
ora2pg-admin 迁移 [子命令] [选项]

子命令:
  结构      迁移数据库结构 (表、视图、序列等)
  数据      迁移数据内容
  全部      执行完整迁移流程

选项:
  --timeout     迁移超时时间 (默认2小时)
  --parallel    并行作业数
  --resume      恢复中断的迁移
  --validate    迁移后验证结果
  --backup      迁移前创建备份
```

## 📁 项目结构

### 源码结构
```
ora2pg-admin/
├── cmd/                    # 命令行入口
│   ├── init.go            # 初始化命令
│   ├── config.go          # 配置命令
│   ├── check.go           # 检查命令
│   ├── migrate.go         # 迁移命令
│   └── root.go            # 根命令
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   │   ├── manager.go     # 配置管理器
│   │   ├── template.go    # 模板引擎
│   │   └── validator.go   # 配置验证
│   ├── service/           # 核心服务
│   │   ├── ora2pg.go      # ora2pg包装服务
│   │   ├── migration.go   # 迁移管理服务
│   │   └── progress.go    # 进度跟踪服务
│   ├── oracle/            # Oracle相关
│   │   ├── client.go      # 客户端检测
│   │   └── connection.go  # 连接测试
│   └── utils/             # 工具函数
│       ├── file.go        # 文件操作
│       ├── logger.go      # 日志管理
│       └── error.go       # 错误处理
├── templates/             # 配置模板
├── docs/                  # 项目文档
│   ├── user-guide.md      # 用户指南
│   ├── examples/          # 配置示例
│   └── troubleshooting.md # 故障排除
├── tests/                 # 测试文件
├── .github/workflows/     # CI/CD 配置
├── Dockerfile            # Docker 配置
├── docker-compose.yml    # 开发环境
├── Makefile             # 构建脚本
├── main.go              # 程序入口
└── README.md            # 项目说明
```

### 用户项目结构
```
我的迁移项目/
├── .ora2pg-admin/          # 配置目录
│   └── config.yaml         # 主配置文件
├── logs/                   # 日志目录
├── output/                 # 迁移输出目录
├── scripts/                # 自定义脚本目录
├── backup/                 # 备份目录
├── docs/                   # 项目文档
├── README.md               # 项目说明
└── .gitignore             # Git 忽略文件
```

## ⚙️ 配置文件

配置文件位于 `.ora2pg-admin/config.yaml`，主要包含：

```yaml
# 项目信息
project:
  name: "项目名称"
  version: "1.0.0"
  description: "项目描述"
  created: "2024-01-01T00:00:00Z"
  updated: "2024-01-01T00:00:00Z"

# Oracle 数据库配置
oracle:
  host: "localhost"
  port: 1521
  sid: "ORCL"                    # 使用SID连接
  service: ""                    # 或使用Service Name
  username: "system"
  password: "${ORACLE_PASSWORD}" # 支持环境变量
  schema: ""                     # 可选，指定模式

# PostgreSQL 数据库配置
postgresql:
  host: "localhost"
  port: 5432
  database: "postgres"
  username: "postgres"
  password: "${PG_PASSWORD}"
  schema: "public"

# 迁移配置
migration:
  types:                   # 迁移对象类型
    - "TABLE"
    - "VIEW"
    - "SEQUENCE"
    - "INDEX"
    - "TRIGGER"
    - "FUNCTION"
    - "PROCEDURE"
  parallel_jobs: 4         # 并行作业数
  batch_size: 1000         # 批处理大小
  output_dir: "output"     # 输出目录
  log_level: "INFO"        # 日志级别

# Oracle 客户端配置
oracle_client:
  home: ""                 # Oracle客户端路径（留空表示自动检测）
  auto_detect: true        # 是否自动检测Oracle客户端
```

## 🔧 环境变量

支持以下环境变量：

```bash
# Oracle 连接
export ORACLE_PASSWORD="your_oracle_password"
export ORACLE_HOME="/opt/oracle/instantclient"

# PostgreSQL 连接
export PG_PASSWORD="your_postgres_password"

# 其他配置
export ORA2PG_ADMIN_LOG_LEVEL="INFO"
export ORA2PG_ADMIN_CONFIG_FILE="custom-config.yaml"
```

## 📚 文档

- [用户使用指南](docs/user-guide.md) - 详细的使用说明和最佳实践
- [配置示例](docs/examples/) - 各种场景的配置示例
- [使用场景](docs/examples/scenarios.md) - 常见使用场景和解决方案
- [故障排除](docs/troubleshooting.md) - 常见问题和解决方法

## 🛠️ 开发

### 本地开发
```bash
# 克隆项目
git clone https://github.com/zaops/ora2pg-admin.git
cd ora2pg-admin

# 安装依赖
go mod download

# 运行测试
make test

# 构建
make build

# 运行
./build/ora2pg-admin --help
```

### 使用 Docker 开发
```bash
# 启动开发环境
docker-compose up -d

# 进入容器
docker-compose exec ora2pg-admin bash

# 运行测试
make test
```

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

### 贡献流程
1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 开发规范
- 遵循 Go 代码规范
- 添加适当的测试用例
- 更新相关文档
- 确保 CI 检查通过

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [ora2pg](https://github.com/darold/ora2pg) - 优秀的 Oracle 到 PostgreSQL 迁移工具
- [Cobra](https://github.com/spf13/cobra) - 强大的 Go CLI 框架
- [Viper](https://github.com/spf13/viper) - 灵活的配置管理库
- [promptui](https://github.com/manifoldco/promptui) - 交互式命令行界面

## 📞 支持

如果您遇到问题或需要帮助：

- 📧 邮箱：support@example.com
- 🐛 问题报告：[GitHub Issues](https://github.com/zaops/ora2pg-admin/issues)
- 💬 讨论：[GitHub Discussions](https://github.com/zaops/ora2pg-admin/discussions)
- 📖 文档：[用户指南](docs/user-guide.md)

## 🚀 路线图

- [ ] Web 管理界面
- [ ] 增量迁移支持
- [ ] 多数据库并行迁移
- [ ] 迁移性能优化
- [ ] 云平台集成
- [ ] 监控和告警功能

---

**Ora2Pg-Admin** - 让 Oracle 到 PostgreSQL 迁移变得简单！ 🚀
