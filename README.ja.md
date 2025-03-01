# gomoqt

MOQT (Media over QUIC Transfork) のGo言語による実装です。

## 概要

この実装はQUICトランスポートを使用したメディアストリーミングアプリケーションを構築するためのライブラリで、[MOQTransforkの仕様](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)に基づいています。

## 機能

- **MOQT プロトコル**: MOQTransforkのプロトコルのコア実装
- **WebTransport サポート**: WebTransportとraw QUICの両方の接続をサポート
- **サンプル実装**: いくつかのコード例を提供

## コンポーネント

### moqt

MOQTransforkの主要なやり取りを実装しています。

### lomc (開発予定)

Low Overhead Media Container の実装です。
**注意:** このパッケージは現在開発中です。

### catalog (開発予定)

MOQ Catalogの実装で、コンテンツの検出と管理を行います。
**注意:** このパッケージは現在開発中です。

## 開発

### 前提条件

- Go 1.22以降
- [just](https://github.com/casey/just) コマンドツール

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

3. 開発環境のセットアップ:
```bash
just dev-setup
```

このコマンドで以下の項目が実行されます:
- 必要な証明書ツール（mkcert）のインストール
- 開発ツール（goimports, golangci-lint）のインストール
- プロジェクトの依存関係のインストール
- 開発用証明書の生成

### 開発用コマンド

#### サンプルの実行
```bash
# エコーサーバーの起動
just run-echo-server

# 別のターミナルでエコークライアントを実行
just run-echo-client
```

#### コードの品質管理
```bash
# コードフォーマット
just fmt

# リンター実行
just lint

# テスト実行
just test

# コード品質チェック（フォーマットとリント）
just check
```

#### ビルドとクリーン
```bash
# コードのビルド
just build

# 生成ファイルの削除
just clean
```

### サンプル

[examples](examples) ディレクトリには、gomoqtの使用方法を示すサンプルアプリケーションが含まれています:

- **エコーサーバーとクライアント** `echo/`: シンプルなエコーサーバーとクライアントの実装
- 追加予定...

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
   just fmt
   just lint
   just test
   ```
5. 変更をコミット (`git commit -m 'Add amazing feature'`)
6. ブランチにプッシュ (`git push origin feature/amazing-feature`)
7. プルリクエストを作成

## ライセンス

このプロジェクトはMITライセンスです。詳細は[LICENSE](LICENSE)を参照してください。

## 参考文献

- [quic-go](https://github.com/quic-go/quic-go) - Go言語によるQUIC実装
- [webtransport-go](https://github.com/quic-go/webtransport-go) - Go言語によるWebTransport実装
- [moq-drafts](https://github.com/kixelated/moq-drafts) - MOQ Transfork仕様