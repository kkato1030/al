# 最新設計のレビュー（package link / shell / activate）

## 1. 良い点・一貫している点

- **config と shell の分離**が明確。link は一級オブジェクト（link.d）、shell は package の一部（shell.d）で、役割が分かりやすい。
- **al package link ＝ al link --package のエイリアス**に統一されており、link.d 周りの操作が一貫している。
- **remove のフラグ 2 本立て**（--keep-shell / --keep-link）で、何を残すか選べる。
- **al activate** は shell.d のみを扱い、.zshrc を触らない方針が維持されている。
- **enable/disable** で shell を無効化してもファイルは残す設計は、trial 運用と相性が良い。
- **after + トポロジカルソート**で読み込み順を制御する仕様が明確。
- **link.d でファイル・ディレクトリ両方**を扱う方針が書かれており、運用の幅が広い。

---

## 2. 曖昧さ・用語の揺れ

### 2.1 「config」と「shell」の表記

- **セクション 2 の見出し**が「パッケージに紐づく **config** の保存形式」のまま。中身は shell.d 用なので、**「パッケージに紐づく shell の保存形式」**の方が一貫する。
- **セクション 3（読み込み順のマニフェスト）**で「この **config** はどのパッケージの **後** に」とある。ここは shell 用マニフェストなので、**「この shell はどのパッケージの 後 に」**に揃えた方がよい。
- **セクション 1（ordering）**の 69 行目「このパッケージの **config** は、指定したツール」→ **「このパッケージの shell は」**に統一した方がよい。

### 2.2 shell.d マニフェストの enable/disable の形

- enable/disable を**どこにどう持つか**が書かれていない。
  - 案: マニフェストの各エントリに `"enabled": true/false` を足す。
  - または: 無効時は `shell.d/disabled/` に移す、など。
- 「マニフェストなどで『無効』とマーク」とあるので、**マニフェストに enabled フラグを持つ**と明記すると実装がぶれない。

---

## 3. 不足している仕様・実装時に決めたい点

### 3.1 既存の `al config` と `al link` の役割分担

- 既存: `al config set/show/alias`（アプリの default_provider / default_profile / default_stage / alias）。**アプリのデフォルト設定専用**。
- 新規: `al link add` / `al link list` / `al link remove` / `al link edit`（link.d で管理するファイル・ディレクトリ）。**link.d のトップレベルコマンド**。
- `al config` と `al link` を分離しているため、**サブコマンド一覧と役割**を設計書か README に書いておくとよい。
  - 例: 「al config ＝ アプリのデフォルト設定」「al link ＝ 管理するリンク（link.d）の追加・一覧・削除・編集」。

### 3.2 `al link --package` のパッケージ指定

- `--package <package_name>` の `<package_name>` は**表示名**で、packages.json 上の複数候補があり得る。
- 複数ヒット時の扱いが書かれていない。**al package link と同様に「複数ヒット時は対話で 1 つ選ぶ」**と明記するとよい。
- エイリアス `al package link add git --path X` では、`git` を渡した時点で対話があればそこで (id, provider, profile) が決まり、`al link add --path X --package <解決後の識別子>` のように渡すか、`--package` が名前を受け取り側で解決するかは実装方針として書いておくとよい。

### 3.3 link.d：パス未存在時の種別判定

- 「ファイル種別は指定されたパスで判断する」とある。**パスがまだ存在しない場合**（新規登録）は、セクション 8 で「実装時に決定」とあるが、**第一段階の案**（例: 末尾が `/` なら dir、それ以外は file）を 1 行書いておくと実装が早い。

### 3.4 トポロジカルソートで同列のときの順序

- 「ファイル名の辞書順」は設計レビュー（セクション 9）にのみ記載。**セクション 1 の「読み込み順」**に「同列はファイル名の辞書順」と 1 行入れておくと、本文だけ読んでも仕様が分かる。

---

## 4. マークアップ・ typo

- **149 行目**: `**al link` に` → 余分なバッククォート。**al link** に、とする。
- **161 行目**: `**--package <package_name>` を指定すると` → バッククォートの位置。**`--package <package_name>`** を指定すると、とする。
- **165 行目**: `**--purge**` → 他と揃えて **`--purge`** とする。
- **170 行目**: `**al package link`** → バッククォートが余分。**al package link** とする。
- **182 行目**: `**--keep-link**` → **`--keep-link`** とする。
- **221 行目**: `**$SHELL` に従う` → **`$SHELL`** に従う、とする。
- **224 行目**: `**--keep-shell` と` → 閉じ ** が抜けている。**`--keep-shell`** と **`--keep-link`** のスコープ、とする。
- **239 行目**: 「`al package link set git`」→ ここは shell の話なので **`al package shell set git`** が正しい。
- **255 行目**: `--keep-link**` → **`--keep-shell` / `--keep-link`** で残す、とする（閉じバッククォートと ** の対応）。

---

## 5. データフロー・削除順序

- セクション 6 のデータフローでは、remove 時に **link.d 紐づきの削除**が描かれていない。現状は shell.d の削除のみ。
- 実装タスク（セクション 7）では「link.d 紐づきの削除（--keep-link 時はスキップ）も追加」とあるので、**データフロー図に「link.d 紐づきの有無を確認し、--keep-link が無ければ symlink と実体を削除」を 1 ステップ足す**と、remove の挙動が図からも追いやすい。

---

## 6. まとめ

| 観点           | 評価 | コメント |
|----------------|------|----------|
| 概念の一貫性   | 良い | link（一級・link.d）と shell（package の一部・shell.d）の分離が明確。 |
| コマンド設計   | 良い | al link --package と al package link のエイリアス関係が分かりやすい。 |
| 削除・残す方針 | 良い | --keep-shell / --keep-link で対象が分かれている。 |
| 用語の揺れ     | 要修正 | セクション 2・3・ordering の「config」を「shell」に揃えるとよい。 |
| 不足仕様       | 軽微 | enable/disable の保存形式、--package の複数ヒット、未存在パスの種別を 1 行ずつ書くとよい。 |
| マークアップ   | 要修正 | 上記 typo を直すと読みやすい。 |

全体として、実装に進めるのに十分な水準にまとまっている。上記の用語統一・マークアップ修正・不足している 1 行仕様を足せば、実装時の迷いがさらに減る。
