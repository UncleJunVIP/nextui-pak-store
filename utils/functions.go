package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-store/models"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
	"image/color"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GetSDRoot() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return os.Getenv("SD_ROOT")
	}

	return models.SDRoot
}

func GetToolRoot() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return os.Getenv("TOOL_ROOT")
	}

	return models.ToolRoot
}

func GetEmulatorRoot() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return os.Getenv("EMULATOR_ROOT")
	}

	return models.EmulatorRoot
}

func FetchStorefront(url string) (models.Storefront, error) {
	logger := common.GetLoggerInstance()

	var data []byte
	var err error

	if override := os.Getenv("STOREFRONT_OVERRIDE"); override != "" {
		data, err = fetch(override)
		if err != nil {
			return models.Storefront{}, err
		}
	} else if os.Getenv("ENVIRONMENT") == "DEV" {
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

	for _, p := range sf.Paks {
		if filepath.Ext(p.ReleaseFilename) == ".pakz" {
			p.IsPakZ = true
		}
	}

	logger.Info("Fetched storefront", zap.String("name", sf.Name))

	return sf, nil
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

func DownloadPakArchive(pak models.Pak) (tempFile string, completed bool, error error) {
	logger := common.GetLoggerInstance()

	releasesStub := fmt.Sprintf("/releases/download/%s/", pak.Version)
	dl := pak.RepoURL + releasesStub + pak.ReleaseFilename
	tmp := filepath.Join("/tmp", pak.ReleaseFilename)

	message := fmt.Sprintf("Downloading %s %s...", pak.StorefrontName, pak.Version)

	res, err := gaba.DownloadManager([]gaba.Download{{
		URL:         dl,
		Location:    tmp,
		DisplayName: message,
	}}, make(map[string]string))

	if err == nil && len(res.Errors) > 0 {
		err = res.Errors[0]
	}

	if err != nil {
		logger.Error("Error downloading", zap.Error(err))
		return "", false, err
	} else if res.Cancelled {
		return "", false, nil
	}

	return tmp, true, nil
}

func RunScript(script models.Script, scriptName string) error {
	logger := common.GetLoggerInstance()

	if script.Path == "" {
		logger.Info("No script to run")
		return nil
	}

	_, err := gaba.ProcessMessage(fmt.Sprintf("%s %s %s...", "Running", scriptName, "Script"), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		logger.Info("Running script", zap.String("path", script.Path), zap.Strings("args", script.Args))

		cmd := exec.Command(script.Path, script.Args...)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			logger.Error("Failed to execute script",
				zap.String("path", script.Path),
				zap.Strings("args", script.Args),
				zap.String("stderr", stderr.String()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to execute script %s: %w", script.Path, err)
		}

		if cmd.ProcessState.ExitCode() != 0 {
			logger.Error("Script returned non-zero exit code",
				zap.String("path", script.Path),
				zap.Strings("args", script.Args),
				zap.Int("exitCode", cmd.ProcessState.ExitCode()),
				zap.String("stderr", stderr.String()))
			return nil, fmt.Errorf("script %s exited with code %d: %s",
				script.Path, cmd.ProcessState.ExitCode(), stderr.String())
		}

		logger.Info("Script executed successfully",
			zap.String("path", script.Path),
			zap.Strings("args", script.Args),
			zap.String("stdout", stdout.String()))

		return nil, nil
	})

	return err
}

func UnzipPakArchive(pak models.Pak, tmp string) error {
	logger := common.GetLoggerInstance()

	pakDestination := ""

	if pak.IsPakZ {
		pakDestination = GetSDRoot()
	} else if pak.PakType == models.PakTypes.TOOL {
		pakDestination = filepath.Join(GetToolRoot(), pak.Name+".pak")
	} else if pak.PakType == models.PakTypes.EMU {
		pakDestination = filepath.Join(GetEmulatorRoot(), pak.Name+".pak")
	}

	_, err := gaba.ProcessMessage(fmt.Sprintf("%s %s...", "Unzipping", pak.StorefrontName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		err := Unzip(tmp, pakDestination, pak, false)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		gaba.ProcessMessage(fmt.Sprintf("Unable to unzip %s", pak.StorefrontName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(3 * time.Second)
			return nil, nil
		})
		logger.Error("Unable to unzip pak", zap.Error(err))
		return err
	}

	return nil
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

func CreateTempQRCode(content string, size int) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)

	if err != nil {
		return "", err
	}

	qr.BackgroundColor = color.Black
	qr.ForegroundColor = color.White
	qr.DisableBorder = true

	tempFile, err := os.CreateTemp("", "qrcode-*")

	err = qr.Write(size, tempFile)

	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	return tempFile.Name(), err
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
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return err
			}

			// Use a temporary file to avoid ETXTBSY error
			tempPath := path + ".tmp"
			tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
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

func IsConnectedToInternet() bool {
	timeout := 5 * time.Second
	_, err := net.DialTimeout("tcp", "8.8.8.8:53", timeout)
	return err == nil
}
