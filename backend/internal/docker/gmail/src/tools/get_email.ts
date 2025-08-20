import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  messageId: z.string().describe("The ID of the email message to retrieve"),
  format: z.enum(["minimal", "full", "raw", "metadata"]).optional().describe("The format of the message to return (default: full)"),
};

export const metadata = {
  name: "get_email",
  description: "Retrieves a specific email message by its ID",
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
  });

  return NewMCPResponse(JSON.stringify(res.data));
}