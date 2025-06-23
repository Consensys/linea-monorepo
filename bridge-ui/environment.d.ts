declare global {
  namespace NodeJS {
    interface ProcessEnv {
      E2E_TEST_SEED_PHRASE: string;
      E2E_TEST_PRIVATE_KEY: string;
      E2E_TEST_WALLET_PASSWORD: string;
      NEXT_PUBLIC_INFURA_ID: string;
      NEXT_PUBLIC_BASE_PATH: string;
    }
  }
}

export {};
