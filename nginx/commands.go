package nginx

import (
	"log"
	"os/exec"
)

func Start() error {
	log.Println("INFO: Starting nginx...")
	if err := shellOut("/opt/nginx/sbin/nginx"); err != nil {
		return err
	}
	log.Println("INFO: nginx started.")
	return nil
}

func Reload() error {
	log.Println("INFO: Reloading nginx...")
	if err := shellOut("/opt/nginx/sbin/nginx -s reload"); err != nil {
		return err
	}
	log.Println("INFO: nginx reloaded.")
	return nil
}

func shellOut(cmd string) error {
	_, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}
