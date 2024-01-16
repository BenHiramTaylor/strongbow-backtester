package utils

var (
	// TradeDirection is an equivalent to an enum for a trade direction, but without the enum as i hate enums in Go
	TradeDirection = tradeDirection{LONG: "LONG", SHORT: "SHORT"}
)

type tradeDirection struct {
	LONG  string
	SHORT string
}
