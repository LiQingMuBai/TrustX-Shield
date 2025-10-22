package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/global"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"
	. "ushield_bot/internal/infrastructure/tools"
	"ushield_bot/internal/request"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func ExtractBundlePackage(_lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {

	fmt.Println("ExtractBundlePackage")
	userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 5
	trxlist, total, err := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)
	if err != nil {

		fmt.Println("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[_lang]["笔"]))

			//builder.WriteString(" TRX ")
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}
	} else {
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
	return msg
}

func EXTRACT_NEXT_BUNDLE_PACKAGE_PAGE(_lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) bool {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]
	if state == nil {
		var state2 global.DepositState
		state2.CurrentPage = 1
		state = &state2
	}
	//if state != nil && state.CurrentPage > 1 {
	state.CurrentPage = state.CurrentPage + 1
	userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch
	info.PageInfo.Page = state.CurrentPage
	info.PageInfo.PageSize = 10
	trxlist, total, _ := userAddressDetectionRepo.GetUserPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)

	fmt.Printf("currentpage : %d", state.CurrentPage)
	fmt.Printf("total: %v\n", total)
	totalPages := (total + 5 - 1) / 5

	fmt.Printf("totalPages : %d", totalPages)
	if int64(state.CurrentPage) > totalPages {
		state.CurrentPage = totalPages
		return true
	}
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
	//}
	fmt.Printf("state: %v\n", state)

	global.DepositStates[callbackQuery.Message.Chat.ID] = state
	return false
}

func EXTRACT_PREV_BUNDLE_PACKAGE_PAGE(_lang string, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, bot *tgbotapi.BotAPI) (*global.DepositState, bool) {
	state := global.DepositStates[callbackQuery.Message.Chat.ID]

	if state != nil && state.CurrentPage == 1 {
		return nil, true
	}
	if state == nil {
		var state global.DepositState
		state.CurrentPage = 1
		global.DepositStates[callbackQuery.Message.Chat.ID] = &state
		userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
		var info request.UserAddressDetectionSearch

		info.Page = 1
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
func CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, _chatID int64, db *gorm.DB) {
	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	addresses, _ := userOperationPackageAddressesRepo.Query(context.Background(), _chatID)

	result := ""
	for _, item := range addresses {
		result += "<code>" + item.Address + "</code>"

		if len(item.Remark) > 0 {
			result += "[" + item.Remark + "]"
		}

		if item.Status == 1 {
			result += "[默认]"
		}
		result += "\n"
	}
	msg := tgbotapi.NewMessage(_chatID, "👇"+global.Translations[_lang]["transaction_package_address_list"]+"\n\n"+result+"\n\n")
	//地址绑定

	msg.ParseMode = "HTML"

	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("⚙地址设置", "click_bundle_package_address_manager_config"),
		//),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕"+global.Translations[_lang]["add_address"], "click_bundle_package_address_manager_add"),

			tgbotapi.NewInlineKeyboardButtonData("➖"+global.Translations[_lang]["remove_address"], "click_bundle_package_address_manager_remove"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(_chatID, 10), "null_bundle_package_address_manager", expiration)
}

func CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_CONFIG(_lang string, cache cache.Cache, bot *tgbotapi.BotAPI, _chatID int64, db *gorm.DB) {
	userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	addresses, _ := userOperationPackageAddressesRepo.Query(context.Background(), _chatID)

	msg := tgbotapi.NewMessage(_chatID, "👇请选择要设置的地址："+"\n")
	//地址绑定

	msg.ParseMode = "HTML"

	var allButtons []tgbotapi.InlineKeyboardButton
	var extraButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, item := range addresses {
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(TruncateString(item.Address), "config_bundle_package_address_"+item.Address))
	}

	extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"))

	for i := 0; i < len(allButtons); i += 1 {
		end := i + 1
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

	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(_chatID, 10), "null_bundle_package_address_manager", expiration)
}
func CONFIG_BUNDLE_PACKAGE_ADDRESS(_lang string, address string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB) {

	msg := tgbotapi.NewMessage(message.Chat.ID, "🔍正在设置地址："+address+"\n")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙设置默认", "set_bundle_package_default_"+address),
			tgbotapi.NewInlineKeyboardButtonData("➖"+global.Translations[_lang]["remove_address"], "remove_bundle_package_"+address),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(message.Chat.ID, 10), "config_bundle_package_address", expiration)
}
func APPLY_BUNDLE_PACKAGE_ADDRESS(_lang string, bundle_address string, cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB) {

	fmt.Printf("address %s\n", bundle_address)

	bundleID := strings.Split(bundle_address, "_")[0]
	address := strings.Split(bundle_address, "_")[1]

	fmt.Printf("address %s\n", address)
	fmt.Printf("bundle_id %s\n", bundleID)

	userOperationBundlesRepo := repositories.NewUserOperationBundlesRepository(db)
	bundlePackage, err := userOperationBundlesRepo.Query(context.Background(), bundleID)

	if err != nil {
		fmt.Println(err)
	}
	userRepo := repositories.NewUserRepository(db)
	user, _ := userRepo.GetByUserID(message.Chat.ID)

	//扣錢
	if bundlePackage.Token == "TRX" {

		balance, _ := tools.SubtractStringNumbers(user.TronAmount, bundlePackage.Amount, 1)
		fmt.Printf("TRX balance %s", balance)
		user.TronAmount = balance
	} else if bundlePackage.Token == "USDT" {
		balance, _ := tools.SubtractStringNumbers(user.Amount, bundlePackage.Amount, 1)
		fmt.Printf("USDT balance %s", balance)

		user.Amount = balance
	}

	err = userRepo.Update2(context.Background(), &user)
	if err != nil {

	}

	//加入訂閲記錄
	userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	var record domain.UserPackageSubscriptions
	record.ChatID = message.Chat.ID
	record.Address = address
	bundle, _ := strconv.ParseInt(bundleID, 10, 64)

	record.BundleID = bundle
	record.Status = 2
	record.Amount = bundlePackage.Amount
	record.Times = tools.ExtractLeadingInt64(bundlePackage.Name)
	record.BundleName = bundlePackage.Name
	err = userPackageSubscriptionsRepo.Create(context.Background(), &record)
	if err != nil {
		return
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"🧾"+global.Translations[_lang]["package_order_purchased_successfully"]+"\n\n"+
		global.Translations[_lang]["package_name"]+"："+strings.ReplaceAll(bundlePackage.Name, "笔", global.Translations[_lang]["笔"])+"\n\n"+
		global.Translations[_lang]["payment_amount"]+"："+bundlePackage.Amount+" "+bundlePackage.Token+"\n\n"+
		global.Translations[_lang]["address"]+"："+address+"\n\n"+
		global.Translations[_lang]["order_id"]+"："+fmt.Sprintf("%d", record.Id)+""+"\n\n")
	msg.ParseMode = "HTML"
	// 当点击"按钮 1"时显示内联键盘
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧾"+global.Translations[_lang]["package_address_list"], "click_bundle_package_address_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_apply_bundle_package_address", expiration)
}

