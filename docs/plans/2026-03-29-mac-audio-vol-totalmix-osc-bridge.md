# mac_audio_vol: macOS Volume Key → TotalMix FX Bridge

## Context
macOSでRME Firefaceを接続すると、システムのボリュームキー（F11/F12/Mute）が無効になる。
TotalMix FXはOSC（Open Sound Control）プロトコルをサポートしているため、ボリュームキーを傍受してOSC経由でTotalMix FXのマスターボリュームを制御するツールを作る。

## アーキテクチャ
Go製の単一バイナリ。3つのコンポーネント:

1. **CGEventTap (C via cgo)** - macOSのメディアキーイベント（NX_SYSDEFINED）を傍受
2. **OSC Client (Go)** - TotalMix FXにUDPでボリューム制御メッセージを送信
3. **Volume State (Go)** - 現在のボリュームレベル管理（0.0〜1.0）

## TotalMix FX OSCプロトコル
- **送信先**: UDP `127.0.0.1:7001`
- **受信元**: UDP `127.0.0.1:9001`（状態同期用）
- `/1/mastervolume` (float32, 0.0〜1.0) - マスターボリューム制御
- `/1/mainDim` (float32) - Dim（Muteの代替として使用）
- ステップ: 0.02（50段階、1キー押下あたり約1%）

## ファイル構成
```
mac_audio_vol/
  go.mod
  main.go           -- エントリポイント、シグナルハンドリング
  keytap.go         -- Go/cgo wrapper、CGEventTapのGoブリッジ
  keytap_darwin.c   -- CGEventTapCreate + コールバック実装
  keytap_darwin.h   -- Cヘッダ
  osc.go            -- OSC client（TotalMix FXとの通信）
  volume.go         -- ボリューム状態管理
```

## 実装ステップ

### Step 1: Go module初期化
- `go mod init mac_audio_vol`
- `github.com/hypebeast/go-osc` を依存に追加

### Step 2: CGEventTap実装 (`keytap_darwin.c`, `keytap_darwin.h`)
- `CGEventTapCreate` でNX_SYSDEFINEDイベントを傍受
- メディアキー（NX_KEYTYPE_SOUND_UP=0, SOUND_DOWN=1, MUTE=7）をフィルタ
- キーイベント検出時にGoのexported関数を呼び出し
- `NULL`を返してイベントを消費（エラー音防止）
- cgo LDFLAGS: `-framework CoreGraphics -framework IOKit`

### Step 3: Goブリッジ (`keytap.go`)
- `//export goMediaKeyCallback` でC→Go呼び出し
- `runtime.LockOSThread()` + `CFRunLoopRun()` でイベントループ実行
- キーイベントをchannelで送信

### Step 4: OSC通信 (`osc.go`)
- `go-osc`ライブラリでUDPメッセージ送信
- ポート9001でリッスンしてTotalMix FXからの状態同期を受信
- 起動時に現在のボリューム値を取得

### Step 5: ボリューム状態管理 (`volume.go`)
- `VolumeState` struct: level, muted, preMuteLevel, step
- Volume Up: `min(level + step, 1.0)` → OSC送信
- Volume Down: `max(level - step, 0.0)` → OSC送信
- Mute: トグル（現在値を保存して0.0を送信 / 復元）

### Step 6: メインエントリ (`main.go`)
- OSCクライアント初期化（127.0.0.1:7001）
- OSCリスナー起動（port 9001）
- CGEventTap起動（別goroutine）
- SIGINT/SIGTERMでクリーンシャットダウン

## 前提条件（ユーザー側の設定）
- TotalMix FXでOSCを有効化: Options > Enable OSC Control
- TotalMix FX incoming port: 7001, outgoing port: 9001
- macOSのアクセシビリティ権限をバイナリに付与

## 検証方法
1. `go build -o mac_audio_vol .` でビルド
2. TotalMix FXでOSCを有効化
3. `./mac_audio_vol` を実行
4. ボリュームキー（F11/F12）を押してTotalMix FXのマスターフェーダーが動くことを確認
5. Muteキーでミュート/アンミュートが動作することを確認
