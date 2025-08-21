import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  id: z.string().min(1).describe("The unique ID of the label to delete (required, cannot be empty)"),
};

export const metadata = {
  name: "delete_label",
  description: "Deletes a Gmail label. Note that system labels (INBOX, SENT, DRAFT, etc.) cannot be deleted. Only user-created labels can be removed.",
  annotations: {
    title: "Delete Gmail label",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function deleteLabel(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  try {
    await gmail.users.labels.delete({
      userId: "me",
      id: args.id,
    });

    return NewMCPResponse(JSON.stringify({
      success: true,
      message: `Label ${args.id} deleted successfully`,
      deletedLabelId: args.id
    }));
  } catch (error: any) {
    if (error.code === 400 && error.message.includes("system label")) {
      throw new Error("Cannot delete system labels (INBOX, SENT, DRAFT, etc.). Only user-created labels can be deleted.");
    }
    throw error;
  }
}