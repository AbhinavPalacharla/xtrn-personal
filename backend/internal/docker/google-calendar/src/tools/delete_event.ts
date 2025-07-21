import { z } from "zod";
import { type InferSchema } from "xmcp";
import { calendar, calendarId } from "../auth";

export const schema = {
  eventId: z.string().describe("ID of the event to delete"),
};

export const metadata = {
  name: "delete_event",
  description: "Deletes an event from the calendar",
  annotations: {
    title: "Delete calendar event",
    destructiveHint: true,
    readOnlyHint: false,
  },
};

export default async function deleteEvent(args: InferSchema<typeof schema>) {
  await calendar.events.delete({
    calendarId,
    eventId: args.eventId,
  });

  return {
    content: [
      {
        type: "text",
        text: `Event deleted: ${args.eventId}`,
      },
    ],
  };
}
