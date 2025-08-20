import { OAuth2Client } from "google-auth-library";
import { InvalidGrantError, MissingTokenFieldsError, UnknownAuthError } from "./utils/error";

const CLIENT_ID = process.env.GOOGLE_CLIENT_ID || process.env.CLIENT_ID;
const CLIENT_SECRET = process.env.GOOGLE_CLIENT_SECRET || process.env.CLIENT_SECRET;
const REFRESH_TOKEN = process.env.TEST_GOOGLE_CALENDAR_REFRESH_TOKEN || process.env.REFRESH_TOKEN;

const client = new OAuth2Client(CLIENT_ID, CLIENT_SECRET, "http://localhost");

let accessToken: string = "";
let expiresAt: number = 0;

const googleOauthClient = async () => {
  const now = Date.now();

  //Get access token if no token or expired
  if (!accessToken || now > expiresAt) {
    try {
      client.setCredentials({ refresh_token: REFRESH_TOKEN });

      const { credentials } = await client.refreshAccessToken();

      if (!credentials.access_token || !credentials.expiry_date) {
        throw new MissingTokenFieldsError();

        // throw {
        //   type: "missing_token_fields",
        //   message: "Missing access token or expiry date from Google",
        // } as const;
      }

      accessToken = credentials.access_token;

      client.setCredentials({ access_token: accessToken });

      return client;
    } catch (err: any) {
      if (err?.response?.data?.error === "invalid_grant") {
        throw new InvalidGrantError();

        // throw {
        //   type: "invalid_grant",
        //   message: "Refresh token is invalid or revoked",
        // } as const;
      }

      throw new UnknownAuthError();

      // throw {
      //   type: "unknown",
      //   message: err?.message || "Unknown error during token refresh",
      // } as const;
    }
  }

  return client;
};

export { googleOauthClient };
