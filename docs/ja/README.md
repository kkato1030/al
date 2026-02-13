# ドキュメント（日本語）

**al** — Mac 環境を trial → stable ワークフローで管理する CLI のドキュメントです。

## 目次

- [はじめに](getting-started.md) — インストール、初期化、activate、最初のパッケージ、sync/backup
- [概念](concepts.md) — Profile、Stage、Provider、link.d、shell.d、sync/backup、extends
- [コマンドリファレンス](command-reference.md) — 全コマンド・サブコマンド
- [Brewfile からの移行](brewfile-migration.md) — Brewfile または現在の brew/mas から取り込み
- [JSON 出力](json-output.md) — 自動化・CI/CD 用の `--json`

## ガイド

- [パッケージ管理](guides/package.md) — add、remove、list、promote、`al package *`
- [Profile と Provider](guides/profile-provider.md) — `al profile *`、`al provider *`
- [Link と Shell](guides/link-shell.md) — `al link *`、`al shell *`、`al package link`
- [Sync と Backup](guides/sync-backup.md) — `al sync`、`al backup`、`--plan`、`--dry-run`
- [Bootstrap と Logs](guides/bootstrap-logs.md) — `al bootstrap *`、`al logs`
- [ワークフロー例](guides/workflow.md) — doctor、review、promote、複数プロファイル

## 他の言語

- [Documentation (English)](../en/README.md)
