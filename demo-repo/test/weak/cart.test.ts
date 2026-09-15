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

  // [7] kills ArithmeticOperator: unitPrice * qty vs / qty
  it("multiplies unitPrice by qty, not divides", () => {
    expect(cartSubtotal([{ unitPrice: 5, qty: 3 }])).toBe(15);
  });

  it("sums multiple items correctly", () => {
    expect(cartSubtotal([{ unitPrice: 4, qty: 2 }, { unitPrice: 3, qty: 1 }])).toBe(11);
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

  // [18][19] kills ArithmeticOperator on percent formula: * (1 - v/100) vs / or * (1 + v/100)
  it("percent coupon reduces subtotal by exact percentage", () => {
    expect(applyCoupon(100, { type: "percent", value: 20 })).toBe(80);
  });

  // [13][15] kills ConditionalExpression/EqualityOperator: flat coupon must not use percent branch
  // [22] kills ArithmeticOperator on flat formula: subtotal - value vs + value
  it("flat coupon subtracts exact amount, not a percentage", () => {
    expect(applyCoupon(100, { type: "flat", value: 15 })).toBe(85);
  });

  it("flat coupon result differs from same-value percent coupon", () => {
    const flat = applyCoupon(200, { type: "flat", value: 50 });
    const pct  = applyCoupon(200, { type: "percent", value: 50 });
    expect(flat).toBe(150);
    expect(pct).toBe(100);
    expect(flat).not.toBe(pct);
  });
});

describe("cartTotal", () => {
  it("returns 0 for an empty cart", () => {
    expect(cartTotal([])).toBe(0);
  });

  it("returns a positive number for a non-empty cart", () => {
    expect(cartTotal([{ unitPrice: 5, qty: 2 }])).toBeGreaterThan(0);
  });

  // [26] kills BlockStatement removal: empty cart with coupon must still return 0, not pass-through
  it("returns 0 for empty cart even when a coupon is supplied", () => {
    expect(cartTotal([], { type: "flat", value: 5 })).toBe(0);
  });
});
