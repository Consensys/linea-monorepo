// NextJS automatically recognises a single middleware.ts file in the project root - https://nextjs.org/docs/app/building-your-application/routing/middleware#convention
import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  // We only want to allow unsafe-eval in local environment for NextJS dev server
  const unsafeScript = process.env.NEXT_PUBLIC_ENVIRONMENT === "local" ? "'unsafe-eval'" : "";

  /**
   * Content Security Policy (CSP) configuration:
   *
   * default-src 'self'
   * - Fallback policy to only allow resources from the same origin, unless overriden by a more specific policy
   *
   * script-src-elem 'self' 'nonce-{nonce}' 'strict-dynamic'
   * - Allow scripts from the same origin, with the provided nonce, and child scripts recursively loaded from a script with nonce
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
    script-src 'self' ${unsafeScript} 'unsafe-inline' https://www.googletagmanager.com https://widget.intercom.io/widget/h5zisg78 https://ajax.cloudflare.com https://js.intercomcdn.com;
    style-src 'self' 'unsafe-inline';
    img-src 'self' blob: data: https:;
    font-src 'self' data: https://cdn.jsdelivr.net;
    connect-src 'self' https:;
    frame-src 'self'
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
   */

  // Replace newline characters and spaces
  const contentSecurityPolicyHeaderValue = cspHeader.replace(/\s{2,}/g, " ").trim();

  const requestHeaders = new Headers(request.headers);
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
