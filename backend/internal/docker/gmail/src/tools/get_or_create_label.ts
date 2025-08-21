import { z } from "zod";
import { type InferSchema } from "xmcp";
import { gmailOauthClient } from "../auth";
import { google } from "googleapis";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  name: z.string().min(1).max(100).describe("Name of the label to find or create (required, 1-100 characters)"),
  messageListVisibility: z.enum(["show", "hide"]).optional().describe("Whether to show the label in the message list (optional, default: 'show')"),
  labelListVisibility: z.enum(["labelShow", "labelHide"]).optional().describe("Whether to show the label in the label list (optional, default: 'labelShow')"),
};

export const metadata = {
  name: "get_or_create_label",
  description: "Gets an existing label by name or creates it if it doesn't exist. This is useful for ensuring a label exists before using it in operations.",
  annotations: {
    title: "Get or create Gmail label",
    readOnlyHint: false,
    idempotentHint: false,
  },
};

export default async function getOrCreateLabel(args: InferSchema<typeof schema>) {
  const client = await gmailOauthClient();
  const gmail = google.gmail({ version: "v1", auth: client });

  try {
    // First, try to find existing labels
    const labelsResponse = await gmail.users.labels.list({
      userId: "me",
    });

    const existingLabel = labelsResponse.data.labels?.find(
      (label) => label.name === args.name
    );

    if (existingLabel) {
      return NewMCPResponse(JSON.stringify({
        success: true,
        label: existingLabel,
        action: "found existing",
        message: `Found existing label '${args.name}' with ID: ${existingLabel.id}`,
        type: existingLabel.type
      }));
    }

    // Create new label if it doesn't exist
    const labelRequest: any = {
      name: args.name,
      messageListVisibility: args.messageListVisibility ?? "show",
      labelListVisibility: args.labelListVisibility ?? "labelShow",
    };

    const newLabel = await gmail.users.labels.create({
      userId: "me",
      requestBody: labelRequest,
    });

    return NewMCPResponse(JSON.stringify({
      success: true,
      label: newLabel.data,
      action: "created new",
      message: `Created new label '${args.name}' with ID: ${newLabel.data.id}`,
      type: newLabel.data.type
    }));
  } catch (error: any) {
    throw new Error(`Failed to get or create label: ${error.message}`);
  }
}