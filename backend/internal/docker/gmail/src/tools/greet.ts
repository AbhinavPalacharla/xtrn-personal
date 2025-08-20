import { z } from "zod";
import { type InferSchema } from "xmcp";
import { NewMCPResponse } from "../utils/error";

export const schema = {
  name: z.string().min(1).max(100).optional().describe("Name to greet (optional, 1-100 characters). If provided, will personalize the greeting message"),
};

export const metadata = {
  name: "greet",
  description: "A simple greeting tool to test the Gmail MCP server functionality and verify authentication is working properly",
  annotations: {
    title: "Greet user",
    readOnlyHint: true,
    idempotentHint: true,
  },
};

export default async function greet(args: InferSchema<typeof schema>) {
  const greeting = args.name ? `Hello, ${args.name}! Welcome to the Gmail MCP server.` : "Hello! Welcome to the Gmail MCP server.";
  
  return NewMCPResponse(greeting);
}