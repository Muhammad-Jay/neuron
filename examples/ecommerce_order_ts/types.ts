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

export interface SystemInput {
  order: OrderInput;
}

export interface ValidateOrderInput {
  order: OrderInput;
}

export interface ValidateOrderOutput {
  valid: boolean;
  status: string;
  order: OrderInput;
}

export interface ParseOrderInput {
  validationData: unknown;
}

export interface ParseOrderOutput {
  order: OrderInput;
}

export interface EnrichCustomerInput {
  customerId: string;
}

export interface EnrichCustomerOutput {
  customerData: {
    tier: string;
    email: string;
    shippingAddress: Address;
  };
  order: OrderInput;
}

export interface CalculateTotalsInput {
  items: OrderItem[];
  customerTier: string;
  shippingState: string;
  email: string;
}

export interface CalculateTotalsOutput {
  totalCents: number;
  taxCents: number;
  order: OrderInput;
}

export interface AuthorizePaymentInput {
  amountCents: number;
  currency: string;
  email: string;
}

export interface AuthorizePaymentOutput {
  paymentIntent: {
    id: string;
    status: string;
  };
  order: OrderInput;
}

export interface CapturePaymentInput {
  paymentIntentId: string;
}

export interface CapturePaymentOutput {
  captureResult: {
    status: string;
    amountReceived: number;
  };
  order: OrderInput;
}

export interface CreateShipmentInput {
  order: OrderInput;
  shippingAddress: Address;
  email: string;
}

export interface CreateShipmentOutput {
  shipment: {
    trackingNumber: string;
    carrier: string;
    labelUrl: string;
  };
  order: OrderInput;
}

export interface SendConfirmationInput {
  trackingNumber: string;
  carrier: string;
  email: string;
  grandTotal: number;
}

export interface SendConfirmationOutput {
  confirmationSent: boolean;
  messageId: string;
}
