# コマンド入出力・対話インターフェース設計ガイドライン

本ドキュメントは、`al` CLI のコマンド入出力および対話インターフェースを統一し、利用者にとって理解しやすく、自動化しやすい形にするための設計ルールを定める。新規コマンドの実装および既存コマンドのリファイン時は、このガイドラインに従うこと。

**関連ドキュメント**

- 実装詳細（`internal/output` パッケージ・AL_DEBUG・スピナー・ツール出力の抑制）は [output-organization.md](output-organization.md) を参照。
- 一覧表示の空メッセージ・見やすさ（折り返し・グリッド等）は [list-improvement.md](list-improvement.md) も参照可。

---

## 1. 入出力

### 1.1 ストリームの使い分け

| 出力先 | 用途 | 例 |
|--------|------|-----|
| **stdout** | コマンドの主な結果・一覧・JSON・eval 用スクリプト | `al package list` の一覧、`al sync --json` の JSON、`al activate zsh` のシェルコード |
| **stderr** | 進捗・警告・デバッグ・対話プロンプト・エラー要約 | 進捗メッセージ、`output.Warning`、確認プロンプト "Continue? [y/N]: "、main の "Error: ..." |

**理由**: パイプラインやリダイレクトで `al list > out.txt` のように stdout だけを取りたい場合に、結果だけがファイルに入り、プロンプトや警告は端末に残るようにするため。

**良い例**

```go
// 一覧は stdout
fmt.Fprintln(w, "Configured profiles:")
for _, p := range profiles {
    fmt.Fprintf(w, "  - %s\n", p.Name)
}

// 警告は stderr
output.Warning("Failed to update provider config: %v", err)

// 確認プロンプトは stderr
fmt.Fprint(os.Stderr, "Continue? [y/N]: ")
```

**避ける例**

```go
// プロンプトを stdout に出さない（パイプ時に結果に混ざる）
fmt.Print("Continue? [y/N]: ")  // 悪い: stdout
```

### 1.2 出力形式（人間向け / 機械向け）

- **デフォルト**: 人間が読むためのテキスト（一覧・サマリ・メッセージ）。
- **機械向け**: グローバルフラグ `--json` を付けたときは JSON のみを stdout に出力する。既存では `al sync --json`・`al doctor --json`・`al diff --json` が対応している。
- **方針**: list/show 系で `--json` を追加する場合は、既存の sync/doctor/diff と同様に「人間向けと JSON の二系統」とし、`--json` 時は JSON のみを stdout に出す（ヘッダーや説明文は出さない）。

**良い例（JSON 時）**

```go
if IsJSONOutput() {
    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(result); err != nil {
        fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
        return err
    }
    return nil
}
// 人間向けテキスト
fmt.Println("Summary: ...")
```

### 1.3 メッセージ種別と表現

| 種別 | 用途 | 出力先 | 実装 |
|------|------|--------|------|
| **成功** | 操作が正常に完了したとき | stdout | `output.Success(format, args...)` |
| **情報** | 進捗・補足説明 | stdout | `output.Info(format, args...)` |
| **警告** | 注意が必要だが処理は続行するとき | stderr | `output.Warning(format, args...)` |
| **エラー** | 致命的エラー（処理を中断するときは `return fmt.Errorf` を優先し、main が "Error: %v" で stderr に出す） | stderr | `output.Error` または `return fmt.Errorf` |

`internal/output` の `Info`・`Success`・`Warning`・`Error` を標準として使う。Success には "✓ "、Warning には "⚠️ "、Error には "✗ " が付く。eval 用に stdout をそのまま使うコマンド（`al activate`）では、スクリプト以外のメッセージは stderr に出す。

**良い例**

```go
output.Success("Installed %s", pkgName)
output.Info("Running bootstrap script...")
output.Warning("Manual packages detected: ensure they are installed.")
return fmt.Errorf("profile '%s' not found", name)  // main が stderr に "Error: ..." と出す
```

