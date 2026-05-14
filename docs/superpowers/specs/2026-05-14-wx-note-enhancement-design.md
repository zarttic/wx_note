# wx_note 全面打磨设计

**日期：** 2026-05-14
**策略：** 方案 C（均衡打磨），每批迭代包含体验+功能+管理三个维度的改进

## 背景

wx_note 是一个微信公众号 Markdown 编辑器，Go 后端 + Vue 3 前端，核心功能已可用。当前痛点：
- 编辑体验不够流畅（无自动保存、无快捷键）
- 内容管理缺失（无标签/分类、无素材库）
- 发布流程不便捷（每次重新填封面/作者、状态不同步）

作为开源项目，采用小步快跑节奏，每批发布都包含多维度亮点。

---

## 第一批改进 ✅ 已完成

### 1. 自动保存 ✅

**问题：** 编辑器没有自动保存，意外关闭会丢失内容。

**实现方案：**
- 编辑器内容变更后 3 秒防抖自动保存
- 新建文章首次自动保存后，路由自动跳转到 `/editor/:id`
- 保存状态指示器（工具栏区域）：
  - 已保存：显示"已自动保存 HH:MM:SS"（绿色）
  - 保存中：显示"保存中..." + 旋转图标（灰色）
  - 保存失败：显示"保存失败" + 错误详情（红色）
  - 空闲：显示"上次保存 HH:MM:SS"（灰色）
- 保留手动保存按钮，手动保存时取消待执行的自动保存
- 自动保存仅更新内容，不改变文章 status
- 加载已有文章时 `suppressAutoSave` 标志防止误触发
- 手动保存进行中时跳过自动保存（防冲突）

**文件：** `frontend/src/views/EditorView.vue`
- 新增 `autoSaveStatus` ref（'idle' | 'saving' | 'saved' | 'error'）
- 新增 `autoSaveError` ref
- 新增 `autoSaveTimer` 防抖计时器
- 拆分 `doSave({ notify, isAuto })` 核心保存逻辑，`saveArticle()` 包装调用
- 工具栏 `autoSaveDisplay` computed 提供状态显示数据

---

### 2. 文章标签系统 ✅

**问题：** 文章只有简单的状态筛选，没有分类或标签，难以管理大量文章。

**数据库变更：**
```sql
CREATE TABLE IF NOT EXISTS tags (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    TEXT NOT NULL,
    UNIQUE(user_id, name)
);
CREATE INDEX IF NOT EXISTS idx_tags_user ON tags(user_id);

CREATE TABLE IF NOT EXISTS article_tags (
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id     INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);
```

**后端变更：**
- 新增文件：`backend/internal/models/tag.go`（Tag + TagWithCount 结构体）
- 新增文件：`backend/internal/repository/tag_repo.go`（List/Create/Delete/GetByArticleID/SetArticleTags/EnsureTags）
- handler.go 新增路由：GET/POST /tags, DELETE /tags/:id
- 文章 API 变更：
  - createArticle/updateArticle 请求体增加 `tag_ids []int64`
  - getArticle/listArticles 返回数据增加 `tags []Tag`
  - listArticles 支持 `tag_id` 查询参数筛选（INNER JOIN article_tags）

**前端变更：**
- `api/client.js` 新增 `tagApi`（list/create/delete）
- `ArticleListView.vue` 增加标签筛选下拉 + 标签列显示
- `EditorView.vue` 增加标签输入区域：
  - 逗号/回车添加标签
  - Backspace 删除末尾标签
  - 自动补全下拉建议（基于用户已有标签，最多 5 个）
  - 保存时传 `tag_ids`

---

### 3. 发布流程简化 ✅

**问题：** 每次发布都要重新输入作者，发布后文章状态不更新。

**实现方案（简化版）：**
- 记住上次作者：`user_configs` 表增加 `last_author` 字段
- 发布成功后自动保存作者到 user_configs
- 编辑器加载时读取 `last_author` 自动填充，优先于 `default_author`
- 发布后更新文章状态为 `published`
- 封面记忆由前端 localStorage 处理（`wx_note_last_cover_name`）

