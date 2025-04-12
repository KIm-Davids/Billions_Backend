package models

type Package string

const (
	TestPackage    Package = "Test"
	ProPackage     Package = "Pro"
	PremiumPackage Package = "Premium"
)

func (pck Package) EnumIndex() string {
	return string(pck) // This now returns "Test", "Pro", or "Premium"
}
