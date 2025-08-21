import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  messageId: z.string().min(1).describe("The unique ID of the email message to modify (required, cannot be empty)"),
  addLabelIds: z.array(z.string()).optional().describe("Array of label IDs to add to the email (optional, e.g., ['IMPORTANT', 'WORK'])"),
  removeLabelIds: z.array(z.string()).optional().describe("Array of label IDs to remove from the email (optional, e.g., ['INBOX', 'UNREAD'])"),
  // Legacy support - if labelIds is provided, it will add those labels
  labelIds: z.array(z.string()).optional().describe("Array of label IDs to add to the email (legacy parameter, use addLabelIds instead)"),
};

export const metadata = {
  name: "modify_email",
  description: "Modifies email labels to move emails between folders, mark as important, archive, or apply other Gmail organizational features",
  annotations: {
    title: "Modify Gmail email labels",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function modifyEmail(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  // Prepare request body
  const requestBody: any = {};
  
  // Handle legacy labelIds parameter
  if (args.labelIds && args.labelIds.length > 0) {
    requestBody.addLabelIds = args.labelIds;
  }
  
  if (args.addLabelIds && args.addLabelIds.length > 0) {
    requestBody.addLabelIds = args.addLabelIds;
  }
  
  if (args.removeLabelIds && args.removeLabelIds.length > 0) {
    requestBody.removeLabelIds = args.removeLabelIds;
  }

  // Ensure we have at least one operation to perform
  if (Object.keys(requestBody).length === 0) {
    throw new Error("At least one of addLabelIds, removeLabelIds, or labelIds must be provided");
  }

  const res = await gmail.users.messages.modify({
    userId: "me",
    id: args.messageId,
    requestBody: requestBody,
  });

  return NewMCPResponse(JSON.stringify(res.data));
}