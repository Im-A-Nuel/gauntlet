# Surviving mutants (run 1789485104500-2077f42)
Trust Score: 74.5% | Threshold: 80%

## src/cart.ts
- [7] line 18 ArithmeticOperator
  original: `return round2(items.reduce((sum, item) => sum + item.unitPrice * item.qty, 0));`
  mutated:  `return round2(items.reduce((sum, item) => sum + item.unitPrice / item.qty, 0));`
  meaning: no test distinguished the mutated line from the original
- [13] line 25 ConditionalExpression
  original: `if (coupon.type === "percent") {`
  mutated:  `if (true) {`
  meaning: no test distinguished the mutated line from the original
- [15] line 25 EqualityOperator
  original: `if (coupon.type === "percent") {`
  mutated:  `if (coupon.type !== "percent") {`
  meaning: no test distinguished the mutated line from the original
- [18] line 26 ArithmeticOperator
  original: `return round2(subtotal * (1 - coupon.value / 100));`
  mutated:  `return round2(subtotal / (1 - coupon.value / 100));`
  meaning: no test distinguished the mutated line from the original
- [19] line 26 ArithmeticOperator
  original: `return round2(subtotal * (1 - coupon.value / 100));`
  mutated:  `return round2(subtotal * (1 + coupon.value / 100));`
  meaning: no test distinguished the mutated line from the original
- [22] line 28 ArithmeticOperator
  original: `return Math.max(0, round2(subtotal - coupon.value));`
  mutated:  `return Math.max(0, round2(subtotal + coupon.value));`
  meaning: no test distinguished the mutated line from the original
- [26] line 32 BlockStatement
  original: `if (isEmpty(items)) {
    return 0;
  }`
  mutated:  `if (isEmpty(items)) {}`
  meaning: no test distinguished the mutated line from the original

## src/pricing.ts
- [33] line 14 EqualityOperator
  original: `return qty >= bulkThreshold;`
  mutated:  `return qty > bulkThreshold;`
  meaning: no test distinguished the mutated line from the original
- [37] line 18 ConditionalExpression
  original: `if (isEligibleForBulkDiscount(input.qty, input.bulkThreshold)) {`
  mutated:  `if (false) {`
  meaning: no test distinguished the mutated line from the original
- [40] line 19 ArithmeticOperator
  original: `return round2(input.unitPrice * (1 - input.bulkDiscountPct / 100));`
  mutated:  `return round2(input.unitPrice * (1 + input.bulkDiscountPct / 100));`
  meaning: no test distinguished the mutated line from the original
- [44] line 27 ArithmeticOperator
  original: `const tax = subtotal * (input.taxRatePct / 100);`
  mutated:  `const tax = subtotal / (input.taxRatePct / 100);`
  meaning: no test distinguished the mutated line from the original
- [45] line 27 ArithmeticOperator
  original: `const tax = subtotal * (input.taxRatePct / 100);`
  mutated:  `const tax = subtotal * (input.taxRatePct * 100);`
  meaning: no test distinguished the mutated line from the original
