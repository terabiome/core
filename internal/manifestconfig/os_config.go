package manifestconfig

import "errors"

type FirewalldTargetAction string

const (
	FirewalldTargetActionDrop    FirewalldTargetAction = "DROP"
	FirewalldTargetActionReject  FirewalldTargetAction = "REJECT"
	FirewalldTargetActionAccept  FirewalldTargetAction = "ACCEPT"
	FirewalldTargetActionUnknown FirewalldTargetAction = ""
)

type FirewalldZoneEntityAction string

const (
	FirewalldZoneEntityActionAdd     FirewalldZoneEntityAction = "add"
	FirewalldZoneEntityActionRemove  FirewalldZoneEntityAction = "remove"
	FirewalldZoneEntityActionUnknown FirewalldZoneEntityAction = ""
)

type KernelModuleAction string

const (
	KernelModuleActionAdd    KernelModuleAction = "add"
	KernelModuleActionRemove KernelModuleAction = "remove"
)

type SystemKernelParameterAction string

const (
	SystemKernelParameterActionAdd    SystemKernelParameterAction = "add"
	SystemKernelParameterActionRemove SystemKernelParameterAction = "remove"
)

// done
type OSConfig struct {
	Shell   ShellConfig   `json:"shell" yaml:"shell"`
	Network NetworkConfig `json:"network" yaml:"network"`
	Disk    DiskConfig    `json:"swap" yaml:"swap"`
	Kernel  KernelConfig  `json:"kernel" yaml:"kernel"`
}

// done - both enable + disable appear -> conflict, throw error
type ShellConfig struct {
	EnableService  bool `json:"enable_service,omitempty" yaml:"enable_service,omitempty"`
	DisableService bool `json:"disable_service,omitempty" yaml:"disable_service,omitempty"`
	StartService   bool `json:"start_service" yaml:"start_service"`
}

func (cfg ShellConfig) SoftValidate() error {
	if cfg.EnableService == true && cfg.DisableService == true {
		return errors.New("shell config: both disable and enable service are present")
	}

	return nil
}

// done
type NetworkConfig struct {
	Tailscale TailscaleConfig `json:"tailscale" yaml:"tailscale"`
	Firewall  FirewallConfig  `json:"firewall" yaml:"firewall"`
}

// done - both enable + disable appear -> conflict, throw error
type TailscaleConfig struct {
	InstallPackage bool   `json:"install_package" yaml:"install_package"`
	EnableService  bool   `json:"enable_service,omitempty" yaml:"enable_service,omitempty"`
	DisableService bool   `json:"disable_service,omitempty" yaml:"disable_service,omitempty"`
	StartService   bool   `json:"start_service" yaml:"start_service"`
	AuthKey        string `json:"auth_key" yaml:"auth_key"`
}

func (cfg TailscaleConfig) SoftValidate() error {
	if cfg.EnableService == true && cfg.DisableService == true {
		return errors.New("tailscale config: both disable and enable service are present")
	}

	return nil
}

// done - both enable + disable appear -> conflict, throw error
type FirewallConfig struct {
	EnableService  bool                  `json:"enable_service,omitempty" yaml:"enable_service,omitempty"`
	DisableService bool                  `json:"disable_service,omitempty" yaml:"disable_service,omitempty"`
	Immediate      bool                  `json:"immediate" yaml:"immediate"`
	Zones          map[string]ZoneConfig `json:"zones" yaml:"zones"`
}

func (cfg FirewallConfig) SoftValidate() error {
	if cfg.EnableService == true && cfg.DisableService == true {
		return errors.New("firewall config: both disable and enable service are present")
	}

	return nil
}

// done
type ZoneConfig struct {
	Interface InterfaceConfig       `json:"interface" yaml:"interface"`
	Services  []ServiceConfig       `json:"services" yaml:"services"`
	SetTarget FirewalldTargetAction `json:"set_service,omitempty" yaml:"set_service,omitempty"`
}

// done
type InterfaceConfig struct {
	Name   string                    `json:"name" yaml:"name"`
	Action FirewalldZoneEntityAction `json:"action,omitempty" yaml:"action,omitempty"`
}

// done
type ServiceConfig struct {
	Name   string                    `json:"name" yaml:"name"`
	Action FirewalldZoneEntityAction `json:"action,omitempty" yaml:"action,omitempty"`
}

// done
type DiskConfig struct {
	Swap SwapConfig `json:"swap" yaml:"swap"`
}

// done
type SwapConfig struct {
	DisableSwap             bool `json:"disable" yaml:"disable"`
	EditFstab               bool `json:"edit_fstab" yaml:"edit_fstab"`
	MaskSwapRelatedServices bool `json:"mask_swap_related_services" yaml:"mask_swap_related_services"`
}

// done
type KernelConfig struct {
	KernelModules          KernelModulesConfig          `json:"kernel_modules" yaml:"kernel_modules"`
	SystemKernelParameters SystemKernelParametersConfig `json:"system_kernel_parameters" yaml:"system_kernel_parameters"`
}

/*
Can refactor Kernel config to this and make things easier. Low priority though.
  kernel:
    kernel_modules:
      immediate: true
      mappings:
        - path: /etc/modules-load.d/k8s.conf
          values:
            - value: br_netfilter
              action: add
            - value: overlay
              action: add
*/

// done
type KernelModulesConfig struct {
	Immediate bool                 `json:"immediate" yaml:"immediate"`
	Values    []KernelModuleConfig `json:"values" yaml:"values"`
}

// done
type KernelModuleConfig struct {
	FilePath string             `json:"path" yaml:"path"` // e.g. /etc/modules-load.d/*.conf
	Action   KernelModuleAction `json:"action" yaml:"action"`
	Value    string             `json:"value" yaml:"value"`
}

// done
type SystemKernelParametersConfig struct {
	Immediate bool                          `json:"immediate" yaml:"immediate"`
	Values    []SystemKernelParameterConfig `json:"values" yaml:"values"`
}

// done
type SystemKernelParameterConfig struct {
	FilePath string                      `json:"path" yaml:"path"` // e.g. /etc/sysctl.d/*.conf
	Action   SystemKernelParameterAction `json:"action" yaml:"action"`
	Key      string                      `json:"key" yaml:"key"`
	Value    string                      `json:"value" yaml:"value"`
}
