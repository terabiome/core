package k3s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-viper/mapstructure/v2"
	"github.com/terabiome/core/pkg/helpers"
	operationspkg "github.com/terabiome/core/pkg/operations"
)

const maxK3sBinaryBytes = 256 << 20

type InstallBinaryOperation struct {
	rawParameters map[string]any
	parameters    installBinaryParameters
}

func (op *InstallBinaryOperation) Name() string {
	return NameInstallBinary
}

func (op *InstallBinaryOperation) PreCheck(env *operationspkg.OperationEnvironment) error {
	if op.parameters.Version == "" {
		return fmt.Errorf("missing k3s version")
	}
	if _, err := k3sArch(env.NodeInfo.Arch); err != nil {
		return err
	}
	return nil
}

func (op *InstallBinaryOperation) Run(ctx context.Context, env *operationspkg.OperationEnvironment) error {
	arch, err := k3sArch(env.NodeInfo.Arch)
	if err != nil {
		return err
	}
	binaryName := "k3s"
	if arch != "amd64" {
		binaryName = fmt.Sprintf("%s-%s", binaryName, arch)
	}

	base := fmt.Sprintf("%s/%s", defaultK3sInstallBaseURL, op.parameters.Version)
	binaryURL := base + "/" + binaryName
	sumURL := fmt.Sprintf("%s/sha256sum-%s.txt", base, arch)

	fmt.Fprintf(env.Journaler, "Downloading K3s binary (%s) from %s\n", arch, binaryURL)

	sumFile, err := helpers.HTTP().Get(ctx, sumURL, 1<<20)
	if err != nil {
		return fmt.Errorf("failed to download k3s checksum: %w", err)
	}
	want, err := helpers.File().Checksum(sumFile, binaryName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(defaultK3sBinaryPath), 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(defaultK3sBinaryPath), err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(defaultK3sBinaryPath), "k3s-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := helpers.HTTP().DownloadTo(ctx, binaryURL, tmpFile, maxK3sBinaryBytes); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download k3s binary: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	got, err := helpers.File().GetSHA256(tmpFile.Name())
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("k3s binary checksum mismatch")
	}

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	if err := os.Rename(tmpFile.Name(), defaultK3sBinaryPath); err != nil {
		return fmt.Errorf("failed to install binary to %s: %w", defaultK3sBinaryPath, err)
	}

	fmt.Fprintf(env.Journaler, "K3s %s binary installed at %s\n", op.parameters.Version, defaultK3sBinaryPath)
	return nil
}

func (op *InstallBinaryOperation) SetParameters(parameters map[string]any) error {
	op.rawParameters = parameters
	return mapstructure.Decode(op.rawParameters, &op.parameters)
}

func NewInstallBinaryOperation() operationspkg.Operation {
	return &InstallBinaryOperation{}
}
