type MCPResponse = {
  isError: boolean;
  content: Array<{
    type: "text";
    text: string;
  }>;
};

type XtrnErrorTypes =
  | "AUTH_ERROR"
  | "AUTH_INVALID_GRANT"
  | "AUTH_MISSING_FIELDS"
  | "AUTH_UNKNOWN_ERROR"
  | "UNKNOWN_ERROR";

type XtrnHeader = {
  /**
    ERROR = MCP Level error - must be handled externally
    LLM_ERROR_RESPONSE = Error that LLM is capable of handling i.e. missing args, etc.
    RESPONSE = Successful tool call response
  */
  xtrn_message_type: "ERROR" | "RESPONSE" | "LLM_ERROR_RESPONSE";
  error_type?: XtrnErrorTypes;
  message?: string;
};

const NewMCPError = (errorType: XtrnErrorTypes, errorMessage: string): MCPResponse => {
  const xtrnHeader: XtrnHeader = {
    xtrn_message_type: "ERROR",
    error_type: errorType,
    message: errorMessage,
  };

  return {
    isError: true,
    content: [
      {
        type: "text",
        text: JSON.stringify(xtrnHeader),
      },
    ],
  };
};

const NewMCPResponse = (res: string): MCPResponse => {
  const xtrnHeader: XtrnHeader = {
    xtrn_message_type: "RESPONSE",
  };

  return {
    isError: false,
    content: [
      {
        type: "text",
        text: JSON.stringify(xtrnHeader),
      },
      {
        type: "text",
        text: res,
      },
    ],
  };
};

const NewMCPLLMErrorResponse = (res: string): MCPResponse => {
  const xtrnHeader: XtrnHeader = {
    xtrn_message_type: "LLM_ERROR_RESPONSE",
  };

  return {
    isError: true,
    content: [
      {
        type: "text",
        text: JSON.stringify(xtrnHeader),
      },
      {
        type: "text",
        text: res,
      },
    ],
  };
};

export class AuthError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "AuthError";
    Object.setPrototypeOf(this, AuthError.prototype);
  }
}

export class MissingTokenFieldsError extends AuthError {
  constructor() {
    super("Missing access token or expiry date from Google");
    this.name = "MissingTokenFieldsError";
    Object.setPrototypeOf(this, MissingTokenFieldsError.prototype);
  }
}

export class InvalidGrantError extends AuthError {
  constructor() {
    super("Refresh token is invalid or revoked");
    this.name = "InvalidGrantError";
    Object.setPrototypeOf(this, InvalidGrantError.prototype);
  }
}

export class UnknownAuthError extends AuthError {
  constructor(message?: string) {
    super(message || "Unknown error during token refresh");
    this.name = "UnknownAuthError";
    Object.setPrototypeOf(this, UnknownAuthError.prototype);
  }
}

/**
 * Type guard to check if error is MissingTokenFieldsError
 */
export function isMissingTokenFieldsError(error: unknown): error is MissingTokenFieldsError {
  return error instanceof MissingTokenFieldsError;
}

/**
 * Type guard to check if error is InvalidGrantError
 */
export function isInvalidGrantError(error: unknown): error is InvalidGrantError {
  return error instanceof InvalidGrantError;
}

/**
 * Type guard to check if error is UnknownAuthError
 */
export function isUnknownAuthError(error: unknown): error is UnknownAuthError {
  return error instanceof UnknownAuthError;
}

/**
 * Handle auth errors and convert to MCP error response
 */
export function handleAuthErrorToMCP(error: unknown) {
  if (isInvalidGrantError(error)) {
    return NewMCPError("AUTH_INVALID_GRANT", error.message);
  }

  if (isMissingTokenFieldsError(error)) {
    return NewMCPError("AUTH_MISSING_FIELDS", error.message);
  }

  if (isUnknownAuthError(error)) {
    return NewMCPError("AUTH_UNKNOWN_ERROR", error.message);
  }

  // Not an auth error, return undefined for caller to handle
  return undefined;
}

export { NewMCPError, NewMCPResponse, NewMCPLLMErrorResponse };

// type MCPResponse = {
//   isError: boolean;
//   content: Array<{
//     type: "text";
//     text: string;
//   }>;
// };

// type XtrnErrorTypes =
//   | "AUTH_ERROR"
//   | "AUTH_INVALID_GRANT"
//   | "AUTH_MISSING_FIELDS"
//   | "AUTH_UNKNOWN_ERROR"
//   | "UNKNOWN_ERROR";

// type XtrnHeader = {
//   xtrn_message_type: "ERROR" | "RESPONSE";
//   error_type?: XtrnErrorTypes;
//   message?: string;
// };

// const NewMCPError = (errorType: XtrnErrorTypes, errorMessage: string): MCPResponse => {
//   const xtrnHeader: XtrnHeader = {
//     xtrn_message_type: "ERROR",
//     error_type: errorType,
//     message: errorMessage,
//   };

