import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  eventId: z.string().describe("ID of the event to update"),
  summary: z.string().optional().describe("New event title"),
  start: z
    .object({
      dateTime: z.string().describe("New start time (ISO format)"),
      timeZone: z.string().optional().describe("Time zone"),
    })
    .optional(),
  end: z
    .object({
      dateTime: z.string().describe("New end time (ISO format)"),
      timeZone: z.string().optional().describe("Time zone"),
    })
    .optional(),
  description: z.string().optional().describe("New description"),
  location: z.string().optional().describe("New location"),
};

export const metadata = {
  name: "update_event",
  description: "Updates an existing event",
  annotations: {
    title: "Update calendar event",
    idempotentHint: false,
    readOnlyHint: false,
  },
};

export default async function updateEvent(args: InferSchema<typeof schema>) {
  const { eventId, ...updates } = args;

  const client = await googleOauthClient();
  const calendar = google.calendar({ version: "v3", auth: client });

  const res = await calendar.events.patch({
    calendarId: "primary",
    eventId,
    requestBody: updates,
  });

  return NewMCPResponse(JSON.stringify(res));

  // return {
  //   content: [
  //     {
  //       type: "text",
  //       text: `Event updated: ${eventId}`,
  //     },
  //   ],
  // };
}
