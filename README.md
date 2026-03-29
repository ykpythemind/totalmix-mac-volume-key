# totalmix-mac-volume-key

macOSのボリュームキー（F11/F12/Mute）でRME TotalMix FXのマスターボリュームを操作するツール。

RME Firefaceなどのオーディオインターフェースを接続するとmacOSのボリュームキーが無効になる問題を解決します。

## 仕組み

- `CGEventTap` でメディアキーイベントを傍受
- OSCプロトコル経由でTotalMix FXのマスターボリュームを制御
- 出力デバイスがRME以外の場合はキーをスルー（通常のシステムボリュームが動作）

## 前提条件

- macOS
- Go (ビルドに必要)
- RME TotalMix FX がインストール済み
- TotalMix FX の OSC 設定:
  1. メニューバーの Options > Enable OSC Control にチェック
  2. Options > Settings > OSC タブを開く
  3. Remote Controller の Incoming Port: `7001`
  4. Remote Controller の Outgoing Port: `9001`
  5. Remote Controller の Host: `127.0.0.1`

## インストール

### ビルド済みバイナリ

[Releases](https://github.com/ykpythemind/totalmix-mac-volume-key/releases) からダウンロードして:

```
# Apple Silicon (M1/M2/M3/M4)
chmod +x totalmix-mac-volume-key-darwin-arm64
sudo cp totalmix-mac-volume-key-darwin-arm64 /usr/local/bin/totalmix-mac-volume-key

# Intel Mac
chmod +x totalmix-mac-volume-key-darwin-amd64
sudo cp totalmix-mac-volume-key-darwin-amd64 /usr/local/bin/totalmix-mac-volume-key
```

### ソースからビルド

Go が必要です。

```
go build -o totalmix-mac-volume-key .
sudo cp totalmix-mac-volume-key /usr/local/bin/
```

### デーモン化（ログイン時に自動起動）

```
totalmix-mac-volume-key --install
```

launchd に登録され、ログイン時に自動起動・クラッシュ時に自動再起動します。

初回実行時にmacOSのアクセシビリティ権限を求められます。
System Settings > Privacy & Security > Accessibility で許可してください。

### アンインストール

```
totalmix-mac-volume-key --uninstall
```

### 手動で実行する場合

```
totalmix-mac-volume-key
```

### オプション

```
-send-port 7001    TotalMix FX の OSC 受信ポート（デフォルト: 7001）
-recv-port 9001    TotalMix FX の OSC 送信ポート（デフォルト: 9001）
-step 0.01         1キー押下あたりのボリュームステップ（デフォルト: 0.01 = 1%）
```

ログは `/tmp/totalmix-mac-volume-key.log` に出力されます。
