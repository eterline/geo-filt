package iplocate

type RegistryOptions struct {
	enableV4 bool
	enableV6 bool
}

func defaultRegistryOptions() *RegistryOptions {
	return &RegistryOptions{
		enableV4: true,
		enableV6: true,
	}
}

func (rg *RegistryOptions) V4Enabled() bool {
	return rg.enableV4
}

func (rg *RegistryOptions) V6Enabled() bool {
	return rg.enableV6
}

func OnlyV4() func(*RegistryOptions) {
	return func(ro *RegistryOptions) {
		*ro = RegistryOptions{
			enableV4: true,
			enableV6: false,
		}
	}
}

func OnlyV6() func(*RegistryOptions) {
	return func(ro *RegistryOptions) {
		*ro = RegistryOptions{
			enableV4: false,
			enableV6: true,
		}
	}
}