**后端变更：**
- `db.go` 增加迁移：ALTER TABLE user_configs ADD COLUMN last_author
- `user.go` 模型增加 LastAuthor 字段
- `user_repo.go` 新增 `SaveLastAuthor()` 方法
- `handler.go` publish handler：status 改为 "published"，保存 last_author
- `handler.go` getConfig：返回 last_author

**前端变更：**
- `EditorView.vue` 发布时 author 使用 `last_author || default_author`
- 新增发布成功弹窗（替代 toast）：
  - 显示文章标题和草稿 ID
  - "前往微信公众平台查看"链接
  - "继续编辑"和"返回列表"按钮
- 封面文件名记忆：localStorage + 提示文字"上次使用：xxx.jpg"

---

## 第二批改进 ✅ 已完成

### 4. 编辑器快捷键体系 ✅

**问题：** 没有自定义快捷键，效率不高。

**实现方案：**
- 全局快捷键（keydown 事件监听 document）：
  - `Ctrl/Cmd + S` → `saveArticle()`
  - `Ctrl/Cmd + Shift + P` → `handlePublish()`（canPublish 为 false 时显示 toast 提示）
  - `Ctrl/Cmd + Shift + V` → 切换 `showPreview` ref
- macOS 检测：`navigator.platform` / `navigator.userAgent`，状态栏显示 Cmd 而非 Ctrl
- 预览面板切换：`v-show="showPreview"`，隐藏时编辑器占满全宽
- 底部状态栏：
  - 左侧：快捷键提示文字
  - 右侧：字数统计（字符数）
- 组件销毁时清理事件监听器（`onBeforeUnmount`）

**文件：** `frontend/src/views/EditorView.vue`

---

### 5. 素材库 ✅

**问题：** 上传的图片没有集中管理，无法复用，每次重新上传。

**数据库变更：**
```sql
CREATE TABLE IF NOT EXISTS media (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    filename   TEXT NOT NULL DEFAULT '',
    size       INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_media_user ON media(user_id, created_at DESC);
```

**后端变更：**
- 新增文件：`backend/internal/models/media.go`（Media 结构体）
- 新增文件：`backend/internal/repository/media_repo.go`（List/GetByID/Create/Delete）
- handler.go 新增路由：GET /media（分页）, DELETE /media/:id
- uploadImage handler：上传成功后自动创建 media 记录（URL + filename + size）

**前端变更：**
- `api/client.js` 新增 `mediaApi`（list/delete）
- 新增页面：`MediaListView.vue`
  - 4 列响应式网格，缩略图 + 文件名 + 日期
  - hover 显示复制 URL 和删除按钮
  - 删除弹窗确认
  - 空状态提示
- `router/index.js` 新增 `/media` 路由
- `App.vue` 导航栏新增素材库入口
- `EditorView.vue` 新增素材库选择器弹窗：
  - 工具栏"素材库"按钮
  - 弹窗显示素材网格，点击图片插入 `![filename](url)` 到 markdown

---

### 6. 排版主题系统 ✅

**问题：** 微信渲染只有一种固定样式，无法切换排版风格。

**实现方案：**
- 主题架构：`defaultStyles` 为完整默认样式映射，`themeStyles` 存储各主题差异覆盖
- `SetTheme(name)` 从 defaultStyles 复制 + 应用主题覆盖，更新全局 `styles` map
- 3 个内置主题：
  - **默认绿**（#07c160 主色，当前样式）
  - **商务蓝**（#1e40af 主色，blockquote 浅蓝背景 #eff6ff）
  - **简约灰**（#6b7280/#374151 主色，code_inline 灰底 #e5e7eb）
- 每个主题仅覆盖与默认值不同的样式键（h2, a, bullet, ordered_num, sup, code_inline, blockquote）
- 向后兼容：`Render(md)` 始终使用默认主题
- 新增 API：`RenderWithTheme(md, theme)`, `AvailableThemes()`, GET /themes
- 无效主题名自动 fallback 到 "default"

