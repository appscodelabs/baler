package baler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	shell "github.com/codeskyblue/go-sh"
	"github.com/mitchellh/ioprogress"
)

func Pack(manifestPath, dest string) error {
	mf, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempDir("", "baler")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	root := tmp + "/" + mf.Name
	layerDir := root + "/__layers"
	err = os.MkdirAll(layerDir, 0755)
	if err != nil {
		return err
	}

	sh := newShell()
	for _, img := range mf.Images {
		fmt.Println("Docker: ", img)
		err = sh.Command("docker", "pull", img).Run()
		if err != nil {
			return err
		}

		name := img[:strings.LastIndex(img, ":")]
		name = strings.Replace(name, "/", "-", 1)
		d := root + "/" + name
		err = os.MkdirAll(d, 0755)
		if err != nil {
			return err
		}

		err = sh.Command("docker", "save", img).WriteOutput(d + "/docker.tar")
		if err != nil {
			return err
		}
		sh.SetDir(d)
		err = sh.Command("tar", "xvf", "docker.tar").Run()
		if err != nil {
			return err
		}
		err = sh.Command("rm", "docker.tar").Run()
		if err != nil {
			return err
		}

		layers, err := ioutil.ReadDir(d)
		if err != nil {
			return err
		}
		for _, layer := range layers {
			if layer.IsDir() {
				fi, err := os.Stat(layerDir + "/" + layer.Name())
				if err == nil && fi.IsDir() {
					err = os.Remove(fmt.Sprintf("%s/%s/layer.tar", d, layer.Name()))
					if err != nil {
						return err
					}
				} else {
					err = os.Mkdir(layerDir+"/"+layer.Name(), 0755)
					if err != nil {
						return err
					}
					err = os.Rename(fmt.Sprintf("%s/%s/layer.tar", d, layer.Name()), fmt.Sprintf("%s/%s/layer.tar", layerDir, layer.Name()))
					if err != nil {
						return err
					}
				}
			}
		}
	}

	sh.SetDir(tmp)
	err = sh.Command("tar", "-czf", mf.Name+".tar.gz", mf.Name).Run()
	if err != nil {
		return err
	}
	return os.Rename(fmt.Sprintf("%s/%s.tar.gz", tmp, mf.Name), fmt.Sprintf("%s/%s.tar.gz", dest, mf.Name))
}

func Unpack(archivePath, dest string) {

}

func Load(archivePath string) {

}

func RemoveImages(manifestPath string) error {
	mf, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}

	sh := newShell()
	for _, img := range mf.Images {
		err = sh.Command("docker", "rmi", img).Run()
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func newShell() *shell.Session {
	session := shell.NewSession()
	session.ShowCMD = true
	return session
}

func loadManifest(manifestPath string) (*Manifest, error) {
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var mf Manifest
	err = json.Unmarshal(data, &mf)
	if err != nil {
		return nil, err
	}
	return &mf, nil
}

func download(rawurl string) (string, error) {
	fmt.Println("Downloading " + rawurl)

	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	f := u.Path[strings.LastIndex(u.Path, "/")+1:]

	resp, err := http.Get(rawurl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bar := ioprogress.DrawTextFormatBar(50)
	progressR := &ioprogress.Reader{
		Reader: resp.Body,
		Size:   resp.ContentLength,
		DrawFunc: ioprogress.DrawTerminalf(os.Stdout, func(progress, total int64) string {
			return fmt.Sprintf(
				"%s %s",
				bar(progress, total),
				ioprogress.DrawTextFormatBytes(progress, total))
		}),
	}
	out, err := os.Create(f)
	if err != nil {
		return "", err
	}
	defer out.Close()
	sz, err := io.Copy(out, progressR)
	if err != nil {
		return "", err
	}
	fmt.Printf("Written %d bytes", sz)
	return f, nil
}
