const isProd = process.env.NEXT_PUBLIC_ENVIRONMENT === "production";
const basePath = isProd ? "/hub/bridge" : "";

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  reactStrictMode: true,
  basePath,
  env: {
    NEXT_PUBLIC_BASE_PATH: basePath,
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "**",
      },
    ],
  },
  turbopack: {
    rules: {
      "*.svg": {
        loaders: ["@svgr/webpack"],
        as: "*.js",
      },
    },
  },
  sassOptions: {
    includePaths: ["src/scss"],
    prependData: `@use 'sass:math'; @import 'breakpoints';`,
  },
  webpack: (config) => {
    config.ignoreWarnings = [...(config.ignoreWarnings || []), { module: /typeorm/ }];

    config.resolve.fallback = {
      ...config.resolve.fallback,
      fs: false,
    };

    config.externals = [...(config.externals || []), "pino-pretty", "lokijs", "encoding"];

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
};

export default nextConfig;
