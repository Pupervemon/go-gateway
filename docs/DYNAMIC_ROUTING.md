# 动态路由配置说明

## 概述

Go Gateway 支持从 Nacos 配置中心动态加载路由配置，无需重启服务即可实现路由的热更新。

## 配置格式

路由配置存储在 Nacos 的 `routes.json` 文件中，格式如下：

```json
[
    {
        "id": "route-id",
        "order": 1,
        "uri": "lb://service-name",
        "predicates": [
            {
                "name": "Path",
                "args": {
                    "_genkey_0": "/api/path1/**",
                    "_genkey_1": "/api/path2/**"
                }
            }
        ],
        "filters": []
    }
]
```

## 字段说明

### id (必填)
- 路由的唯一标识符
- 类型：string
- 示例：`"user-service-route"`

### order (可选)
- 路由的优先级，数值越小优先级越高
- 类型：int
- 默认值：0
- 示例：`100`

### uri (必填)
- 目标服务的URI
- 支持两种格式：
  1. `lb://service-name` - 使用 Nacos 服务发现进行负载均衡
  2. `https://domain.com/path` - 直接转发到外部服务
- 类型：string
- 示例：
  - `"lb://user-service"` - 转发到用户服务
  - `"https://api.dify.ai/v1"` - 转发到 Dify API

### predicates (必填)
- 路由匹配条件，目前支持 Path 匹配
- 类型：数组

#### Path Predicate
- `name`: 必须为 `"Path"`
- `args`: 路径参数
  - `_genkey_0`, `_genkey_1`, ...: 匹配的路径
  - 支持通配符 `**` 匹配多级路径

### filters (可选)
- 路由过滤器（暂未实现）
- 类型：数组

## 路由示例

### 1. 内部服务（使用负载均衡）

```json
{
    "id": "user-service",
    "order": 100,
    "uri": "lb://user-service",
    "predicates": [
        {
            "name": "Path",
            "args": {
                "_genkey_0": "/api/user/**",
                "_genkey_1": "/api/admin/users/**"
            }
        }
    ],
    "filters": []
}
```

### 2. 外部服务（直接URL）

```json
{
    "id": "dify-api-route",
    "uri": "https://api.dify.ai/v1",
    "predicates": [
        {
            "name": "Path",
            "args": {
                "_genkey_0": "/api/dify/**"
            }
        }
    ],
    "filters": []
}
```

### 3. 多路径匹配

```json
{
    "id": "common-service",
    "uri": "lb://common-service",
    "predicates": [
        {
            "name": "Path",
            "args": {
                "_genkey_0": "/api/notifications/**",
                "_genkey_1": "/api/file/**",
                "_genkey_2": "/api/announcements/**"
            }
        }
    ],
    "filters": []
}
```

## 路径匹配规则

- `/api/user/**` - 匹配 `/api/user/` 下的所有路径
- `/api/user/profile` - 精确匹配 `/api/user/profile`
- `/api/user/{id}/profile` - 使用 Gin 的参数匹配（自动转换）

## 热更新

当 Nacos 中的 `routes.json` 配置发生变化时，网关会自动：
1. 监听配置变更
2. 重新加载路由配置
3. 更新内存中的路由规则

无需重启网关服务即可生效。

## 配置步骤

1. 登录 Nacos 控制台
2. 在配置管理中新建配置
3. Data ID: `routes.json`
4. Group: `DEFAULT_GROUP` (或根据配置修改)
5. 配置格式: `JSON`
6. 将路由配置粘贴到配置内容中
7. 发布配置

## 注意事项

1. 路由匹配是按 order 字段排序的，order 值越小优先级越高
2. 路径 `/api` 前缀会在网关层面统一处理，配置时不需要包含
3. 外部服务（如 Dify）的请求会自动移除 `/api` 前缀
4. 确保所有路径都以 `/api` 开头，以匹配网关的路由组