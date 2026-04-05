FFmpeg GUI Tool (Wails + Nuxt 4)

一個基於 Wails (Go) + Nuxt 4 的跨平台 FFmpeg 圖形化工具，提供影片轉檔、處理與即時串流（RTMP / SRT）功能。


---

✨ Features

🎬 基礎功能

影片 / 音訊轉檔（MP4, MKV, WebM, MP3...）

視訊壓縮（CRF / Bitrate）

解析度調整（720p / 1080p / 自訂）

幀率調整（FPS）

音訊抽取 / 替換

字幕燒錄（soft / hard subtitle）

GIF / WebP 轉換


🚀 進階功能

自訂 FFmpeg 指令（raw command）

批次處理（Batch Processing）

任務佇列（Queue）

即時進度顯示

錯誤 log / stdout 顯示


📡 串流功能（Streaming）

支援將影片或即時來源推流至串流平台：

RTMP 推流（YouTube / Twitch / 自建伺服器）

SRT 推流（低延遲傳輸）

支援來源：

本地影片檔案

WebCam / Capture Device

RTSP / HTTP 串流




---

🧱 Tech Stack

Frontend

Nuxt 4

Vue 3

UnoCSS


Backend

Wails v2

Go

FFmpeg



---

📦 Installation

1. 安裝 FFmpeg

# macOS
brew install ffmpeg

# Ubuntu
sudo apt install ffmpeg

# Windows
# 下載 https://ffmpeg.org/download.html

確認：

ffmpeg -version


---

2. 安裝 Wails

go install github.com/wailsapp/wails/v2/cmd/wails@latest


---

3. 安裝前端依賴

cd frontend
pnpm install


---

🚀 Development

wails dev

自動啟動 Go backend + Nuxt frontend

支援 hot reload



---

🏗 Build

wails build

輸出：

macOS .app

Windows .exe

Linux binary



---

📁 Project Structure

.
├── backend/
│   ├── main.go
│   ├── ffmpeg/
│   │   ├── runner.go
│   │   ├── command.go
│   │   └── stream.go
│
├── frontend/
│   ├── app.vue
│   ├── pages/
│   ├── components/
│   └── composables/
│
├── wails.json
└── README.md


---

⚙️ FFmpeg Integration

基本轉檔

ffmpeg -i input.mp4 -c:v libx264 -crf 23 output.mp4

WebP 轉換

ffmpeg -i input.gif -vcodec libwebp output.webp


---

📡 Streaming Usage

🔴 RTMP 推流

ffmpeg -re -i input.mp4 \
  -c:v libx264 -preset veryfast -b:v 3000k \
  -c:a aac -b:a 128k \
  -f flv rtmp://<server>/<app>/<stream_key>

🟣 SRT 推流

ffmpeg -re -i input.mp4 \
  -c:v libx264 -b:v 2500k \
  -c:a aac \
  -f mpegts \
  "srt://<host>:<port>?mode=caller&latency=200"


---

🧠 Backend Design (Go + Wails)

任務模型

type Task struct {
    ID        string
    Command   string
    Status    string
    Progress  float64
}

執行 FFmpeg

cmd := exec.Command("ffmpeg", args...)
stdout, _ := cmd.StdoutPipe()
cmd.Start()

透過解析 stderr / stdout 取得：

progress

fps

bitrate



---

🎨 Frontend Design (Nuxt 4)

核心頁面

/convert：轉檔

/stream：推流

/tasks：任務管理


Composable

export const useFFmpeg = () => {
  const startTask = async (payload) => {
    return await window.go.main.App.StartTask(payload)
  }
}


---

🔐 注意事項

RTMP key 請妥善保管

SRT 建議設定 latency

高 bitrate 需要穩定網路



---

🧩 Roadmap

[ ] HLS / DASH 支援

[ ] GPU 加速（NVENC / VAAPI）

[ ] OBS-like preview

[ ] 插件系統

[ ] 雲端任務同步



---

📄 License

MIT


---

🙌 Credits

FFmpeg

Wails

Nuxt



---

如果你有需求（例如 WebRTC / 直播聊天室整合），可以再擴展此專案。

我已經幫你把完整 README.md 建在右側 canvas，內容是偏工程實戰導向的版本（不是空泛介紹），包含：

FFmpeg GUI 功能設計（轉檔 / WebP / 任務佇列）

RTMP / SRT 推流範例（可直接用）

Wails + Go backend 架構

Nuxt 4 frontend 組織方式

專案結構

指令與實際 code snippet


如果你要更進一步優化，我可以幫你再升級幾個方向（這些會讓專案更「像產品」而不是 demo）：

1. 推流 UI 設計（很關鍵）

Stream Key 管理

Server preset（YouTube / Twitch / 自建）

bitrate / resolution slider



2. FFmpeg command builder（核心價值）

用 UI 組 command（避免讓使用者寫 CLI）

例如：

codec selector

CRF vs bitrate toggle




3. 即時 log parsing（你 Go 很適合做）

解析：

frame=

fps=

time=


前端做 progress bar



4. 串流 + 本地預覽（進階）

用：

ffplay

或 WebRTC bridge（你之前有做直播，這塊你會很強）






