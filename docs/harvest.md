# Harvest: shokudo-nexus

## 使えたもの
- [x] Makefile (make check / make quality)
- [x] lint設定 (golangci-lint + oxlint + biome + buf)
- [x] CI YAML
- [x] CLAUDE.md (30行、50行以下達成)
- [ ] ADR テンプレート (ADR未作成)
- [x] 品質チェックリスト (make quality)
- [x] E2Eテスト雛形 (test/contract/smoke.sh)
- [x] Hooks（PostToolUse golangci-lint + oxlint）
- [x] lefthook
- [x] startup.sh

## 使えなかったもの（理由付き）
- ADR: 作成されなかった。gRPC+Firestoreの技術選定判断があったがADR化されず

## テンプレート改善提案

| 対象ファイル | 変更内容 | 根拠 |
|-------------|---------|------|
| buf.yaml テンプレート | managed フィールドコメントアウト済み（v5.8で対応） | buf v2でlint時にエラー |
| service.proto テンプレート | DeleteレスポンスにEmpty型を使わない | buf lint STANDARD違反 |
| idea-work SKILL.md | FusionService引数変更時にmain.goの更新を配線確認で検出すべき | コンパイルエラーがShip時に発見された |

## メトリクス

| 項目 | 値 |
|------|-----|
| Issue (closed/total) | 5/5 |
| PR merged | 5 |
| テスト数 | 60+ |
| CI失敗数 | 1 |
| ADR数 | 0 |
| テンプレート実装率 | 90% |
| CLAUDE.md行数 | 30 |

## 次のPJへの申し送り
- gRPC PJでは proto定義変更→gen→service→main.go の配線を必ず確認
- Firestore エミュレータがないとリポジトリ層の統合テストが書けない。CI環境整備が課題
