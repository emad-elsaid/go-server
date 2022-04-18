package main

import (
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"os"
)

func Helpers() {
	HELPER("partial", func(path string, data interface{}) (template.HTML, error) {
		return template.HTML(partial(path, data)), nil
	})

	HELPER("sha256", func() interface{} {
		cache := map[string]string{}
		return func(p string) (string, error) {
			if v, ok := cache[p]; ok {
				return v, nil
			}

			f, err := os.Open(p)
			if err != nil {
				return "", err
			}

			d, err := io.ReadAll(f)
			if err != nil {
				return "", err
			}

			cache[p] = fmt.Sprintf("%x", sha256.Sum256(d))
			return cache[p], nil
		}
	}())
}
