import type { TelegramRequest, TelegramResponse } from "../types";

export function chatId(req: TelegramRequest): TelegramResponse {
  return {
    method: "sendMessage",
    chat_id: req.message!.chat.id,
    parse_mode: "HTML",
    text: `Your chat ID: <code>${req.message!.chat.id}</code>`,
  };
}
