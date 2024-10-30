package helper

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
	NGN = "NGN"
)

func IsSupported(currency string) bool {
	switch currency {
	case USD, EUR, CAD, NGN:
		return true
	}

	return false
}
