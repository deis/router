package nginx

import (
	"log"
	"os"
	"os/exec"
)

const (
	nginxBinary = "/opt/nginx/sbin/nginx"
)

// Start nginx.
func Start() error {
	log.Println("INFO: Starting nginx...")
	cmd := exec.Command(nginxBinary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Println("INFO: nginx started.")
	return nil
}

// Reload nginx configuration.
func Reload() error {
	log.Println("INFO: Reloading nginx...")
	cmd := exec.Command(nginxBinary, "-s", "reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Println("INFO: nginx reloaded.")
	return nil
}
