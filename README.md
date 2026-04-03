# Zipic

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

🚀 A fast, modern image compression tool. Compress images in browser with adjustable quality, batch processing, and ZIP download.

快速高效的在线图片压缩工具，支持批量处理和ZIP打包下载。

高速で効率的な画像圧縮ツール、一括処理とZIP ダウンロード対応。

![20260403200959](https://cdn.jsdelivr.net/gh/yezihack/assets/b/20260403200959.png)

---

## [English](#english) | [中文](#中文) | [日本語](#日本語)

---

<a name="english"></a>

## English

### Features

- 🚀 **Fast Compression** - Efficient image compression with adjustable quality (10-100%)
- 📁 **Batch Processing** - Compress up to 10 images at once
- 📦 **ZIP Download** - Automatically package multiple images into ZIP
- 🌐 **Multi-language** - Supports English, Chinese, Japanese
- 💾 **History Records** - Local storage of up to 100 compression records
- 🔒 **Privacy First** - Files auto-deleted after 1 day, no data collection
- 📱 **Responsive Design** - Works on desktop and mobile devices

### Tech Stack

- **Backend**: Go 1.22+, Gin, imaging
- **Frontend**: Vue 3, Element Plus, TypeScript, vue-i18n
- **Build**: Single binary with embedded frontend

### Quick Start

```bash
# Build
./build.ps1  # Windows
./build.sh   # Linux/Mac

# Run
./backend/bin/zipic.exe

# Access
http://localhost:8040
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/compress` | Compress single image |
| POST | `/api/batch-compress` | Batch compress images |
| GET | `/api/download?filename=xxx` | Download single file |
| GET | `/api/download-zip?batch_id=xxx` | Download ZIP package |
| GET | `/api/preview?filename=xxx` | Preview image |
| GET | `/health` | Health check |

### Configuration

Default port: `8040`

Custom port via environment variable:
```bash
PORT=9000 ./zipic
```

---

<a name="中文"></a>

## 中文

### 功能特点

- 🚀 **快速压缩** - 高效图片压缩，质量可调（10-100%）
- 📁 **批量处理** - 一次最多压缩 10 张图片
- 📦 **ZIP 下载** - 多张图片自动打包为 ZIP
- 🌐 **多语言支持** - 支持中文、英文、日文
- 💾 **历史记录** - 本地存储最多 100 条压缩记录
- 🔒 **隐私优先** - 文件 1 天后自动删除，不收集数据
- 📱 **响应式设计** - 适配桌面和移动设备

### 技术栈

- **后端**: Go 1.22+, Gin, imaging
- **前端**: Vue 3, Element Plus, TypeScript, vue-i18n
- **构建**: 单一二进制文件，内嵌前端

### 快速开始

```bash
# 构建
./build.ps1  # Windows
./build.sh   # Linux/Mac

# 运行
./backend/bin/zipic.exe

# 访问
http://localhost:8040
```

### API 接口

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/api/compress` | 压缩单张图片 |
| POST | `/api/batch-compress` | 批量压缩图片 |
| GET | `/api/download?filename=xxx` | 下载单个文件 |
| GET | `/api/download-zip?batch_id=xxx` | 下载 ZIP 包 |
| GET | `/api/preview?filename=xxx` | 预览图片 |
| GET | `/health` | 健康检查 |

### 配置

默认端口：`8040`

通过环境变量自定义端口：
```bash
PORT=9000 ./zipic
```

---

<a name="日本語"></a>

## 日本語

### 機能

- 🚀 **高速圧縮** - 効率的な画像圧縮、品質調整可能（10-100%）
- 📁 **一括処理** - 一度に最大10枚の画像を圧縮
- 📦 **ZIP ダウンロード** - 複数画像を自動的にZIPにパッケージ
- 🌐 **多言語対応** - 日本語、英語、中国語をサポート
- 💾 **履歴記録** - ローカルに最大100件の圧縮履歴を保存
- 🔒 **プライバシー重視** - ファイルは1日後に自動削除、データ収集なし
- 📱 **レスポンシブデザイン** - デスクトップとモバイルに対応

### 技術スタック

- **バックエンド**: Go 1.22+, Gin, imaging
- **フロントエンド**: Vue 3, Element Plus, TypeScript, vue-i18n
- **ビルド**: 単一バイナリ、フロントエンド埋め込み

### クイックスタート

```bash
# ビルド
./build.ps1  # Windows
./build.sh   # Linux/Mac

# 実行
./backend/bin/zipic.exe

# アクセス
http://localhost:8040
```

### API エンドポイント

| メソッド | エンドポイント | 説明 |
|----------|----------------|------|
| POST | `/api/compress` | 単一画像の圧縮 |
| POST | `/api/batch-compress` | 一括画像圧縮 |
| GET | `/api/download?filename=xxx` | ファイルダウンロード |
| GET | `/api/download-zip?batch_id=xxx` | ZIP ダウンロード |
| GET | `/api/preview?filename=xxx` | 画像プレビュー |
| GET | `/health` | ヘルスチェック |

### 設定

デフォルトポート：`8040`

環境変数でポート指定：
```bash
PORT=9000 ./zipic
```

---

## License / 许可证 / ライセンス

MIT License

---

## Contributing / 贡献 / 貢献

Issues and Pull Requests are welcome!

欢迎提交 Issue 和 Pull Request！

IssueやPull Requestを歓迎します！