import { google } from "googleapis";
import dotenv from "dotenv";

if (process.env.ENV_PATH) {
  dotenv.config({
    path: [process.env.ENV_PATH!],
  });
} else {
  console.error("[INFO] NO ENV PATH DETECTED");
}

// Or statements for testing purposes read from env file
const CLIENT_ID = process.env.CLIENT_ID || process.env.GOOGLE_CLIENT_ID;
const CLIENT_SECRET = process.env.CLIENT_SECRET || process.env.GOOGLE_CLIENT_SECRET;
const REFRESH_TOKEN = process.env.REFRESH_TOKEN || process.env.TEST_GOOGLE_CALENDAR_REFRESH_TOKEN;

console.log(`ENV INFO: ${[CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN]}`);

if (!CLIENT_ID || !CLIENT_SECRET || !REFRESH_TOKEN) {
  throw new Error("Missing env vars");
}

console.log(CLIENT_ID, CLIENT_SECRET, REFRESH_TOKEN);

const oauthClient = new google.auth.OAuth2(CLIENT_ID, CLIENT_SECRET, "http://localhost");

oauthClient.setCredentials({ refresh_token: REFRESH_TOKEN });

export const calendar = google.calendar({ version: "v3", auth: oauthClient });
export const calendarId = "primary";
