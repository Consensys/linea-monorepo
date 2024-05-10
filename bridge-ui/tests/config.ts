import 'dotenv/config';

const config = {
  PRIVATE_KEY: process.env.E2E_TEST_PRIVATE_KEY || '',
  URL: process.env.E2E_TEST_URL || '',
};

export default config;
