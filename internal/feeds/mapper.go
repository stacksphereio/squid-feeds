package feeds

// CountryToRegion maps ISO 2-letter country codes to regions
var CountryToRegion = map[string]string{
	// Europe
	"GB": "EU", "FR": "EU", "DE": "EU", "ES": "EU", "IT": "EU",
	"NL": "EU", "BE": "EU", "SE": "EU", "NO": "EU", "FI": "EU",
	"PL": "EU", "IE": "EU", "PT": "EU", "AT": "EU", "CH": "EU",
	"DK": "EU", "GR": "EU", "CZ": "EU", "RO": "EU", "HU": "EU",

	// North America
	"US": "NA", "CA": "NA", "MX": "NA",

	// South America
	"BR": "SA", "AR": "SA", "CL": "SA", "PE": "SA", "CO": "SA",
	"VE": "SA", "EC": "SA", "UY": "SA", "PY": "SA", "BO": "SA",

	// Asia
	"JP": "ASIA", "CN": "ASIA", "IN": "ASIA", "SG": "ASIA",
	"HK": "ASIA", "KR": "ASIA", "TH": "ASIA", "ID": "ASIA",
	"MY": "ASIA", "PH": "ASIA", "VN": "ASIA", "TW": "ASIA",
	"AE": "ASIA", "SA": "ASIA", "IL": "ASIA", "TR": "ASIA",
}

// MapCountryToRegion returns the region for a given country code
// Returns empty string if country is not mapped
func MapCountryToRegion(countryCode string) string {
	return CountryToRegion[countryCode]
}
