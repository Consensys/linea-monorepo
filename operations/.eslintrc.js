module.exports = {
  env: {
    commonjs: true,
    mocha: false,
    es2021: false,
    jest: true,
  },
  extends: ["../.eslintrc.js"],
  plugins: ["prettier"],
  parserOptions: {
    sourceType: "module",
  },
  rules: {
    "prettier/prettier": "error",
  },
};
