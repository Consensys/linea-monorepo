module.exports = {
  root: true,
  env: {
    node: true,
    browser: false,
    es2022: true,
  },
  ignorePatterns: ["dist", "coverage"],
  extends: ["eslint:recommended", "plugin:@typescript-eslint/recommended", "plugin:prettier/recommended"],
  parser: "@typescript-eslint/parser",
  plugins: ["@typescript-eslint", "prettier"],
  parserOptions: {
    ecmaVersion: 12,
    sourceType: "module",
  },
};
