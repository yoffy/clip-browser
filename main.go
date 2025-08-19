package main

import (
	"flag"
	"fmt"
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

	http.HandleFunc("/open", openHandler)
	http.HandleFunc("/clip", clipHandler)

	log.Printf("Listening %s\n", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("Listen failed: %v\n", err)
	}
}

// URLを開くハンドラ
func openHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("empty url\t%d\t%s\t%s\n", http.StatusBadRequest, r.RemoteAddr, r.URL)
		return
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("bad request\t%d\t%s\t%s\n", http.StatusBadRequest, r.RemoteAddr, r.URL)
		return
	}

	err := openURL(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		log.Printf("%v\t%d\t%s\t%s\n", err, http.StatusInternalServerError, r.RemoteAddr, r.URL)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("open\t%d\t%s\t%s\n", http.StatusOK, r.RemoteAddr, r.URL)
}

// クリップボードに文字列をコピーするハンドラ
func clipHandler(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		log.Printf("empty text\t%d\t%s\t%s\n", http.StatusBadRequest, r.RemoteAddr, r.URL)
		return
	}

	err := clipboard.WriteAll(text)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		log.Printf("%v\t%d\t%s\t%s\n", err, http.StatusInternalServerError, r.RemoteAddr, r.URL)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("clip\t%d\t%s\t%s\n", http.StatusOK, r.RemoteAddr, r.URL)
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
