(function(window) {
  window.__env = window.__env || {};

  // API URL for pipeline injection
  // The pipeline should use envsubst to replace this value before serving
  window.__env.apiUrl = '${API_URL}';

})(this);
