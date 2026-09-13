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
  return round2(items.reduce((sum, item) => sum + item.unitPrice * item.qty, 0));
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

export function cartTotal(items: CartItem[], coupon?: Coupon): number {
  if (isEmpty(items)) {
    return 0;
  }
  return applyCoupon(cartSubtotal(items), coupon);
}
