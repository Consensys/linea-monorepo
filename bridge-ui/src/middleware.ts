// NextJS automatically recognises a single middleware.ts file in the project root - https://nextjs.org/docs/app/building-your-application/routing/middleware#convention
import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const nonce = Buffer.from(crypto.randomUUID()).toString("base64");

  /**
   * Content Security Policy (CSP) configuration:
   *
   * default-src 'self'
   * - Fallback policy to only allow resources from the same origin, unless overriden by a more specific policy
   *
   * script-src-elem 'self' 'nonce-{nonce}' 'unsafe-eval' <HTTPS_URL>
   * - Allow scripts from the same origin, with the provided nonce  and from specified domains.
   * - `unsafe-eval` is required to enable evaluation of arbitrary strings in JS, and the React component does not load otherwise
   * - We would prefer to use `strict-dynamic` instead of specifying HTTPS_URLs
   *    - However the Usabilla script dynamically loads JS from Usabilla domain, and we are unable to provide the CSP nonce to each instance in which JS is dynamically loaded
   *    - If 'strict-dynamic' is used then HTTPS_URLs are ignored
   *    - As a compromise, we use script-src with 'strict-dynamic' as a fallback
   * - script-src-elem applies to scripts created via HTML elements, but does not cover inline event handlers or scripts injected via JS APIs like 'new Function()'
   *
   * style-src 'self' 'unsafe-inline'
   * - Allow styles from the same origin and inline styles.
   *
   * img-src
   * - Control image source
   * - We allow `https:` here because we cannot sustainably maintain a whitelist for token image sources used by our widgets, especially when new tokens come out everyday and some introduce new image sources
   *
   * font-src
   * - Control font source
   *
   * connect-src
   * - Controls all outbound network requests, including fetch(), WebSockets, EventSource, navigator.sendBeacon
   *
   * frame-src
   * - Control source for iframes
   *
   * object-src 'none'
   * - Disallow object, embed, and applet elements (should not appear in modern frontends)
   *
   * base-uri 'self'
   * - Control <base href="..."> to be from same origin as the page
   *
   * form-action 'self'
   * - Restrict form submissions to the same origin.
   *
   * frame-ancestors 'none'
   * - Disallow this site from being embedded in iframes (similar to X-Frame-Options: DENY).
   *
   * block-all-mixed-content
   * - Block all mixed (HTTP over HTTPS) content.
   *
   * upgrade-insecure-requests
   * - Automatically upgrade HTTP requests to HTTPS.
   */
  const cspHeader = `
    default-src 'self';
    script-src 'self' 'nonce-${nonce}' 'strict-dynamic' 'unsafe-eval';
    script-src-elem 'self' 'nonce-${nonce}' 'unsafe-eval'
      https://www.googletagmanager.com https://w.usabilla.com
      https://js.hsleadflows.net https://js.hsadspixel.net https://js.hubspot.com https://js.hs-analytics.net https://js.hs-banner.com
      https://widget.intercom.io https://js.intercomcdn.com;
    style-src 'self' 'unsafe-inline';
    img-src 'self' blob: data: https:;
    font-src 'self' data: https://cdn.jsdelivr.net;
    connect-src 'self' 
      https://*.infura.io https://*.alchemyapi.io 
      wss://*.infura.io wss://*.alchemyapi.io 
      https://*.quiknode.pro wss://*.quiknode.pro 
      https://www.googletagmanager.com 
      https://*.walletconnect.org wss://*.walletconnect.org
      https://*.dynamic.xyz https://app.dynamicauth.com https://*.dynamicauth.com https://dynamic-static-assets.com
      https://api.onramper.com https://*.onramper.com
      https://api.layerswap.io https://*.layerswap.io
      https://api.li.fi https://*.li.fi https://li.quest
      https://*.hubspot.com https://api.hubapi.com https://js.hs-banner.com https://js.hs-banner.com
      https://iris-api-sandbox.circle.com https://iris-api.circle.com
      https://price.api.cx.metamask.io
      https://registry.npmjs.org
      https://www.google-analytics.com
      https://rpc.linea.build
      https://bsc-dataseed.binance.org https://bsc-dataseed.bnbchain.org https://bsc-rpc.publicnode.com https://bsc-dataseed1.defibit.io
      https://arb1.arbitrum.io https://arbitrum-one-rpc.publicnode.com https://arbitrum.drpc.org
      https://rpc-gel.inkonchain.com
      https://rpc.superposition.so
      https://cronos.drpc.org https://evm.cronos.org
      https://api.lens.matterhosted.dev
      https://mainnet.corn-rpc.com
      https://swell-mainnet.alt.technology
      https://rpc.soniclabs.com https://sonic.drpc.org
      https://evm-rpc.sei-apis.com
      https://mainnet.aurora.dev
      https://mode.drpc.org https://mainnet.mode.network
      https://rpc.api.moonbeam.network https://moonbeam.drpc.org https://moonbeam-rpc.publicnode.com
      https://rpc.fantom.network https://rpcapi.fantom.network https://fantom-rpc.publicnode.com https://fantom.drpc.org
      https://rpc.gnosischain.com https://gnosis-rpc.publicnode.com https://gnosis.drpc.org;
    frame-src 'self' 
      https://www.googletagmanager.com 
      https://*.walletconnect.com 
      https://buy.onramper.com/;
    object-src 'none';
    base-uri 'self';
    form-action 'self';
    frame-ancestors 'none';
    block-all-mixed-content;
    upgrade-insecure-requests;
  `;

  /**
   * Purposely excluded URLs from CSP whitelist because they seem suspicious
   *
   * base-uri
   * - https://d6tizftlrpuof.cloudfront.net/live/
   *
   * script-src
   * - https://snap.licdn.com
   *
   */

  // Replace newline characters and spaces
  const contentSecurityPolicyHeaderValue = cspHeader.replace(/\s{2,}/g, " ").trim();

  const requestHeaders = new Headers(request.headers);
  // Pass nonce to <Script> elements in layout.tsx to bypass CSP
  requestHeaders.set("x-nonce", nonce);
  requestHeaders.set("Content-Security-Policy", contentSecurityPolicyHeaderValue);
  // Set response headers so that browsers enforce CSP
  const responseHeaders = new Headers();
  responseHeaders.set("Content-Security-Policy", contentSecurityPolicyHeaderValue);

  const response = NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  });
  response.headers.set("Content-Security-Policy", contentSecurityPolicyHeaderValue);

  return response;
}

// Filter Middleware to run on specific paths - https://nextjs.org/docs/14/app/building-your-application/configuring/content-security-policy#adding-a-nonce-with-middleware
export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder
     */
    {
      source: "/((?!api|_next/static|_next/image|favicon.ico|public/).*)",
      // Skip running Middleware if request includes NextJS prefetch headers
      missing: [
        { type: "header", key: "next-router-prefetch" },
        { type: "header", key: "purpose", value: "prefetch" },
      ],
    },
  ],
};
