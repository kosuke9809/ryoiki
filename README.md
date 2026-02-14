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

## Usage

```bash
# workspaceを作成
ry add ./ws-auth --purpose "OAuth2実装"

# 一覧表示
ry list

# 全体の詳細ステータス
ry status

# workspaceに切り替え
cd $(ry switch ws-auth)

# メタデータを更新
ry describe ws-auth --purpose "OAuth2実装 + セッション管理"

# 単一workspaceの詳細
ry show ws-auth

# workspaceを削除（ディレクトリも）
ry forget ws-auth --remove-dir
```

## Commands

| Command | Description |
|---------|-------------|
| `ry add <path>` | workspace作成（`jj workspace add`ラッパー） |
| `ry list` | 全workspaceの一覧表示 |
| `ry status` | 全workspaceの詳細ステータス表示 |
| `ry switch <name>` | workspaceのパスを出力（`cd $(ry switch name)`で切り替え） |
| `ry show <name>` | 単一workspaceの詳細情報 |
| `ry describe <name>` | workspaceのメタデータを設定・更新 |
| `ry forget <name>` | workspace削除 |
| `ry version` | バージョン表示 |

## Requirements

- [jj (Jujutsu)](https://github.com/jj-vcs/jj) がインストール済みであること
