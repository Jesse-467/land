# Land 论坛系统

## 项目简介

Land 是一个基于 Go 语言和 Gin 框架开发的论坛系统。支持用户注册登录、发帖、评论、投票、社区管理、访问量统计、缓存与防雪崩、MySQL/Redis 混合索引优化等功能，适合学习。

---

## 技术栈

-   **Go 1.18+**
-   **Gin**：高性能 Web 框架
-   **Gorm**：ORM 框架
-   **MySQL**：主数据存储
-   **Redis**：缓存、计数、排行榜
-   **JWT**：用户认证
-   **Snowflake**：分布式 ID 生成
-   **Zap**：日志
-   **Go Playground Validator**：参数校验
-   **Swagger**（推荐集成）：API 文档

---

## 目录结构与模块说明

```
land/
├── conf/           # 配置文件（config.yaml等）
├── controllers/    # 路由控制器（API接口实现，按业务拆分）
├── dao/            # 数据访问层
│   ├── mysql/      # MySQL相关操作
│   └── redis/      # Redis相关操作
├── docs/           # 项目文档
├── logic/          # 业务逻辑层（服务/聚合/一致性等）
├── logger/         # 日志组件
├── main.go         # 启动入口
├── middlewares/    # Gin中间件（JWT、限流等）
├── models/         # 数据模型（参数、表结构、DTO等）
├── pkg/            # 工具包（JWT、雪花ID等）
├── routers/        # 路由注册
├── settings/       # 配置加载
└── log/            # 日志文件
```

### 主要目录说明

