package additional

import (
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func MenuNavigate(_lang string, db *gorm.DB, _chatID int64, bot *tgbotapi.BotAPI) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ushield_additional_services_sim"], "click_sim"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ushield_additional_services_visa"], "click_visa"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ushield_additional_services_energy_financing"], "click_energy_financing"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ushield_additional_services_sns"], "click_sns"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["ushield_additional_services_ecs"], "click_ecs"),
		),
	)

	dictRepo := repositories.NewSysDictionariesRepo(db)
	ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
	ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

	msg := tgbotapi.NewMessage(_chatID, global.Translations[_lang]["ushield_additional_services_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"

	bot.Send(msg)
}
