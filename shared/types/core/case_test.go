package core

import "testing"

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"customerId", "customer_id"},
		{"customerEmail", "customer_email"},
		{"amountCents", "amount_cents"},
		{"shippingAddress", "shipping_address"},
		{"validationData", "validation_data"},
		{"grandTotal", "grand_total"},
		{"JSONData", "json_data"},
		{"URL", "url"},
		{"id", "id"},
		{"order_id", "order_id"},
		{"priceCents", "price_cents"},
		{"_id", "_id"},
	}
	for _, tt := range tests {
		if got := CamelToSnake(tt.in); got != tt.want {
			t.Errorf("CamelToSnake(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
