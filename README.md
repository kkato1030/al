# al - Mac Management Tools

Mac のパッケージ（Homebrew / mas）と設定（dotfiles・シェル）を一元管理する CLI。Profile と Stage（trial / stable）で環境を分離し、試用後に本番へ昇格させる運用をサポートする。

---

## Quick Start

### インストール

```bash
curl -fsSL https://raw.githubusercontent.com/kkato1030/al/main/install.sh | bash
al version
```

### 初期化

```bash
al init
```

以下が行われる: 設定ディレクトリ `~/.al`（または `$AL_HOME`）の作成、profile `core`（stage: trial）の作成、provider `brew` の登録、デフォルト設定（profile=core, provider=brew, stage=trial）の保存。

### パッケージの追加

```bash
al add jq
al add --prv mas "Xcode" --id 497799835
```

デフォルトは brew。未インストールの場合はその場でインストールされる。`--profile` / `--prf`、`--provider` / `--prv`、`--stage` / `-s` で登録先を指定可能。

### シェルの有効化

`.zshrc` または `.bashrc` に次の 1 行を追加する（al は設定ファイルを自動編集しない）:

```bash
eval "$(al activate zsh)"   # zsh
eval "$(al activate bash)"  # bash
```

有効になる内容: shell.d の有効スニペットの source、`brew install/uninstall` および `mas install/uninstall` 実行時の al 利用の案内（`y` で従来どおり実行）。

### 環境の同期・バックアップ

- GitHub のリポジトリから `~/.al` を clone して適用: `al sync owner/repo`（初回は clone、以降は適用のみ）
- 現在の設定を GitHub に push: `al backup --init`（リポジトリが無ければ作成して push）

---

## 機能一覧

| 機能 | 説明 | 主なコマンド |
|------|------|----------------|
| パッケージ管理 | Homebrew / mas の追加・削除・昇格・一覧・アップグレード | `al add`, `al remove`, `al promote`, `al list`, `al upgrade` |
| trial レビュー | trial のレビュー期限切れパッケージの対話的解決（remove / promote / postpone） | `al review` |
| Profile | 用途別の環境分離（例: work / private） | `al profile add`, `al profile list`, `al profile show` |
| link.d | dotfiles を `~/.al/link.d/` に集約し、ユーザパスを symlink に | `al link add/list/remove/edit` |
| shell.d | パッケージごとのシェルスニペットと読み込み順の管理 | `al shell show/add/edit/remove/enable/disable` |
| bootstrap | 新規 PC セットアップ用のワンオフシェルスクリプト（`~/.al/bootstrap/script.sh`） | `al bootstrap add/edit/remove/show` |
| sync / backup | GitHub との clone 適用・push | `al sync [owner/repo]`, `al backup` |
| Brewfile 取り込み | 既存 Brewfile を al の管理下に登録 | `al import Brewfile --prf <profile>` |
| 設定 | デフォルト profile / provider / stage、バックアップ先、エイリアス | `al config set/show`, `al config alias list` |

エイリアス一覧は `al config alias list`。例: `al add` → `al package add`、`al promote` → trial から promote_to への移動。

---

## 概念

### Profile と Stage

- **Profile**: 環境の単位（例: core, work, private）。各 profile は独立したパッケージ一覧を持つ。`al add --prf <profile> ...` で登録先を指定する。
- **Stage**: 試用（trial）か本番（stable）かを表す。Profile 名は `profile_name.stage_name`（例: `core.trial`, `core.stable`）。`al init` で作成されるのは `core`（stage: trial）。カスタム profile は `al profile add` で追加する。

### promote（昇格）

Profile に `promote_to` を設定すると、パッケージを trial から stable などへ移動できる。`al package move <name> --to <profile>` またはエイリアス `al promote <name>`（パッケージが属する profile の promote_to へ移動）。同一パッケージ（同一 ID・provider）が trial と stable の両方に同時に存在することはない。

### trial のレビュー期限（review）

Profile に `review_days` を設定すると、その profile のパッケージがレビュー対象になる。`review_days` が無い profile はレビュー対象外。パッケージの `review_by`（レビュー期限日時）が無いか期限を過ぎていれば「期限切れ」として扱う。`al activate` 実行時に期限切れがあると stderr に一覧と「`al review` で解決」の案内を表示する。`al review` では各パッケージについて **remove**（使わなかったので削除）、**promote**（stable へ昇格）、**postpone**（同じ日数だけ延期＝`review_by` を今＋review_days に更新）のいずれかを選べる。postpone 時は「本当に？ それが必要なの？」の確認プロンプトが出る。

### Provider

パッケージのインストール元。**brew**（Homebrew formula/cask）、**mas**（Mac App Store）、**manual**（登録のみ、インストールは行わない）。`al init` で brew が追加される。mas は `al provider add mas`。デフォルト provider は `al config set --default-provider` で変更可能。

