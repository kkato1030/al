# はじめに

## インストール

**最新版**

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
al version
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

## 初期化

```bash
al init
```

以下が行われる: 設定ディレクトリ `~/.al`（または `$AL_HOME`）の作成、profile `core`（stage: trial）の作成、provider `brew` の登録、デフォルト設定（profile=core, provider=brew, stage=trial）の保存。

### ガイド付き初期化（推奨）

初めて `al` を使う場合は、対話的なガイド付き初期化がおすすめです:

```bash
al init --guided
```

以下の質問に答えるだけで、自分に合った設定を自動生成できます:

1. **プロファイル設定**: Single profile (core only) / Multiple profiles (core + additional)
2. **追加プロファイル名**: Multiple profiles を選んだ場合、カンマ区切りで入力（例: work, personal）
3. **trial ワークフローの有効化**: 実験的パッケージの試用期間を設けるか
4. **レビュー期間**: trial を有効にした場合、期限日数（1/7/14/30/60日）を選択

### パッケージの追加

```bash
al add jq
al add --prv mas "Xcode" --id 497799835
```

デフォルトは brew。未インストールの場合はその場でインストールされる。`--profile`/`--prf`、`--provider`/`--prv`、`--stage`/`-s` で登録先を指定可能。

### シェルの有効化

`.zshrc` または `.bashrc` に次の 1 行を追加する（al は設定ファイルを自動編集しない）:

```bash
eval "$(al activate zsh)"   # zsh
eval "$(al activate bash)"  # bash
```

有効になる内容: Homebrew の初期化（`brew shellenv`）、shell.d の有効スニペットの source、`brew install/uninstall` および `mas install/uninstall` 実行時の al 利用の案内。Apple Silicon（`/opt/homebrew`）と Intel Mac（`/usr/local`）の両方に対応。

### 環境の同期・バックアップ

- GitHub のリポジトリから `~/.al` を clone して適用: `al sync owner/repo`（初回は clone、以降は適用のみ）
- 現在の設定を GitHub に push: `al backup --init`（リポジトリが無ければ作成して push）

次: [概念](concepts.md) | [コマンドリファレンス](command-reference.md)
