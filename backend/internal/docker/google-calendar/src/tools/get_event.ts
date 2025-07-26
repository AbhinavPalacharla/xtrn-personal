import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";

export const schema = {
  eventId: z.string().describe("ID of the event to retrieve"),
};

export const metadata = {
  name: "get_event",
  description: "Retrieves details of a specific event",
  annotations: {
    title: "Get calendar event",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function getEvent(args: InferSchema<typeof schema>) {
  const client = await googleOauthClient();
  const calendar = google.calendar({ version: "v3", auth: client });

  const res = await calendar.events.get({
    calendarId: "primary",
    eventId: args.eventId,
  });

  return {
    content: [
      {
        type: "text",
        text: JSON.stringify(res.data, null, 2),
      },
    ],
  };
}
