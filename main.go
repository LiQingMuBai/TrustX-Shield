package main

import (
	"fmt"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 存储用户的选择状态（实际应用中应使用数据库或缓存）
var userSelections = make(map[int64]map[int]bool) // chatID -> 选项ID -> 是否选中

func main() {
	bot, _ := tgbotapi.NewBotAPI("8325896700:AAFHojiHdLiTFYFuO27fhVNudqSa_Ptz8gc")

	for update := range bot.GetUpdatesChan(tgbotapi.NewUpdate(0)) {
		if update.CallbackQuery != nil {
			query := update.CallbackQuery
			data := query.Data // 格式如: "check:1", "check:2"
			chatID := query.Message.Chat.ID
			messageID := query.Message.MessageID

			fmt.Printf("messageID: %d, data: %s\n", messageID, data)

			// 解析操作
			if strings.HasPrefix(data, "check:") {
				optionID := 0
				_, _ = fmt.Sscanf(data, "check:%d", &optionID)

				// 初始化用户选择映射
				if _, exists := userSelections[chatID]; !exists {
					userSelections[chatID] = make(map[int]bool)
				}

				// 切换状态
				currentState := userSelections[chatID][optionID]
				userSelections[chatID][optionID] = !currentState

				// 生成新的键盘
				newKeyboard := buildCheckboxKeyboard(userSelections[chatID])

				// 编辑消息，更新按钮
				editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, newKeyboard)
				bot.Send(editMsg)

				// 回应点击（可选提示）
				//bot.AnswerCallbackQuery(tgbotapi.NewCallbackQuery(query.ID, "状态已更新"))
			}
		}

		if update.Message != nil {
			// 发送初始的可勾选按钮
			keyboard := buildCheckboxKeyboard(nil)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请选择选项：")
			msg.ReplyMarkup = &keyboard
			bot.Send(msg)
		}
	}
}

// 构建带有勾选状态的键盘
func buildCheckboxKeyboard(selected map[int]bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 1; i <= 3; i++ {
		label := fmt.Sprintf("选项 %d", i)
		if selected != nil && selected[i] {
			label = "✅ " + label
		} else {
			label = "□ " + label // 或者用空格
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("check:%d", i))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
