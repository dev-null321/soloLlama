# soloLlama

A tiny self-hosted web UI for chatting with local Llama / Mistral / Gemma models managed by **Ollama**.  
Written in pure Go—builds to a single executable that embeds the entire HTML/JS front-end.

---

## Features

| ✔ | Description |
|---|-------------|
| **Streaming** answers token-by-token |
| **Conversation context** (messages array passed to Ollama) |
| Upload a **`.gguf`** file + optional **Modelfile** and create an Ollama tag from the UI |
| One-click model switcher |
| Runs on Windows, macOS, Linux; cross-compile from any host |


---

## Prerequisites

| Requirement | Notes |
|-------------|-------|
| **Go 1.21.6+** | `https://go.dev/dl` |
| **Ollama**  | Needs to be reachable on `localhost:11434`.<br>• native install on macOS/Linux<br>• **Windows:** run in WSL 2 or Docker: `docker run -d -p 11434:11434 ollama/ollama` |

---

## Build

```bash
# clone
git clone https://github.com/yourname/soloLlama.git
cd soloLlama

# build native binary (Linux/macOS host)
go build -o soloLlama   # builds for your host OS

# cross-compile to Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o soloLlama.exe




https://github.com/user-attachments/assets/a0ca8ee9-5863-4348-b51b-60afda26038c

