import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  id: z.string().min(1).describe("The unique ID of the label to update (required, cannot be empty)"),
  name: z.string().min(1).max(100).optional().describe("New name for the label (optional, 1-100 characters)"),
  messageListVisibility: z.enum(["show", "hide"]).optional().describe("Whether to show the label in the message list (optional)"),
  labelListVisibility: z.enum(["labelShow", "labelHide"]).optional().describe("Whether to show the label in the label list (optional)"),
};

export const metadata = {
  name: "update_label",
  description: "Updates an existing Gmail label's properties such as name, visibility in message list, and visibility in label list.",
  annotations: {
    title: "Update Gmail label",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function updateLabel(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  // Prepare updates with only the fields that were provided
  const updates: any = {};
  if (args.name) updates.name = args.name;
  if (args.messageListVisibility) updates.messageListVisibility = args.messageListVisibility;
  if (args.labelListVisibility) updates.labelListVisibility = args.labelListVisibility;

  // Ensure we have at least one field to update
  if (Object.keys(updates).length === 0) {
    throw new Error("At least one field (name, messageListVisibility, or labelListVisibility) must be provided");
  }

  const res = await gmail.users.labels.update({
    userId: "me",
    id: args.id,
    requestBody: updates,
  });

  return NewMCPResponse(JSON.stringify({
    success: true,
    label: res.data,
    message: `Label updated successfully`,
    updatedFields: Object.keys(updates)
  }));
}