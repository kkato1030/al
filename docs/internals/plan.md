# 実装計画（shell.d / link.d / activate）

GitHub issue #37 に基づく、アプリ固有設定・セットアップコマンドまわりの実装計画。

## 方針の整理

- **shell.d**: パッケージごとの shell 用スニペット。`al activate <shell>` で source される。`~/.al/shell.d/` に配置。
- **link.d**: 管理する設定ファイル・ディレクトリの実体。ユーザ向けパス（例: `~/.gitconfig`）はここへの symlink。`~/.al/link.d/` に配置。
- **al config**: 既存のまま。アプリのデフォルト設定（default_provider / default_profile / default_stage / alias）専用。

---

## 1. al activate

- **コマンド**: `al activate zsh` / `al activate bash` など。
- **挙動**: shell.d 内の有効な設定をトポロジカルソート（`--after` 依存）で並べ、source するための shell コードを標準出力。
- ユーザは `.zshrc` 等に `eval "$(al activate zsh)"` を 1 行書く。al は .zshrc を編集しない。

## 2. al package shell

- **al package shell show <pkg>**: 内容表示。
- **al package shell set <pkg> [--after <dep_pkg>]**: 内容設定と読み込み順（後ろに読みたいパッケージ）の指定。
- **al package shell edit <pkg>**: `EDITOR`（未設定時は vim）で編集。`.zsh` / `.bash` は `$SHELL` から推定。
- **al package shell enable <pkg>** / **al package shell disable <pkg>**: `al activate` で source するかどうか。無効でもファイルは残す。
- shell.d 用マニフェスト（例: `.manifest.json`）に `after` と `enabled` を持たせる。同列はファイル名辞書順。

## 3. al link（link.d のトップレベルコマンド）

- **al link add [--path] <path> [--package <pkg>]**: link.d に追加。単体または `--package` でパッケージに紐づけ。種別はパスで判定（未存在時は末尾 `/` なら dir、それ以外は file）。
- **al link list [--package <pkg>]**: 管理対象一覧。
- **al link remove --path <path> [--package <pkg>] [--purge]**: 管理から外す。デフォルトは copy-back（元ファイルがあれば復元）。`--purge` で実体も削除。
- **al link edit --path <path> [--package <pkg>]**: link.d 内の実体を `EDITOR` で編集。
- **al package link**: **al link --package <package_name>** のエイリアス。例: `al package link add git --path ~/.gitconfig` ＝ `al link add --path ~/.gitconfig --package git`。`--package` の名前で複数ヒットする場合は対話で 1 つ選択。

## 4. al package remove のフラグ

- **--keep-shell**: パッケージ削除時に shell.d の内容は残す。
- **--keep-link**: パッケージ削除時に link.d 紐づきは残す（紐づきを外して単体の link として残す）。
- デフォルト: 両方とも削除（shell.d と link.d の紐づきを消す）。remove のデータフローに「link.d 紐づきの有無を確認し、--keep-link が無ければ symlink と実体を削除」を含める。

## 5. link.d の構造

- 各管理対象は `~/.al/link.d/<id>/` のようなサブディレクトリに格納。
- メタデータはマニフェスト（例: `.manifest.json`）で管理（パッケージ紐づき、file/dir 種別など）。

## 6. 実装時の注意

- shell.d 用語: 設計書内の「config」を「shell」に揃える（パッケージに紐づく shell の保存形式・読み込み順など）。
- link.d 用語: 一級オブジェクトは「link（link.d）」で統一。
- サブコマンド一覧と役割を README または設計書に明記: `al config` ＝ アプリのデフォルト設定、`al link` ＝ link.d の追加・一覧・削除・編集。
