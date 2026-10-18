package k3s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/terabiome/internal/operations"
)

type InstallBinaryOperation struct {
	rawParameters map[string]any
	parameters    installBinaryParameters
}

// Name implements [operations.Operation].
func (op *InstallBinaryOperation) Name() string {
	return string(operations.OperationNameK3sInstallBinary)
}

// PreCheck implements [operations.Operation].
func (op *InstallBinaryOperation) PreCheck(env *operations.OperationEnvironment) error {
	if op.parameters.Version == "" {
		return fmt.Errorf("missing k3s version")
	}
	return nil
}

// Run implements [operations.Operation].
func (op *InstallBinaryOperation) Run(ctx context.Context, env *operations.OperationEnvironment) error {
	arch := runtime.GOARCH
	switch arch {
	case "arm":
		arch = "armhf" // K3s uses armhf for 32-bit ARM
	case "aarch64":
		arch = "arm64"
	}

	url := fmt.Sprintf("%s/%s/k3s-%s", defaultK3sInstallBaseURL, op.parameters.Version, arch)

	fmt.Fprintf(env.Journaler, "Downloading K3s binary (%s) from %s\n", arch, url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download k3s binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download k3s binary: HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(defaultK3sBinaryPath), "k3s-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Cleanup if we fail

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save binary content: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), defaultK3sBinaryPath); err != nil {
		return fmt.Errorf("failed to install binary to %s: %w", defaultK3sBinaryPath, err)
	}

	fmt.Fprintf(env.Journaler, "K3s v%s binary installed at %s\n", op.parameters.Version, defaultK3sBinaryPath)
	return nil
}

// SetParameters implements [operations.Operation].
func (op *InstallBinaryOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewInstallBinaryOperation() operations.Operation {
	return &InstallBinaryOperation{}
}
