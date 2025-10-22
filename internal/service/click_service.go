package service

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
	"ushield_bot/internal/request"
)

func CLICK_DEPOSIT_USDT_RECORDS(_lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)

	//trxDepositRepo := repositories.NewUserTRXDepositsRepository(db)
	var info request.UserUsdtDepositsSearch
	info.PageInfo.Page = 1
	info.PageInfo.PageSize = 10
	//trxlist, _, _ := trxDepositRepo.GetUserTrxDepositsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)
	usdtlist, _, _ := usdtDepositRepo.GetUserUsdtDepositsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)

	var builder strings.Builder
	builder.WriteString("\n") // 添加分隔符

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	for _, word := range usdtlist {
		builder.WriteString("[")
		builder.WriteString(word.CreatedDate)
		builder.WriteString("]")
		builder.WriteString("+")
		builder.WriteString(word.Amount)
		builder.WriteString(" USDT ")
		builder.WriteString(" （订单 #TOPUP- ")
		builder.WriteString(word.OrderNO)
		builder.WriteString("）")

		builder.WriteString("\n") // 添加分隔符
	}
	//
	//// 去除最后一个空格
	result = strings.TrimSpace(builder.String())
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["deposit_records"]+"\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "prev_deposit_usdt_page"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "next_deposit_usdt_page"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
}

func ClickBusinessCooperation(_lang string, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["promotion_link"]+"："+"https://t.me/ushield_bot?start="+strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)+"\n\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
}

func ClickOfficalChannel(_lang string, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["join_vip_cooperation"]+"https://t.me/ushield1\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
}

func ClickCallCenter(_lang string, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "📞"+global.Translations[_lang]["support"]+"：@TrustXService1\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
}

func CLICK_DEPOSIT_TRX_RECORDS(_lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	//usdtDepositRepo := repositories.NewUserUSDTDepositsRepository(db)
	//usdtlist, _ := usdtDepositRepo.ListAll(context.Background(), callbackQuery.Message.Chat.ID, 1)

	trxDepositRepo := repositories.NewUserTRXDepositsRepository(db)
	var info request.UserTrxDepositsSearch
	info.PageInfo.Page = 1
	info.PageInfo.PageSize = 10
	trxlist, _, _ := trxDepositRepo.GetUserTrxDepositsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)

	var builder strings.Builder
	builder.WriteString("\n") // 添加分隔符
	//- [6.29] +3000 TRX（订单 #TOPUP-92308）
	for _, word := range trxlist {
		builder.WriteString("[")
		builder.WriteString(word.CreatedDate)
		builder.WriteString("]")
		builder.WriteString("+")
		builder.WriteString(word.Amount)
		builder.WriteString(" TRX ")
		builder.WriteString(" （订单 #TOPUP- ")
		builder.WriteString(word.OrderNO)
		builder.WriteString("）")

		builder.WriteString("\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["deposit_records"]+"\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "prev_deposit_trx_page"),
			tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "next_deposit_trx_page"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
}
func CLICK_MY_RECEPIT(_lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["billing"]+"\n\n📌 "+
		global.Translations[_lang]["balance"]+"：\n\n- TRX："+user.TronAmount+"\n- USDT："+user.Amount+"\n")

	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬇️"+global.Translations[_lang]["trx_deposit_records"], "click_deposit_trx_records"),
			tgbotapi.NewInlineKeyboardButtonData("⬇️"+global.Translations[_lang]["usdt_deposit_records"], "click_deposit_usdt_records"),
		),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	bot.Send(msg)
}
