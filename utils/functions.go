package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FetchStorefront(url string) (models.Storefront, error) {
	data, err := fetch(url)
	if err != nil {
		return models.Storefront{}, err
	}

	var sf models.Storefront
	if err := json.Unmarshal(data, &sf); err != nil {
		return models.Storefront{}, err
	}

	return sf, nil
}

func FetchPakJson(url string) (models.Pak, error) {
	data, err := fetch(url)
	if err != nil {
		return models.Pak{}, err
	}

	var pak models.Pak
	if err := json.Unmarshal(data, &pak); err != nil {
		return models.Pak{}, err
	}

	return pak, nil
}

func DownloadPakArchive(pak models.Pak, action string) (string, error) {
	logger := common.GetLoggerInstance()

	releasesStub := fmt.Sprintf("/releases/download/%s/", pak.Version)
	dl := pak.RepoURL + releasesStub + pak.ReleaseFilename
	tmp := filepath.Join("/tmp", pak.ReleaseFilename)

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{
		"--message", fmt.Sprintf("%s %s %s...", action, pak.StorefrontName, pak.Version),
		"--timeout", "-1"}
	cmd := exec.CommandContext(ctxWithCancel, "minui-presenter", args...)

	var stdoutbuf, stderrbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &stderrbuf

	err := cmd.Start()
	if err != nil && cmd.ProcessState.ExitCode() != -1 {
		logger.Fatal("Error launching download screen... That's pretty dumb!", zap.Error(err))
	}

	err = DownloadFile(dl, tmp)
	if err != nil {
		logger.Error("Unable to download pak", zap.String("url", dl), zap.Error(err))
		cancel()
		return "", err
	}

	cancel()

	return tmp, nil
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func DownloadTempFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	tempFile, err := os.CreateTemp("", "download-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func DownloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(path), f.Mode())
			if err != nil {
				return err
			}

			// Use a temporary file to avoid ETXTBSY error
			tempPath := path + ".tmp"
			tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(tempFile, rc)
			tempFile.Close() // Close the file before attempting to rename it

			if err != nil {
				os.Remove(tempPath) // Clean up on error
				return err
			}

			// Now rename the temporary file to the target path
			err = os.Rename(tempPath, path)
			if err != nil {
				os.Remove(tempPath) // Clean up on error
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
