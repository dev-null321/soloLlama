# soloLlama

A tiny self-hosted web UI for chatting with local Llama / Mistral / Gemma models managed by **Ollama**.  
Written in pure Go—builds to a single executable that embeds the entire HTML/JS front-end.

---

## Features

| Description |
|-------------|
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
| **Ollama**     | Needs to be reachable on `localhost:11434`.<br>• native install on macOS/Linux<br>• **Windows:** run in WSL 2 or Docker: `docker run -d -p 11434:11434 ollama/ollama` |
| **Linux**      | If you don't have xdg-utils installed please install it. 
| **soloLlama**  |`http://localhost:8081`


---

## Build

```bash
# clone
https://github.com/dev-null321/soloLlama.git
cd soloLlama

# build native binary (Linux/macOS host)
go build -o soloLlama   # builds for your host OS

# cross-compile to Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o soloLlama.exe
```

https://github.com/user-attachments/assets/5d218d22-00be-4939-8197-30b01501193a
