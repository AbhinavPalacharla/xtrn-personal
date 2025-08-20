import { z } from "zod";
import { type ToolMetadata, type InferSchema } from "xmcp";
import { NewMCPError, NewMCPLLMErrorResponse, NewMCPResponse } from "../utils/error";

// Define the schema for tool parameters
export const schema = {
  name: z.string().describe("The name of the user to greet"),
};

// Define tool metadata
export const metadata: ToolMetadata = {
  name: "greet",
  description: "Greet the user",
  annotations: {
    title: "Greet the user",
    readOnlyHint: true,
    destructiveHint: false,
    idempotentHint: true,
  },
};

// Tool implementation
export default async function greet({ name }: InferSchema<typeof schema>) {
  try {
    const result = `Hello, ${name}!`;

    return NewMCPResponse(result);
  } catch (e: unknown) {
    // fallback for unexpected errors that LLM can handle
    if (e instanceof Error) {
      return NewMCPLLMErrorResponse(`${e.name} - ${e.message}`);
    }

    // fallback for unknown errors that LLM cannot handle
    return NewMCPError("UNKNOWN_ERROR", "An unknown error occurred");
  }
}
