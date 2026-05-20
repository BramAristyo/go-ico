package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bramAristyo/go-ico/pkg"
)

func main() {
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
	})
	fmt.Println("server running at http://localhost:8080")
	http.ListenAndServe(":8099", nil)
}