**后端变更：**
- `renderer/wechat.go` 重构样式系统：
  - `defaultStyles` 保留原始完整映射
  - `themeStyles` 注册各主题覆盖
  - `init()` 初始化 styles 和注册主题
  - `copyStyles()` 浅拷贝辅助函数
  - `SetTheme()` / `AvailableThemes()` / `RenderWithTheme()` / `RenderWithFootnotesAndTheme()`
- `handler.go` 新增 GET /themes 路由和 listThemes handler
- preview handler：读取 `theme` 字段，调用 `RenderWithTheme`
- publish handler：读取 `theme` 表单字段，调用 `RenderWithTheme`

**前端变更：**
- `api/client.js`：preview/publish 支持 theme 参数
- `EditorView.vue`：
  - 新增 `currentTheme` ref（默认 'default'）和 `themeOptions` 常量
  - 预览面板顶部主题切换条：3 个色块按钮 + ring 选中效果
  - `watch(currentTheme)` 触发 `updatePreview()`
  - 发布时传 `theme: currentTheme.value`

---

## 额外改进 ✅ 已完成

### 文章删除弹窗确认 ✅

**问题：** 原有删除确认使用行内文字，体验不佳。

**方案：** 改为模态弹窗确认
- 弹窗显示警告图标 + "确定要删除文章「xxx」吗？此操作无法撤销。"
- "取消"和"确认删除"按钮
- 删除中显示 loading 状态
- 点击遮罩层关闭弹窗

**文件：** `frontend/src/views/ArticleListView.vue`
- `confirmDeleteId` → `deleteTarget`（存储完整文章对象以显示标题）
- 移除行内确认 UI，改为 modal-overlay 弹窗

---

## 第三批改进 ✅ 已完成

### 7. 文章历史版本 ✅

**问题：** 文章修改后无法查看或恢复之前的内容。

**实现方案：**
- 新增 `article_revisions` 表，每次保存文章自动记录快照
- 每篇文章保留最近 30 个版本，自动清理旧版本
- 恢复版本前自动创建当前内容快照（可再次恢复回去）
- 编辑器工具栏"历史"按钮打开版本面板

**数据库变更：**
```sql
CREATE TABLE IF NOT EXISTS article_revisions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL DEFAULT '',
    markdown   TEXT NOT NULL DEFAULT '',
    word_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_revisions_article ON article_revisions(article_id, created_at DESC);
```

**后端变更：**
- 新增文件：`backend/internal/models/revision.go`（Revision 结构体）
- 新增文件：`backend/internal/repository/revision_repo.go`（Create/List/GetByID/DeleteOldRevisions）
- handler.go 新增路由：GET /articles/:id/revisions, GET /revisions/:id, POST /articles/:id/revisions/:rid/restore
- updateArticle handler：保存成功后自动创建版本快照 + 清理旧版本
- restoreRevision handler：恢复前先创建当前快照，再更新文章内容

**前端变更：**
- `api/client.js` 新增 `revisionApi`（list/get/restore）
- `EditorView.vue` 新增版本历史面板：
  - 左侧版本列表（标题 + 字数 + 日期）
  - 右侧版本内容预览
  - "恢复此版本"按钮
  - 空状态提示

---

### 8. 拖拽排序 ✅

**问题：** 文章和模板列表无法手动调整排序。

**实现方案：**
- 使用原生 HTML5 Drag and Drop API，无需额外依赖
- 拖拽手柄（GripVertical 图标），hover 时显示
- 拖拽中半透明效果，放置目标高亮
- 拖拽完成后立即调用后端保存排序

**后端变更：**
- `article_repo.go` 新增 `UpdateSortOrder()` 方法
- `template_repo.go` 新增 `UpdateSortOrder()` 方法
- `db.go` 新增迁移：articles 表增加 sort_order 列
- handler.go 新增路由：PUT /articles/reorder, PUT /templates/reorder
- 文章列表查询排序改为 `sort_order ASC, updated_at DESC`

**前端变更：**
- `api/client.js` 新增 `articleApi.reorder()`, `templateApi.reorder()`
- `ArticleListView.vue` 新增拖拽手柄列 + 行级拖拽
- `TemplateListView.vue` 新增卡片级拖拽 + 手柄

---

### 9. 一键排版 ✅

**问题：** 中文 Markdown 内容常见排版问题：中英文无空格、半角标点、多余空行。

