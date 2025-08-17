"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import { useState } from "react";

export default function Chat() {
  const [input, setInput] = useState("");
  const { messages, sendMessage } = useChat({
    transport: new DefaultChatTransport({
      api: "http://localhost:8080/hello",
      prepareSendMessagesRequest: ({ messages }) => {
        return {
          body: {
            message: messages[messages.length - 1],
          },
        };
      },
    }),
  });

  return (
    <div>
      {messages.map((message, index) => (
        <div key={index}>
          {message.role === "user" ? "User: " : "AI: "}
          {message.parts.map((part) => {
            switch (part.type) {
              case "text":
                return <div key={`${message.id}-text`}>{part.text}</div>;
            }
          })}
        </div>
      ))}

      <form
        onSubmit={(e) => {
          e.preventDefault();
          sendMessage({ text: input });
          setInput("");
        }}
      >
        <input value={input} placeholder="say something..." onChange={(e) => setInput(e.currentTarget.value)} />
        <button type="submit">Enter</button>
      </form>
    </div>
  );
}
