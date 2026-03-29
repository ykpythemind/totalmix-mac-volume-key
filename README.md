# totalmix-mac-volume-key

macOSのボリュームキー（F11/F12/Mute）でRME TotalMix FXのマスターボリュームを操作するツール。

RME Firefaceなどのオーディオインターフェースを接続するとmacOSのボリュームキーが無効になる問題を解決します。

## 仕組み

- `CGEventTap` でメディアキーイベントを傍受
- OSCプロトコル経由でTotalMix FXのマスターボリュームを制御
- 出力デバイスがRME以外の場合はキーをスルー（通常のシステムボリュームが動作）

## 前提条件

- macOS
- RME TotalMix FX がインストール済み
- TotalMix FX の OSC 設定:
  - Options > Settings > OSC タブ
  - OSC Control を有効化
  - Remote Controller の Incoming Port: `7001`
  - Remote Controller の Outgoing Port: `9001`
  - Remote Controller の Host: `127.0.0.1`

## ビルド

```
go build -o totalmix-mac-volume-key .
```

## 使い方

```
./totalmix-mac-volume-key
```

初回実行時にmacOSのアクセシビリティ権限を求められます。
System Settings > Privacy & Security > Accessibility で許可してください。

### オプション

```
-send-port 7001    TotalMix FX の OSC 受信ポート（デフォルト: 7001）
-recv-port 9001    TotalMix FX の OSC 送信ポート（デフォルト: 9001）
-step 0.01         1キー押下あたりのボリュームステップ（デフォルト: 0.01 = 1%）
```
