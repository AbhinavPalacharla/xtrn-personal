import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";

export const schema = {
  timeMin: z.string().describe("Start of time range (ISO format)"),
  timeMax: z.string().describe("End of time range (ISO format)"),
  maxResults: z.number().optional().describe("Maximum number of results"),
  orderBy: z.enum(["startTime", "updated"]).optional().describe("Sort order"),
};

export const metadata = {
  name: "list_events",
  description: "Lists events within a specified time range",
  annotations: {
    title: "List calendar events",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function listEvents(args: InferSchema<typeof schema>) {
  const client = await googleOauthClient();
  const calendar = google.calendar({ version: "v3", auth: client });

  const res = await calendar.events.list({
    calendarId: "primary",
    timeMin: args.timeMin,
    timeMax: args.timeMax,
    maxResults: args.maxResults ?? 10,
    orderBy: args.orderBy ?? "startTime",
    singleEvents: true,
  });

  return {
    content: [
      {
        type: "text",
        text: JSON.stringify(res.data.items ?? [], null, 2),
      },
    ],
  };
}
