package config

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lijinglin3/clash/component/mmdb"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/log"
)

func downloadMMDB(path string) (err error) {
	resp, err := http.Get("https://cdn.jsdelivr.net/gh/Dreamacro/maxmind-geoip@release/Country.mmdb")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)

	return err
}

func initMMDB() error {
	if _, err := os.Stat(constant.Path.MMDB()); os.IsNotExist(err) {
		log.Infoln("Can't find MMDB, start download")
		if err := downloadMMDB(constant.Path.MMDB()); err != nil {
			return fmt.Errorf("can't download MMDB: %s", err.Error())
		}
	}

	if !mmdb.Verify() {
		log.Warnln("MMDB invalid, remove and download")
		if err := os.Remove(constant.Path.MMDB()); err != nil {
			return fmt.Errorf("can't remove invalid MMDB: %s", err.Error())
		}

		if err := downloadMMDB(constant.Path.MMDB()); err != nil {
			return fmt.Errorf("can't download MMDB: %s", err.Error())
		}
	}

	return nil
}

// Init prepare necessary files
func Init(dir string) error {
	// initial homedir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return fmt.Errorf("can't create config directory %s: %s", dir, err.Error())
		}
	}

	// initial config.yaml
	if _, err := os.Stat(constant.Path.Config()); os.IsNotExist(err) {
		log.Infoln("Can't find config, create a initial config file")
		f, err := os.OpenFile(constant.Path.Config(), os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("can't create file %s: %s", constant.Path.Config(), err.Error())
		}
		f.Write([]byte(`mixed-port: 7890`))
		f.Close()
	}

	// initial mmdb
	if err := initMMDB(); err != nil {
		return fmt.Errorf("can't initial MMDB: %w", err)
	}
	return nil
}
