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
});
