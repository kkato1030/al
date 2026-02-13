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

## 2. コマンド入出力・対話

**ルール**: コマンド・サブコマンドの**入出力**（引数・フラグ・stdout/stderr・メッセージ形式・`--json`）および**対話**（確認プロンプト・対話モード・TUI）を追加・変更する場合は、**docs/command-io-guidelines.md に従う**こと。

- 新規コマンドは、ガイドラインの「入力設計」「出力形式」「メッセージ種別」「対話インターフェース」に合わせて実装する。
- 既存コマンドのリファインは、ガイドライン策定後に順次行う。§8 チェックリストに「入出力・対話は command-io-guidelines に準拠しているか」を追加してもよい。

---

## 3. 開発時の Makefile 利用

開発・検証時は **Makefile のターゲットを優先して使う**こと。以下を開発フローに組み込む。

| ターゲット | 用途 | いつ使うか |
|------------|------|------------|
| `make help` | 利用可能なターゲット一覧 | 忘れたとき |
| `make fmt` | コードフォーマット（`go fmt ./...`） | 編集後・PR 前 |
| `make vet` | 静的解析（`go vet`） | 編集後・PR 前 |
| `make lint` | golangci-lint（未インストール時はスキップ） | 品質チェック |
| `make test` | テスト実行（`go test -v ./...`）。ローカルでは e2e はスキップ。e2e を CI 相当で試す場合は `GITHUB_ACTIONS=true go test -v ./e2e/` | 機能変更後・PR 前 |
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

## 4. プロジェクト概要

- **al**: Mac のパッケージ（Homebrew / mas）と設定（dotfiles・シェル）を一元管理する CLI
- **構成**: Cobra によるサブコマンド、設定は `~/.al`（または `$AL_HOME`）
- **主要概念**: Profile / Stage（trial・stable）/ Provider（brew, mas, manual）/ link.d / shell.d / sync・backup

詳細は README.md の「概念」「機能一覧」を参照。

---

## 5. リポジトリ構成

- **`cmd/`**: コマンド定義（Cobra）。`root.go` がエントリで、`config/`・`link/`・`package/`・`profile/`・`provider/` がサブコマンド
- **`internal/`**: 設定・プロバイダ・Brewfile パース・UI など本番用ロジック（外部から import しない想定）
- **`main.go`**: エントリポイント。バージョン・ビルド情報は ldflags で注入（Makefile 参照）
- **`e2e/`**: e2e テスト。CI の macOS runner でのみ実行される（ローカルでは `GITHUB_ACTIONS` 未設定のためスキップ）。
- **`testdata/`**: テスト用データ（例: Brewfile）
- **`docs/`**: 設計・計画メモ（必要に応じて参照）。jj コマンドのチートシートは `docs/jj-cheat-sheet.md` を参照
- **`.github/workflows/`**: CI（単体テストは ubuntu、e2e は macos-latest）
- **`.goreleaser.yml`**: リリース・ビルド設定

コマンドを追加・変更するときは `cmd/` と README.md の両方を整合させる。

---

## 6. テスト・品質

- **単体テスト**: `make test` または `go test -v ./...`。ローカルでは e2e パッケージは `GITHUB_ACTIONS` 未設定のためスキップされる。
- **e2e テスト**: `e2e/` パッケージ。**GitHub Actions の macos-latest ジョブでのみ実行**される。CI では ubuntu ジョブで e2e を除外した `go test`、macOS ジョブで `go test -v ./e2e/` を実行する。e2e をローカルで実行したい場合は `GITHUB_ACTIONS=true go test -v ./e2e/`（macOS 推奨）。
- カバレッジ: `make test-coverage` → `coverage.html` を確認
- フォーマット: `make fmt` で統一
- 静的解析: `make vet`、可能なら `make lint`（golangci-lint）

新規機能・コマンド追加時は、対応するテストを追加し、`make test` が通る状態を保つこと。

---

## 7. その他の注意事項

- **エイリアス**: `al add` などは `al config alias list` で確認できるエイリアス。実体は `al package add` 等。エイリアスを変える場合は README の「エイリアス」の記述も更新する。
- **バージョン**: バイナリのバージョンは `main.Version`（ldflags）で渡している。リリース手順（GoReleaser 等）と Makefile の `VERSION` を整合させる。
- **互換性**: 設定ファイル（`~/.al` 内）の形式や CLI の破壊的変更を行う場合は、CHANGELOG や README で明示し、可能なら移行手順を書く。
- **ドキュメントの優先順位**: 利用者向けの一次情報は README.md。AGENTS.md は開発・AI 向けの手順とルールである。

