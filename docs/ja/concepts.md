# 概念

## Profile と Stage

- **Profile**: 環境の単位（例: core, work, private）。各 profile は独立したパッケージ一覧を持つ。`al add --prf <profile> ...` で登録先を指定する。
- **Stage**: 試用（trial）か本番（stable）かを表す。Profile 名は `profile_name.stage_name`（例: `core.trial`, `core.stable`）。`al init` で作成されるのは `core`（stage: trial）。カスタム profile は `al profile add` で追加する。

## promote（昇格）

Profile に `promote_to` を設定すると、パッケージを trial から stable などへ移動できる。`al package move <name> --to <profile>` またはエイリアス `al promote <name>`（パッケージが属する profile の promote_to へ移動）。同一パッケージ（同一 ID・provider）が trial と stable の両方に同時に存在することはない。

## trial のレビュー期限（review）

Profile に `review_days` を設定すると、その profile のパッケージがレビュー対象になる。パッケージの `review_by`（レビュー期限日時）が無いか期限を過ぎていれば「期限切れ」として扱う。`al activate` 実行時に期限切れがあると stderr に一覧と「`al review` で解決」の案内を表示する。`al review` では各パッケージについて **remove**（削除）、**promote**（stable へ昇格）、**postpone**（延期）のいずれかを選べる。

## Provider

パッケージのインストール元。**brew**（Homebrew formula/cask）、**mas**（Mac App Store）、**manual**（登録のみ、インストールは行わない）。`al init` で brew が追加される。mas は `al provider add mas`。デフォルト provider は `al config set --default-provider` で変更可能。

provider 間の依存関係は `providers.json` の `depends_on` で管理される。デフォルトでは **mas は brew に依存**し、`al provider add mas` や `al sync` では依存順（brew → mas）で処理される。

**brew tap**: `brew tap` で追加する tap はパッケージとは別に provider brew 側で管理する（`~/.al/brew-taps.json`）。`al list` には tap は出さない。`al add homebrew/cask-fonts` のように tap を指定すると brew-taps に登録される。tap の更新は `al provider upgrade` または `al upgrade` で行う。`al provider prune` では、packages に `owner/repo/toolname` 形式のパッケージが1つもない tap を untap し、brew-taps から削除する（homebrew/core と homebrew/cask は対象外）。

## link.d

設定ファイル・ディレクトリの実体を `~/.al/link.d/<name>/content` に置き、ユーザが参照するパス（例: `~/.config/foo`）をその content への symlink にする。sync / backup で他マシンや GitHub と揃えられる。追加は `al link add <name> <user_path>`。パッケージに紐づける場合は `al package link add/remove/edit`（link 名 = パッケージ名の 1:1 想定）。

## shell.d

パッケージごとのシェル設定を `~/.al/shell.d/<パッケージ識別子>/` に置き、`al activate` の出力でまとめて source する。読み込み順は after で制御。有効/無効は `al shell enable/disable`。

## bootstrap

新規 PC のセットアップで、al の他の機能では自動化できないワンオフのシェル実行を行うためのスクリプト。`~/.al/bootstrap/script.sh` に 1 本のスクリプトが保存される。`al bootstrap add` で作成、`edit` で編集、`remove` で削除、`show` で内容表示。sync / backup で他マシンと共有可能。

## extends

Profile が別の profile を継承する指定。例: `work` が `core.stable` を extends する場合、`al sync --profile work` では work と core.stable のパッケージが適用される。

## sync と backup

- **sync**: `~/.al` が無い場合は指定した `owner/repo` を clone してから、provider の確保・パッケージのインストール・link.d の symlink 適用を行う。既に存在する場合は適用のみ。最後に `~/.al/bootstrap/script.sh` が存在すれば実行する。`--all` で AutoSync 有効な profile をすべて対象、`--profile <name>` で指定 profile とその extends のみ。`--pkg-only` / `--link-only` でパッケージのみ / link のみ。**`--plan`** でドライラン（変更を適用せずプレビュー。アップグレード可能パッケージもチェック。`AL_DEBUG=1` でデバッグログ有効化）。manual provider のパッケージがある場合は、インストールを促す警告を表示する。
- **backup**: `~/.al` を commit して GitHub に push。`--init` でリポジトリが無ければ作成。`--repo owner/repo` で保存先を指定。デフォルトは `gh` で取得したユーザの `dotal`。**`--dry-run`** で実際に commit・push せずプレビュー。
