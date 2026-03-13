const { createProxyMiddleware } = require("http-proxy-middleware");

module.exports = function (app) {
  // Enable Cross-Origin Isolation for SharedArrayBuffer
  app.use((req, res, next) => {
    res.setHeader("Cross-Origin-Opener-Policy", "same-origin");
    // res.setHeader("Cross-Origin-Embedder-Policy", "require-corp");
    next();
  });

  // Proxy carto maps to bypass COEP
  app.use(
    "/carto-proxy",
    createProxyMiddleware({
      target: "https://basemaps.cartocdn.com",
      changeOrigin: true,
      pathRewrite: {
        "^/carto-proxy": "",
      },
      onProxyRes: function (proxyRes) {
        proxyRes.headers["Cross-Origin-Resource-Policy"] = "cross-origin";
      },
    }),
  );

  // Proxy API requests to the mock server
  app.use(
    "/api",
    createProxyMiddleware({
      target: "http://127.0.0.1:8081",
      changeOrigin: true,
    }),
  );
};
