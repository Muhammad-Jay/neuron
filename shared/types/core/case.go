package core

import "unicode"

// CamelToSnake converts camelCase and PascalCase identifiers to lower
// snake_case while preserving existing underscores and digits, so already
// canonical identifiers pass through unchanged.
//
//	CamelToSnake("customerId")   // "customer_id"
//	CamelToSnake("JSONData")     // "json_data"
//	CamelToSnake("order_id")     // "order_id"
//	CamelToSnake("URL")          // "url"
func CamelToSnake(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes)+len(runes)/3)
	for i, r := range runes {
		switch {
		case r == '_' || unicode.IsDigit(r):
			out = append(out, r)
		case unicode.IsUpper(r):
			if i > 0 {
				prev := runes[i-1]
				boundary := isLowerOrDigit(prev) ||
					(unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]))
				if boundary {
					out = append(out, '_')
				}
			}
			out = append(out, unicode.ToLower(r))
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

func isLowerOrDigit(r rune) bool {
	return unicode.IsLower(r) || unicode.IsDigit(r)
}
