import { describe, it, expect } from "vitest";
import { isEmpty, cartSubtotal, applyCoupon, cartTotal } from "../../src/cart";

// Deliberately shallow, see test/weak/pricing.test.ts for the rationale.

describe("isEmpty", () => {
  it("returns a boolean", () => {
    expect(typeof isEmpty([])).toBe("boolean");
  });
});

describe("cartSubtotal", () => {
  it("returns a positive number for a non-empty cart", () => {
    const subtotal = cartSubtotal([{ unitPrice: 5, qty: 2 }]);
    expect(subtotal).toBeGreaterThan(0);
  });
});

describe("applyCoupon", () => {
  it("returns a number with no coupon", () => {
    expect(typeof applyCoupon(100)).toBe("number");
  });

  it("returns a number with a percent coupon", () => {
    expect(typeof applyCoupon(100, { type: "percent", value: 10 })).toBe("number");
  });

  it("returns a number with a flat coupon", () => {
    expect(typeof applyCoupon(100, { type: "flat", value: 10 })).toBe("number");
  });
});

describe("cartTotal", () => {
  it("returns 0 for an empty cart", () => {
    expect(cartTotal([])).toBe(0);
  });

  it("returns a positive number for a non-empty cart", () => {
    expect(cartTotal([{ unitPrice: 5, qty: 2 }])).toBeGreaterThan(0);
  });
});
