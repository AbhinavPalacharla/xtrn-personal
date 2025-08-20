import { z } from "zod";
import { type InferSchema } from "xmcp";
import { googleOauthClient } from "../auth";
import { google } from "googleapis";
import { handleAuthErrorToMCP, NewMCPError, NewMCPLLMErrorResponse, NewMCPResponse } from "../utils/error";

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
  reminders: z.object({
    useDefault: z.boolean().optional().describe("Use default reminders (true) or custom reminders (false)"),
    overrides: z.array(z.object({
      method: z.enum(["email", "popup"]).describe("Reminder method"),
      minutes: z.number().describe("Minutes before event to send reminder")
    })).optional().describe("Custom reminder overrides - only used when useDefault is false")
  }).optional().describe("Event reminders/notifications configuration"),
};

export const metadata = {
  name: "create_event",
  description: "Creates a new event in Google Calendar with optional reminders/notifications",
  annotations: {
    title: "Create a calendar event with reminders",
    idempotentHint: false,
    readOnlyHint: false,
  },
};

export default async function createEvent(args: InferSchema<typeof schema>) {
  try {
    const client = await googleOauthClient();
    const calendar = google.calendar({ version: "v3", auth: client });

    // Prepare the event object for the Google Calendar API
    const eventBody: any = {
      summary: args.summary,
      start: args.start,
      end: args.end,
      description: args.description,
      location: args.location,
    };

    // Handle reminders if provided
    if (args.reminders) {
      eventBody.reminders = {
        useDefault: args.reminders.useDefault ?? false,
      };
      
      // If using custom reminders, add the overrides
      if (!args.reminders.useDefault && args.reminders.overrides) {
        eventBody.reminders.overrides = args.reminders.overrides;
      }
    }

    const res = await calendar.events.insert({
      calendarId: "primary",
      requestBody: eventBody,
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
