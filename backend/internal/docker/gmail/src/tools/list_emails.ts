import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  maxResults: z.number().optional().describe("Maximum number of emails to return (default: 10)"),
  query: z.string().optional().describe("Gmail search query to filter emails"),
  labelIds: z.array(z.string()).optional().describe("Label IDs to filter emails by"),
};

export const metadata = {
  name: "list_emails",
  description: "Lists emails from the user's Gmail inbox with optional filtering",
  annotations: {
    title: "List Gmail emails",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function listEmails(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  const res = await gmail.users.messages.list({
    userId: "me",
    maxResults: args.maxResults ?? 10,
    q: args.query,
    labelIds: args.labelIds,
  });

  return NewMCPResponse(JSON.stringify(res.data));
}