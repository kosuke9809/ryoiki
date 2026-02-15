# ryoiki

jj (Jujutsu) workspace管理に特化したCLIツール。

## Features

- workspace名だけで素早く切り替え
- 各workspaceに用途・説明をメタデータとして紐づけ
- 全workspaceの状態（commit, change, 説明, 用途）を一望
- workspace削除時にディレクトリも一緒にクリーンアップ
- マルチAIエージェント開発での並行作業workspace管理に最適

## Install

```bash
go install github.com/kosuke9809/ryoiki@latest
```

### シェル統合（推奨）

`ryoiki switch` で直接 `cd` できるように、シェルの設定ファイルに以下を追加:

**zsh** (`~/.zshrc`):
```bash
eval "$(ryoiki init zsh)"
```

**bash** (`~/.bashrc`):
```bash
eval "$(ryoiki init bash)"
```

**fish** (`~/.config/fish/config.fish`):
```fish
ryoiki init fish | source
```

## Usage

```bash
# workspaceを作成
ryoiki add ./ws-auth --purpose "OAuth2実装"

# 一覧表示
ryoiki list

# 全体の詳細ステータス
ryoiki status

# workspaceに切り替え（init設定済みなら直接cd）
ryoiki switch ws-auth

# メタデータを更新
ryoiki describe ws-auth --purpose "OAuth2実装 + セッション管理"

# 単一workspaceの詳細
ryoiki show ws-auth

# workspaceを削除（ディレクトリも）
ryoiki forget ws-auth --remove-dir
```

## Commands

| Command | Description | Options |
|---------|-------------|---------|
| `ryoiki add <path>` | workspace作成（`jj workspace add`ラッパー） | `-n, --name` 名前指定 / `-r, --revision` 親リビジョン / `-p, --purpose` 用途 |
| `ryoiki list` | 全workspaceの一覧表示 | `--json` JSON出力 |
| `ryoiki status` | 全workspaceの詳細ステータス表示 | `--json` JSON出力 |
| `ryoiki init <shell>` | シェル統合スクリプトを出力（`eval "$(ryoiki init zsh)"`） | |
| `ryoiki switch <name>` | workspaceに切り替え（init設定済みなら直接cd） | |
| `ryoiki show <name>` | 単一workspaceの詳細情報 | |
| `ryoiki describe <name>` | workspaceのメタデータを設定・更新 | `-p, --purpose` 用途（必須） |
| `ryoiki forget <name>` | workspace削除 | `-f, --force` 確認スキップ / `--remove-dir` ディレクトリも削除 |
| `ryoiki tenkai` | インタラクティブTUIを起動 | |
| `ryoiki version` | バージョン表示 | |

## TUI Mode

`ryoiki tenkai` でインタラクティブTUIを起動できます。

```bash
ryoiki tenkai
```

### キーバインド

| Key | Action |
|-----|--------|
| `j` / `k` | カーソル移動 |
| `Enter` | 詳細表示 |
| `Esc` | 一覧に戻る |
| `d` | purpose（用途）を設定 |
| `f` | workspaceを削除 |
| `a` | workspaceを追加 |
| `r` | リフレッシュ |
| `q` | 終了 |

## Requirements

- [jj (Jujutsu)](https://github.com/jj-vcs/jj) がインストール済みであること
