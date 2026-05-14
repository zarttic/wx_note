# wx_note 全面打磨设计

**日期：** 2026-05-14
**分支：** feature/enhance-wx-note
**策略：** 方案 C（均衡打磨），每批迭代包含体验+功能+管理三个维度的改进

## 背景

wx_note 是一个微信公众号 Markdown 编辑器，Go 后端 + Vue 3 前端，核心功能已可用。当前痛点：
- 编辑体验不够流畅（无自动保存、无快捷键）
- 内容管理缺失（无标签/分类、无素材库）
- 发布流程不便捷（每次重新填封面/作者、状态不同步）

作为开源项目，采用小步快跑节奏，每批发布都包含多维度亮点。

---

## 第一批改进

### 1. 自动保存

**问题：** 编辑器没有自动保存，意外关闭会丢失内容。

**方案：**
- 编辑器内容变更后 3 秒防抖自动保存
- 新建文章首次自动保存后，路由自动跳转到 `/editor/:id`
- 保存状态指示器（工具栏区域）：
  - 已保存：显示"已自动保存 HH:MM:SS"
  - 保存中：显示"保存中..." + 旋转图标
  - 保存失败：显示"保存失败" + 错误提示
- 保留手动保存按钮，但改为次要操作
- 自动保存仅更新内容，不改变文章 status

**后端变更：** 无，现有 PUT /articles/:id API 已支持
**前端变更：** EditorView.vue 中 watch markdown 触发自动保存，增加保存状态枚举

---

### 2. 文章标签系统

**问题：** 文章只有简单的状态筛选，没有分类或标签，难以管理大量文章。

**方案：**
- 多对多标签：一篇文章可有多个标签，一个标签可属于多篇文章
- 编辑器右侧发布栏增加标签输入（逗号分隔输入，自动去重）
- 文章列表页增加标签筛选器
- 文章列表页每篇文章显示标签徽章
- 标签自动补全：基于用户已有标签

**数据库变更：**
```sql
CREATE TABLE tags (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    TEXT NOT NULL,
    UNIQUE(user_id, name)
);
CREATE INDEX idx_tags_user ON tags(user_id);

CREATE TABLE article_tags (
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id     INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);
```

**后端变更：**
- 新增 tags CRUD API：GET/POST /tags, DELETE /tags/:id
- 文章 API 增加 tags 字段：创建/更新时传入 tag_ids，返回时包含 tags 数组
- 文章列表 API 支持按 tag_id 筛选

**前端变更：**
- ArticleListView 增加标签筛选下拉
- EditorView 发布栏增加标签输入组件
- 标签自动补全组件

---

### 3. 发布流程简化

**问题：** 每次发布都要重新上传封面、输入作者，发布后文章状态不更新。

**方案：**
- 记住上次发布设置：
  - 后端 `user_configs` 表增加 `last_cover_url` 和 `last_author` 字段
  - 发布成功后自动保存本次设置到 user_configs
  - 编辑器加载时读取上次设置自动填充
- 发布后更新文章状态：
  - 已有文章：更新 status 为 `published`，保存 `draft_media_id`
  - 新建文章：创建时直接标记为 `published`
- 发布成功弹窗（替代 toast）：
  - 显示文章标题和草稿 ID
  - 包含"前往微信公众平台查看"链接
  - 包含"继续编辑"和"返回列表"按钮

**后端变更：**
- user_configs 表增加 last_cover_url、last_author 字段
- publish handler 发布成功后更新文章状态和用户配置
- getConfig API 返回 last_cover_url 和 last_author

**前端变更：**
- EditorView 发布栏自动填充上次设置
- 新增发布成功弹窗组件

---

## 第二批改进

### 4. 编辑器快捷键体系

**问题：** 没有自定义快捷键，效率不高。

**方案：**
- 全局快捷键：
  - `Ctrl/Cmd + S` — 保存文章
  - `Ctrl/Cmd + Shift + P` — 发布到草稿箱
  - `Ctrl/Cmd + Shift + V` — 切换预览面板显隐
- md-editor-v3 内置快捷键保留（加粗 Ctrl+B、斜体 Ctrl+I 等）
- 编辑器底部状态栏显示快捷键提示
- 快捷键不与 md-editor-v3 冲突（全局操作 vs 编辑区内操作）

**前端变更：** EditorView 添加 keydown 事件监听

---

### 5. 素材库

**问题：** 上传的图片没有集中管理，无法复用，每次重新上传。

**方案：**
- 新增 media 表存储图片记录
- 编辑器图片上传成功后自动记录到素材库
- 新增素材库页面（路由 `/media`），网格视图展示已上传图片
- 编辑器插入图片增加"从素材库选择"选项
- 图片信息：URL、文件名、上传时间

**数据库变更：**
```sql
CREATE TABLE media (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    filename   TEXT NOT NULL DEFAULT '',
    size       INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_media_user ON media(user_id, created_at DESC);
```

**后端变更：**
- 新增 media CRUD API：GET /media, DELETE /media/:id
- uploadImage handler 上传成功后自动创建 media 记录

**前端变更：**
- 新增 MediaListView 页面
- EditorView 图片插入增强（本地上传 + 素材库选择）
- App.vue 导航栏增加素材库入口

---

### 6. 排版主题系统

**问题：** 微信渲染只有一种固定样式，无法切换排版风格。

**方案：**
- 渲染器支持主题概念：一组样式映射预设
- 内置 3 个主题：
  - **默认绿**（当前样式，#07c160 主色）
  - **商务蓝**（#1e40af 主色，正式风格）
  - **简约灰**（黑白灰，极简风格）
- 预览面板顶部增加主题切换按钮，实时预览
- 主题只影响渲染输出，不影响编辑器
- 后端 preview API 增加可选 `theme` 参数，默认 `default`
- publish 时使用当前选择的主题渲染

**后端变更：**
- renderer/wechat.go 的 styles map 改为按主题加载
- 新增主题注册机制，支持通过函数注册新主题
- preview handler 增加 theme query parameter

**前端变更：**
- 预览面板顶部增加主题切换器（3 个色块按钮）
- 切换主题时重新调用 preview API
- 发布时传递当前主题

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
