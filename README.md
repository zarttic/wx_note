# wx_note — 微信公众号 Markdown 编辑器

> 舒适的 Markdown 编辑体验 + 一键发布到微信公众号草稿箱

## 功能

- **Markdown 编辑** — md-editor-v3 (CodeMirror) 编辑器，语法高亮
- **实时预览** — 模拟手机屏幕，微信排版效果所见即所得
- **自动保存** — 3 秒防抖自动保存，意外关闭不丢内容
- **排版主题** — 内置默认绿、商务蓝、简约灰三套主题
- **文章标签** — 多对多标签系统，支持按标签筛选
- **素材库** — 编辑器上传图片自动归档，可从素材库选择插入
- **一键发布** — 直接调用微信官方 API，发布到草稿箱
- **快捷键** — Ctrl/Cmd+S 保存、Ctrl/Cmd+Shift+P 发布、Ctrl/Cmd+Shift+V 切换预览
- **模板系统** — 创建和管理 Markdown 模板，一键应用

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + Vite + md-editor-v3 + Pinia + Tailwind CSS + Lucide Icons |
| 后端 | Go (Gin + SQLite + goldmark 自定义渲染器) |
| 微信 API | 官方微信 API（获取 token → 上传素材 → 新增草稿） |

## 快速开始

### 1. 公众号准备

登录 [微信公众平台](https://mp.weixin.qq.com) → 设置与开发 → 基本配置，获取：
- **AppID**
- **AppSecret**

### 2. 一键启动

```bash
chmod +x start.sh
./start.sh
```

然后打开浏览器访问 **http://localhost:5173**

### 3. 手动启动

**后端：**
```bash
cd backend
go build -o wx-note ./cmd/server/
./wx-note
```
默认监听 `0.0.0.0:8100`，可通过 `PORT` 环境变量修改。数据存储在 `./data/` 目录。

**前端：**
```bash
cd frontend
npm install
npm run dev
```

## 使用流程

1. **注册/登录** — 首次使用创建账号
2. **配置公众号** — 进入设置页，填入 AppID 和 AppSecret，点击验证连接
3. **编辑文章** — 左侧写 Markdown，右侧实时预览微信排版效果
4. **选择封面** — 点击封面区域上传图片（建议 900×383 或 2.35:1 比例）
5. **添加标签** — 在发布栏输入标签，按回车或逗号添加
6. **切换主题** — 预览面板顶部选择排版主题
7. **一键发布** — 点击「发布草稿」按钮
8. **公众号后台** — 登录 mp.weixin.qq.com 查看草稿并手动发布

## 项目结构

```
wx_note/
├── backend/
│   ├── cmd/server/main.go          # 入口
│   ├── internal/
│   │   ├── api/handler.go          # API 路由与处理
│   │   ├── middleware/auth.go      # JWT 认证中间件
│   │   ├── models/                 # 数据模型
│   │   │   ├── article.go
│   │   │   ├── tag.go
│   │   │   ├── media.go
│   │   │   ├── template.go
│   │   │   └── user.go
│   │   ├── repository/             # 数据库访问层
│   │   │   ├── db.go               # 数据库初始化
│   │   │   ├── schema.sql          # 表结构定义
│   │   │   ├── article_repo.go
│   │   │   ├── tag_repo.go
│   │   │   ├── media_repo.go
│   │   │   ├── template_repo.go
│   │   │   └── user_repo.go
│   │   ├── renderer/wechat.go      # goldmark 微信渲染器（内联样式 + section 标签）
│   │   └── wechat/client.go        # 微信 API 客户端
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.vue                 # 导航布局
│   │   ├── router/index.js         # 路由配置
│   │   ├── stores/auth.js          # Pinia 认证状态
│   │   ├── api/client.js           # API 客户端
│   │   └── views/
│   │       ├── EditorView.vue      # 编辑器 + 预览 + 发布
│   │       ├── ArticleListView.vue # 文章管理（搜索/筛选/标签）
│   │       ├── MediaListView.vue   # 素材库
│   │       ├── TemplateListView.vue
│   │       ├── TemplateEditView.vue
│   │       ├── SettingsView.vue    # 公众号配置 + 个人信息
│   │       ├── LoginView.vue
│   │       └── RegisterView.vue
│   └── vite.config.js
├── start.sh
└── README.md
```

## API 路由

### 公开路由

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 用户注册 |
| POST | `/api/auth/login` | 用户登录 |
| GET | `/api/health` | 健康检查 |

### 认证路由

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/logout` | 退出登录 |
| GET/PUT | `/api/user/profile` | 个人信息 |
| GET/PUT | `/api/user/config` | 公众号配置 |
| GET | `/api/articles` | 文章列表（支持 search/status/tag_id 筛选） |
| POST | `/api/articles` | 创建文章（支持 tag_ids） |
| GET/PUT/DELETE | `/api/articles/:id` | 文章详情/更新/删除 |
| GET | `/api/tags` | 标签列表（含文章计数） |
| POST | `/api/tags` | 创建标签 |
| DELETE | `/api/tags/:id` | 删除标签 |
| GET | `/api/media` | 素材库列表 |
| DELETE | `/api/media/:id` | 删除素材 |
| POST | `/api/editor/preview` | Markdown 预览（支持 theme 参数） |
| POST | `/api/editor/upload-image` | 上传图片到微信 |
| POST | `/api/editor/verify` | 验证微信连接 |
| POST | `/api/editor/publish` | 发布到草稿箱（支持 theme 参数） |
| GET | `/api/themes` | 获取可用主题列表 |
| GET/POST | `/api/templates` | 模板列表/创建 |
| GET/PUT/DELETE | `/api/templates/:id` | 模板详情/更新/删除 |
| GET | `/api/templates/categories/all` | 模板分类列表 |

## 排版主题

内置三套微信排版主题，可在预览面板实时切换：

| 主题 | 主色 | 风格 |
|------|------|------|
| 默认 | #07c160 微信绿 | 经典微信风格 |
| 商务 | #1e40af 深蓝 | 正式商务排版 |
| 简约 | #6b7280 灰色 | 极简黑白灰 |

## 快捷键

| 快捷键 | 功能 |
|--------|------|
| Ctrl/Cmd + S | 保存文章 |
| Ctrl/Cmd + Shift + P | 发布到草稿箱 |
| Ctrl/Cmd + Shift + V | 切换预览面板 |

## 注意事项

- 发布功能仅支持已配置 AppID + AppSecret 的公众号（订阅号/服务号均可）
- 草稿箱中的文章需要在公众号后台手动确认发布
- 微信渲染器使用 section + strong 替代 h1/h2/h3 标签，使用 flex 布局替代 table，确保微信客户端正确显示
- 链接自动转换为脚注形式（微信不支持外链跳转）
