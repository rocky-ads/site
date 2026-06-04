package location

import (
	"net/url"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

const (
	UnitMiles = "mi"
	UnitKm    = "km"
)

var mileRegions = map[string]struct{}{
	"US": {},
	"GB": {},
	"LR": {},
	"MM": {},
}

var canadaTimezones = map[string]struct{}{
	"America/Toronto":      {},
	"America/Vancouver":    {},
	"America/Winnipeg":     {},
	"America/Halifax":      {},
	"America/St_Johns":     {},
	"America/Edmonton":     {},
	"America/Regina":       {},
	"America/Whitehorse":   {},
	"America/Yellowknife":  {},
	"America/Iqaluit":      {},
	"America/Moncton":      {},
	"America/Glace_Bay":    {},
	"America/Goose_Bay":    {},
}

// DistanceUnitFromPhone returns mi or km based on the phone number region.
func DistanceUnitFromPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return UnitMiles
	}
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return UnitMiles
	}
	region := phonenumbers.GetRegionCodeForNumber(num)
	if usesMilesRegion(region) {
		return UnitMiles
	}
	return UnitKm
}

// DistanceUnitFromTimezone returns mi or km from an IANA timezone name.
func DistanceUnitFromTimezone(timezone string) string {
	tz := decodeTimezone(timezone)
	if tz == "Europe/London" {
		return UnitMiles
	}
	if strings.HasPrefix(tz, "America/") {
		if _, canada := canadaTimezones[tz]; !canada {
			return UnitMiles
		}
	}
	return UnitKm
}

func usesMilesRegion(region string) bool {
	region = strings.ToUpper(strings.TrimSpace(region))
	_, ok := mileRegions[region]
	return ok
}

func decodeTimezone(timezone string) string {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return ""
	}
	decoded, err := url.QueryUnescape(timezone)
	if err != nil {
		return timezone
	}
	return decoded
}
