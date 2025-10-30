import axios from "axios";
import { ILogger } from "../logging/ILogger";
import { getCurrentUnixTimestampSeconds } from "../utils/time";
import { IOAuth2TokenClient, OAuth2TokenResponse } from "../core/client/IOAuth2TokenClient";
import { IRetryService } from "../core/services/IRetryService";

export class OAuth2TokenClient implements IOAuth2TokenClient {
  private bearerToken?: string;
  private tokenExpiresAtSeconds?: number;

  constructor(
    private readonly logger: ILogger,
    private readonly retryService: IRetryService,
    private readonly tokenUrl: string,
    private readonly clientId: string,
    private readonly clientSecret: string,
    private readonly audience: string,
    private readonly grantType: string = "client_credentials",
    private readonly expiryBufferSeconds: number = 60,
  ) {}

  async getBearerToken(): Promise<string | undefined> {
    // Serve cached token while it remains valid.
    if (
      this.bearerToken &&
      this.tokenExpiresAtSeconds &&
      getCurrentUnixTimestampSeconds() < this.tokenExpiresAtSeconds - this.expiryBufferSeconds
    ) {
      this.logger.info("getBearerToken cache-hit");
      return this.bearerToken;
    }

    this.logger.info("getBearerToken requesting new token");
    const { data } = await this.retryService.retry(() =>
      axios.post<OAuth2TokenResponse>(
        this.tokenUrl,
        {
          client_id: this.clientId,
          client_secret: this.clientSecret,
          audience: this.audience,
          grant_type: this.grantType,
        },
        {
          headers: {
            "content-type": "application/json",
          },
        },
      ),
    );

    if (!data?.access_token) {
      this.logger.error("Failed to retrieve OAuth2 access token");
      return undefined;
    }

    const tokenType = data.token_type ?? "Bearer";
    this.bearerToken = `${tokenType} ${data.access_token}`.trim();

    if (data?.expires_at) {
      const tokenExpiresAtSecondsCandidate = data?.expires_at;
      if (tokenExpiresAtSecondsCandidate < getCurrentUnixTimestampSeconds()) {
        this.logger.error(`OAuth2 access token already expired at ${tokenExpiresAtSecondsCandidate}`);
        return undefined;
      }
      this.tokenExpiresAtSeconds = tokenExpiresAtSecondsCandidate;
    } else if (data?.expires_in) {
      this.tokenExpiresAtSeconds = getCurrentUnixTimestampSeconds() + data?.expires_in;
    } else {
      this.logger.error(`OAuth2 access token did not provide expiry data`);
      return undefined;
    }

    this.logger.info(
      `getBearerToken successfully retrived new OAuth2 Bearer token tokenExpiresAtSeconds=${this.tokenExpiresAtSeconds}`,
    );

    return this.bearerToken;
  }
}
