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
	result := new(VendorManager)

	return &VendorManager{
		aem: aem,

		javaManager: NewJavaManager(result),
		vaultCLI:    NewVaultCLI(result),
		quickstart:  NewQuickstart(result),
		oakRun:      NewOakRun(result),
		sdk:         NewSDK(result),
	}
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

// TODO make this method idempotent
func (vm *VendorManager) Prepare() error {
	// validation phase (quick feedback)
	sdk, err := vm.quickstart.IsDistSDK()
	if err != nil {
		return err
	}
	// preparation phase (slow feedback)
	if err := vm.aem.BaseOpts().Prepare(); err != nil {
		return err
	}
	if err := vm.javaManager.Prepare(); err != nil {
		return err
	}
	if sdk {
		if err := vm.sdk.Prepare(); err != nil {
			return err
		}
	}
	if err := vm.oakRun.Prepare(); err != nil {
		return err
	}
	if err := vm.vaultCLI.Prepare(); err != nil {
		return err
	}
	return nil
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
