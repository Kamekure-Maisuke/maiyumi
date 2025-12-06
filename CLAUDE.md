# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## コード規約
- Goの標準かつ可読性が高い方法で記述。
- 動的計画法を用いた最も効率的なアルゴリズムで記述。
- 極力標準ライブラリを利用。
- 標準ライブラリを利用するには、`~/.local/share/mise/installs/go/1.25.5` を確認してから実装。
  - バージョン互換性の違いをなくすため。
- HTMLはW3Cに沿った標準的な記法で記述。

## 技術スタック
- Go 1.25.5

## プロジェクト構成

```
go.mod
main.go: エンドポイント
templates: viewファイル
|--index.tmpl
```
