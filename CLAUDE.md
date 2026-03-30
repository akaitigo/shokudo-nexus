# shokudo-nexus

## 概要
余剰食品と子ども食堂のリアルタイム需要をgRPCでマッチングするサプライチェーン管理プラットフォーム。消費期限の短い生鮮食品の廃棄を最小化し、地域内での食の循環を最適化する。

## 技術スタック
- Go (API, gRPC サーバー)
- TypeScript / React (フロントエンド)
- Firestore (GCP)
- GCP Cloud Run
- buf (Proto管理)
- gRPC-Web (フロントエンド通信)

## コーディングルール
- Go: 標準の Go スタイルガイドに従うこと。golangci-lint + gofumpt で自動チェック
- TypeScript/React: `~/.claude/rules/typescript.md` のルールに従うこと。`any` 型禁止、`as` 型アサーション最小限
- Proto/gRPC: `~/.claude/rules/proto.md` のルールに従うこと。フィールド番号の再利用禁止、buf lint/breaking を実行

## ビルド & テスト

```bash
# バックエンド
cd backend && make check

# フロントエンド
cd frontend && make check

# Proto lint
buf lint

# 全チェック
make check
```

## ディレクトリ構造

```
shokudo-nexus/
├── backend/
│   ├── cmd/shokudo-nexus/   # エントリポイント
│   │   └── main.go
│   ├── internal/            # 内部パッケージ
│   ├── pkg/                 # 公開パッケージ
│   ├── scripts/             # ビルド・lint スクリプト
│   │   └── post-lint.sh
│   ├── go.mod
│   └── Makefile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── lib/
│   │   └── types/
│   ├── public/
│   ├── test/e2e/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── biome.json
│   ├── playwright.config.ts
│   └── Makefile
├── proto/
│   └── shokudo/v1/          # gRPC サービス定義
│       └── service.proto
├── test/
│   └── contract/            # 契約テスト
│       └── smoke.sh
├── docs/
│   └── quality-override.md
├── .github/
│   ├── workflows/
│   │   └── dependabot-auto-merge.yml
│   ├── ISSUE_TEMPLATE/
│   └── pull_request_template.md
├── .claude/
│   ├── CLAUDE.md
│   ├── settings.json
│   ├── startup.sh
│   └── scripts/
│       └── post-lint.sh
├── buf.yaml
├── lefthook.yml
├── Makefile
├── PRD.md
├── README.md
└── LICENSE
```

## 環境変数

```bash
# Firestore
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
FIRESTORE_EMULATOR_HOST=localhost:8080  # ローカル開発時

# gRPC
GRPC_HOST=localhost
GRPC_PORT=9090

# Firebase Auth
FIREBASE_API_KEY=your-firebase-api-key
```
