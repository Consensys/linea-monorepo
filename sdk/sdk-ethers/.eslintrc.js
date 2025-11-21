// eslintrc.js - ULTIMATE CLEAN CONFIGURATION

module.exports = {
  // Inherit base rules from the project root configuration.
  extends: "../../.eslintrc.js",
  
  // Environment configurations.
  env: {
    // Enables Node.js global variables and Node.js scope analysis.
    // Note: 'node: true' implicitly enables 'commonjs: true' and 'es2021: true'.
    node: true, 
    // Enables Jest testing global variables.
    jest: true,
  },
  
  parserOptions: {
    // Specifies that the code uses ES Modules syntax (import/export).
    sourceType: "module",
    // Specifies the ECMAScript version used for parsing (optional, defaults to latest if not set).
    // Adding ecmaVersion for explicit clarity on the target JavaScript version.
    ecmaVersion: 2021, 
  },
  
  rules: {
    // Enforce consistency between ESLint and Prettier formatting rules.
    "prettier/prettier": "error",
  },
};
