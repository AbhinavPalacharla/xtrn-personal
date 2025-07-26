type AppError = { type: "invalid_grant"; message: string } | { type: "missing_token_fields"; message: string };

const isAppError = (err: unknown): err is AppError => {
  return typeof err === "object" && err !== null && "type" in err && typeof (err as any).type === "string";
};

const NewMCPError = (errorType: string, errorMessage: string) => {
  return {
    isError: false,
    content: [
      {
        type: "text",
        text: JSON.stringify({
          is_error: true,
          content: [
            {
              type: "error",
              error_code: errorType,
              error: errorMessage,
            },
          ],
        }),
      },
    ],
  };
};

const NewMCPResponse = (res: object) => {
  return {
    isError: false,
    content: [
      {
        type: "text",
        text: JSON.stringify(res),
      },
    ],
  };
};

export { isAppError, NewMCPError, AppError, NewMCPResponse };
