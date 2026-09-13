import type { NextConfig } from 'next';
const config: NextConfig = {
  outputFileTracingIncludes: { '/*': ['./sample-runs/**/*.json'] },
  poweredByHeader: false,
};
export default config;
