package catfee

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	_rd "ushield_bot/internal/infrastructure/3rd"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"
	"ushield_bot/internal/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func CustodyAddressCond(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["catfee_custody_address_tips"]+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
}

func CustodyAddressAdd(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	_address := message.Text
	_chatID := message.Chat.ID
	if !tools.IsValidAddress(_address) {
		msg := tgbotapi.NewMessage(_chatID, "❌"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	//要查下是否已经有绑定的地址

	total, _ := userSmartTransactionAddressesRepo.Count(context.Background(), _chatID)

	if total >= 5 {
		msg := tgbotapi.NewMessage(_chatID, "<b>"+global.Translations[_lang]["catfee_energy_address_limit_tips"]+"</b>"+"\n")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return
	}
	record, _ := userSmartTransactionAddressesRepo.Query(context.Background(), _address)

	if record.ID > 0 {
		msg := tgbotapi.NewMessage(_chatID, "❌"+"<b>"+global.Translations[_lang]["catfee_add_address_already_exit_tips"]+"</b>"+"\n")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		return
	}
	var userAddress domain.UserSmartTransactionAddresses

	userAddress.Status = "0"
	userAddress.CreatedAt = time.Now()
	userAddress.ChatID = strconv.FormatInt(_chatID, 10)
	userAddress.QuotaMode = "UNLIMITED"
	userAddress.Address = _address
	userSmartTransactionAddressesRepo.Create(context.Background(), &userAddress)

	//添加成功
	msg := tgbotapi.NewMessage(_chatID, "✅"+"<b>"+global.Translations[_lang]["address_added_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func CustodyRemoveAddressCond(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["energy_address_remove_tips"]+"\n")
	msg.ParseMode = "HTML"
	bot.Send(msg)
	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
}

func CustodyAddressRemove(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, message *tgbotapi.Message, catfee *_rd.CatfeeService) {

	_address := message.Text
	_chatID := message.Chat.ID
	fmt.Printf("删除用户id %d，地址 %s\v", _chatID, _address)
	if !tools.IsValidAddress(_address) {
		msg := tgbotapi.NewMessage(_chatID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return
	}

	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	err := userSmartTransactionAddressesRepo.Remove(context.Background(), strconv.FormatInt(_chatID, 10), _address)

	if err != nil {
		fmt.Printf("删除地址失败%v\n", err)
	}

	code, err := catfee.MateOpenBasicDelete(_address)

	if err != nil {
		fmt.Printf("catfee.MateOpenBasicDelete: %v\n", err)
	}
	fmt.Printf("catfee删除状态 %d\n", code)

	msg := tgbotapi.NewMessage(_chatID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func CustodyAddressDisable(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, catfee *_rd.CatfeeService) {

	_address := callbackQuery.Message.Text
	_chatID := callbackQuery.Message.Chat.ID
	fmt.Printf("暂停用户id %d，地址 %s\v", _chatID, _address)
	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	err := userSmartTransactionAddressesRepo.Disable(context.Background(), strconv.FormatInt(_chatID, 10), _address)

	if err != nil {
		fmt.Printf("暂停地址失败%v\n", err)
	}
	code, err := catfee.MateOpenBasicDisable(_address)

	if err != nil {

	}
	fmt.Printf("catfee暂停地址失败 %d\n", code)

	msg := tgbotapi.NewMessage(_chatID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func CustodyAddressEnable(_lang string, cache cache.Cache, db *gorm.DB, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, catfee *_rd.CatfeeService) {

	_address := callbackQuery.Message.Text
	_chatID := callbackQuery.Message.Chat.ID
	fmt.Printf("启用用户id %d，地址 %s\v", _chatID, _address)
	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)

	err := userSmartTransactionAddressesRepo.Enable(context.Background(), strconv.FormatInt(_chatID, 10), _address)

	if err != nil {
		fmt.Printf("启用地址失败%v\n", err)
	}
	code, err := catfee.MateOpenBasicDisable(_address)

	if err != nil {

	}
	fmt.Printf("catfee启用地址失败 %d\n", code)

	msg := tgbotapi.NewMessage(_chatID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	msg.ParseMode = "HTML"
	bot.Send(msg)

}

func CatfeeAddressPrevePage(_lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) (*global.DepositState, bool) {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]

	if state != nil && state.CurrentPage == 1 {
		return nil, true
	}
	if state == nil {
		var state global.DepositState
		state.CurrentPage = 1
		global.DepositStates[callbackQuery.Message.Chat.ID] = &state
		userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch

		info.Page = 1
		info.PageSize = 5
		trxlist, _, _ := userAddressDetectionRepo.GetUserSmartTransactionPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)

		var builder strings.Builder
		builder.WriteString("\n") // 添加分隔符
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[_lang]["笔"]))
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["deduction_records"]+"\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "prev_bundle_package_page"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "next_bundle_package_page"),
			),
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	} else {
		state.CurrentPage = state.CurrentPage - 1
		userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch
		info.PageInfo.Page = state.CurrentPage
		info.PageSize = 5
		trxlist, _, _ := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)
		var builder strings.Builder
		builder.WriteString("\n") // 添加分隔符
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[_lang]["笔"]))
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}

		// 去除最后一个空格
		result := strings.TrimSpace(builder.String())
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["deduction_records"]+"\n\n "+
			result+"\n")
		msg.ParseMode = "HTML"
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "prev_bundle_package_page"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "next_bundle_package_page"),
			),
			tgbotapi.NewInlineKeyboardRow(
				//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		bot.Send(msg)
	}
	return state, false
}
