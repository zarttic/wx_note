# 📝 wx_note — 微信公众号 Markdown 编辑器

> 舒适的 Markdown 编辑体验 + 一键发布到微信公众号草稿箱

## 功能

- ✏️ **Markdown 编辑** — CodeMirror 6 编辑器，语法高亮
- 👀 **实时预览** — 模拟手机屏幕，微信排版主题所见即所得
- 🚀 **一键发布** — 直接调用微信官方 API，发布到草稿箱
- 📷 **封面图片** — 自动上传并设置为文章封面

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + Vite + CodeMirror 6 + markdown-it + Tailwind CSS |
| 后端 | Python FastAPI |
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

### 3. 手动启动（前后端分离）

**后端：**
```bash
cd backend
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
cp .env.example .env
# 编辑 .env 填入 app_id 和 app_secret
uvicorn main:app --reload --host 127.0.0.1 --port 8100
```

**前端：**
```bash
cd frontend
npm install
npm run dev
```

## 使用流程

1. **编辑文章** — 左侧写 Markdown，右侧实时预览微信排版效果
2. **选择封面** — 点击封面区域上传图片（建议使用 900×383 或 2.35:1 比例）
3. **配置账号** — 点击「配置」标签，填入 AppID 和 AppSecret
4. **一键发布** — 点击「发布到草稿箱」按钮
5. **公众号后台** — 登录 mp.weixin.qq.com 查看草稿并手动发布

## 项目结构

```
wx_note/
├── backend/              # Python FastAPI 后端
│   ├── main.py           # API 路由
│   ├── wechat_api.py     # 微信官方 API 客户端
│   ├── requirements.txt
│   └── .env              # 公众号配置（不提交到 git）
├── frontend/             # Vue 3 前端
│   ├── src/
│   │   ├── App.vue       # 主界面（编辑器 + 预览 + 发布）
│   │   ├── components/
│   │   │   └── MarkdownEditor.vue  # CodeMirror 编辑器
│   │   └── api.js        # API 封装
│   └── vite.config.js
├── start.sh              # 一键启动脚本
└── README.md
```

## API 路由

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/config` | 获取公众号配置 |
| POST | `/api/config` | 更新公众号配置 |
| POST | `/api/verify` | 验证 app_id/app_secret |
| POST | `/api/preview` | Markdown → 微信 HTML 预览 |
| POST | `/api/upload-image` | 上传图片到微信素材库 |
| POST | `/api/publish` | 一键发布到草稿箱 |

## 注意事项

- 发布功能仅支持已配置 AppID + AppSecret 的公众号（订阅号/服务号均可）
- 草稿箱中的文章需要在公众号后台手动确认发布
- `.env` 和 `config.json` 包含敏感信息，已加入 `.gitignore`
