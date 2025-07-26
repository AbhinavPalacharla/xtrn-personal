import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { isAppError, AppError, NewMCPError, NewMCPResponse } from "../utils/error";

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

    return NewMCPResponse({
      is_error: false,
      content: [
        {
          type: "text",
          text: `Event created with ID: ${res.data.id}`,
        },
      ],
    });
  } catch (e: any) {
    if (isAppError(e)) {
      return NewMCPError(e.type, e.message);
    }
  }
}
