import axios from "axios";
import { getCurrentUnixTimestampSeconds } from "../../utils/viem/time";
import { ILogger } from "../../logging/ILogger";
import { IOAuth2TokenClient, OAuth2TokenResponse } from "../../core/client/IOAuth2TokenClient";

export class OAuth2TokenClient implements IOAuth2TokenClient {
  private bearerToken?: string;
  private tokenExpiresAtSeconds?: number;

  constructor(
    private readonly logger: ILogger,
    private readonly tokenUrl: string,
    private readonly clientId: string,
    private readonly clientSecret: string,
    private readonly audience: string,
    private readonly grantType: string = "client_credentials",
    private readonly expiryBufferSeconds: number = 60,
    private readonly defaultExpiresInSeconds: number = 3_600,
  ) {}

  async getBearerToken(): Promise<string> {
    // Serve cached token while it remains valid.
    if (
      this.bearerToken &&
      this.tokenExpiresAtSeconds &&
      getCurrentUnixTimestampSeconds() < this.tokenExpiresAtSeconds - this.expiryBufferSeconds
    ) {
      return this.bearerToken;
    }

    const { data } = await axios.post<OAuth2TokenResponse>(
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
    );

    if (!data?.access_token) {
      this.logger.error("Failed to retrieve OAuth2 access token");
    }

    const tokenType = data.token_type ?? "Bearer";
    this.bearerToken = `${tokenType} ${data.access_token}`.trim();

    if (data?.expires_at) {
      this.tokenExpiresAtSeconds = data?.expires_at;
    } else if (data?.expires_in) {
      this.tokenExpiresAtSeconds = getCurrentUnixTimestampSeconds() + data?.expires_in;
    } else {
      this.tokenExpiresAtSeconds = getCurrentUnixTimestampSeconds() + this.defaultExpiresInSeconds;
    }

    return this.bearerToken;
  }
}
