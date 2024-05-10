/** @type {import("prettier").Config} */
module.exports = {
  ...require("../.prettierrc.js"),
  plugins: ["prettier-plugin-solidity"],
  overrides: [
    {
      files: "*.sol",
      options: {
        parser: "solidity-parse",
        bracketSpacing: true,
        printWidth: 120,
        singleQuote: false,
        tabWidth: 2,
        useTabs: false,
      },
    },
  ],
};
