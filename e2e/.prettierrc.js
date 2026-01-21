/** @type {import("prettier").Config} */
const baseConfig = require("../.prettierrc.js");

module.exports = {
  // Merkezi ayarları dahil et
  ...baseConfig,

  // Buraya bu projeye özel (override) ayarlar ekleyebilirsin
  // Örnek:
  // trailingComma: "all",
  // printWidth: 120,

  // Eklenti desteği (Özellikle monorepo yapılarında kritik)
  plugins: [
    ...(baseConfig.plugins || []),
    // "prettier-plugin-packagejson", // JSON dosyalarını otomatik sıralar
  ],
};