//   return {
//     isError: true,
//     content: [
//       // XTRN header as first content field always
//       {
//         type: "text",
//         text: JSON.stringify(xtrnHeader),
//       },
//     ],
//   };
// };

// const NewMCPResponse = (res: string): MCPResponse => {
//   const xtrnHeader: XtrnHeader = {
//     xtrn_message_type: "RESPONSE",
//   };

//   return {
//     isError: false,
//     content: [
//       {
//         type: "text",
//         text: JSON.stringify(xtrnHeader),
//       },
//       {
//         type: "text",
//         text: res,
//       },
//     ],
//   };
// };

// export class AuthError extends Error {
//   constructor(message: string) {
//     super(message);
//     this.name = "AuthError";
//     Object.setPrototypeOf(this, AuthError.prototype);
//   }
// }

// export class MissingTokenFieldsError extends AuthError {
//   constructor() {
//     super("Missing access token or expiry date from Google");
//     this.name = "MissingTokenFieldsError"; // assign here, no declaration
//     Object.setPrototypeOf(this, MissingTokenFieldsError.prototype);
//   }
// }

// export class InvalidGrantError extends AuthError {
//   constructor() {
//     super("Refresh token is invalid or revoked");
//     this.name = "InvalidGrantError";
//     Object.setPrototypeOf(this, InvalidGrantError.prototype);
//   }
// }

// export class UnknownAuthError extends AuthError {
//   constructor(message?: string) {
//     super(message || "Unknown error during token refresh");
//     this.name = "UnknownAuthError";
//     Object.setPrototypeOf(this, UnknownAuthError.prototype);
//   }
// }

// export function handleAuthErrorToMCP(error: unknown) {
//   if (error instanceof InvalidGrantError) {
//     return NewMCPError("AUTH_INVALID_GRANT", error.message);
//   }

//   if (error instanceof MissingTokenFieldsError) {
//     return NewMCPError("AUTH_MISSING_FIELDS", error.message);
//   }

//   if (error instanceof UnknownAuthError) {
//     return NewMCPError("AUTH_UNKNOWN_ERROR", error.message);
//   }

//   // Not an auth error, return undefined so caller can handle or fallback
//   return undefined;
// }
//////////////////////////////////////////////////////
// type AppError = { type: "invalid_grant"; message: string } | { type: "missing_token_fields"; message: string };

// const isAppError = (err: unknown): err is AppError => {
//   return typeof err === "object" && err !== null && "type" in err && typeof (err as any).type === "string";
// };

// type MCPResponse = {
//   isError: boolean;
//   content: Array<{
//     type: "text";
//     text: string;
//   }>;
// };

// type XtrnErrorTypes = "AUTH_ERROR";

// type XtrnHeader = {
//   xtrn_message_type: "ERROR" | "RESPONSE";
//   error_type?: XtrnErrorTypes;
//   message?: string;
// };

// const NewMCPError = (errorType: XtrnErrorTypes, errorMessage: string): MCPResponse => {
//   const xtrnHeader: XtrnHeader = {
//     xtrn_message_type: "ERROR",
//     error_type: errorType,
//     message: errorMessage,
//   };

//   return {
//     isError: true,
//     content: [
//       // XTRN header as first content field always
//       {
//         type: "text",
//         text: JSON.stringify(xtrnHeader),
//       },
//     ],
//   };
// };

// const NewMCPResponse = (res: string): MCPResponse => {
//   const xtrnHeader: XtrnHeader = {
//     xtrn_message_type: "RESPONSE",
//   };

//   return {
//     isError: false,
//     content: [
//       {
//         type: "text",
//         text: JSON.stringify(xtrnHeader),
//       },
//       {
//         type: "text",
//         text: res,
//       },
//     ],
//   };

//   // res.content = [
//   //   {
//   //     type: "text",
//   //     text: JSON.stringify({
//   //       xtrn_message_type: "RESPONSE",
//   //     }),
//   //   },
//   //   ...res.content,
//   // ];
//   // return res;
// };

// // const NewMCPResponse = (res: object) => {
// //   return {
// //     isError: false,
// //     content: [
// //       // XTRN header
// //       {
// //         type: "text",
// //         text: JSON.stringify({
// //           xtrn_message_type: "RESPONSE",
// //         }),
// //       },
// //     ],
// //   };
// // };

// // const NewMCPError = (errorType: string, errorMessage: string) => {
// //   return {
// //     isError: false,
// //     content: [
// //       {
// //         type: "text",
// //         text: JSON.stringify({
// //           is_error: true,
// //           content: [
// //             {
// //               type: "error",
// //               error_code: errorType,
// //               error: errorMessage,
// //             },
// //           ],
// //         }),
// //       },
// //     ],
// //   };
// // };

// // const NewMCPResponse = (res: object) => {
// //   return {
// //     isError: false,
// //     content: [
// //       {
// //         type: "text",
// //         text: JSON.stringify(res),
// //       },
// //     ],
// //   };
// // };

// export { isAppError, NewMCPError, AppError, NewMCPResponse };
