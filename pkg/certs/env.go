package sslcerts

import (
	"os"
	"path"
)

func init() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}
	certsPath := path.Join(path.Dir(execPath), "certs")
	if st, err := os.Stat(certsPath); err != nil {
		return
	} else {
		if st.IsDir() {
			original := os.Getenv("SSL_CERT_DIR")
			if original == "" {
				os.Setenv("SSL_CERT_DIR", certsPath)
			} else {
				os.Setenv("SSL_CERT_DIR", original+":"+certsPath)
			}
		}
	}
}
