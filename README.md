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

`ryoiki switch` で直接 `cd` できるように、以下のコマンドを実行:

```bash
ryoiki init zsh    # ~/.zshrc に自動追記
ryoiki init bash   # ~/.bashrc に自動追記
ryoiki init fish   # ~/.config/fish/config.fish に自動追記
```

手動で設定したい場合は `--print` フラグでスクリプトを出力:

```bash
eval "$(ryoiki init zsh --print)"
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
| `ryoiki add [path]` | workspace作成（`jj workspace add`ラッパー） | `-n, --name` 名前指定（パス省略時は `~/.ryoiki/<repo>-<repohash>/<name>`） / `-r, --revision` 親リビジョン / `-p, --purpose` 用途 |
| `ryoiki list` | 全workspaceの一覧表示 | `--json` JSON出力 |
| `ryoiki status` | 全workspaceの詳細ステータス表示 | `--json` JSON出力 |
| `ryoiki init <shell>` | シェル統合をセットアップ（デフォルト: 設定ファイルに自動追記） | `--print` stdout にスクリプト出力 |
| `ryoiki switch <name>` | workspaceに切り替え（init設定済みなら直接cd） | |
| `ryoiki show <name>` | 単一workspaceの詳細情報 | |
| `ryoiki describe <name>` | workspaceのメタデータを設定・更新 | `-p, --purpose` 用途（必須） |
| `ryoiki forget <name>` | workspace削除（ディレクトリも削除） | `-f, --force` 確認スキップ / `--keep-dir` ディレクトリを保持 |
| `ryoiki tenkai` | インタラクティブTUIを起動 | |
| `ryoiki version` | バージョン表示 | |

## TUI Mode

`ryoiki tenkai` でインタラクティブTUIを起動できます。

```bash
ryoiki tenkai
```

### レイアウト

- 通常サイズの端末では 3 ペイン表示:
  - 左上: workspace 一覧
  - 左下: 選択 workspace の detail
  - 右上: 選択 workspace の `jj log`（graph）
  - 右下: 選択 workspace の stdout/stderr tail
- 端末幅が狭い場合は自動で 1 ペイン表示にフォールバックします。

### キーバインド

| Key | Action |
|-----|--------|
| `j` / `k` | カーソル移動 |
| `Enter` | 詳細表示（1ペイン時） |
| `Esc` | 一覧に戻る |
| `d` | purpose（用途）を設定 |
| `f` | workspaceを削除 |
| `a` | workspaceを追加（path first） |
| `A` | workspaceを追加（name first） |
| `s` | 選択workspaceへ切り替え（`cd`） |
| `/` | fuzzy検索 |
| `L` | stdoutペイン表示の切り替え |
| `r` | リフレッシュ |
| `?` | ヘルプ表示 |
| `q` | 終了 |

`s` で切り替えるには最新のシェル統合が必要です。動かない場合は再設定してください。

```bash
ryoiki init zsh
source ~/.zshrc
```

## Requirements

- [jj (Jujutsu)](https://github.com/jj-vcs/jj) がインストール済みであること
