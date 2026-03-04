# al — Mac 環境管理 CLI

[English](README.md)

**al** は、Mac のパッケージ（Homebrew / mas）と設定（dotfiles・シェル）を **trial → stable** のワークフローで一元管理する CLI です。Profile と Stage で環境を分離し、試用後に本番へ昇格させる運用をサポートします。

---

## Quick Start

**インストール**

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
al version
```

**初期化**（対話式は `al init --guided`）

```bash
al init
```

**パッケージの追加**

```bash
al add jq
```

**シェルの有効化** — `.zshrc` または `.bashrc` に追加:

```bash
eval "$(al activate zsh)"   # または bash
```

**GitHub から同期** / **GitHub にバックアップ**

```bash
al sync owner/repo          # リポジトリの設定を適用
al backup --init           # リポジトリ作成と push（初回）
```

---

## 機能概要

| 領域 | 主なコマンド |
|------|----------------|
| パッケージ | `al add`, `al remove`, `al promote`, `al list`, `al upgrade` |
| 差分・レビュー | `al diff`, `al review`（trial 期限） |
| 診断 | `al doctor` |
| Profile | `al profile add/list/show` |
| dotfiles | `al link add/list/remove/edit`, link.d |
| シェルスニペット | `al shell show/add/edit/enable/disable`, shell.d |
| 同期・バックアップ | `al sync [owner/repo]`, `al backup` |
| 取り込み | `al import [Brewfile] --prf <profile>` |

一覧と詳細: **[Documentation (English)](docs/en/README.md)** \| **[ドキュメント（日本語）](docs/ja/README.md)**

---

## 概念（要約）

- **Profile**: 環境の単位（例: core, work）。各 profile は独立したパッケージ一覧を持つ。`--prf` / `--profile` で指定。
- **Stage**: **trial**（試用）と **stable**（本番）。昇格は `al promote <名前>` または `al package move <名前> --to <profile>`。
- **Provider**: パッケージの取得元 — **brew**、**mas**、**manual**。デフォルトは brew。

詳細: [Concepts (en)](docs/en/concepts.md) \| [概念 (ja)](docs/ja/concepts.md)

---

## コマンド一覧

| コマンド | 説明 |
|----------|------|
| `al init` | 初回セットアップ。`--guided` で対話式。 |
| `al activate zsh`/`bash` | シェル連携（rc に `eval "$(al activate zsh)"` を追加）。 |
| `al review` | 期限切れ trial パッケージの解決。 |
| `al doctor` | 環境の診断。 |
| `al sync [owner/repo]` | 設定の適用（必要なら clone）。`--plan` でプレビュー。`--prv brew` で brew のみを同期。 |
| `al diff` | システムとプロファイルの比較。 |
| `al backup` | `~/.al` を commit して GitHub に push。 |
| `al upgrade` | provider とパッケージのアップグレード。 |
| `al config` / `al profile` / `al provider` / `al package` / `al link` / `al shell` / `al bootstrap` / `al logs` | 各サブコマンド。 |

完全なリファレンス: [docs/en/command-reference.md](docs/en/command-reference.md) \| [docs/ja/command-reference.md](docs/ja/command-reference.md)

---

## インストール方法

**最新版**

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
```

**特定バージョン**

```bash
AL_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
```

**インストール先**（デフォルト: `/usr/local/bin`）

```bash
AL_INSTALL_DIR=$HOME/bin curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
```

**セルフアップデート**

```bash
al update
```

---

## 開発

ビルド・テスト・貢献の手順は [AGENTS.md](AGENTS.md) を参照。ビルド: `make build`、テスト: `make test`。e2e テストは CI（macOS runner）で実行されます。
