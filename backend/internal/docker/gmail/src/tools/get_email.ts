import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  messageId: z.string().min(1).describe("The unique ID of the email message to retrieve (required, cannot be empty)"),
  format: z.enum(["minimal", "full", "raw", "metadata"]).optional().describe("The format of the message to return (optional, default: 'full'). 'minimal' returns basic metadata, 'full' returns complete message content, 'raw' returns raw message data, 'metadata' returns only message headers and metadata"),
  metadataHeaders: z.array(z.string()).optional().describe("Array of specific header names to include in the response (optional, e.g., ['From', 'Subject', 'Date']). Only applies when format is 'metadata'"),
};

export const metadata = {
  name: "get_email",
  description: "Retrieves a specific email message by its ID with various format options and metadata filtering",
  annotations: {
    title: "Get Gmail email",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function getEmail(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  const res = await gmail.users.messages.get({
    userId: "me",
    id: args.messageId,
    format: args.format ?? "full",
    metadataHeaders: args.metadataHeaders,
  });

  return NewMCPResponse(JSON.stringify(res.data));
}