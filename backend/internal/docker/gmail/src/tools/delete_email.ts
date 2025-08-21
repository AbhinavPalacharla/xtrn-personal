import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  messageId: z.string().min(1).describe("The unique ID of the email message to permanently delete (required, cannot be empty)"),
};

export const metadata = {
  name: "delete_email",
  description: "Permanently deletes an email message from Gmail. This action cannot be undone and the email will be removed from all folders and labels.",
  annotations: {
    title: "Delete Gmail email",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function deleteEmail(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  await gmail.users.messages.delete({
    userId: "me",
    id: args.messageId,
  });

  return NewMCPResponse(JSON.stringify({ 
    success: true, 
    message: `Email ${args.messageId} deleted successfully`,
    deletedMessageId: args.messageId 
  }));
}