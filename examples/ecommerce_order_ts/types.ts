// Shared types for the order-processing pipeline.

export interface Address {
  street: string;
  city: string;
  state: string;
  zip: string;
}

export interface OrderItem {
  sku: string;
  name: string;
  qty: number;
  priceCents: number;
}

export interface OrderInput {
  id: string;
  customerId: string;
  customerEmail: string;
  currency: string;
  items: OrderItem[];
  total: number;
  shippingAddress: Address;
}
