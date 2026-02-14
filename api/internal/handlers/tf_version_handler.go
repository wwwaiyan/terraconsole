package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/terraconsole/api/internal/config"
)

type TFVersionHandler struct {
	cfg *config.Config
}

func NewTFVersionHandler(cfg *config.Config) *TFVersionHandler {
	return &TFVersionHandler{cfg: cfg}
}

type TFVersion struct {
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
}

// ListVersions returns available terraform versions from HashiCorp releases
func (h *TFVersionHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://releases.hashicorp.com/terraform/index.json")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch versions"})
		return
	}
	defer resp.Body.Close()

	var index struct {
		Versions map[string]struct {
			Version string `json:"version"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to parse versions"})
		return
	}

	installed := h.getInstalledVersions()

	var versions []TFVersion
	for v := range index.Versions {
		if strings.Contains(v, "-") {
			continue
		}
		_, isInstalled := installed[v]
		versions = append(versions, TFVersion{
			Version:   v,
			Installed: isInstalled,
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	if len(versions) > 50 {
		versions = versions[:50]
	}

	writeJSON(w, http.StatusOK, versions)
}

// ListInstalledVersions returns locally installed terraform versions
func (h *TFVersionHandler) ListInstalledVersions(w http.ResponseWriter, r *http.Request) {
	installed := h.getInstalledVersions()

	var versions []TFVersion
	for v, path := range installed {
		versions = append(versions, TFVersion{
			Version:   v,
			Installed: true,
			Path:      path,
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	writeJSON(w, http.StatusOK, versions)
}

// InstallVersion downloads and installs a specific terraform version
func (h *TFVersionHandler) InstallVersion(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")

	versionDir := filepath.Join(h.cfg.TerraformDir, version)
	binaryPath := filepath.Join(versionDir, "terraform")

	if _, err := os.Stat(binaryPath); err == nil {
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Version already installed",
			"path":    binaryPath,
			"version": version,
		})
		return
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH

	url := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip",
		version, version, osName, arch)

	resp, err := http.Get(url)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to download terraform"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Version %s not found", version)})
		return
	}

	if err := os.MkdirAll(versionDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create directory"})
		return
	}

	zipPath := filepath.Join(versionDir, "terraform.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save download"})
		return
	}
	io.Copy(zipFile, resp.Body)
	zipFile.Close()

	if err := unzipFile(zipPath, versionDir); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to extract terraform"})
		return
	}

	os.Chmod(binaryPath, 0755)
	os.Remove(zipPath)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Version installed successfully",
		"path":    binaryPath,
		"version": version,
	})
}

func (h *TFVersionHandler) getInstalledVersions() map[string]string {
	installed := make(map[string]string)

	entries, err := os.ReadDir(h.cfg.TerraformDir)
	if err != nil {
		return installed
	}

	for _, entry := range entries {
		if entry.IsDir() {
			binaryPath := filepath.Join(h.cfg.TerraformDir, entry.Name(), "terraform")
			if _, err := os.Stat(binaryPath); err == nil {
				installed[entry.Name()] = binaryPath
			}
		}
	}

	return installed
}

func unzipFile(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		destPath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}
