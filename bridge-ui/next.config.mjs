/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  reactStrictMode: true,
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**",
      },
    ],
  },
  sassOptions: {
    prependData: `@use 'sass:math'; @import 'src/scss/breakpoints';`,
  },
  webpack: (config) => {
    const warning = [...(config.ignoreWarnings || []), { module: /typeorm/ }];

    config.ignoreWarnings = warning;

    config.resolve.fallback = {
      fs: false,
    };
    config.externals.push("pino-pretty", "lokijs", "encoding");

    const fileLoaderRule = config.module.rules.find((rule) => rule.test?.test?.(".svg"));

    config.module.rules.push(
      {
        ...fileLoaderRule,
        test: /\.svg$/i,
        resourceQuery: /url/,
      },
      {
        test: /\.svg$/i,
        issuer: fileLoaderRule.issuer,
        resourceQuery: { not: [...fileLoaderRule.resourceQuery.not, /url/] },
        use: ["@svgr/webpack"],
      },
    );

    fileLoaderRule.exclude = /\.svg$/i;

    return config;
  },
  async headers() {
    /**
     * Mitigate XSS and data injection attack.
     * Restrict what content the browser is allowed to load.
     * - default-src 'self' -> Default policy to only allow resources from same origin
     * - script-src -> Ideally not https: as this allows all HTTPS sources, but it turns out that Google Tag Manager uses many domains to load scripts
     */
    const cspHeader = `
      default-src 'self';
      connect-src 'self' https:;
      script-src 'self' 'unsafe-eval' 'unsafe-inline' https:;
      style-src 'self' 'unsafe-inline';
      img-src 'self' blob: data: https:;
      font-src 'self' data: https://cdn.jsdelivr.net;
      object-src 'none';
      base-uri 'self';
      form-action 'self';
      frame-ancestors 'none';
      upgrade-insecure-requests;
    `;
    return [
      {
        source: "/(.*)",
        headers: [
          {
            key: "Content-Security-Policy",
            value: cspHeader.replace(/\n/g, ""),
          },
        ],
      },
    ];
  },
};

export default nextConfig;
