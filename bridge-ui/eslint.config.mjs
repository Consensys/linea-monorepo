import { nextjs } from "@consensys/eslint-config";

// eslint-disable-next-line import/no-anonymous-default-export
export default [
  ...nextjs,
  {
    languageOptions: {
      parserOptions: {
        project: "./tsconfig.json",
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    ignores: ["test/advancedFixtures.ts"]
  }
];
