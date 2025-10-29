# gomoqt

MOQT（Media over QUIC Transport）のGo言語による実装で、特にMOQ Lite仕様に準拠した効率的なメディアストリーミングをQUIC上で実現します。

[![codecov](https://codecov.io/gh/OkutaniDaichi0106/gomoqt/branch/main/graph/badge.svg?token=4LZCD3FEU3)](https://codecov.io/gh/OkutaniDaichi0106/gomoqt)

## 概要

この実装は[MOQ Lite仕様](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)に従い、QUICトランスポートを使用したリアルタイムメディアストリーミングアプリケーションを構築するための基盤を提供します。MOQ Liteは、QUIC上のMedia over QUICトランスポートプロトコルの簡略版で、低遅延と低複雑性を実現しながら、QUICベースのメディア配信の核となる利点を保持するよう設計されています。

## 機能

この実装には以下が含まれます：

- **MOQ Lite プロトコル**: MOQ Lite仕様のコア実装
- **WebTransport サポート**: ブラウザ内のWebTransport接続への完全対応
- **Raw QUIC サポート**: ネイティブアプリケーション向けの直接QUIC接続
- **トラック管理**: メディアトラック処理のPublisher/Subscriberパターン
- **マルチプレックスストリーミング**: 複数のメディアトラックの効率的なマルチプレックス化
- **サンプルアプリケーション**: さまざまなユースケースを示す完全な実装例

- **リアルタイムストリーミング**:
	ライブイベント、会議、低遅延監視などのインタラクティブなユースケースにおいて、エンドツーエンドの遅延を最小化。ユーザー体験に応答性が重要なシナリオに最適です。

- **安定したストリーミング**:
	様々なネットワーク条件下での堅牢な再生。QUIC/WebTransportの基盤の上に構築され、途切れを減らしパケット損失からの復帰を改善します。

- **効率的なコンテンツ配信**:
	プロトコルレベルの最適化とマルチプレックスにより、多数の同時視聴者やストリームをサービス提供する際の接続オーバーヘッドとインフラコストを削減。

- **滑らかな再生**:
	ジッターとバッファ管理により、バッファリングを削減し、視聴者に滑らかで継続的な再生体験を提供。

- **最適化された映像品質**:
	帯域幅が限られた環境下で使用可能な品質を優先するアダプティブ配信パターン。デバイスの種類を問わず一貫したユーザー体験を実現。

## コンポーネント

| コンポーネント | 説明 |
|-----------|------|
| **moqt** | MOQ Liteの主要なやり取りを実装しています。 |
| **moq-web** | ブラウザ向けWebTransportをサポートするTypeScript/JavaScript実装です。 |
| **interop** | 異なるプラットフォーム間でのMOQ実装の相互運用性を検証するためのテストツールとサンプルです。 |

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

3. Mageビルドツールのインストール:
   ```bash
   go install github.com/magefile/mage@latest
   ```

注：開発環境セットアップコマンド（dev-setup、証明書生成など）は引き続きJustfileで利用可能です。コアビルドコマンド（test、lint、fmt、build、clean）はMageに移行しました。

### 開発用コマンド

#### サンプルの実行

```bash
# Interopサーバーを起動
just interop-server

# 別のターミナルでInteropクライアントを実行
just interop-client
```

#### コード品質
```bash
# コードをフォーマット
mage fmt

# リンターを実行（golangci-lintが必要: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest）
mage lint

# テストを実行
mage test
```

#### ビルド & クリーン
```bash
# コードをビルド
mage build

# 生成ファイルをクリーンアップ
mage clean

# 利用可能なコマンドを表示
mage help
```

### サンプル

[examples](examples) ディレクトリには、gomoqtの使用方法を示すサンプルアプリケーションが含まれています：

- **Interop Server and Client** (`interop/`): 異なるMOQ実装間での相互運用性テスト
- **ブロードキャスト例** (`examples/broadcast/`): ブロードキャスト機能のデモンストレーション
- **エコー例** (`examples/echo/`): シンプルなエコーサーバーとクライアント実装
- **Native QUIC** (`examples/native_quic/`): 直接QUIC接続の例
- **リレー** (`examples/relay/`): メディアストリーミング向けリレー機能

### ドキュメント

- [GoDoc](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt)
- [MOQ Lite仕様](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)
- [実装状況](moqt/README.md) - 詳細な実装進捗

## 仕様準拠

この実装はMOQ Lite仕様をターゲットとし、QUIC上のMedia over QUICトランスポートへの簡略化されたアプローチを提供します。現在の実装状況は[moqtパッケージREADME](moqt/README.md)に記載されており、仕様セクション別に実装機能の詳細な追跡が含まれています。

## コントリビューション

貢献を歓迎します！以下の方法でご協力いただけます：

1. リポジトリをフォーク
2. 機能ブランチを作成（`git checkout -b feature/amazing-feature`）
3. 変更を加える
4. コード品質を確認：
   ```bash
   mage fmt
   mage lint
   mage test
   ```
5. 変更をコミット（`git commit -m 'Add amazing feature'`）
6. ブランチにプッシュ（`git push origin feature/amazing-feature`）
7. プルリクエストを作成

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)を参照してください。

## 謝辞

- [quic-go](https://github.com/quic-go/quic-go) - Go言語によるQUIC実装
- [webtransport-go](https://github.com/quic-go/webtransport-go) - Go言語によるWebTransport実装
- [MOQ Lite仕様](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html) - この実装が従う仕様