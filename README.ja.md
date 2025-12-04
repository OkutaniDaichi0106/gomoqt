# gomoqt

MOQT (Media over QUIC Transfork) のGo言語による実装です。

[![Go Reference](https://pkg.go.dev/badge/github.com/OkutaniDaichi0106/gomoqt.svg)](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt)
[![codecov](https://codecov.io/gh/OkutaniDaichi0106/gomoqt/branch/main/graph/badge.svg?token=4LZCD3FEU3)](https://codecov.io/gh/OkutaniDaichi0106/gomoqt)

## 目次

- [概要](#概要)
- [機能](#機能)
- [コンポーネント](#コンポーネント)
- [開発](#開発)
- [サンプル](#サンプル)
- [ドキュメント](#ドキュメント)
- [コントリビューション](#コントリビューション)
- [ライセンス](#ライセンス)
- [参考文献](#参考文献)

## 概要

この実装はQUICトランスポートを使用したメディアストリーミングアプリケーションを構築するためのライブラリで、[MOQTransforkの仕様](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)に基づいています。

## 機能

- **MOQT プロトコル**: MOQTransforkのプロトコルのコア実装
- **WebTransport サポート**: WebTransportとraw QUICの両方の接続をサポート
- **サンプル実装**: いくつかのコード例を提供

## コンポーネント

### moqt

MOQ Liteの主要なやり取りを実装しています。

### moq-web

ブラウザ向けWebTransportをサポートするTypeScript/JavaScript実装です。

### interop

異なるプラットフォーム間でのMOQ実装の相互運用性を検証するためのテストツールとサンプルです。

## 開発

### 前提条件

- Go 1.25.0以降
- [Mage](https://magefile.org/) ビルドツール（`go install github.com/magefile/mage@latest` でインストール）

### はじめ方

1. リポジトリのクローン:

```bash
git clone https://github.com/OkutaniDaichi0106/gomoqt.git
cd gomoqt
```

2. パッケージのインストール:

```bash
go get github.com/OkutaniDaichi0106/gomoqt
```

3. Mageツールのインストール:

```bash
go install github.com/magefile/mage@latest
```

### 開発用コマンド

#### サンプルの実行

```bash
# 相互運用性テスト用サーバーの起動
mage interop:server

# 別のターミナルで相互運用性テスト用クライアントを実行（Go）
mage interop:client go

# またはTypeScriptクライアントを実行
mage interop:client ts
```

#### コードの品質管理

```bash
# コードフォーマット
mage fmt

# リンター実行
mage lint

# コード品質チェック（フォーマットとリント）
mage check

# 全テスト実行
mage test:all

# カバレッジ付きテスト実行
mage test:coverage
```

#### ビルドとクリーン

```bash
# コードのビルド
mage build

# 生成ファイルの削除
mage clean
```

### サンプル

[examples](examples)
ディレクトリには、gomoqtの使用方法を示すサンプルアプリケーションが含まれています:

- **Interopサーバーとクライアント** `cmd/interop/`: 異なるMOQ実装間の相互運用性テスト
- **ブロードキャスト** `examples/broadcast/`: ブロードキャスト機能のデモ
- **証明書** `examples/cert/`: 証明書管理のサンプル
- **エコー** `examples/echo/`: シンプルなエコーサーバーとクライアントの実装
- **ネイティブQUIC** `examples/native_quic/`: QUICプロトコルを使用した直接通信
- **リレー** `examples/relay/`: リレーサーバーの実装

### ドキュメント

- [Goドキュメント](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt)
- [MOQTransforkの仕様](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)

## コントリビューション

ご協力お待ちしております！日本語対応できます！
参加方法は以下を参考にして下さい！

1. リポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更を加える
4. コード品質の確認:
   ```bash
   mage fmt
   mage lint
   mage test:all
   ```
5. 変更をコミット (`git commit -m 'Add amazing feature'`)
6. ブランチにプッシュ (`git push origin feature/amazing-feature`)
7. プルリクエストを作成

## ライセンス

このプロジェクトはMITライセンスです。詳細は[LICENSE](LICENSE)を参照してください。

## 参考文献

- [quic-go](https://github.com/quic-go/quic-go) - Go言語によるQUIC実装
- [webtransport-go](https://github.com/quic-go/webtransport-go) -
  Go言語によるWebTransport実装
- [moq-drafts](https://github.com/kixelated/moq-drafts) - MOQ Transfork仕様
