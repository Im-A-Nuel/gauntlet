import { describe, it, expect } from "vitest";
import {
  isEmpty,
  cartSubtotal,
  applyCoupon,
  shippingFee,
  cartTotal,
} from "../../src/cart";

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
  it("distinguishes a percentage from the same numeric flat value", () => {
    expect(applyCoupon(200, { type: "percent", value: 10 })).toBe(180);
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

describe("shippingFee", () => {
  it.each([
    [500, 0],
    [499.99, 0.99],
    [400, 0.99],
    [399.99, 1.49],
    [300, 1.49],
    [299.99, 1.99],
    [200, 1.99],
    [199.99, 2.49],
    [150, 2.49],
    [149.99, 2.99],
    [100, 2.99],
    [99.99, 4.99],
    [50, 4.99],
    [49.99, 9.99],
    [0.01, 9.99],
    [0, 0],
    [-1, 0],
  ])("returns the exact fee at subtotal %s", (subtotal, expected) => {
    expect(shippingFee(subtotal)).toBe(expected);
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