-   **controllers/**：每个业务领域一个控制器，负责参数校验、权限、调用 logic 层、返回响应。
-   **logic/**：业务聚合与一致性保障，复杂业务流程、缓存一致性、延迟双删等均在此实现。
-   **dao/mysql/**、**dao/redis/**：数据访问层，所有数据库/缓存操作集中于此，便于维护和扩展。
-   **middlewares/**：JWT 认证、限流、日志等 Gin 中间件。
-   **models/**：所有数据结构定义，包括表结构、请求参数、响应结构体等。
-   **pkg/**：通用工具包，如 JWT 生成解析、雪花 ID 生成等。
-   **settings/**：配置加载与管理。

---

## 架构设计与核心特性

### 1. 认证与安全

-   JWT 认证，所有敏感操作需登录
-   密码加密存储，防止明文泄露
-   限流中间件，防止接口被刷
-   详细的参数校验与错误码体系

### 2. 缓存与一致性

-   Redis 缓存帖子详情、访问量、投票等热点数据
-   缓存雪崩防护：所有缓存均带有随机 TTL（±10~25%），防止大面积同时过期
-   缓存穿透防护：不存在标记，防止恶意请求击穿数据库
-   延迟双删、强一致性接口，保证缓存与数据库一致
-   支持手动/定时同步访问量

### 3. 访问量统计与防刷

-   访问量计数存储于 Redis，支持同一用户 24 小时内不重复计数
-   定时/手动同步访问量到 MySQL，保证数据持久化
-   支持访问量排行榜

### 4. MySQL/Redis 混合索引优化

-   帖子列表支持按时间、访问量、分数排序
-   按时间/访问量排序用 MySQL 索引，按分数用 Redis
-   支持分页、社区筛选、搜索
-   支持 use_index 参数灵活切换

### 5. 代码规范与可维护性

-   分层清晰，接口/逻辑/数据访问分离
-   结构体、接口、错误码、日志等均有详细注释
-   便于二次开发和扩展

---

## 数据结构与表结构简述

### 用户表（user）

| 字段        | 类型     | 说明               |
| ----------- | -------- | ------------------ |
| id          | bigint   | 自增主键           |
| user_id     | bigint   | 用户 ID（雪花 ID） |
| username    | varchar  | 用户名             |
| password    | varchar  | 密码（加密）       |
| email       | varchar  | 邮箱               |
| gender      | tinyint  | 性别（0=未知）     |
| create_time | datetime | 注册时间           |
| update_time | datetime | 更新时间           |

### 社区表（community）

| 字段           | 类型      | 说明     |
| -------------- | --------- | -------- |
| id             | int       | 自增主键 |
| community_id   | int       | 社区 ID  |
| community_name | varchar   | 社区名   |
| introduction   | varchar   | 简介     |
| create_time    | timestamp | 创建时间 |
| update_time    | timestamp | 更新时间 |

### 帖子表（post）

| 字段         | 类型     | 说明               |
| ------------ | -------- | ------------------ |
| id           | bigint   | 自增主键           |
| post_id      | bigint   | 帖子 ID（雪花 ID） |
| title        | varchar  | 标题               |
| content      | varchar  | 内容               |
| author_id    | bigint   | 作者 ID            |
| community_id | bigint   | 社区 ID            |
| status       | tinyint  | 帖子状态           |
| view_count   | bigint   | 访问量             |
| create_time  | datetime | 创建时间           |
| update_time  | datetime | 更新时间           |

### 评论表（comment）

| 字段        | 类型     | 说明      |
| ----------- | -------- | --------- |
| id          | bigint   | 自增主键  |
| comment_id  | bigint   | 评论 ID   |
| content     | text     | 评论内容  |
| post_id     | bigint   | 帖子 ID   |
| author_id   | bigint   | 作者 ID   |
| parent_id   | bigint   | 父评论 ID |
| status      | tinyint  | 评论状态  |
| create_time | datetime | 创建时间  |
| update_time | datetime | 更新时间  |

### 投票表（vote）

| 字段        | 类型     | 说明                |
| ----------- | -------- | ------------------- |
| post_id     | bigint   | 帖子 ID             |
| user_id     | bigint   | 用户 ID             |
| direction   | tinyint  | 1=赞，-1=踩，0=取消 |
| create_time | datetime | 投票时间            |

---

## API 接口文档（详细）

### 用户相关

#### 1. 注册

-   **POST** `/auth/register`
-   **参数（JSON）**:
    -   username: string
    -   password: string
    -   re_password: string
    -   email: string
-   **返回**: 注册成功/失败，错误码
-   **示例**:

```json
{
    "username": "alice",
    "password": "123456",
    "re_password": "123456",
    "email": "alice@example.com"
}
```

#### 2. 登录

-   **POST** `/auth/login`
-   **参数（JSON）**:
    -   username: string
    -   password: string
-   **返回**:
    -   user_id
    -   user_name
    -   token
-   **示例**:

```json
{
    "username": "alice",
    "password": "123456"
}
```

---

### 社区相关

#### 1. 社区列表

-   **GET** `/api/v1/community`
-   **返回**: 社区列表

#### 2. 社区详情

-   **GET** `/api/v1/community/:id`
-   **返回**: 社区详细信息

---

### 帖子相关

#### 1. 创建帖子

-   **POST** `/api/v1/post`
-   **参数（JSON）**:
    -   title: string
    -   content: string
    -   community_id: int
-   **权限**: 需登录
-   **返回**: 创建成功/失败

#### 2. 获取帖子详情

-   **GET** `/api/v1/post/:id`
-   **返回**: 帖子详细信息（含作者、社区、访问量、投票数等）

#### 3. 获取帖子列表（推荐新版）

-   **GET** `/api/v1/posts2/`
-   **参数（Query）**:
    -   page: int，页码，默认 1
    -   size: int，每页条数，默认 50，最大 100
    -   order: string，排序方式（time/score/view）
    -   community_id: int，社区 ID（可选）
    -   search: string，搜索关键词（可选）
    -   use_index: bool，是否用 MySQL 索引优化，默认 true
-   **返回**: 分页帖子列表，含作者、社区、访问量、投票数等
-   **排序说明**:
    -   order=time：按创建时间倒序
    -   order=score：按分数倒序（Redis）
    -   order=view：按访问量倒序
-   **示例**:

```
GET /api/v1/posts2/?page=1&size=20&order=view&community_id=2
```

#### 4. 更新帖子

-   **PUT** `/api/v1/post`
-   **参数（JSON）**:
    -   post_id: int
    -   title: string
    -   content: string
    -   community_id: int
-   **权限**: 需登录，作者本人可操作
-   **返回**: 更新成功/失败
-   **一致性**: 延迟双删保证缓存一致性

#### 5. 更新帖子（强一致性）

-   **PUT** `/api/v1/post/consistency`
-   **同上，强一致性版本**

#### 6. 清除帖子缓存

-   **DELETE** `/api/v1/post/:id/cache`