**避ける例**

```go
fmt.Printf("Package '%s' has been successfully added to profile '%s' with provider '%s'\n", ...)  // 長文
fmt.Println("Warning: ...")  // output.Warning を使う
```

### 1.4 空状態・not found

- **一覧が空のとき**: メッセージを統一する。
  - 設定・登録済みリソースの一覧で「何もない」: **"No &lt;resource&gt; configured"**（例: "No profiles configured", "No packages configured"）。
  - 検索・フィルタ結果が 0 件: **"No &lt;resource&gt; found"**（例: "No packages found matching the specified filters", "No logs found"）。
- **エラーとしての not found**: 操作対象が存在しないときは `return fmt.Errorf("... not found")` に統一し、main の "Error: %v" に任せる。文言は「何が」見つからないかを含める（例: `"profile '%s' not found"`）。

**良い例**

```go
if len(profiles) == 0 {
    fmt.Println("No profiles configured")
    return nil
}
// ...
if profile == nil {
    return fmt.Errorf("profile '%s' not found", name)
}
```

**避ける例**

```go
fmt.Println("No templates available")   // 他と揃えるなら "No templates configured" など方針を統一
return fmt.Errorf("not found")           // 何が not found か明示する
```

### 1.5 成功メッセージ

簡潔な定型を推奨する。長い説明は避け、必要な情報（リソース名・ID・プロファイル等）だけを短く入れる。

**推奨パターン**

- 追加: "Added &lt;resource&gt; &lt;id&gt;" または "Added &lt;resource&gt; &lt;id&gt; to &lt;profile&gt;"
- インストール: "Installed &lt;name&gt;"
- 削除: "Removed &lt;name&gt;"
- 更新・昇格: "Updated &lt;name&gt;" / "Promoted &lt;name&gt; to &lt;profile&gt;"

**良い例**

```go
output.Success("Installed %s", pkgName)
output.Success("Added package %s to profile %s", pkgName, profileName)
```

**避ける例**

```go
fmt.Printf("Package '%s' (ID: %s) has been successfully added to profile '%s' with provider '%s'\n", ...)
```

### 1.6 入力設計

- **引数**: 必須の場合は `cobra.ExactArgs(n)` または `cobra.MinimumNArgs(1)` などで明示。省略可能で対話に fallback する場合は `cobra.MaximumNArgs(1)` 等とし、Long で "If ... is not provided, interactive mode will be used." と書く。
- **フラグ**:
  - グローバル: `--json`（root の PersistentFlags）は全コマンドで共通。
  - コマンド固有: プロファイル指定は `--profile`、短縮は `--prf`。プロバイダ指定は `--provider`、短縮は `--prv`。他サブコマンドで同じ概念を指すフラグを追加する場合は、この命名に揃えることを推奨する。

---

## 2. 対話インターフェース

### 2.1 確認プロンプト（y/N）

- **表記**: **`[y/N]`** に統一する。大文字の N は「デフォルトが No」であることを示す。
- **文言**:
  - 「続行しますか？」系: 短い定型にする。例: **"Continue? [y/N]: "**
  - 対象を明示する場合: 一文で何についての確認か分かるようにする。例: **"Import %s %s to profile '%s'? [y/N]: "**
  - 二重確認が必要な場合も冗長にせず、簡潔に。例: **"Postpone anyway? [y/N]: "**（"Really? Do you really need it?" は避ける）。

**良い例**

```go
fmt.Fprint(os.Stderr, "Continue? [y/N]: ")
fmt.Fprintf(os.Stderr, "Import %s %s to profile '%s'? [y/N]: ", provider, name, profile)
```

**避ける例**

```go
fmt.Print("Do you want to continue? [y/N]: ")   // "Continue? [y/N]: " に統一
fmt.Print("  Really? Do you really need it? [y/N]: ")  // "Postpone anyway? [y/N]: " など簡潔に
```

