(function(window) {
  window.__env = window.__env || {};

  // apiUrl is intentionally empty: Cloudflare Pages proxies /api/* to the gateway via _redirects
  window.__env.apiUrl = '';

})(this);
