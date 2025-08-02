import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { handleAuthErrorToMCP, NewMCPError, NewMCPResponse } from "../utils/error";

// import {
//   MissingTokenFieldsError,
//   InvalidGrantError,
//   UnknownAuthError,
// } from "../auth/errors"; // Adjust path accordingly

// import { NewMCPError, NewMCPResponse } from "../utils/error";

export const schema = {
  summary: z.string().describe("Event title"),
  start: z.object({
    dateTime: z.string().describe("Start time (ISO format)"),
    timeZone: z.string().optional().describe("Time zone"),
  }),
  end: z.object({
    dateTime: z.string().describe("End time (ISO format)"),
    timeZone: z.string().optional().describe("Time zone"),
  }),
  description: z.string().optional().describe("Event description"),
  location: z.string().optional().describe("Event location"),
};

export const metadata = {
  name: "create_event",
  description: "Creates a new event in Google Calendar",
  annotations: {
    title: "Create a calendar event",
    idempotentHint: false,
    readOnlyHint: false,
  },
};

export default async function createEvent(args: InferSchema<typeof schema>) {
  try {
    const client = await googleOauthClient();
    const calendar = google.calendar({ version: "v3", auth: client });

    const res = await calendar.events.insert({
      calendarId: "primary",
      requestBody: args,
    });

    return NewMCPResponse(JSON.stringify(res));
  } catch (e: unknown) {
    const mcpError = handleAuthErrorToMCP(e);
    if (mcpError) return mcpError;

    // fallback for unexpected errors
    if (e instanceof Error) {
      return NewMCPError("UNKNOWN_ERROR", e.message);
    }

    // fallback for unknown non-Error thrown values
    return NewMCPError("UNKNOWN_ERROR", "An unknown error occurred");

    // if (e instanceof InvalidGrantError) {
    //   return NewMCPError("AUTH_INVALID_GRANT", e.message);
    // } else if (e instanceof MissingTokenFieldsError) {
    //   return NewMCPError("AUTH_MISSING_FIELDS", e.message);
    // } else if (e instanceof UnknownAuthError) {
    //   return NewMCPError("AUTH_UNKNOWN_ERROR", e.message);
    // } else if (e instanceof Error) {
    //   // fallback for unexpected errors
    // }
  }
}

// import { z } from "zod";
// import { type InferSchema } from "xmcp";
// import { googleOauthClient } from "../auth";
// import { google } from "googleapis";
// import { isAppError, AppError, NewMCPError, NewMCPResponse } from "../utils/error";

// export const schema = {
//   summary: z.string().describe("Event title"),
//   start: z.object({
//     dateTime: z.string().describe("Start time (ISO format)"),
//     timeZone: z.string().optional().describe("Time zone"),
//   }),
//   end: z.object({
//     dateTime: z.string().describe("End time (ISO format)"),
//     timeZone: z.string().optional().describe("Time zone"),
//   }),
//   description: z.string().optional().describe("Event description"),
//   location: z.string().optional().describe("Event location"),
// };

// export const metadata = {
//   name: "create_event",
//   description: "Creates a new event in Google Calendar",
//   annotations: {
//     title: "Create a calendar event",
//     idempotentHint: false,
//     readOnlyHint: false,
//   },
// };

// export default async function createEvent(args: InferSchema<typeof schema>) {
//   try {
//     const client = await googleOauthClient();
//     const calendar = google.calendar({ version: "v3", auth: client });

//     const res = await calendar.events.insert({
//       calendarId: "primary",
//       requestBody: args,
//     });

//     return NewMCPResponse(JSON.stringify(res));
//   } catch (e: any) {
//     return NewMCPResponse(JSON.stringify(e));
//     // // if (isAppError(e)) {
//     // //   return NewMCPError(e.type, e.message);
//     // // }

//     // if (isAppError(e)) {
//     //   return NewMCPError("AUTH_ERROR", e.message);
//     // }
//   }
// }
