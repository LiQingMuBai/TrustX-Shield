package service

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
	"strconv"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
)

func ExtractAddressManager(_lang string, message *tgbotapi.Message, db *gorm.DB, bot *tgbotapi.BotAPI) {
	if IsValidAddress(message.Text) || IsValidEthereumAddress(message.Text) {
		userRepo := repositories.NewUserAddressMonitorRepo(db)
		var record domain.UserAddressMonitor
		record.ChatID = message.Chat.ID
		record.Address = message.Text
		record.Status = 1
		if IsValidAddress(message.Text) {
			record.Network = "tron"
		}
		if IsValidAddress(message.Text) {
			record.Network = "ethereum"
		}
		errsg := userRepo.Create(context.Background(), &record)
		if errsg != nil {
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_added_success"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
	}
}

func ADDRESS_LIST_TRACE(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB) {
	userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)
	addresses, _ := userAddressEventRepo.Query(context.Background(), callbackQuery.Message.Chat.ID)
	// 初始化结果字符串
	var result string

	if len(addresses) > 0 {
		// 遍历数组并拼接字符串
		for i, item := range addresses {
			if i > 0 {
				result += " ✅\n" // 添加分隔符
			}

			restDays := fmt.Sprintf("%d", 31-item.Days)

			result += "<code>" + item.Address + "</code>" + global.Translations[_lang]["remaining_days_1"] + restDays + global.Translations[_lang]["remaining_days_2"]
		}
		result += " ✅\n\n" // 添加分隔符
	} else {
		result += "\n" + global.Translations[_lang]["NAACUAMS"] + "\n\n"
	}
	//查看余额
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
	if IsEmpty(user.Amount) {
		user.Amount = "0"
	}

	if IsEmpty(user.TronAmount) {
		user.TronAmount = "0"
	}

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "📊 "+global.Translations[_lang]["currently_monitoring_addresses"]+"\n"+
		result+
		//"💰 当前余额："+"\n- "+user.TronAmount+" TRX \n - "+user.Amount+" USDT \n"+
		"💰"+global.Translations[_lang]["balance"]+": "+" "+
		"-TRX： "+user.TronAmount+"    "+
		"-USDT： "+user.Amount+"\n"+
		global.Translations[_lang]["freeze_alert_service_monitoring_tips"])
	msg.ParseMode = "HTML"

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🛑"+global.Translations[_lang]["stop_monitoring"], "stop_freeze_risk"),
			tgbotapi.NewInlineKeyboardButtonData("🔗"+global.Translations[_lang]["secondary_contact"], "click_backup_account"),
			//tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
			//tgbotapi.NewInlineKeyboardButtonData("地址管理", "user_backup_notify"),
		),
		tgbotapi.NewInlineKeyboardRow(

			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "address_list_trace", expiration)
}

func ADDRESS_MANAGER(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, chatID int64, db *gorm.DB) {
	userAddressRepo := repositories.NewUserAddressMonitorRepo(db)

	addresses, _ := userAddressRepo.Query(context.Background(), chatID)

	result := ""
	for _, item := range addresses {
		result += "<code>" + item.Address + "</code>" + "\n"
	}
	msg := tgbotapi.NewMessage(chatID, "预警地址列表"+"\n"+result)
	//地址绑定

	msg.ParseMode = "HTML"

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[_lang]["add_address"], "address_manager_add"),
			//tgbotapi.NewInlineKeyboardButtonData("设置钱包", "address_manager"),
			tgbotapi.NewInlineKeyboardButtonData("➖"+global.Translations[_lang]["remove_address"], "address_manager_remove"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(chatID, 10), "address_manager", expiration)
}
