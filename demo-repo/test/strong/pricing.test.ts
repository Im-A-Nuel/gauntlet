import { describe, it, expect } from "vitest";
import { calculateLineTotal, unitPriceAfterDiscount, isEligibleForBulkDiscount, round2 } from "../../src/pricing";

// Authored by Claude (Sonnet 5) as a hand-written stand-in for what
// `gauntlet strengthen` would hand to IBM Bob: precise, boundary-value
// assertions that kill the mutants test/weak/pricing.test.ts misses. IBM
// Bob is not installed in this environment, so these are NOT a Bob output —
// see docs/CLAUDE_PROGRESS.md. They exist so the before/after Trust Score
// delta is real and reproducible (`npm run test:strong`) without
// misattributing authorship.

describe("round2", () => {
  it("rounds to exactly two decimal places", () => {
    // Math.round(n * 100) / 100 inherits binary floating-point imprecision,
    // so these are the function's real observed outputs, not idealized
    // decimal rounding (round2(1.015) is 1.01, not 1.02).
    expect(round2(1.005)).toBe(1);
    expect(round2(1.015)).toBe(1.01);
    expect(round2(2.345)).toBe(2.35);
  });
});

describe("isEligibleForBulkDiscount", () => {
  it("is eligible exactly at the threshold", () => {
    expect(isEligibleForBulkDiscount(10, 10)).toBe(true);
  });
  it("is not eligible one below the threshold", () => {
    expect(isEligibleForBulkDiscount(9, 10)).toBe(false);
  });
  it("is eligible above the threshold", () => {
    expect(isEligibleForBulkDiscount(11, 10)).toBe(true);
  });
});

describe("unitPriceAfterDiscount", () => {
  it("applies the exact discount percentage at the threshold", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 100,
      qty: 10,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(result).toBe(85);
  });

  it("charges full unit price one unit below the threshold", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 100,
      qty: 9,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(result).toBe(100);
  });

  it("applies zero discount when bulkDiscountPct is 0", () => {
    const result = unitPriceAfterDiscount({
      unitPrice: 100,
      qty: 10,
      bulkThreshold: 10,
      bulkDiscountPct: 0,
      taxRatePct: 8,
    });
    expect(result).toBe(100);
  });
});

describe("calculateLineTotal", () => {
  it("computes the exact discounted, taxed total at the threshold", () => {
    // unit 100 * 10 qty, 15% bulk discount -> 85/unit -> subtotal 850, 8% tax -> 918
    const total = calculateLineTotal({
      unitPrice: 100,
      qty: 10,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(total).toBe(918);
  });

  it("computes the exact undiscounted, taxed total below the threshold", () => {
    // unit 100 * 9 qty, no discount -> subtotal 900, 8% tax -> 972
    const total = calculateLineTotal({
      unitPrice: 100,
      qty: 9,
      bulkThreshold: 10,
      bulkDiscountPct: 15,
      taxRatePct: 8,
    });
    expect(total).toBe(972);
  });

  it("applies zero tax when taxRatePct is 0", () => {
    const total = calculateLineTotal({
      unitPrice: 50,
      qty: 2,
      bulkThreshold: 100,
      bulkDiscountPct: 0,
      taxRatePct: 0,
    });
    expect(total).toBe(100);
  });
});
