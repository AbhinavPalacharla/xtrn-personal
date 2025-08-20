import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  maxResults: z.number().min(1).max(100).optional().describe("Maximum number of emails to return (optional, default: 10, min: 1, max: 100)"),
  query: z.string().optional().describe("Gmail search query to filter emails (optional, e.g., 'from:example@gmail.com', 'subject:meeting', 'has:attachment')"),
  labelIds: z.array(z.string()).optional().describe("Array of Gmail label IDs to filter emails by (optional, e.g., ['INBOX', 'IMPORTANT', 'SENT'])"),
  includeSpamTrash: z.boolean().optional().describe("Whether to include emails from SPAM and TRASH in the results (optional, default: false)"),
  pageToken: z.string().optional().describe("Page token for pagination to get the next page of results (optional)"),
};

export const metadata = {
  name: "list_emails",
  description: "Lists emails from the user's Gmail inbox with optional filtering, search queries, and label-based filtering",
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
    includeSpamTrash: args.includeSpamTrash ?? false,
    pageToken: args.pageToken,
  });

  return NewMCPResponse(JSON.stringify(res.data));
}