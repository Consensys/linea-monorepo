const { ethers } = require('ethers');
const fs = require('fs');
const path = require('path');

// Configuration with environment variables
const config = {
  rpcUrl: process.env.RPC_URL || 'http://localhost:8545',
  privateKey: process.env.PRIVATE_KEY,
  toAddress: process.env.TO_ADDRESS,
  transferAmount: process.env.TRANSFER_AMOUNT || '0.001',
  intervalMs: parseInt(process.env.INTERVAL_MS) || 1000,
  logToFile: process.env.LOG_TO_FILE === 'true'
};

class Logger {
  constructor() {
    this.logDir = '/app/logs';
    if (config.logToFile) {
      this.ensureLogDir();
      this.logFile = path.join(this.logDir, `transfer-${new Date().toISOString().split('T')[0]}.log`);
    }
  }

  ensureLogDir() {
    if (!fs.existsSync(this.logDir)) {
      fs.mkdirSync(this.logDir, { recursive: true });
    }
  }

  log(message) {
    const timestamp = new Date().toISOString();
    const logMessage = `[${timestamp}] ${message}`;

    console.log(logMessage);

    if (config.logToFile && this.logFile) {
      fs.appendFileSync(this.logFile, logMessage + '\n');
    }
  }

  error(message) {
    const timestamp = new Date().toISOString();
    const logMessage = `[${timestamp}] ERROR: ${message}`;

    console.error(logMessage);

    if (config.logToFile && this.logFile) {
      fs.appendFileSync(this.logFile, logMessage + '\n');
    }
  }
}

class TransferBot {
  constructor() {
    this.logger = new Logger();
    this.provider = new ethers.JsonRpcProvider(config.rpcUrl);
    this.wallet = new ethers.Wallet(config.privateKey, this.provider);
    this.isRunning = false;
    this.transactionCount = 0;
    this.startTime = Date.now();
  }

  async initialize() {
    try {
      // Test connection with timeout
      const networkPromise = this.provider.getNetwork();
      const timeoutPromise = new Promise((_, reject) =>
        setTimeout(() => reject(new Error('Connection timeout')), 10000)
      );

      const network = await Promise.race([networkPromise, timeoutPromise]);
      this.logger.log(`Connected to network: ${network.name} (Chain ID: ${network.chainId})`);

      // Check wallet balance
      const balance = await this.provider.getBalance(this.wallet.address);
      this.logger.log(`Wallet address: ${this.wallet.address}`);
      this.logger.log(`Wallet balance: ${ethers.formatEther(balance)} ETH`);

      if (balance === 0n) {
        throw new Error('Wallet has no balance');
      }

      // Validate recipient address
      if (!ethers.isAddress(config.toAddress)) {
        throw new Error('Invalid recipient address');
      }

      return true;
    } catch (error) {
      this.logger.error(`Initialization failed: ${error.message}`);
      return false;
    }
  }

  async sendTransfer() {
    try {
      const tx = {
        to: config.toAddress,
        value: ethers.parseEther(config.transferAmount),
        gasLimit: 21000,
      };

      this.logger.log(`Sending transaction #${this.transactionCount + 1}...`);

      const txResponse = await this.wallet.sendTransaction(tx);
      this.transactionCount++;

      this.logger.log(`✅ Transaction sent: ${txResponse.hash}`);
      this.logger.log(`   Amount: ${config.transferAmount} ETH | To: ${config.toAddress}`);

      return txResponse;
    } catch (error) {
      this.logger.error(`Transaction failed: ${error.message}`);

      if (error.code === 'INSUFFICIENT_FUNDS') {
        this.logger.log('⚠️  Insufficient funds. Stopping bot...');
        this.stop();
      }
    }
  }

  start() {
    if (this.isRunning) {
      this.logger.log('Bot is already running');
      return;
    }

    this.isRunning = true;
    this.logger.log('🚀 Starting transfer bot...');
    this.logger.log(`   Sending ${config.transferAmount} ETH every ${config.intervalMs}ms`);
    this.logger.log(`   From: ${this.wallet.address}`);
    this.logger.log(`   To: ${config.toAddress}`);

    this.intervalId = setInterval(async () => {
      if (this.isRunning) {
        await this.sendTransfer();
      }
    }, config.intervalMs);

    // Log stats every minute
    this.statsInterval = setInterval(() => {
      const uptime = Math.floor((Date.now() - this.startTime) / 1000);
      const rate = this.transactionCount / (uptime / 60);
      this.logger.log(`📊 Stats: ${this.transactionCount} transactions | ${uptime}s uptime | ${rate.toFixed(2)} tx/min`);
    }, 60000);
  }

  stop() {
    if (!this.isRunning) {
      return;
    }

    this.isRunning = false;
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
    if (this.statsInterval) {
      clearInterval(this.statsInterval);
    }

    const uptime = Math.floor((Date.now() - this.startTime) / 1000);
    this.logger.log(`🛑 Bot stopped. Total: ${this.transactionCount} transactions in ${uptime}s`);
  }

  // Health check endpoint
  getStatus() {
    return {
      running: this.isRunning,
      transactionCount: this.transactionCount,
      uptime: Date.now() - this.startTime
    };
  }
}

// Health check server (optional)
if (process.env.ENABLE_HEALTH_CHECK === 'true') {
  const http = require('http');
  let botInstance;

  const server = http.createServer((req, res) => {
    if (req.url === '/health') {
      const status = botInstance ? botInstance.getStatus() : { running: false };
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(status));
    } else {
      res.writeHead(404);
      res.end('Not Found');
    }
  });

  server.listen(3000, () => {
    console.log('Health check server running on port 3000');
  });
}

// Main execution
async function main() {
  // Validate required environment variables
  if (!config.privateKey) {
    console.error('❌ PRIVATE_KEY environment variable is required');
    process.exit(1);
  }

  if (!config.toAddress) {
    console.error('❌ TO_ADDRESS environment variable is required');
    process.exit(1);
  }

  const bot = new TransferBot();

  // Make bot available for health checks
  if (process.env.ENABLE_HEALTH_CHECK === 'true') {
    global.botInstance = bot;
  }

  // Initialize the bot
  const initialized = await bot.initialize();
  if (!initialized) {
    process.exit(1);
  }

  // Start the bot
  bot.start();

  // Handle graceful shutdown
  const shutdown = (signal) => {
    console.log(`\n📡 Received ${signal} signal...`);
    bot.stop();
    process.exit(0);
  };

  process.on('SIGINT', () => shutdown('SIGINT'));
  process.on('SIGTERM', () => shutdown('SIGTERM'));
}

// Run the script
main().catch(error => {
  console.error('Script failed:', error);
  process.exit(1);
});
