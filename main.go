package main

import (
	"archive/zip"
	_ "embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bramAristyo/go-ico/pkg"
	"github.com/joho/godotenv"
)

//go:embed static/index.html
var indexHTML []byte

//go:embed static/favicon.ico
var faviconICO []byte

func basicMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		start := time.Now()
		next(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	}
}

func main() {
	_ = godotenv.Load()
	mux := http.NewServeMux()

	mux.HandleFunc("/", basicMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	}))

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(faviconICO)
	})

	mux.HandleFunc("/upload", basicMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit total request size (e.g., 50MB)
		r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
		if err := r.ParseMultipartForm(50 << 20); err != nil {
			http.Error(w, "upload too large or invalid", http.StatusBadRequest)
			return
		}

		sizeParam := r.FormValue("size")
		if sizeParam == "" {
			sizeParam = "32"
		}

		sz, err := strconv.Atoi(strings.TrimSpace(sizeParam))
		if err != nil || sz <= 0 {
			http.Error(w, "invalid size", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["images"]
		if len(files) == 0 {
			http.Error(w, "no images uploaded", http.StatusBadRequest)
			return
		}

		if len(files) == 1 {
			f, err := files[0].Open()
			if err != nil {
				http.Error(w, "open error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()

			src, err := pkg.DecodeImage(f)
			if err != nil {
				http.Error(w, "decode error: "+err.Error(), http.StatusBadRequest)
				return
			}

			name := strings.TrimSuffix(files[0].Filename, filepath.Ext(files[0].Filename)) + ".ico"
			w.Header().Set("Content-Type", "image/x-icon")
			w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)

			img := pkg.ResizeImage(src, sz)
			pkg.EncodeICO(w, img)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="icons.zip"`)

		zw := zip.NewWriter(w)
		defer zw.Close()

		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				continue
			}

			src, err := pkg.DecodeImage(f)
			f.Close()
			if err != nil {
				continue
			}

			name := strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename)) + ".ico"
			entry, err := zw.Create(name)
			if err != nil {
				continue
			}

			pkg.EncodeICO(entry, pkg.ResizeImage(src, sz))
		}
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8099"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("server running at http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("could not listen on %s: %v", port, err)
	}
}