func DispatchOthers(_lang string, bundleID string, cache cache.Cache, bot *tgbotapi.BotAPI, _chatID int64, db *gorm.DB) {
	//userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

	//addresses, _ := userOperationPackageAddressesRepo.Query(context.Background(), _chatID)

	//msg := tgbotapi.NewMessage(_chatID, "我们设置了 "+"<b>「仅允许派送至已管理的地址」</b>"+" 的安全规则。这样可以更有效地保障您的资产安全，避免因误操作导致能量丢失。\n\n"+
	//	"如果您尚未添加可用的接收地址，请前往<b>【首页】 ➝ 【添加地址】</b> 进行添加，完成后即可正常使用派送功能。\n\n📌 安全提示：建议定期检查并更新您的地址列表，确保所有地址均为您可控的合法地址。"+"\n\n"+
	//	"👇请选择要派送的地址或输入需派发的地址："+"\n\n")
	//地址绑定
	msg := tgbotapi.NewMessage(_chatID, global.Translations[_lang]["enter_address"]+"\n\n")
	msg.ParseMode = "HTML"

	var allButtons []tgbotapi.InlineKeyboardButton
	var extraButtons []tgbotapi.InlineKeyboardButton
	var keyboard [][]tgbotapi.InlineKeyboardButton
	//for _, item := range addresses {
	//	allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(TruncateString(item.Address), "dispatch_others_"+bundleID+"_"+item.Address))
	//}

	extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package"))

	for i := 0; i < len(allButtons); i += 1 {
		end := i + 1
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

	msg.ReplyMarkup = inlineKeyboard

	bot.Send(msg)

	expiration := 1 * time.Minute // 短时间缓存空值

	//设置用户状态
	cache.Set(strconv.FormatInt(_chatID, 10), "dispatch_others", expiration)
}

func ExtractBundlePackageST(_lang string, db *gorm.DB, callbackQuery *tgbotapi.CallbackQuery) tgbotapi.MessageConfig {

	fmt.Println("ExtractBundlePackageST")
	//userAddressDetectionRepo := repositories.NewUserPackageSubscriptionsRepository(db)
	userAddressDetectionRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)
	var info request.UserAddressDetectionSearch

	info.Page = 1
	info.PageSize = 10000
	trxlist, total, err := userAddressDetectionRepo.GetUserSmartTransactionPackageSubscriptionsInfoList(context.Background(), info, callbackQuery.Message.Chat.ID)
	if err != nil {

		fmt.Println("能量笔数套餐空", err)
	}
	var builder strings.Builder
	if total > 0 {
		//- [6.29] +3000 TRX（订单 #TOPUP-92308）
		for _, word := range trxlist {
			builder.WriteString("[")
			builder.WriteString(word.CreatedDate)
			builder.WriteString("]")
			builder.WriteString(" -")
			builder.WriteString(strings.ReplaceAll(word.BundleName, "笔", global.Translations[_lang]["笔"]))

			//builder.WriteString(" TRX ")
			//builder.WriteString(" （能量笔数套餐）")

			builder.WriteString("\n") // 添加分隔符
		}
	} else {
		builder.WriteString("\n") // 添加分隔符
	}

	// 去除最后一个空格
	result := strings.TrimSpace(builder.String())

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🧾"+global.Translations[_lang]["deduction_records"]+"\n\n "+
		result+"\n")
	msg.ParseMode = "HTML"
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "prev_bundle_package_page"),
		//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "next_bundle_package_page"),
		//),
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
			tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_bundle_package_ST"),
		),
	)
	msg.ReplyMarkup = inlineKeyboard
	return msg
}
