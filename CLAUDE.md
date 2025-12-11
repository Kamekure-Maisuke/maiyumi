# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## コード規約
- Goの標準かつ可読性が高い方法で記述。
- 動的計画法を用いた最も効率的なアルゴリズムで記述。
- 極力標準ライブラリを利用。
- 標準ライブラリを利用するには、`~/.local/share/mise/installs/go/1.25.5` を確認してから実装。
  - バージョン互換性の違いをなくすため。
- HTMLはW3Cに沿った標準的な記法で記述。
- CSSはstatic/cssディレクトリ。BEM + ITCSS設計。記述はシンプルかつ効率的にすること。
- repository層のテストは極力書く。ケースは最低限。

## 技術スタック
- Go 1.25.5

## プロジェクト構成

```
|--.github
|--data.db
|--dump
|  |--dump_20251206_114340.sql
|--go.mod
|--go.sum
|--main.go
|--model
|  |--model.go
|--repository
|  |--adjustment.go
|  |--session.go
|  |--talent.go
|  |--user.go
|--scripts
|  |--dump_db.go
|--static
|  |--css
|  |  |--components
|  |  |  |--button.css
|  |  |  |--card.css
|  |  |  |--container.css
|  |  |  |--form.css
|  |  |  |--nav.css
|  |  |  |--stat.css
|  |  |  |--table.css
|  |  |--elements
|  |  |  |--base.css
|  |  |--generic
|  |  |  |--reset.css
|  |  |--main.css
|  |  |--settings
|  |  |  |--variables.css
|  |  |--utilities
|  |  |  |--helpers.css
|--templates
|  |--index.tmpl
|  |--login.tmpl
|  |--register.tmpl
|  |--talent_detail.tmpl
|  |--talent_form.tmpl
|  |--talents.tmpl
```
