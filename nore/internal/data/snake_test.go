package data

import (
	"reflect"
	"testing"
)

func TestDeepSnakeCase(t *testing.T) {
	in := map[string]any{
		"customerId":    "c1",
		"customerEmail": "a@b.c",
		"shippingAddress": map[string]any{
			"city": "SF", "zipCode": 94102,
		},
		"items": []any{
			map[string]any{"sku": "s", "priceCents": 100},
		},
		"already_snake": true,
	}
	want := map[string]any{
		"customer_id":    "c1",
		"customer_email": "a@b.c",
		"shipping_address": map[string]any{
			"city": "SF", "zip_code": 94102,
		},
		"items": []any{
			map[string]any{"sku": "s", "price_cents": 100},
		},
		"already_snake": true,
	}
	got := DeepSnakeCase(in)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DeepSnakeCase\ngot:  %#v\nwant: %#v", got, want)
	}

	// The original input must not be mutated.
	if in["customerId"] != "c1" {
		t.Errorf("DeepSnakeCase mutated its input")
	}
}

func TestSnakeMapIdempotent(t *testing.T) {
	in := map[string]any{"order": map[string]any{"customer_id": "c1"}}
	if !reflect.DeepEqual(SnakeMap(in), SnakeMap(SnakeMap(in))) {
		t.Errorf("SnakeMap should be idempotent")
	}
}
