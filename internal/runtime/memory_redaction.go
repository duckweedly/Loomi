package runtime

import "github.com/sheridiany/loomi/internal/productdata"

func RedactMemoryText(value string) string {
	return productdata.RedactEventText(value)
}
