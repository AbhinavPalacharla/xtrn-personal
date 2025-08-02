import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

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
  const client = await googleOauthClient();
  const calendar = google.calendar({ version: "v3", auth: client });

  const res = await calendar.events.delete({
    calendarId: "primary",
    eventId: args.eventId,
  });

  return NewMCPResponse(JSON.stringify(res));

  // return {
  //   content: [
  //     {
  //       type: "text",
  //       text: `Event deleted: ${args.eventId}`,
  //     },
  //   ],
  // };
}
