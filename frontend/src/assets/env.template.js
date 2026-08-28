(function(window) {
  window.__env = window.__env || {};

  // API URL injected by the CD pipeline via envsubst
  window.__env.apiUrl = '${API_URL}';

})(this);
