// Package cost converts API usage cost ticks into USD.
package cost

// USDPerTick converts usage cost ticks to USD (1 tick == 1e-10 USD).
const USDPerTick = 1e-10

// FromTicks returns USD and whether a cost was reported.
func FromTicks(ticks int64, present bool) (float64, bool) {
	if !present {
		return 0, false
	}
	return float64(ticks) * USDPerTick, true
}
