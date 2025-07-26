// import { google } from "googleapis";
// import dotenv from "dotenv";

// if (process.env.ENV_PATH) {
//   dotenv.config({
//     path: [process.env.ENV_PATH!],
//   });
// } else {
//   console.error("[INFO] NO ENV PATH DETECTED");
// }

// // Or statements for testing purposes read from env file
// const CLIENT_ID = process.env.CLIENT_ID || process.env.GOOGLE_CLIENT_ID;
// const CLIENT_SECRET = process.env.CLIENT_SECRET || process.env.GOOGLE_CLIENT_SECRET;
// const REFRESH_TOKEN = process.env.REFRESH_TOKEN || process.env.TEST_GOOGLE_CALENDAR_REFRESH_TOKEN;

// console.log(`ENV INFO: ${[CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN]}`);

// if (!CLIENT_ID || !CLIENT_SECRET || !REFRESH_TOKEN) {
//   throw new Error("Missing env vars");
// }

// console.log(CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN);

// const oauthClient = new google.auth.OAuth2(CLIENT_ID, CLIENT_SECRET, "http://localhost");

// oauthClient.setCredentials({ refresh_token: REFRESH_TOKEN });

// export const calendar = google.calendar({ version: "v3", auth: oauthClient });
// export const calendarId = "primary";

import { OAuth2Client } from "google-auth-library";

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
        throw {
          type: "missing_token_fields",
          message: "Missing access token or expiry date from Google",
        } as const;
      }

      accessToken = credentials.access_token;

      client.setCredentials({ access_token: accessToken });

      return client;
    } catch (err: any) {
      if (err?.response?.data?.error === "invalid_grant") {
        throw {
          type: "invalid_grant",
          message: "Refresh token is invalid or revoked",
        } as const;
      }

      throw {
        type: "unknown",
        message: err?.message || "Unknown error during token refresh",
      } as const;
    }
  }

  return client;
};

export { googleOauthClient };
