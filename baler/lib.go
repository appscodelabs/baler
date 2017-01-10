package baler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	shell "github.com/codeskyblue/go-sh"
	"github.com/mitchellh/ioprogress"
)

func Pack(manifestPath, dest string) error {
	mf, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}

	// https://github.com/hashicorp/consul-template/issues/58
	tmp, err := ioutil.TempDir(dest, "baler")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	fmt.Println("Using temporary dir:", tmp)
	fmt.Println("-------------------------------------------------------------------------------------")

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

func Unpack(archivePath, dest string) error {
	var err error
	if strings.HasPrefix(archivePath, "http://") || strings.HasPrefix(archivePath, "https://") {
		archivePath, err = download(archivePath)
		if err != nil {
			return err
		}
	}

	archivePath, err = filepath.Abs(archivePath)
	if err != nil {
		return err
	}

	fi, err := os.Stat(archivePath)
	if err == nil && fi.IsDir() {
		return fmt.Errorf("%s is not a file", archivePath)
	}
	mfName := filepath.Base(archivePath)
	if strings.Contains(mfName, ".") {
		mfName = mfName[:strings.Index(mfName, ".")]
	}

	// https://github.com/hashicorp/consul-template/issues/58
	tmp, err := ioutil.TempDir(dest, "baler")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	fmt.Println("Using temporary dir:", tmp)
	fmt.Println("-------------------------------------------------------------------------------------")

	fmt.Println("[*] Extracting archive ...")
	sh := newShell()
	sh.SetDir(tmp)
	sh.Command("tar", "-xzf", archivePath).Run()
	fmt.Println("_____________________________________________________________________________________")
	fmt.Println()

	root := tmp + "/" + mfName
	layerDir := root + "/__layers"
	fi, err = os.Stat(layerDir)
	if os.IsNotExist(err) || (err == nil && !fi.IsDir()) {
		return errors.New("Layers are missing.")
	}
	os.MkdirAll(dest+"/"+mfName, 0755)

	images, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}
	for _, img := range images {
		if img.Name() == "__layers" || !img.IsDir() {
			continue
		}

		fmt.Println()
		fmt.Println("[*] Image:", img.Name())
		fmt.Println("-------------------------------------------------------------------------------------")
		fmt.Println("[-] copying layers ...")

		d := root + "/" + img.Name()
		mfBytes, err := ioutil.ReadFile(d + "/manifest.json")
		if err != nil {
			return err
		}
		var imgMFs []ImageManifest
		err = json.Unmarshal(mfBytes, &imgMFs)
		if err != nil {
			return err
		}
		for _, imgMF := range imgMFs {
			for _, layer := range imgMF.Layers {
				layer = layer[:len(layer)-len("/layer.tar")]
				err = sh.Command("cp", "-r", fmt.Sprintf("%s/%s/layer.tar", layerDir, layer), fmt.Sprintf("%s/%s/layer.tar", d, layer)).Run()
				if err != nil {
					return err
				}
			}
		}
		fmt.Println()

		fmt.Println("[-] repackaging image ...")
		sh.SetDir(d)
		err = sh.Command("tar", "-czf", fmt.Sprintf("%s/%s/%s.tar", dest, mfName, img.Name()), ".").Run()
		if err != nil {
			return err
		}
		fmt.Println("_____________________________________________________________________________________")
		fmt.Println("")
	}
	return err
}

func Load(archivePath string) error {
	var err error
	if strings.HasPrefix(archivePath, "http://") || strings.HasPrefix(archivePath, "https://") {
		archivePath, err = download(archivePath)
		if err != nil {
			return err
		}
	}

	fi, err := os.Stat(archivePath)
	if err == nil && fi.IsDir() {
		return fmt.Errorf("%s is not a file", archivePath)
	}
	mfName := filepath.Base(archivePath)
	if strings.Contains(mfName, ".") {
		mfName = mfName[:strings.Index(mfName, ".")]
	}

	destDir, _ := os.Getwd()
	Unpack(archivePath, destDir)
	d := destDir + "/" + mfName
	sh := newShell()
	for {
		restart := false
		images, err := ioutil.ReadDir(d)
		if err != nil {
			return err
		}
		for _, img := range images {
			if !img.IsDir() {
				err = sh.Command("docker", "load", "-i", d+"/"+img.Name()).SetTimeout(120 * time.Second).Run()
				restart = err != nil
			}
		}
		if restart {
			sh.Command("systemctl", "restart", "docker").Run()
		} else {
			break
		}
	}
	return err
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

func loadManifest(manifestPath string) (*BalerManifest, error) {
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var mf BalerManifest
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
