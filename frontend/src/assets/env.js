(function(window) {
  window.__env = window.__env || {};

  // apiUrl is empty: Cloudflare Pages proxies /api/* to the gateway via _redirects.
  // For local development, configure proxy.conf.json in Angular CLI.
  window.__env.apiUrl = '';

})(this);
