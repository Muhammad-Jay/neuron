import { Service, record, string } from "@neuron/sdk";

export const createShipment = Service("create-shipment")
  .version("1.0.0")
  .description("Create shipment")
  .executor("set", {
    shipment: {
      tracking_number: "1Z999AA10123456784",
      carrier: "UPS",
      label_url: "https://shipping.example.com/label/1Z999AA10123456784",
    },
  })
  .inputSchema({
    order: record().required(),
    shippingAddress: record().required(),
    email: string().email(),
  })
  .outputSchema({
    shipment: record(),
  });
