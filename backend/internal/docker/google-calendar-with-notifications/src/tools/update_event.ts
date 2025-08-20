import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { handleAuthErrorToMCP, NewMCPError, NewMCPLLMErrorResponse, NewMCPResponse } from "../utils/error";

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
  reminders: z.object({
    useDefault: z.boolean().optional().describe("Use default reminders (true) or custom reminders (false)"),
    overrides: z.array(z.object({
      method: z.enum(["email", "popup"]).describe("Reminder method"),
      minutes: z.number().describe("Minutes before event to send reminder")
    })).optional().describe("Custom reminder overrides - only used when useDefault is false")
  }).optional().describe("Event reminders/notifications configuration"),
};

export const metadata = {
  name: "update_event",
  description: "Updates an existing event including reminders/notifications",
  annotations: {
    title: "Update calendar event with reminders",
    idempotentHint: false,
    readOnlyHint: false,
  },
};

export default async function updateEvent(args: InferSchema<typeof schema>) {
  try {
    const { eventId, reminders, ...updates } = args;

    const client = await googleOauthClient();
    const calendar = google.calendar({ version: "v3", auth: client });

    // Prepare the update object
    const updateBody: any = { ...updates };

    // Handle reminders if provided
    if (reminders) {
      updateBody.reminders = {
        useDefault: reminders.useDefault ?? false,
      };
      
      // If using custom reminders, add the overrides
      if (!reminders.useDefault && reminders.overrides) {
        updateBody.reminders.overrides = reminders.overrides;
      }
    }

    const res = await calendar.events.patch({
      calendarId: "primary",
      eventId,
      requestBody: updateBody,
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
