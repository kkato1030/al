# Link と Shell

## link.d

dotfiles の実体を `~/.al/link.d/<名前>/content` に置き、参照するパス（例: `~/.config/foo`）をその content への symlink にする。sync/backup で他マシンと揃えられる。

- **追加**: `al link add <名前> <ユーザパス>`
- **一覧**: `al link list`
- **削除**: `al link remove <名前>`（オプションで実体をユーザパスへ copy-back）
- **編集**: `al link edit <名前>`

パッケージに紐づける場合（link 名 = パッケージ名の想定）: `al package link add/remove/edit`。

## shell.d

パッケージごとのシェルスニペットは `~/.al/shell.d/<パッケージ識別子>/` にあり、`al activate` の出力で source される。読み込み順は `after` で制御。

- **表示**: `al shell show <パッケージ名>`
- **追加**: `al shell add <パッケージ名>` — 作成して編集
- **編集**: `al shell edit <パッケージ名>`（未作成時はエラー。add を使う）
- **削除**: `al shell remove <パッケージ名>`
- **有効/無効**: `al shell enable <パッケージ名>`、`al shell disable <パッケージ名>`（無効でもファイルは保持）

`.zshrc`/`.bashrc` に `eval "$(al activate zsh)"`（または bash）を追加するとシェルで有効になる。
