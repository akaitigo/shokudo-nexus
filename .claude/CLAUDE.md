# shokudo-nexus アーキテクチャ

## アーキテクチャ概要

モノレポ構成。バックエンドは Go で gRPC サービスを実装し、フロントエンドは TypeScript/React で gRPC-Web 経由で通信する。

```
[React SPA] --gRPC-Web--> [Go gRPC Server] ---> [Firestore]
                              |
                         [Cloud Run]
```

## 主要な設計判断

<!-- ADR で記録する。ここにはサマリーのみ -->

- **バックエンド言語: Go** -- gRPC エコシステムの成熟度。Kotlin/Quarkus から変更（tech-reeval による）
- **データベース: Firestore** -- スキーマレスで素早いMVP開発。リアルタイムリスナー対応
- **Proto管理: buf** -- lint + breaking change検出 + フォーマット統一

## 外部サービス連携

| サービス | 用途 | 認証方式 |
|---------|------|---------|
| Firestore | データストア | サービスアカウント |
| Firebase Auth | ユーザー認証 | API Key + OAuth |
| Cloud Run | ホスティング | GCP IAM |

## ドメインモデル概要

- **FoodItem**: 余剰食品（名前、カテゴリ、消費期限、数量、提供元）
- **Shokudo**: 子ども食堂（名前、住所、連絡先、必要食材リスト）
- **FusionRequest**: 拠点間の食材融通リクエスト
- **DonationRecord**: 寄付履歴（レポート用）