provider 間の依存関係は `providers.json` の `depends_on` で管理される。デフォルトでは **mas は brew に依存**（`depends_on: ["brew"]`）し、`al provider add mas` や `al sync` では依存順（brew → mas）で処理される。

**brew tap**: `brew tap` で追加する tap はパッケージとは別に provider brew 側で管理する（`~/.al/brew-taps.json`）。`al list` には tap は出さない。`al add homebrew/cask-fonts` のように tap を指定すると brew-taps に登録される。tap の更新（Homebrew と全 tap の最新化）は `al provider upgrade` または `al upgrade` で行う（brew の場合は内部で `brew update` が実行される）。`al provider prune` では、packages に `owner/repo/toolname` 形式のパッケージが1つもない tap を untap し、brew-taps から削除する。

### link.d

設定ファイル・ディレクトリの実体を `~/.al/link.d/<name>/content` に置き、ユーザが参照するパス（例: `~/.config/foo`）をその content への symlink にする。sync / backup で他マシンや GitHub と揃えられる。追加は `al link add <name> <user_path>`。パッケージに紐づける場合は `al package link add/remove/edit`（link 名 = パッケージ名の 1:1 想定）。

### shell.d

パッケージごとのシェル設定を `~/.al/shell.d/<パッケージ識別子>/` に置き、`al activate` の出力でまとめて source する。読み込み順は after で制御。有効/無効は `al shell enable/disable`。

### bootstrap

新規 PC のセットアップで、al の他の機能では自動化できないワンオフのシェル実行を行うためのスクリプト。`~/.al/bootstrap/script.sh` に 1 本のスクリプトが保存される。`al bootstrap add` で作成、`edit` で編集、`remove` で削除、`show` で内容表示。sync / backup で他マシンと共有可能。

### extends

Profile が別の profile を継承する指定。例: `work` が `core.stable` を extends する場合、`al sync --profile work` では work と core.stable のパッケージが適用される。

### sync と backup

- **sync**: `~/.al` が無い場合は指定した `owner/repo` を clone してから、provider の確保・パッケージのインストール・link.d の symlink 適用を行う。既に存在する場合は適用のみ。最後に `~/.al/bootstrap/script.sh` が存在すれば実行する。`--all` で AutoSync 有効な profile をすべて対象、`--profile <name>` で指定 profile とその extends のみ。`--pkg-only` / `--link-only` でパッケージのみ / link のみ。**`--plan` でドライラン（変更を適用せず、何が変更されるかをプレビュー）**。manual provider のパッケージがある場合は、インストールを促す警告を表示する。
- **backup**: `~/.al` を commit して GitHub に push。`--init` でリポジトリが無ければ作成。`--repo owner/repo` で保存先を指定。デフォルトは `gh` で取得したユーザの `dotal`。

---

## インストール

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

## コマンドリファレンス

### ルート

| コマンド | 説明 |
|----------|------|
| `al init` | 初回セットアップ（profile core, provider brew, デフォルト設定） |
| `al activate zsh` / `bash` | シェル用コードを出力。`.zshrc` 等に `eval "$(al activate zsh)"` を追加する。trial のレビュー期限切れがあると stderr に案内を表示 |
| `al review` | レビュー期限切れの trial パッケージを対話で解決（remove / promote / postpone） |
| `al sync [owner/repo]` | `~/.al` が無ければ clone し、provider/パッケージ/link を適用。最後に bootstrap スクリプトがあれば実行。`--plan` で変更内容をプレビュー（実行なし）。`--all` で全プロファイル、`--profile <name>` で特定プロファイルを対象にできる |
| `al backup` | `~/.al` を commit して GitHub に push（`--init` でリポジトリ作成） |
| `al update` | al 本体を最新版に更新 |
| `al upgrade` | 全 provider と全パッケージをアップグレード（`-y` で確認スキップ） |
| `al version` | バージョン表示 |

### エイリアス（`al config alias list` で一覧）

`al add` → `al package add`、`al remove` → `al package remove`、`al list` → `al package list`、`al promote` → `al package move {args} --to package.promote_to`、`al import` → `al package import {args}`、`al pkg` → `al package`、`al prf` → `al profile`、`al prv` → `al provider`

### al config

| サブコマンド | 説明 |
|--------------|------|
| `al config set` | `--default-provider`, `--default-profile`, `--default-stage`, `--backup-repo` を設定 |
| `al config show` | 現在の設定を表示 |
| `al config alias list` | エイリアス一覧 |

### al profile

| サブコマンド | 説明 |
|--------------|------|
| `al profile add [名前]` | profile を追加（`-t` でテンプレートから作成） |
| `al profile list` | 一覧 |
| `al profile show [名前]` | 詳細 |
| `al profile remove <名前>` | 削除 |
| `al profile template` | 利用可能なテンプレート一覧 |

### al provider

