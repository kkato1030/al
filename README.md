# al - Mac Management Tools

`al` は Mac のパッケージや設定を管理するためのツールです。新しいパッケージや設定を試用してから本格採用する「trial/core モデル」により、安定した環境を維持しながら柔軟に実験できます。

## 概要

`al` を使うことで、以下のような Mac 環境の管理が可能になります：

- Homebrew パッケージの管理
- dotfiles（設定ファイル）の管理
- アプリケーション設定の管理
- シェル設定の管理
- その他 Mac 環境に関連するあらゆるパッケージ・設定の管理

## 主要な概念

### Trial/Core モデル

`al` は **trial/core モデル**を採用しています。このモデルでは、新しいパッケージや設定はまず **trial** に追加され、実際に使用してみてから判断します。十分に検証され、常用する価値があると判断した場合のみ、**core** に昇格させます。

**重要な原則**: trial と core は排他的です。同じパッケージ・設定が trial と core の両方に存在することはできません。

このモデルのメリット：

- **慎重な採用**: 新しいパッケージや設定を即座に本番環境に追加せず、実際の使用経験を積んでから判断できる
- **環境の安定性**: core に含まれるパッケージ・設定は、十分に検証された信頼できるもののみが含まれる
- **柔軟な実験**: trial で自由に試行錯誤を行い、不要なものは core に昇格させずに削除できる

### Profile

`al` では、用途に応じて複数の **profile** を作成し、環境を分離できます。各 profile は独立した trial/core のセットを持ちます。

- **デフォルト profile**: `trial` と `core` は常に存在するデフォルトの profile です。これらは変更・削除できません。
- **カスタム profile**: `work` や `private` などの profile を任意に作成・切り替え・削除できます。

例えば：
- **work** profile: 仕事用の環境に必要なパッケージ・設定
- **private** profile: プライベート用の環境に必要なパッケージ・設定

## 基本的な使い方

### パッケージ・設定の追加

新しいパッケージや設定を試す場合は、add で追加します：

```bash
al add <package>
```

### Trial から Core への昇格

trial で十分に検証し、常用する価値があると判断した場合、core に昇格させます：

```bash
al promote <package>
```

### Profile の管理

カスタム profile を作成・確認・削除できます：

```bash
al profile create <profile-name>
al profile list
al profile delete <profile-name>
```

Profile を指定する場合は、各種コマンドで `-p (--profile)` を指定します：

```bash
# work にインストールする場合
al add <package> -p work
```

## 使用例

### 例1: 新しいパッケージを試す

1. 新しいパッケージを trial に追加
2. 一定期間使用して評価
3. 気に入ったら `al promote` で core に昇格
4. 気に入らなければ trial から削除

### 例2: 仕事用とプライベート用で環境を分離

1. `work` profile を作成して仕事用のパッケージ・設定を管理
2. `private` profile を作成してプライベート用のパッケージ・設定を管理
3. 必要に応じて profile を切り替えて使用
