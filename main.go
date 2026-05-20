package main

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bramAristyo/go-ico/pkg"
)

//go:embed static/index.html
var indexHTML []byte

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseMultipartForm(50 << 20)

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
	})

	fmt.Println("server running at http://localhost:8099")
	http.ListenAndServe(":8099", nil)
}
