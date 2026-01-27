/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  reactStrictMode: true,
  sassOptions: {
    prependData: `@use 'sass:math'; @use 'src/scss/breakpoints' as *;`,
  },
  // Transpile workspace packages
  transpilePackages: [
    "@consensys/linea-contract-integrity-verifier",
    "@consensys/linea-contract-integrity-verifier-ethers",
    "@consensys/linea-contract-integrity-verifier-viem",
  ],
  // External packages for server components
  serverExternalPackages: ["ethers", "viem"],
};

export default nextConfig;
