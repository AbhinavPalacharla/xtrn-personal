import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { handleAuthErrorToMCP, NewMCPError, NewMCPLLMErrorResponse, NewMCPResponse } from "../utils/error";

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
  try {
    const client = await googleOauthClient();
    const calendar = google.calendar({ version: "v3", auth: client });

    const res = await calendar.events.get({
      calendarId: "primary",
      eventId: args.eventId,
    });

    return NewMCPResponse(JSON.stringify(res));
  } catch (e: unknown) {
    const mcpError = handleAuthErrorToMCP(e);
    if (mcpError) return mcpError;

    // fallback for unexpected errors that LLM can handle
    if (e instanceof Error) {
      return NewMCPLLMErrorResponse(`${e.name} - ${e.message}`);
    }

    // fallback for unknown errors that LLM cannot handle
    return NewMCPError("UNKNOWN_ERROR", "An unknown error occurred");
  }
}
