# shokudo-nexus

余剰食品と子ども食堂のリアルタイム需要をgRPCでマッチングするサプライチェーン管理プラットフォーム。

## 概要

消費期限の短い生鮮食品の廃棄を最小化し、地域内での食の循環を最適化します。

### 主な機能

- 余剰食品のリアルタイム在庫登録・通知
- 複数拠点間での食材融通・リクエスト
- 寄付履歴の可視化と自治体向けレポート

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| バックエンド | Go, gRPC |
| フロントエンド | TypeScript, React |
| データベース | Firestore (GCP) |
| インフラ | GCP Cloud Run |
| Proto管理 | buf |

## Quick Start

### 前提条件

- Go 1.23+
- Node.js 22+
- buf CLI
- GCP アカウント（Firestore）

### Clone & Setup

```bash
git clone https://github.com/akaitigo/shokudo-nexus.git
cd shokudo-nexus
cp .env.example .env  # 環境変数を編集
```

### バックエンド

```bash
cd backend
go mod tidy
make build
make test
```

### フロントエンド

```bash
cd frontend
npm install
npm run dev
```

### Proto

```bash
buf lint
buf generate
```

### 全チェック

```bash
make check
```

## 環境変数

```bash
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
FIRESTORE_EMULATOR_HOST=localhost:8080  # ローカル開発時
GRPC_HOST=localhost
GRPC_PORT=9090
FIREBASE_API_KEY=your-firebase-api-key
```

> **警告**: 本番利用前に以下の対応が必要です:
> - Firebase Auth の認証フロー実装
> - Firestore セキュリティルールの設定
> - TLS 証明書の設定
> - 適切なアクセス制御の実装

## ライセンス

MIT