**实现方案：**
- 纯前端实现，不需要后端支持
- 编辑器工具栏新增"排版"按钮（Sparkles 图标）
- 点击后对 markdown 内容执行格式化，无变化时提示"已符合规范"

**格式化规则：**
- 中英文间自动加空格：`你好world` → `你好 world`
- 中文数字间自动加空格：`123个` → `123 个`
- 半角标点转全角（仅中文语境）：逗号、句号、冒号、分号
- 去除全角标点前后的多余空格
- 多余空行合并为单个空行
- 行尾空格清理
- 跳过代码块（```...```）不处理

**文件：** `frontend/src/utils/formatter.js`
- `formatMarkdown(text)` 入口函数
- 自动识别并跳过代码块区域

---

## 第四批改进 ✅ 已完成

### 10. 发布流程修复 ✅

**问题：** 当前发布存在多个严重缺陷：
- 每次发布都创建新文章记录，导致重复条目
- 封面图不持久化
- AppSecret 更新 bug

**实现方案：**

**10a. 发布关联已有文章 ✅：**
- 前端发布时传 `article_id` 参数（取自 `articleId.value`）
- 后端 publish handler：如果传了 article_id 且文章存在，更新该文章的 status/draft_media_id
- 无 article_id 时仍创建新文章（兼容直接发布流程）

**10b. 封面图持久化 ✅：**
- 保存文章时 payload 包含 `cover_url` 字段
- 加载文章时从 `article.cover_url` 恢复封面预览
- 无 cover_url 时清除旧的封面预览

**10c. AppSecret 更新修复 ✅：**
- 后端 `UpdateConfig` 改为条件更新：secret 非空时更新 wechat_secret，为空时跳过
- 前端已有"留空保持不变"提示文字

---

### 11. N+1 查询优化 ✅

**问题：** listArticles 每篇文章单独查询标签，reorder 逐条 UPDATE。

**实现方案：**

**11a. 文章列表标签批量加载 ✅：**
- `tag_repo.go` 新增 `GetByArticleIDs` 方法
- 单次 IN 查询 + Go 端分组返回 `map[int64][]Tag`
- `listArticles` handler 用批量查询替代循环

**11b. 批量排序更新 ✅：**
- `article_repo.go` / `template_repo.go` 新增 `BatchUpdateSortOrder` 方法
- 使用事务批量 UPDATE
- reorder handler 改为调用批量方法

---

### 12. 编辑器体验增强 ✅

**问题：** 编辑器缺少粘贴图片、查找替换、任务列表、下载功能。

**实现方案：**

**12a. 粘贴图片支持 ✅：**
- MdEditor 添加 `@on-paste-image="handleEditorUploadImage"` prop
- 复用现有上传逻辑

**12b. 查找替换 ✅：**
- toolbar 配置添加 `'find'` 项

**12c. 任务列表 ✅：**
- toolbar 配置添加 `'task'` 项

**12d. 下载 Markdown ✅：**
- 工具栏添加"下载"按钮（Download 图标）
- Blob + 触发下载，文件名从标题生成

---

### 13. 从文章创建模板 ✅

**问题：** 模板必须从模板管理页面手动创建。

**实现方案 ✅：**
- 编辑器工具栏添加"存模板"按钮（Bookmark 图标）
- 弹窗输入模板名称和分类
- 调用 `templateApi.create()` 创建模板

---

### 14. 素材库直传 ✅

**问题：** 素材库页面没有上传入口。

**实现方案 ✅：**
- 页面头部添加"上传图片"按钮
- 支持多选文件，逐个调用 `editorApi.uploadImage` 上传
- 上传中显示进度（X/Y）
- 上传完成后自动刷新列表

---

### 15. WeChat Token 缓存 ✅

**问题：** 每次 API 请求创建新 wechat.Client，token 不共享。

**实现方案 ✅：**
- Handler 新增 `wechatClients sync.Map` 字段
- `getWechatClient(uid)` 辅助方法：优先从缓存取，否则创建并存储
- 30 分钟 TTL 后自动清理
- uploadImage、verifyWechat 改用缓存 client
