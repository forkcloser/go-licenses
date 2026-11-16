package licenses

import "slices"

type LicenseRestrictiveness string

const (
	RestrictionsShareLicense LicenseRestrictiveness = "ShareLicense"
	RestrictionsShareCode    LicenseRestrictiveness = "ShareCode"
	RestrictionsUnknown      LicenseRestrictiveness = "Unknown"
	RestrictionsNotAllowed   LicenseRestrictiveness = "NotAllowed"
)

// LicenseTypeRestrictiveness is what the license types ask together: the
// most demanding of what each asks, and RestrictionsUnknown for none.
func LicenseTypeRestrictiveness(licenseTypes ...Type) LicenseRestrictiveness {
	if len(licenseTypes) == 0 {
		return RestrictionsUnknown
	}

	// The most demanding first.
	for _, asked := range []LicenseRestrictiveness{RestrictionsNotAllowed, RestrictionsUnknown, RestrictionsShareCode} {
		if slices.ContainsFunc(
			licenseTypes,
			func(licenseType Type) bool { return restrictivenessOf(licenseType) == asked },
		) {
			return asked
		}
	}

	return RestrictionsShareLicense
}

// restrictivenessOf is what one license type asks.
func restrictivenessOf(licenseType Type) LicenseRestrictiveness {
	switch licenseType {
	case Notice, Permissive, Unencumbered:
		return RestrictionsShareLicense
	case Restricted, Reciprocal:
		return RestrictionsShareCode
	case Unknown:
		return RestrictionsUnknown
	case ByExceptionOnly, Forbidden:
		return RestrictionsNotAllowed
	}

	// A type that is none of the constants.
	return RestrictionsNotAllowed
}
