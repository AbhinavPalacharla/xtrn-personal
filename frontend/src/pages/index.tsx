import { useChat } from "@ai-sdk/react";
import { useState } from "react";

const API_BASE_URL = "http://127.0.0.1:8080";
const USER_ID = "4ZV1dBca-cm1rKWg4RyVo";

const Index = () => {
  const [chatID, setChatID] = useState<string | null>(null);

  const { messages, input, handleInputChange, handleSubmit, isLoading } = useChat({
    api: chatID ? `${API_BASE_URL}/chat/${chatID}` : `${API_BASE_URL}/chat`,
    headers: {
      "x-xtrn-user-id": USER_ID,
    },
    experimental_prepareRequestBody: ({ messages }) => {
      return messages[messages.length - 1].content;
    },
    initialMessages: [],
    onResponse: (res) => {
      const chatIDHeader = res.headers.get("x-xtrn-chat-id");
      console.log(`Chat ID Header recieved - ${chatIDHeader}`);
      if (chatIDHeader) {
        setChatID(chatIDHeader);
      }
      console.log(`SET CHAT ID: ${chatID}`);
    },
  });

  // return (
  //   <div>
  //     <h1>Hello</h1>
  //   </div>
  // );
  //
  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      {/* Show chatID if present */}
      {chatID && (
        <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-2 rounded mb-2 text-xs">
          <strong>Chat ID:</strong> <code>{chatID}</code>
        </div>
      )}
      <ul className="space-y-4">
        {messages.map((m) => (
          <li key={m.id} className="space-y-2">
            {m.role === "user" ? (
              <div className="bg-blue-100 p-3 rounded text-right ml-auto max-w-[80%]">
                <div className="text-xs text-gray-500 uppercase mb-1">You</div>
                <div>{m.content}</div>
              </div>
            ) : (
              <div className="bg-gray-100 p-3 rounded text-left mr-auto max-w-[80%]">
                <div className="text-xs text-gray-500 uppercase mb-1">{m.role}</div>
                {/* If you have structured parts, you could map here. */}
                <div className="whitespace-pre-wrap">{m.content}</div>
              </div>
            )}
          </li>
        ))}
      </ul>

      <form onSubmit={handleSubmit} className="flex gap-2 items-center">
        <input
          value={input}
          onChange={handleInputChange}
          placeholder="Type a message..."
          className="flex-1 border px-3 py-2 rounded"
          disabled={isLoading}
        />
        <button
          type="submit"
          disabled={isLoading || !input.trim()}
          className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          {isLoading ? "..." : "Send"}
        </button>
      </form>
    </div>
  );
};

export default Index;

// import { useChat } from "@ai-sdk/react";
// import { useState } from "react";

// const Index = () => {
//   const [chatID, setChatID] = useState<string | null>(null);

//   const { messages, input, handleInputChange, isLoading } = useChat({
//     api: chatID ? `http://localhost:8080/chat/${chatID}` : "/chat",
//     headers: {
//       "x-xtrn-user-id": "4ZV1dBca-cm1rKWg4RyVo",
//     },
//     initialMessages: [],
//     onResponse: (res) => {
//       const retChatID = res.headers.get("x-xtrn-chat-id");
//       if (retChatID && !chatID) {
//         setChatID(retChatID);
//         window.history.replaceState(null, "", `/chat/${retChatID}`);
//       }
//     },
//   });

//   return (
//     <div className="max-w-2xl mx-auto p-4 space-y-4">
//       <ul className="space-y-4">
//         {messages.map((m) => (
//           <li key={m.id} className="space-y-2">
//             {m.role === "user" ? (
//               <div className="bg-blue-100 p-3 rounded text-right ml-auto max-w-[80%]">
//                 <div className="text-xs text-gray-500 uppercase mb-1">You</div>
//                 <div>{m.content}</div>
//               </div>
//             ) : (
//               <div className="bg-gray-100 p-3 rounded text-left mr-auto max-w-[80%]">
//                 <div className="text-xs text-gray-500 uppercase mb-1">{m.role}</div>
//                 {m.parts?.map((part, idx) => {
//                   if (part.type === "text") {
//                     return (
//                       <p key={idx} className="whitespace-pre-wrap">
//                         {part.text}
//                       </p>
//                     );
//                   }

//                   if (part.type === "tool-invocation") {
//                     const call = part.toolInvocation;
//                     return (
//                       <div
//                         key={idx}
//                         className="bg-yellow-100 border-l-4 border-yellow-400 px-4 py-2 rounded text-sm my-2"
//                       >
//                         🛠 <strong>{call.toolName}</strong> (ID: <code>{call.toolCallId}</code>)<br />
//                         <div className="text-xs text-gray-600">State: {call.state}</div>
//                         <div className="mt-1">
//                           <strong>Args:</strong>
//                           <pre className="bg-white border p-2 rounded text-sm overflow-x-auto">
//                             {JSON.stringify(call.args, null, 2)}
//                           </pre>
//                         </div>
//                         {"result" in call && call.result && (
//                           <div className="mt-2">
//                             ✅ <strong>Result:</strong>
//                             <pre className="bg-white border p-2 rounded text-sm overflow-x-auto">
//                               {JSON.stringify(call.result, null, 2)}
//                             </pre>
//                           </div>
//                         )}
//                       </div>
//                     );
//                   }

//                   if (part.type === "reasoning") {
//                     return (
//                       <div key={idx} className="italic text-sm text-gray-600">
//                         🤔 {part.reasoning}
//                       </div>
//                     );
//                   }

//                   return null;
//                 })}
//               </div>
//             )}
//           </li>
//         ))}
//       </ul>

//       <form
//         onSubmit={async (e: React.FormEvent) => {
//           e.preventDefault();

//           const newMessage = {
//             role: "user",
//             content: input,
//           };

//           const res = await fetch(chatID ? `http://localhost:8080/chat/${chatID}` : `http://localhost:8080/chat`, {
//             headers: {
//               "Content-Type": "application/json",
//             },
//             body: JSON.stringify({ message: newMessage }),
//           });

//           const retChatID = res.headers.get("x-xtrn-chat-id");

//           if (retChatID && !chatID) {
//             setChatID(retChatID);
//             window.history.replaceState(null, "", `/chat/${retChatID}`);
//           }
//         }}
//         className="flex gap-2 items-center"
//       >
//         <input
//           value={input}
//           onChange={handleInputChange}
//           placeholder="Type a message..."
//           className="flex-1 border px-3 py-2 rounded"
//         />
//         <button
//           type="submit"
//           disabled={isLoading}
//           className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
//         >
//           {isLoading ? "..." : "Send"}
//         </button>
//       </form>
//     </div>
//   );
// };

// export default Index;
