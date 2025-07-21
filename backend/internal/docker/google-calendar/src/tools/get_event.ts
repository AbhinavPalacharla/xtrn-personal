import { z } from "zod";
import { type InferSchema } from "xmcp";
import { calendar, calendarId } from "../auth";

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
  const res = await calendar.events.get({
    calendarId,
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
