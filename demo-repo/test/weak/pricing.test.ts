import { describe, it, expect } from "vitest";
import { calculateLineTotal, unitPriceAfterDiscount, isEligibleForBulkDiscount, round2 } from "../../src/pricing";

// Deliberately shallow: exercises every branch (high line coverage) but
// asserts only loose properties, not exact boundary behavior. This is the
// "weak" baseline suite the demo's first Trust Score run measures against.

describe("round2", () => {
  it("returns a number", () => {
    expect(typeof round2(1.005)).toBe("number");
  });
});

describe("isEligibleForBulkDiscount", () => {
  it("returns a boolean", () => {
    expect(typeof isEligibleForBulkDiscount(5, 10)).toBe("boolean");
  });

  // [33] kills EqualityOperator: >= vs >; exact boundary qty === bulkThreshold must be true
  it("returns true when qty equals bulkThreshold (boundary)", () => {
    expect(isEligibleForBulkDiscount(10, 10)).toBe(true);
  });

  it("returns false when qty is one below bulkThreshold", () => {
    expect(isEligibleForBulkDiscount(9, 10)).toBe(false);
  });
});

describe("unitPriceAfterDiscount", () => {
  it("returns a positive number for a normal order", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 10,
      qty: 20,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(result).toBeGreaterThan(0);
  });

  it("returns a positive number for a small order", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 10,
      qty: 1,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(result).toBeGreaterThan(0);
  });

  // [37][40] kills ConditionalExpression (if false) and ArithmeticOperator (* (1+pct/100)):
  // eligible order must return the discounted price, not full price and not inflated price
  it("applies bulk discount and returns exact discounted unit price", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 10,
      qty: 5,
      bulkThreshold: 5,
      bulkDiscountPct: 20,
      taxRatePct: 0,
    });
    expect(result).toBe(8);
  });

  it("returns full unit price when qty is below bulk threshold", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 10,
      qty: 4,
      bulkThreshold: 5,
      bulkDiscountPct: 20,
      taxRatePct: 0,
    });
    expect(result).toBe(10);
  });
});

describe("calculateLineTotal", () => {
  it("returns a positive total", () => {
    const total = calculateLineTotal({
      unitPrice: 10,
      qty: 20,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(total).toBeGreaterThan(0);
  });

  // [44][45] kills ArithmeticOperator on tax: * (taxRatePct/100) vs / or * (taxRatePct*100)
  // unitPrice=10, qty=5, threshold=5, discount=20% -> discountedUnit=8, subtotal=40
  // tax=40*(10/100)=4, total=44
  it("computes line total with correct tax: subtotal * (taxRatePct / 100)", () => {
    const total = calculateLineTotal({
      unitPrice: 10,
      qty: 5,
      bulkThreshold: 5,
      bulkDiscountPct: 20,
      taxRatePct: 10,
    });
    expect(total).toBe(44);
  });

  it("computes line total with zero tax as exact subtotal", () => {
    const total = calculateLineTotal({
      unitPrice: 10,
      qty: 5,
      bulkThreshold: 5,
      bulkDiscountPct: 20,
      taxRatePct: 0,
    });
    expect(total).toBe(40);
  });
});
