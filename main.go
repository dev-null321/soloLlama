package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AskRequest is the inbound payload carrying model and history
type AskRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatRequest is the payload sent to Ollama API
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

var ollamaCmd *exec.Cmd

func startOllama() error {
	ollamaCmd = exec.Command("ollama", "serve")
	err := ollamaCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ollama: %v", err)
	}
	time.Sleep(3 * time.Second)
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return fmt.Errorf("ollama not responding: %v", err)
	}
	resp.Body.Close()
	return nil
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if ollamaCmd != nil && ollamaCmd.Process != nil {
			ollamaCmd.Process.Kill()
		}
		os.Exit(0)
	}()
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("Please open browser and navigate to %s\n", url)
		return
	}
	if err != nil {
		fmt.Printf("Error opening browser: %v\n", err)
	}
}

// handleAsk streams clean text from Ollama, parsing JSON lines
func handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var askReq AskRequest
	if err := json.NewDecoder(r.Body).Decode(&askReq); err != nil {
		http.Error(w, "Malformed request", http.StatusBadRequest)
		return
	}
	if askReq.Model == "" {
		http.Error(w, "Model is required", http.StatusBadRequest)
		return
	}

	// prepare payload with streaming enabled
	payload := ChatRequest{
		Model:    askReq.Model,
		Messages: askReq.Messages,
		Stream:   true,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Encoding payload failed", http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewReader(buf))
	if err != nil {
		http.Error(w, fmt.Sprintf("Ollama unreachable: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk struct {
			Message ChatMessage `json:"message"`
		}
		if err := decoder.Decode(&chunk); err != nil {
			break
		}
		w.Write([]byte(chunk.Message.Content))
		flusher.Flush()
	}
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		http.Error(w, "Ollama unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func handleCreateModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	modelName := r.FormValue("modelName")
	if modelName == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}
	gguf, hdr, err := r.FormFile("modelFile")
	if err != nil {
		http.Error(w, "Missing .gguf file", http.StatusBadRequest)
		return
	}
	defer gguf.Close()
	path := filepath.Join(".", hdr.Filename)
	f, _ := os.Create(path)
	defer f.Close()
	defer os.Remove(path)
	io.Copy(f, gguf)
	mfPath := ""
	mf, mfHdr, err := r.FormFile("modelfile")
	if err == nil {
		defer mf.Close()
		mfPath = filepath.Join(".", mfHdr.Filename)
		f2, _ := os.Create(mfPath)
		defer f2.Close()
		defer os.Remove(mfPath)
		io.Copy(f2, mf)
	}
	var cmd *exec.Cmd
	if mfPath != "" {
		cmd = exec.Command("ollama", "create", modelName, "-f", mfPath)
	} else {
		tmp := fmt.Sprintf("FROM %s", path)
		tp := filepath.Join(".", "tempfile_"+modelName)
		os.WriteFile(tp, []byte(tmp), 0644)
		defer os.Remove(tp)
		cmd = exec.Command("ollama", "create", modelName, "-f", tp)
	}
	outb, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ollama create error: %s", string(outb)), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Model created successfully"))
}

func handleDebug(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		fmt.Fprintf(w, "❌ Ollama connection failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Fprintf(w, "✅ Ollama running (status: %d)\n", resp.StatusCode)
	b, _ := io.ReadAll(resp.Body)
	fmt.Fprintf(w, "Models:\n%s\n", string(b))
}

func main() {
	log.Println("Starting Ollama...")
	if err := startOllama(); err != nil {
		log.Fatalf("Error: %v", err)
	}
	setupSignalHandler()
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ask", handleAsk)
	http.HandleFunc("/models", handleModels)
	http.HandleFunc("/create-model", handleCreateModel)
	http.HandleFunc("/debug", handleDebug)
	openBrowser("http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