### 2.2 確認プロンプトの出力先

確認プロンプトの文言は **stderr** に出す。ユーザーが stdin をパイプしている場合は対話しない前提とする。

**実装例**

```go
fmt.Fprint(os.Stderr, "Continue? [y/N]: ")
```

### 2.3 入力の読み方（y/N）

- y/N の読み取りは **bufio.Scanner または共通ヘルパー**に統一する。`ReadString('\n')` と `Scanner.Scan()` の混在を避ける。
- 受け付ける値: **"y" または "yes"**（大文字可）で true、それ以外または空 Enter は false。

**共通ヘルパー例（方針）**

```go
// internal/prompt または internal/ui に置く想定
func Confirm(w io.Writer, prompt string) (bool, error) {
    fmt.Fprint(w, prompt)
    scanner := bufio.NewScanner(os.Stdin)
    if !scanner.Scan() {
        return false, scanner.Err()
    }
    s := strings.TrimSpace(strings.ToLower(scanner.Text()))
    return s == "y" || s == "yes", nil
}
```

**使用例**

```go
ok, err := prompt.Confirm(os.Stderr, "Continue? [y/N]: ")
if err != nil {
    return err
}
if !ok {
    fmt.Fprintln(os.Stderr, "Cancelled.")
    return nil
}
```

### 2.4 --yes / -y

確認を求めるコマンドは、**--yes / -y** で確認をスキップできるようにする。既存では `al upgrade -y`・`al provider upgrade -y`・`al provider prune -y` 等が対応している。`al import`・`al update`・`al review` など、確認プロンプトがあるコマンドでも、必要に応じて `-y` を追加することを検討する。

### 2.5 対話モードのトリガー

- 引数省略で対話に入る場合: コマンドの Long で **"If ... is not provided, interactive mode will be used."** と明記する。
- フラグで対話を明示する場合: **`--interactive`** / **`-i`** など一貫した名前を使う（例: `al package import -i`）。

### 2.6 TUI とライン入力の使い分け

- **複数選択・一覧から選ぶ**: **Bubble Tea（internal/ui）** を優先する。例: プロファイル選択、テンプレート選択、init --guided。
- **単純な 1 行入力**: 名前・説明などは `fmt.Print("Label: ")` + `bufio.Scanner` でもよい。その場合、オプション項目の説明は **(optional, press Enter to skip)** に統一する。

**良い例**

```go
fmt.Print("Description (optional, press Enter to skip): ")
```

### 2.7 選択肢の表示

キー＋説明は、**短いキーを [ ] で囲み、続けて説明**する形式に統一する。

**良い例**

```go
fmt.Print("[r]emove  [p]romote  [s]postpone: ")
```

**避ける例**

```go
fmt.Print("remove (r) / promote (p) / postpone (s): ")  // キーを [ ] で囲む形式に揃える
```

---

## 3. チェックリスト（実装・リファイン時）

- [ ] 結果・一覧・JSON は stdout、プロンプト・警告・進捗は stderr に振り分けているか
- [ ] 成功・情報・警告は `internal/output` の Info/Success/Warning を使っているか（activate の stdout は除く）
- [ ] 空状態メッセージは "No ... configured" / "No ... found" のルールに沿っているか
- [ ] 成功メッセージは簡潔な定型か
- [ ] 確認プロンプトは `[y/N]` で stderr に出力しているか
- [ ] y/N の読み取りは Scanner または共通ヘルパーで統一しているか
- [ ] 確認をスキップする場合は `--yes`/`-y` を用意しているか（該当コマンド）
- [ ] 対話モードのトリガー（引数省略 or `-i`）を Long に明記しているか
- [ ] オプション入力の説明は "(optional, press Enter to skip)" に統一しているか
- [ ] 選択肢は "[key]description" 形式で表示しているか
