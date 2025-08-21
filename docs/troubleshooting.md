# 故障排除指南

本文档提供了使用 Ora2Pg-Admin 过程中常见问题的解决方案。

## 环境相关问题

### Q1: 提示"未找到 Oracle 客户端"

**错误信息：**
```
❌ Oracle客户端: 未安装
💡 建议:
  1. 请安装Oracle Instant Client
  2. 设置ORACLE_HOME环境变量
  3. 将Oracle客户端路径添加到PATH环境变量
```

**解决方案：**

1. **下载并安装 Oracle Instant Client**
   - 访问 [Oracle 官网](https://www.oracle.com/database/technologies/instant-client.html)
   - 下载适合您操作系统的 Instant Client
   - 解压到合适的目录

2. **设置环境变量**
   ```bash
   # Linux/macOS
   export ORACLE_HOME=/opt/oracle/instantclient_19_8
   export PATH=$ORACLE_HOME:$PATH
   export LD_LIBRARY_PATH=$ORACLE_HOME:$LD_LIBRARY_PATH
   
   # Windows (PowerShell)
   $env:ORACLE_HOME = "C:\instantclient_19_8"
   $env:PATH = "$env:ORACLE_HOME;$env:PATH"
   ```

3. **验证安装**
   ```bash
   ora2pg-admin 检查 环境
   ```

### Q2: 提示"ora2pg工具未找到"

**错误信息：**
```
❌ ora2pg工具: 未找到
💡 解决建议:
  1. 确认ora2pg已正确安装
  2. 将ora2pg添加到PATH环境变量
  3. 检查Perl环境是否正确配置
```

**解决方案：**

1. **安装 ora2pg**
   ```bash
   # 使用 CPAN
   cpan Ora2Pg
   
   # 或从源码安装
   wget https://github.com/darold/ora2pg/archive/v24.3.tar.gz
   tar -xzf v24.3.tar.gz
   cd ora2pg-24.3
   perl Makefile.PL
   make && sudo make install
   ```

2. **检查 Perl 环境**
   ```bash
   perl -v
   perl -MOra2Pg -e "print 'Ora2Pg installed successfully\n'"
   ```

3. **验证 ora2pg**
   ```bash
   ora2pg --version
   ```

## 连接相关问题

### Q3: Oracle 数据库连接失败

**错误信息：**
```
❌ Oracle数据库连接测试
状态: ❌ 连接失败
错误: ORA-12541: TNS:no listener
```

**解决方案：**

1. **检查网络连通性**
   ```bash
   ping oracle-server.example.com
   telnet oracle-server.example.com 1521
   ```

2. **验证连接参数**
   - 确认主机名、端口号正确
   - 检查 SID 或 Service Name 是否正确
   - 验证用户名和密码

3. **检查 Oracle 服务状态**
   ```bash
   # 在 Oracle 服务器上检查监听器状态
   lsnrctl status
   ```

4. **使用 tnsping 测试**
   ```bash
   tnsping oracle-server.example.com:1521/ORCL
   ```

### Q4: PostgreSQL 数据库连接失败

**错误信息：**
```
❌ PostgreSQL连接失败: psql执行失败
```

**解决方案：**

1. **安装 PostgreSQL 客户端**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql-client
   
   # CentOS/RHEL
   sudo yum install postgresql
   
   # macOS
   brew install postgresql
   ```

2. **检查连接参数**
   ```bash
   psql -h postgres-server.example.com -p 5432 -U postgres -d postgres
   ```

3. **检查防火墙设置**
   ```bash
   # 检查端口是否开放
   nmap -p 5432 postgres-server.example.com
   ```

## 配置相关问题

### Q5: 配置文件解析失败

**错误信息：**
```
❌ 解析配置文件失败
请检查配置文件的语法是否正确
```

**解决方案：**

1. **检查 YAML 语法**
   ```bash
   # 使用在线工具验证 YAML 语法
   # 或使用 yamllint
   yamllint .ora2pg-admin/config.yaml
   ```

2. **常见语法错误**
   - 缩进必须使用空格，不能使用制表符
   - 冒号后面必须有空格
   - 字符串包含特殊字符时需要引号

3. **重新生成配置**
   ```bash
   # 备份现有配置
   cp .ora2pg-admin/config.yaml .ora2pg-admin/config.yaml.backup
   
   # 重新配置
   ora2pg-admin 配置 数据库
   ```

### Q6: 环境变量未生效

**错误信息：**
```
❌ 数据库凭据无效
请检查用户名和密码
```

**解决方案：**

1. **检查环境变量设置**
   ```bash
   echo $ORACLE_PASSWORD
   echo $PG_PASSWORD
   ```

2. **正确设置环境变量**
   ```bash
   # 临时设置
   export ORACLE_PASSWORD="your_password"
   
   # 永久设置（添加到 ~/.bashrc 或 ~/.zshrc）
   echo 'export ORACLE_PASSWORD="your_password"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **在配置文件中使用环境变量**
   ```yaml
   oracle:
     password: "${ORACLE_PASSWORD}"
   postgresql:
     password: "${PG_PASSWORD}"
   ```

## 迁移相关问题

### Q7: 迁移过程中断

**错误信息：**
```
⚠️ 收到中断信号，正在停止迁移...
❌ 迁移被用户取消
```

**解决方案：**

1. **使用恢复功能**
   ```bash
   ora2pg-admin 迁移 全部 --resume
   ```

2. **检查中断原因**
   - 查看日志文件了解中断原因
   - 检查系统资源使用情况
   - 确认网络连接稳定性

3. **调整超时设置**
   ```bash
   ora2pg-admin 迁移 全部 --timeout=6h
   ```

### Q8: 迁移性能慢

**问题描述：**
迁移速度很慢，需要很长时间才能完成。

**解决方案：**

1. **调整并行度**
   ```bash
   ora2pg-admin 迁移 全部 --parallel=8
   ```

2. **优化批处理大小**
   ```yaml
   migration:
     batch_size: 10000  # 增加批处理大小
   ```

3. **分阶段迁移**
   ```bash
   # 先迁移结构
   ora2pg-admin 迁移 结构
   
   # 再迁移数据
   ora2pg-admin 迁移 数据
   ```

4. **系统优化**
   - 增加内存分配
   - 使用 SSD 存储
   - 优化网络带宽

### Q9: 数据类型转换错误

**错误信息：**
```
ERROR: column "date_field" is of type timestamp without time zone but expression is of type character varying
```

**解决方案：**

1. **检查数据类型映射**
   - 查看 ora2pg 文档了解数据类型映射规则
   - 必要时手动调整生成的 SQL

2. **自定义数据类型转换**
   ```yaml
   advanced:
     data_types:
       auto_convert: true
   ```

3. **手动处理特殊情况**
   - 在 `scripts/` 目录创建自定义转换脚本
   - 在迁移前后执行数据清理

## 权限相关问题

### Q10: 权限不足错误

**错误信息：**
```
ORA-00942: table or view does not exist
ERROR: permission denied for table
```

**解决方案：**

1. **检查 Oracle 用户权限**
   ```sql
   -- 授予必要的权限
   GRANT SELECT ANY TABLE TO migration_user;
   GRANT SELECT ANY DICTIONARY TO migration_user;
   ```

2. **检查 PostgreSQL 用户权限**
   ```sql
   -- 授予数据库权限
   GRANT ALL PRIVILEGES ON DATABASE target_db TO migration_user;
   GRANT ALL ON SCHEMA public TO migration_user;
   ```

3. **使用具有足够权限的用户**
   - Oracle: 使用 SYSTEM 或具有 DBA 权限的用户
   - PostgreSQL: 使用 postgres 超级用户或数据库所有者

## 日志和调试

### 启用详细日志

```bash
# 启用详细输出
ora2pg-admin 迁移 全部 --verbose

# 设置日志级别
export ORA2PG_ADMIN_LOG_LEVEL=DEBUG
```

### 查看日志文件

```bash
# 查看应用日志
tail -f logs/ora2pg-admin.log

# 查看 ora2pg 日志
tail -f logs/ora2pg-*.log

# 查看迁移输出
ls -la output/
```

### 生成诊断报告

```bash
# 生成环境诊断报告
ora2pg-admin 检查 环境 --verbose > diagnostic-report.txt

# 包含连接测试
ora2pg-admin 检查 连接 --verbose >> diagnostic-report.txt
```

## 获取帮助

如果以上解决方案无法解决您的问题，请：

1. **查看详细文档**
   - [用户使用指南](user-guide.md)
   - [配置示例](examples/)

2. **收集诊断信息**
   - 错误日志
   - 配置文件
   - 环境信息
   - 操作步骤

3. **联系技术支持**
   - 📧 邮箱：support@example.com
   - 🐛 GitHub Issues：[报告问题](https://github.com/your-org/ora2pg-admin/issues)
   - 💬 讨论区：[GitHub Discussions](https://github.com/your-org/ora2pg-admin/discussions)

## 预防措施

为了避免常见问题，建议：

1. **迁移前充分测试**
   - 在测试环境先执行完整流程
   - 验证所有依赖工具正常工作
   - 确认网络连接稳定

2. **做好备份**
   - 备份源数据库
   - 备份目标数据库
   - 保存配置文件

3. **监控资源使用**
   - 监控 CPU、内存、磁盘使用情况
   - 确保有足够的磁盘空间
   - 监控网络带宽使用

4. **制定应急预案**
   - 准备回滚方案
   - 制定问题升级流程
   - 准备技术支持联系方式