---

## 8. チェックリスト（コマンド変更時）

- [ ] `cmd/` の実装を変更した
- [ ] README.md のコマンドリファレンス・機能一覧・使用例を更新した
- [ ] 入出力・対話を変更した場合、docs/command-io-guidelines.md に準拠しているか確認した
- [ ] `make fmt && make vet && make test` を実行した
- [ ] （任意）`make lint` を実行した
- [ ] （任意）`make build-release` でビルドし、`bin/al` で代表的なコマンドを手動確認した

以上に従うことで、利用者への情報提供を継続し、開発時は Makefile を軸にした一貫した手順で作業できます。

---

## 9. 開発の標準フロー（jj + GitHub）

本リポジトリは Jujutsu（jj）でバージョン管理する想定である。変更をコミットして PR まで出す一連の流れは以下を標準とする。jj の用語・詳細は `docs/jj-cheat-sheet.md` を参照。

1. **変更を加える**  
   コードやドキュメントを編集する。ワーキングコピー（@）に変更が載る。

2. **change に説明を付ける**  
   `jj describe -m "コミットメッセージ"` で現在の change（@）に説明を設定する。

3. **change を確定する**  
   `jj new` を引数なしで実行する。@ が新しい空の change に移り、直前の change が確定した状態になる。

4. **リモートに push する**  
   `jj git push -c @-` を実行する。確定した change（@ の親）に対してブックマークが自動作成され、そのブランチがリモートに push される。

5. **PR を作成する**  
   `gh pr create --base main --head <push で作成されたブランチ名> --title "..." --body "..."` で PR を作成する。ブランチ名は push 時のメッセージで表示される（例: `push-<change_id>`）。

6. **マージ後に取り込む**  
   マージ後、`jj git fetch` でリモートの最新を取得する。main を進めたい場合は `jj edit main` や `jj bookmark move main -r origin/main` などで main を更新する。

**補足**: 状態確認には `jj status`・`jj log -n 5`、取り消しには `jj undo` を利用する。

---

## 10. 開発フローで用いるツールの allowlist

以下に挙げるツール・コマンドは、開発フロー（コード編集・ビルド・テスト・jj 操作・PR 作成・fetch）の文脈で **ユーザーに確認せずに実行してよい** ものとする。AI エージェントはこれらを必要に応じて実行してよい。

**Cursor での設定**: ターミナルで確認なしに実行させるには次を利用する。
- **ルール**: `.cursor/rules/development-allowlist.mdc`（alwaysApply）でエージェントに許可を伝える。
- **ワークスペース**: `.vscode/settings.json` の `cursor.terminal.allowList` に許可コマンドの配列を記載済み。
- **参照用**: `.cursor/allowlist.json` に同じコマンド一覧を保持。設定 UI に手動で追加する場合はここを参照。

| 種別 | ツール／コマンド | 用途・備考 |
|------|------------------|-------------|
| ビルド・テスト | `make`（Makefile のターゲット） | `fmt`, `vet`, `lint`, `test`, `build`, `run`, `clean` 等（§3 参照）。Go 系はすべて make 経由とし、直接 `go` は使わない。足りなければ新規 Make ターゲットを検討する。 |
| VCS（jj） | `jj` | `status`, `log`, `describe`, `new`, `edit`, `undo`, `git push`, `git fetch`, `bookmark` 等（§9・`docs/jj-cheat-sheet.md` 参照） |
| GitHub | `gh` | `pr create`, `auth status`, `pr view`, `pr list` 等、PR 作成・状態確認 |
| Git（参照） | `git` | `status`, `branch`, `log` 等、状態確認のための読み取り系 |
| ウェブ検索 | （エージェント組み込み） | ドキュメント・API・技術情報の確認は確認せず実行してよい |
| シェル | `cd`, 上記コマンドの実行 | リポジトリルートへの移動および上記ツールの起動 |

**注意**: 上記以外のコマンド（例: システム設定変更・ネットワーク・パッケージインストール・任意のスクリプト）は、必要であればユーザーに確認してから実行すること。
