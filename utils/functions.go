package utils

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func FetchStorefront(url string) (models.Storefront, error) {
	var data []byte
	var err error

	if os.Getenv("ENVIRONMENT") == "DEV" {
		data, err = os.ReadFile("storefront.json")
		if err != nil {
			return models.Storefront{}, fmt.Errorf("failed to read local storefront.json", err)
		}
	} else {
		data, err = fetch(url)
		if err != nil {
			return models.Storefront{}, err
		}
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

func ParseJSONFile(filePath string, out *models.Pak) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func DownloadPakArchive(pak models.Pak, action string) (string, error) {
	logger := common.GetLoggerInstance()

	releasesStub := fmt.Sprintf("/releases/download/%s/", pak.Version)
	dl := pak.RepoURL + releasesStub + pak.ReleaseFilename
	tmp := filepath.Join("/tmp", pak.ReleaseFilename)

	message := ""

	if action == "Updating" {
		message = fmt.Sprintf("%s %s to %s...", action, pak.StorefrontName, pak.Version)
	} else {
		message = fmt.Sprintf("%s %s %s...", action, pak.StorefrontName, pak.Version)
	}

	_, err := gabagool.DownloadManager([]gabagool.Download{{
		URL:         dl,
		Location:    tmp,
		DisplayName: message,
	}}, make(map[string]string))
	if err != nil {
		logger.Error("Error downloading", zap.Error(err))
		return "", err
	}

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
	} else if resp.ContentLength <= 0 {
		return "", fmt.Errorf("empty response")
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

func Unzip(src, dest string, pak models.Pak, isUpdate bool) error {
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
		if isUpdate && ShouldIgnoreFile(f.Name, pak) {
			return nil
		}

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

func ShouldIgnoreFile(filePath string, pak models.Pak) bool {
	for _, ignorePattern := range pak.UpdateIgnore {
		match, err := filepath.Match(ignorePattern, filePath)
		if err == nil && match {
			return true
		}

		parts := strings.Split(filePath, string(os.PathSeparator))
		for i := 0; i < len(parts); i++ {
			if i > 0 && strings.HasSuffix(parts[i-1], ".pak") {
				break
			}

			partialPath := strings.Join(parts[:i+1], string(os.PathSeparator))
			match, err := filepath.Match(ignorePattern, partialPath)
			if err == nil && match {
				return true
			}
		}
	}

	return false
}