| サブコマンド | 説明 |
|--------------|------|
| `al provider add <名前>` | 登録（brew / mas など）。依存関係（例: mas → brew）も自動で解決して先に処理 |
| `al provider list` | 一覧 |
| `al provider upgrade [provider-name]` | provider のアップグレード（未指定時は全 provider）。brew の場合は `brew update` により Homebrew と全 tap が更新される |
| `al provider prune` | brew のみ: brew-taps に登録されている tap のうち、packages に `owner/repo/toolname` 形式のパッケージが1つもない tap を untap し、brew-taps からも削除。homebrew/core と homebrew/cask は対象外。`--dry-run` で確認、`-y` で確認スキップ |

### al package

| サブコマンド | 説明 |
|--------------|------|
| `al package add [名前]` | 追加（`--provider`, `--profile`, `--stage`, `--id` など） |
| `al package list` | 一覧（`--profile` で絞り込み）。brew tap は表示しない（provider 側で管理） |
| `al package show <名前>` | 詳細 |
| `al package remove <名前>` | 削除 |
| `al package move <名前> --to <profile>` | 別 profile へ移動 |
| `al package import [Brewfile]` | Brewfile から取り込み（`--prf`, `--install`, `--dry-run` など） |
| `al package search <検索語>` | 検索 |
| `al package upgrade` | 登録パッケージのアップグレード |
| `al shell` | shell.d の show/add/edit/remove/enable/disable |
| `al package link` | link.d の add/remove/edit |

### al link

| サブコマンド | 説明 |
|--------------|------|
| `al link add <名前> <ユーザパス>` | link.d に登録し、ユーザパスを symlink に |
| `al link list` | 一覧 |
| `al link remove <名前>` | 削除（オプションで実体をユーザパスへ copy-back） |
| `al link edit <名前>` | ユーザパスなどの編集 |

### al shell

パッケージごとのシェルスニペット（`~/.al/shell.d/<パッケージ識別子>/`）を管理。

| サブコマンド | 説明 |
|--------------|------|
| `al shell show <パッケージ名>` | シェルスニペットの内容と設定を表示 |
| `al shell add <パッケージ名>` | 新規シェルスニペットを作成し、エディタで編集 |
| `al shell edit <パッケージ名>` | 既存シェルスニペットを編集（未作成時はエラー。`add` を利用） |
| `al shell remove <パッケージ名>` | シェルスニペットを削除 |
| `al shell enable <パッケージ名>` | `al activate` での読み込みを有効化 |
| `al shell disable <パッケージ名>` | `al activate` での読み込みを無効化（ファイルは保持） |

### al bootstrap

新規 PC のセットアップで、al の他の機能では自動化できないワンオフのシェル実行用。スクリプトは `~/.al/bootstrap/script.sh` に保存される。

| サブコマンド | 説明 |
|--------------|------|
| `al bootstrap add` | スクリプトを作成（未作成時のみ初期内容で作成） |
| `al bootstrap edit` | EDITOR でスクリプトを編集（未作成時は add と同様に作成してから開く） |
| `al bootstrap remove` | スクリプトを削除 |
| `al bootstrap show` | スクリプト内容を表示 |

---

## Brewfile からの移行

既存の Brewfile を al に取り込む。デフォルトは登録のみ（既存環境を al に寄せる想定）。

**事前準備**: 登録先 profile を用意（`al profile add <profile>`）。Brewfile の種類に応じて `al provider add brew` / `al provider add mas` を実行。

```bash
al import Brewfile --prf core --dry-run   # 確認
al import Brewfile --prf core             # 登録のみ
al import Brewfile --prf core --install  # 未インストール分もインストール
```

| オプション | 説明 |
|------------|------|
| `--profile`, `--prf` | 登録先 profile（必須） |
| `-s`, `--stage` | stage 名 |
| `--dry-run` | 書き込まずパース結果と登録予定の一覧のみ表示 |
| `--install` | 未インストールのパッケージを brew/mas でインストール |
| `--overwrite` | 同一 id・provider・profile の既存登録を上書き |
| `--verbose` | 対応外の行をスキップした理由を表示 |

**対応行**: `tap "user/repo"`、`brew "formula"`、`cask "name"`、`mas "App Name", id: 1234567890`。vscode / go / cargo / flatpak 等はスキップ（`--verbose` で確認可能）。

---

## 使用例

**試用から本番への昇格**

1. `al add <パッケージ名>` で trial に追加
2. 評価後、`al promote <パッケージ名>` で昇格、不要なら `al remove <パッケージ名>`

**仕事用・プライベートの分離**

1. `al profile add work -e core.stable -p core.stable`
2. `al profile add private`
3. `al add <パッケージ> --prf work` / `al add <パッケージ> --prf private`
4. `al sync --profile work` または `al sync --all` で環境を揃える

**変更内容を事前確認（プランモード）**

1. `al sync --plan` でシステムへの変更を確認（実行なし）
2. 問題がなければ `al sync` で実際に適用

**GitHub での共有・復元**

1. `al backup --init` でリポジトリ作成と push
2. 別マシンで `al sync owner/dotal` で clone して適用
3. 変更後は `al backup` で push
