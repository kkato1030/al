# AGENTS.md — 開発者・AI エージェント向けガイド

このファイルは、本リポジトリ（al）の開発・改修時に従うべき方針と手順をまとめたものです。人間の開発者および AI エージェントは、変更を加える際にこの内容を参照してください。

---

## 1. コマンド変更時の README 更新（必須）

**ルール**: `al` の CLI コマンド・サブコマンド・オプション・エイリアス・挙動に変更を加えた場合は、**必ず README.md を更新**すること。

- **対象**: ルートコマンド・`al config`・`al profile`・`al provider`・`al package`・`al link` の追加・削除・名前変更・フラグ変更・説明の変更
- **更新箇所**: README.md の「コマンドリファレンス」「機能一覧」「Quick Start」「使用例」など、影響する記述すべて
- **目的**: 利用者が README だけで正しく `al` を使える状態を維持する

作業の流れの例:
1. `cmd/` 以下でコマンド実装を変更
2. 変更内容を README.md に反映（表・説明文・使用例）
3. 必要なら「Brewfile からの移行」「概念」の記述も整合させる

---

## 2. 開発時の Makefile 利用

開発・検証時は **Makefile のターゲットを優先して使う**こと。以下を開発フローに組み込む。

| ターゲット | 用途 | いつ使うか |
|------------|------|------------|
| `make help` | 利用可能なターゲット一覧 | 忘れたとき |
| `make fmt` | コードフォーマット（`go fmt ./...`） | 編集後・PR 前 |
| `make vet` | 静的解析（`go vet`） | 編集後・PR 前 |
| `make lint` | golangci-lint（未インストール時はスキップ） | 品質チェック |
| `make test` | テスト実行（`go test -v ./...`） | 機能変更後・PR 前 |
| `make test-coverage` | カバレッジ取得＋HTML レポート | テスト追加・リファクタ時 |
| `make build` | バイナリビルド（`bin/al`） | 動作確認 |
| `make build-dev` | 開発用ビルド（fmt + vet 後にビルド） | 日常の動作確認 |
| `make build-release` | リリース用ビルド（fmt + vet + test 後） | リリース・配布用 |
| `make run ARGS="..."` | ビルドせず実行（例: `make run ARGS="version"`） | 手軽な動作確認 |
| `make install` | GOPATH/bin にインストール | ローカルで `al` を常時使う場合 |
| `make clean` | ビルド成果物・キャッシュ削除 | クリーンビルドしたいとき |
| `make build-darwin` | macOS amd64/arm64 向けクロスビルド | 配布用バイナリ作成 |

**推奨フロー（例）**:
- コードを触ったら: `make fmt && make vet && make test`
- PR を出す前に: 上記に加えて `make lint`（golangci-lint がある場合）
- リリース前: `make build-release` でビルドし、`bin/al` で動作確認

変数:
- `VERSION?=0.1.0` でバージョン指定可能（未指定時は 0.1.0）
- `ARGS` は `make run` 用（例: `make run ARGS="sync --help"`）

---

## 3. プロジェクト概要

- **al**: Mac のパッケージ（Homebrew / mas）と設定（dotfiles・シェル）を一元管理する CLI
- **構成**: Cobra によるサブコマンド、設定は `~/.al`（または `$AL_HOME`）
- **主要概念**: Profile / Stage（trial・stable）/ Provider（brew, mas, manual）/ link.d / shell.d / sync・backup

詳細は README.md の「概念」「機能一覧」を参照。

---

## 4. リポジトリ構成

- **`cmd/`**: コマンド定義（Cobra）。`root.go` がエントリで、`config/`・`link/`・`package/`・`profile/`・`provider/` がサブコマンド
- **`internal/`**: 設定・プロバイダ・Brewfile パース・UI など本番用ロジック（外部から import しない想定）
- **`main.go`**: エントリポイント。バージョン・ビルド情報は ldflags で注入（Makefile 参照）
- **`testdata/`**: テスト用データ（例: Brewfile）
- **`docs/`**: 設計・計画メモ（必要に応じて参照）。jj コマンドのチートシートは `docs/jj-cheat-sheet.md` を参照
- **`.github/workflows/`**: CI（例: release）
- **`.goreleaser.yml`**: リリース・ビルド設定

コマンドを追加・変更するときは `cmd/` と README.md の両方を整合させる。

---

## 5. テスト・品質

- テスト: `make test` または `go test -v ./...`
- カバレッジ: `make test-coverage` → `coverage.html` を確認
- フォーマット: `make fmt` で統一
- 静的解析: `make vet`、可能なら `make lint`（golangci-lint）

新規機能・コマンド追加時は、対応するテストを追加し、`make test` が通る状態を保つこと。

---

## 6. その他の注意事項

- **エイリアス**: `al add` などは `al config alias list` で確認できるエイリアス。実体は `al package add` 等。エイリアスを変える場合は README の「エイリアス」の記述も更新する。
- **バージョン**: バイナリのバージョンは `main.Version`（ldflags）で渡している。リリース手順（GoReleaser 等）と Makefile の `VERSION` を整合させる。
- **互換性**: 設定ファイル（`~/.al` 内）の形式や CLI の破壊的変更を行う場合は、CHANGELOG や README で明示し、可能なら移行手順を書く。
- **ドキュメントの優先順位**: 利用者向けの一次情報は README.md。AGENTS.md は開発・AI 向けの手順とルールである。

---

## 7. チェックリスト（コマンド変更時）

- [ ] `cmd/` の実装を変更した
- [ ] README.md のコマンドリファレンス・機能一覧・使用例を更新した
- [ ] `make fmt && make vet && make test` を実行した
- [ ] （任意）`make lint` を実行した
- [ ] （任意）`make build-release` でビルドし、`bin/al` で代表的なコマンドを手動確認した

以上に従うことで、利用者への情報提供を継続し、開発時は Makefile を軸にした一貫した手順で作業できます。
