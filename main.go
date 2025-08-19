package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
)

var addr = flag.String("addr", ":8080", "Address the listening interface.")

func main() {
	flag.Parse()

	http.HandleFunc("/browse", browseHandler)
	http.HandleFunc("/clip", clipHandler)

	log.Printf("Listening %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("Listen failed: %v\n", err)
	}
}

// URLを開くハンドラ
func browseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("%v\t%d\t%s\t%s\n", err, http.StatusInternalServerError, r.RemoteAddr, "")
		return
	}
	if len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("empty url\t%d\t%s\t%s\n", http.StatusBadRequest, r.RemoteAddr, "")
		return
	}

	url := string(body)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("bad request\t%d\t%s\t%s\n", http.StatusBadRequest, r.RemoteAddr, url)
		return
	}

	err = openURL(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		log.Printf("%v\t%d\t%s\t%s\n", err, http.StatusInternalServerError, r.RemoteAddr, url)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("browse\t%d\t%s\t%s\n", http.StatusOK, r.RemoteAddr, url)
}

// クリップボードに文字列をコピーするハンドラ
func clipHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("%v\t%d\t%s\t%s\n", err, http.StatusInternalServerError, r.RemoteAddr, "")
		return
	}
	if len(body) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("empty text\t%d\t%s\t%s\n", http.StatusBadRequest, r.RemoteAddr, "")
		return
	}

	text := string(body)
	err = clipboard.WriteAll(text)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		log.Printf("%v\t%d\t%s\t%s\n", err, http.StatusInternalServerError, r.RemoteAddr, text)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("clip\t%d\t%s\t%s\n", http.StatusOK, r.RemoteAddr, text)
}

// OSごとにブラウザを開く
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // linux and others
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Run()
}
