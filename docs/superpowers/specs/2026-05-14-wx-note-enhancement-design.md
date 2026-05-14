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

## 第三批改进（远景）

### 7. 文章历史版本
- 新增 article_revisions 表，每次保存记录快照
- 编辑器可查看和回滚历史版本

### 8. 拖拽排序
- 文章列表支持拖拽调整排序
- 模板列表支持拖拽排序

### 9. 一键排版
- 自动优化 Markdown 内容：中英文间加空格、标点修正、段落间距统一
- 前端实现，不需要后端支持
