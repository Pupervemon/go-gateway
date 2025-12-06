# 配置文件说明

## config.template.yaml - 配置模板文件

### 📝 用途
这是一个配置文件的**模板示例**，用于指导开发者如何正确配置本地环境。

### 🔧 使用说明
1. **复制模板**：
   ```bash
   cp config/config.template.yaml config/config.yaml
   ```

2. **修改配置**：
   - 根据实际环境修改 `config.yaml` 中的值
   - 不要直接修改模板文件

3. **配置项说明**：
   ```yaml
   server:
     port: 8080                    # 网关监听端口

   nacos:
     host: "127.0.0.1"            # Nacos 服务器地址
     port: 8848                   # Nacos 端口
     namespace: "public"          # 命名空间ID
     group: "DEFAULT_GROUP"       # 配置组名
     dataid: "gateway.yaml"       # 远程配置文件名
   ```

### ⚠️ 注意事项
- `config.template.yaml` 是模板，不会被 git 忽略
- `config.yaml` 是实际使用的配置，已在 .gitignore 中忽略
- 请勿将包含敏感信息的配置文件提交到版本控制