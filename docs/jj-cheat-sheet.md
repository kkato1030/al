# jj コマンド チートシート

[JJ Reference](https://justinpombrio.net/src/jj-cheat-sheet.pdf)（justinpombrio.net, Feb 2025）に基づくリファレンス。Jujutsu（jj）は DAG ベースの VCS で、Git と連携して利用できる。

---

## モデルと用語

- **change（チェンジ）**: DAG のノード。各 change は以下を持つ。
  - リポジトリディレクトリ内のファイルシステムの状態
  - ファイルコンフリクト（git と異なり、jj の利用はブロックされない）
  - 1 つ以上の親 change（ルートは親なし・空ディレクトリ）
  - 説明文（コミットメッセージに相当。空文字可）
- **@（ワーキングチェンジ）**: 現在の「ワーキングコピーリビジョン」。git の HEAD に相当。
- **bookmark**: change に付く一意のラベル。Git 連携時はブランチ名として扱われる。
- **BOOKMARK@REMOTE**: リモートのブックマークの最終既知位置（例: `feat-ui@origin`）。

---

## 基本ルール

- @ をある change に向けると、作業ディレクトリがその change の内容に更新される。
- @ が指していた change を削除すると、@ はその親から分かれた新しい空 change に移る。
- ファイル変更も説明もなく、@ でもブックマークでも参照されていない change は静かに消える。
- change は diff を表す。change を移すとその diff を新しい親に適用しようとし、マージコンフリクトの原因になりうる。
- 多くのコマンドはデフォルトで @ に作用する。`-r` / `--revision` で対象 change を指定できる。

---

## ファイルコンフリクト

- ワーキング change（@）にコンフリクトがある場合、マーカー（`<<<<<<<`, `=======`, `>>>>>>>` 等）を削除してファイルを編集すれば解消。
- バイナリは使いたい版でファイルを置き換える。`jj restore` が便利なことがある。

---

## グローバル設定

```bash
jj config set --user user.name MY_NAME
jj config set --user user.email MY_EMAIL
jj config set --user ui.editor MY_EDITOR
jj config edit --user   # 設定ファイルを直接編集
```

`--user` の代わりに `--repo` でリポジトリ固有設定（優先される）。

---

## リポジトリ

| コマンド | 説明 |
|----------|------|
| `jj git init` | Git 連携リポジトリを新規作成 |
| `jj git clone URL [DESTINATION]` | URL からクローン |
| `jj git init --colocate` | 既存の Git リポジトリを jj リポジトリとしても使う |

---

## ローカル編集の主なコマンド

| コマンド | 説明 |
|----------|------|
| `jj` | 重要な change を表示 |
| `jj undo` | 直前のコマンドを取り消す |
| `jj new` | 現在の親から新しい change を生成（@ がそこに移る） |
| `jj new p q` | 親に `p` と `q` を持つ新しい change を生成（マージ） |
| `jj status` | 現在の change・親・ファイル変更を表示 |
| `jj describe -m "edit foo"` | 現在の change の説明を設定 |
| `jj show` | 指定 change の説明などを表示 |
| `jj edit q` | @ を change `q` に移す（作業ディレクトリが `q` の内容になる） |
| `jj restore --from q (paths..)` | 指定パスを change `q` の内容で復元 |
| `jj backout` | 現在の change の内容を打ち消す change を追加 |
| `jj abandon q` | change `q` を破棄（子があると別の親に繋ぎ変え等） |
| `jj diff (paths..)` | 変更間の diff を表示 |
| `jj squash` | 変更を親に取り込む（squash） |

### ブックマーク

| コマンド | 説明 |
|----------|------|
| `jj bookmark list` | 全ブックマークを表示 |
| `jj bookmark create feat/ui` | ブックマークを作成（現在の @ に） |
| `jj bookmark delete feat/ui` | ブックマークを削除 |
| `jj bookmark move feat/ui` | ブックマークを現在の @ に移動 |
| `jj bookmark rename feat/ui feat/ux` | ブックマーク名を変更 |
| `jj bookmark set BOOKMARK` | 作成または移動のどちらか有効な方を実行 |

### エイリアス的なコマンド

- **`jj commit`**: `jj describe` + `jj new` の短縮。
- **`jj bookmark set BOOKMARK`**: ブックマークの作成または移動。

---

## Git 連携

### jj git push

- ローカルの change をリモートにコピーする。
- ローカルで push 後に変更した change は、リモートでは別の新しい change になる（git の force push で置き換わるイメージ）。
- プライマリブランチに push された change は immutable になり、編集するには `--ignore-immutable` が必要。
- ブックマークもリモートにコピーされる。同じ名前のブックマークがローカルとリモートの両方にある場合、リモートの「最後に見た位置」と現在のリモートの位置が一致しないと push は失敗し、先に `jj git fetch` するよう促される。

### jj git fetch

- リモートの change をローカルに取り込む。
- リモートで変更された change はローカルでは新しい change として現れる。
- ローカルのブックマークは、リモートでそのブックマークが指している change に合わせて進められる。
- リモートのブックマークが指す change が、ローカルでそのブックマークが指す change の子孫でない場合、同じ名前のブックマークがもう一つ作られて「ブックマークコンフリクト」になる（git の pull でマージコンフリクトに似た状況）。

---

## ブックマークコンフリクトの解消

1. **マージする**: `jj new CHANGE-ID-1 CHANGE-ID-2` でマージ change を作り、ファイルコンフリクトを解消したあと `jj bookmark move BOOKMARK-NAME` でブックマークを更新。change ID は `jj bookmark list BOOKMARK-NAME` で確認。
2. **片方だけ使う**: 残したい change に `jj bookmark move BOOKMARK-NAME -r CHANGE-ID` でブックマークを移す。
3. **リベースする**: `jj rebase -b CHANGE-ID-2 -d CHANGE-ID-1` のあと `jj bookmark move BOOKMARK-NAME -r CHANGE-ID-2`。2 番目の change と、そこから分岐した以降の change が 1 番目の後に並ぶ。

---

## 記法の対応（チートシート図での意味）

- `r`, `q`, `p` — change
- `@` — ワーキング change（ワーキングコピーリビジョン）
- `"edit foo"` — change の説明
- `feat/ui` — ブックマーク
- `Files1`, `Files2` — ファイルシステムの状態
- `Edit1`, `Edit2` — 2 つの change 間の diff

---

**出典**: [JJ Reference (PDF)](https://justinpombrio.net/src/jj-cheat-sheet.pdf) — justinpombrio.net & lark.gay, Feb 2025
