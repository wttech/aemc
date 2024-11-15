package pkg

// VendorManager manages third-party tools like JDK, OakRun, and Vault CLI
type VendorManager struct {
	aem *AEM

	javaManager *JavaManager
	oakRun      *OakRun
	quickstart  *Quickstart
	sdk         *SDK
	vaultCLI    *VaultCLI
}

func NewVendorManager(aem *AEM) *VendorManager {
	result := &VendorManager{aem: aem}
	result.javaManager = NewJavaManager(result)
	result.sdk = NewSDK(result)
	result.quickstart = NewQuickstart(result)
	result.oakRun = NewOakRun(result)
	result.vaultCLI = NewVaultCLI(result)
	return result
}

func (vm *VendorManager) InstanceJar() (string, error) {
	sdk, err := vm.quickstart.IsDistSDK()
	if err != nil {
		return "", err
	}
	if sdk {
		return vm.sdk.QuickstartJar()
	}
	return vm.quickstart.FindDistFile()
}

func (vm *VendorManager) PrepareWithChanged() (bool, error) {
	changed := false

	javaChanged, err := vm.javaManager.PrepareWithChanged()
	changed = changed || javaChanged
	if err != nil {
		return changed, err
	}

	sdk, err := vm.quickstart.IsDistSDK()
	if err != nil {
		return false, err
	}
	if sdk {
		sdkChanged, err := vm.sdk.PrepareWithChanged()
		changed = changed || sdkChanged
		if err != nil {
			return changed, err
		}
	}

	oakRunChanged, err := vm.oakRun.PrepareWithChanged()
	changed = changed || oakRunChanged
	if err != nil {
		return changed, err
	}

	vaultCLIChanged, err := vm.vaultCLI.PrepareWithChanged()
	changed = changed || vaultCLIChanged
	if err != nil {
		return changed, err
	}

	return changed, nil
}

func (vm *VendorManager) JavaManager() *JavaManager {
	return vm.javaManager
}

func (vm *VendorManager) OakRun() *OakRun {
	return vm.oakRun
}

func (vm *VendorManager) Quickstart() *Quickstart {
	return vm.quickstart
}

func (vm *VendorManager) SDK() *SDK {
	return vm.sdk
}

func (vm *VendorManager) VaultCLI() *VaultCLI {
	return vm.vaultCLI
}
