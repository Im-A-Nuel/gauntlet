import { describe, it, expect } from "vitest";
import { isEmpty, cartSubtotal, applyCoupon, cartTotal } from "../../src/cart";

// Authored by Claude (Sonnet 5) — see test/strong/pricing.test.ts header for
// why these exist and what they are (and are not).

describe("isEmpty", () => {
  it("is true for an empty array", () => {
    expect(isEmpty([])).toBe(true);
  });
  it("is false for a single item", () => {
    expect(isEmpty([{ unitPrice: 1, qty: 1 }])).toBe(false);
  });
});

describe("cartSubtotal", () => {
  it("sums unitPrice times qty across items", () => {
    expect(
      cartSubtotal([
        { unitPrice: 5, qty: 2 },
        { unitPrice: 3, qty: 4 },
      ]),
    ).toBe(22);
  });
  it("is 0 for an empty cart", () => {
    expect(cartSubtotal([])).toBe(0);
  });
});

describe("applyCoupon", () => {
  it("returns the subtotal unchanged with no coupon", () => {
    expect(applyCoupon(100, undefined)).toBe(100);
  });
  it("applies an exact percent discount", () => {
    expect(applyCoupon(100, { type: "percent", value: 10 })).toBe(90);
  });
  it("applies an exact flat discount", () => {
    expect(applyCoupon(100, { type: "flat", value: 30 })).toBe(70);
  });
  it("clamps a flat discount larger than the subtotal to exactly 0", () => {
    expect(applyCoupon(20, { type: "flat", value: 50 })).toBe(0);
  });
  it("does not go negative at the exact clamp boundary", () => {
    expect(applyCoupon(50, { type: "flat", value: 50 })).toBe(0);
  });
});

describe("cartTotal", () => {
  it("is exactly 0 for an empty cart regardless of coupon", () => {
    expect(cartTotal([], { type: "percent", value: 50 })).toBe(0);
  });
  it("applies the coupon to a non-empty cart's subtotal", () => {
    expect(
      cartTotal([{ unitPrice: 10, qty: 10 }], {
        type: "percent",
        value: 20,
      }),
    ).toBe(84.99);
  });
});
