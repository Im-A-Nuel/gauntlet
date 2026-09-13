export interface PricingInput {
  unitPrice: number;
  qty: number;
  bulkThreshold: number;
  bulkDiscountPct: number;
  taxRatePct: number;
}

export function round2(n: number): number {
  return Math.round(n * 100) / 100;
}

export function isEligibleForBulkDiscount(qty: number, bulkThreshold: number): boolean {
  return qty >= bulkThreshold;
}

export function unitPriceAfterDiscount(input: PricingInput): number {
  if (isEligibleForBulkDiscount(input.qty, input.bulkThreshold)) {
    return round2(input.unitPrice * (1 - input.bulkDiscountPct / 100));
  }
  return input.unitPrice;
}

export function calculateLineTotal(input: PricingInput): number {
  const discountedUnitPrice = unitPriceAfterDiscount(input);
  const subtotal = discountedUnitPrice * input.qty;
  const tax = subtotal * (input.taxRatePct / 100);
  return round2(subtotal + tax);
}
