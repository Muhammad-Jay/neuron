import { Service, record, string } from "@neuron/sdk";

export const enrichCustomer = Service("enrich-customer")
  .version("1.0.0")
  .description("Enrich with customer data")
  .executor("set", {
    customer_data: {
      tier: "gold",
      email: "customer@example.com",
      shipping_address: { state: "CA", city: "San Francisco", zip: "94102" },
    },
  })
  .inputSchema({
    customerId: string().required(),
  })
  .outputSchema({
    customerData: record(),
  });
