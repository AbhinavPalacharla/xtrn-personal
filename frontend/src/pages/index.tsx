"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import type { UIMessage } from "ai";
import { useState } from "react";

// No tools passed: use the base UIMessage type
type Msg = UIMessage;
type Part = Msg["parts"][number];

export default function Chat() {
  const [input, setInput] = useState("");

  const { messages, sendMessage } = useChat<Msg>({
    transport: new DefaultChatTransport({
      api: "http://localhost:8080/chats",
      prepareSendMessagesRequest: ({ messages }) => ({
        body: { message: messages[messages.length - 1] },
      }),
    }),
    onError: (e) => console.error("useChat error:", e),
  });

  const renderPart = (part: Part, key: string) => {
    switch (part.type) {
      case "text":
        return <div key={key}>{part.text}</div>;

      case "dynamic-tool": {
        const name = part.toolName ?? "tool";
        switch (part.state) {
          case "input-streaming":
            return (
              <div key={key}>
                <em>{name}</em>: preparing input…
              </div>
            );
          case "input-available":
            return (
              <div key={key}>
                <strong>↪ {name} request</strong>
                <pre className="whitespace-pre-wrap break-words">{JSON.stringify(part.input, null, 2)}</pre>
              </div>
            );
          case "output-available":
            return (
              <div key={key}>
                <strong>⬅ {name} response</strong>
                <pre className="whitespace-pre-wrap break-words">{JSON.stringify(part.output, null, 2)}</pre>
              </div>
            );
          case "output-error":
            return (
              <div key={key} style={{ color: "crimson" }}>
                <strong>{name} error:</strong> {String(part.errorText ?? "An error occurred")}
              </div>
            );
          default:
            return <div key={key}>{name}…</div>;
        }
      }

      // Optional: show step boundaries if your backend emits them
      case "step-start":
        return <hr key={key} />;

      default:
        // Debug fallback so you can see any other parts you stream (sources, files, etc.)
        return (
          <pre key={key} className="whitespace-pre-wrap break-words text-xs opacity-70">
            {JSON.stringify(part, null, 2)}
          </pre>
        );
    }
  };

  return (
    <div>
      {messages.map((m, mi) => (
        <div key={m.id ?? mi} style={{ marginBottom: 16 }}>
          <div style={{ fontWeight: 600, marginBottom: 4 }}>{m.role === "user" ? "User" : "AI"}:</div>
          {m.parts.map((p, pi) => renderPart(p as Part, `${m.id ?? mi}-${pi}`))}
        </div>
      ))}

      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (input.trim()) {
            sendMessage({ text: input });
            setInput("");
          }
        }}
      >
        <input value={input} placeholder="say something..." onChange={(e) => setInput(e.currentTarget.value)} />
        <button type="submit">Enter</button>
      </form>
    </div>
  );
}
