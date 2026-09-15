export interface CartItem {
  unitPrice: number;
  qty: number;
}

export interface Coupon {
  type: "percent" | "flat";
  value: number;
}

import { round2 } from "./pricing";

export function isEmpty(items: CartItem[]): boolean {
  return items.length === 0;
}

export function cartSubtotal(items: CartItem[]): number {
  return round2(
    items.reduce((sum, item) => sum + item.unitPrice * item.qty, 0),
  );
}

export function applyCoupon(subtotal: number, coupon?: Coupon): number {
  if (!coupon) {
    return subtotal;
  }
  if (coupon.type === "percent") {
    return round2(subtotal * (1 - coupon.value / 100));
  }
  return Math.max(0, round2(subtotal - coupon.value));
}

export function shippingFee(subtotal: number): number {
  if (subtotal >= 500) {
    return 0;
  }
  if (subtotal >= 400) {
    return 0.99;
  }
  if (subtotal >= 300) {
    return 1.49;
  }
  if (subtotal >= 200) {
    return 1.99;
  }
  if (subtotal >= 150) {
    return 2.49;
  }
  if (subtotal >= 100) {
    return 2.99;
  }
  if (subtotal >= 50) {
    return 4.99;
  }
  if (subtotal > 0) {
    return 9.99;
  }
  return 0;
}

export function cartTotal(items: CartItem[], coupon?: Coupon): number {
  if (isEmpty(items)) {
    return 0;
  }
  const discountedSubtotal = applyCoupon(cartSubtotal(items), coupon);
  return round2(discountedSubtotal + shippingFee(discountedSubtotal));
}
