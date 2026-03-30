# shokudo-nexus

余剰食品と子ども食堂をgRPCでリアルタイムマッチングするプラットフォーム。食品廃棄最小化。

## 技術スタック
- Go: gRPCサービス
- TypeScript/React: gRPC-Web SPA
- Firestore / GCP Cloud Run
- Proto: buf lint + breaking

## ルール
- Go: golangci-lint + gofumpt
- TypeScript: `~/.claude/rules/typescript.md`
- Proto: `~/.claude/rules/proto.md`

## コマンド
```
make check     # lint → test → build
make quality   # 品質ゲート
```

## 構造
```
backend/cmd/ internal/domain/ service/ repository/
frontend/src/ components/ hooks/ lib/ types/
proto/shokudo/v1/
```

## 環境変数
`.env.example` を参照
