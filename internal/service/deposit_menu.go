package service

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
)

func DEPOSIT_AMOUNT(_lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	trxSubscriptionsRepo := repositories.NewUserTRXSubscriptionsRepository(db)

	trxlist, _ := trxSubscriptionsRepo.ListAll(context.Background())

	var allButtons []tgbotapi.InlineKeyboardButton
	var extraButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, trx := range trxlist {
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData("💰"+trx.Name, "deposit_trx_"+trx.Amount))
	}

	extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[_lang]["switch_to_usdt_deposit"], "forward_deposit_usdt"), tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"))

	for i := 0; i < len(allButtons); i += 2 {
		end := i + 2
		if end > len(allButtons) {
			end = len(allButtons)
		}
		row := allButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}

	for i := 0; i < len(extraButtons); i += 1 {
		end := i + 1
		if end > len(extraButtons) {
			end = len(extraButtons)
		}
		row := extraButtons[i:end]
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(row...))
	}

	// 3. 创建键盘标记
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	userRepo := repositories.NewUserRepository(db)

	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
		"🆔"+global.Translations[_lang]["user_id"]+": "+user.Associates+"\n"+
			"👤"+global.Translations[_lang]["username"]+": @"+user.Username+"\n"+
			"💰"+global.Translations[_lang]["balance"]+": "+"\n"+
			"- TRX：   "+user.TronAmount+"\n"+
			"-  USDT："+user.Amount)

	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}
