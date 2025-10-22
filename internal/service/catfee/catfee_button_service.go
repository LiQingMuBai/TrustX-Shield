package catfee

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"ushield_bot/internal/cache"
	"ushield_bot/internal/global"
	trxfee "ushield_bot/internal/infrastructure/3rd"
	"ushield_bot/internal/infrastructure/repositories"
	"ushield_bot/internal/infrastructure/tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// ⚠️剩余笔数不足10笔，请及时充值
// ➖➖➖➖➖➖➖➖➖➖
// 托管模式：智能托管
// 已用笔数：16 笔
// 剩余笔数：4 笔

// 🔋托管地址【2】
// ➖➖➖➖➖➖➖➖➖➖
// TSwA...ZGCCTV  已用 - 6
// TXLE...3n2222  已用 - 8
func CLICK_BUNDLE_PACKAGE_ADDRESS_STATS_ST(_lang string, cache cache.Cache, db *gorm.DB, chatID int64, bot *tgbotapi.BotAPI) {
	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)
	addresses, _ := userSmartTransactionAddressesRepo.List(context.Background(), strconv.FormatInt(chatID, 10))
	var allButtons []tgbotapi.InlineKeyboardButton
	var builder strings.Builder
	for _, st_address := range addresses {
		fmt.Println(st_address)

		builder.WriteString("<code>" + st_address.Address + "</code>")
		builder.WriteString(global.Translations[_lang]["used"])
		builder.WriteString("-")
		builder.WriteString(strconv.Itoa(st_address.UsedCount))
		builder.WriteString("\n") // 添加分隔符

		label := global.Translations[_lang]["catfee_custody_address_energy"]

		if st_address.Status == "1" {
			label = "✅ " + label
		} else {
			label = "□" + label // 或者用空格
		}
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(tools.TruncateString(st_address.Address), "noop"), tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("custody_address_check_%d_%s", st_address.ID, st_address.Status)))
	}

	allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["catfee_add_address"], "catfee_add_address"), tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["catfee_remove_address"], "catfee_remove_address"), tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["back_homepage"], "back_bundle_package_ST"))

	fmt.Printf("按钮数量 %d\n", len(allButtons))
	// 调用函数，按每行 2 个排列
	keyboard := LayoutButtonsInRowsOfTwo(allButtons)

	originStr := global.Translations[_lang]["catfee_custody_address_list_head"]

	userRepo := repositories.NewUserRepository(db)

	user, _ := userRepo.GetByUserID(chatID)

	totalTimes := user.StTimes
	usedTimes := user.UsedStTimes
	restTimes := totalTimes - usedTimes

	targetStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(originStr, "{total_times}", strconv.FormatInt(totalTimes, 10)), "{used_times}", strconv.FormatInt(usedTimes, 10)), "{rest_times}", strconv.FormatInt(restTimes, 10))

	//🔋托管地址【2】
	//➖➖➖➖➖➖➖➖➖➖
	//TSwA...ZGCCTV  已用 - 6
	//TXLE...3n2222  已用 - 8
	//
	//🔹【转有U地址】 消耗  65K能量  扣1笔
	//🔹【转无U地址】 消耗131K能量  扣2笔

	//➖➖➖➖➖➖➖➖➖➖
	//TSwA...ZGCCTV  已用 - 6
	//TXLE...3n2222  已用 - 8
	custodyOriginStr := global.Translations[_lang]["catfee_custody_address_count"]
	custodyTargetStr := strings.ReplaceAll(custodyOriginStr, "{custody_address_count}", strconv.Itoa(len(addresses)))

	msg := tgbotapi.NewMessage(chatID, custodyTargetStr+"\n"+targetStr+"\n"+builder.String()+global.Translations[_lang]["catfee_custody_address_energy_rule"])
	// 3. 创建键盘标记
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
	//expiration := 1 * time.Minute // 短时间缓存空值
	//设置用户状态
	//cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "apply_ST_bundle_package_"+bundleID, expiration)

	//// 发送初始的可勾选按钮
	//keyboard := buildCheckboxKeyboard(nil)
	//msg := tgbotapi.NewMessage(chatID, "请选择选项：")
	//msg.ReplyMarkup = &keyboard
	//bot.Send(msg)
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

// 存储用户的选择状态（实际应用中应使用数据库或缓存）
var userSelections = make(map[int64]map[int]bool) // chatID -> 选项ID -> 是否选中

func CheckOption(_lang string, db *gorm.DB, chatID int64, messageID int, _data string, bot *tgbotapi.BotAPI, catfeeClient *trxfee.CatfeeService) {

	userSmartTransactionAddressesRepo := repositories.NewUserSmartTransactionAddressesRepository(db)
	result := strings.ReplaceAll(_data, "custody_address_check_", "")

	ID := strings.Split(result, "_")[0]
	status := strings.Split(result, "_")[1]

	fmt.Printf("用户：%s，当前状态：%s\n", ID, status)
	record, _ := userSmartTransactionAddressesRepo.Find(context.Background(), ID)
	if status == "1" {
		fmt.Printf("用户ID %d，当前状态：%s，地址：%s 需要暂停为3", chatID, status, record.Address)
		userSmartTransactionAddressesRepo.Disable(context.Background(), strconv.FormatInt(chatID, 10), record.Address)
		//暂停

		catfeeClient.MateOpenBasicDisable(record.Address)

	}
	if status == "3" {
		//判断下是否次数不足，不能开启
		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(chatID)

		if user.StTimes <= user.UsedStTimes {
			fmt.Printf("\n 无法开启用户%s伴侣，当前托管笔数 %d，已用笔数%d\n", user.Associates, user.StTimes, user.UsedStTimes)
			msg := tgbotapi.NewMessage(chatID, global.Translations[_lang]["catfee_energy_address_buy_error"])
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
				))
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return
		}

		fmt.Printf("用户ID %d，当前状态：%s，地址：%s 需要启动为1", chatID, status, record.Address)
		userSmartTransactionAddressesRepo.Enable(context.Background(), strconv.FormatInt(chatID, 10), record.Address)
		//启动
		catfeeClient.MateOpenBasicEnable(record.Address)
	}

	if status == "0" {

		//判断下是否次数不足，不能开启
		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(chatID)

		if user.StTimes <= user.UsedStTimes {
			fmt.Printf("\n 无法开启用户%s伴侣，当前托管笔数 %d，已用笔数%d\n", user.Associates, user.StTimes, user.UsedStTimes)
			msg := tgbotapi.NewMessage(chatID, global.Translations[_lang]["catfee_energy_address_buy_error"])
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔢"+global.Translations[_lang]["smart_transaction_address_list"], "click_bundle_package_address_stats_ST"),
				))
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return
		}

		//增加
		fmt.Printf("用户ID %d，当前状态：%s，地址：%s 需要启动为1", chatID, status, record.Address)
		userSmartTransactionAddressesRepo.Enable2(context.Background(), strconv.FormatInt(chatID, 10), record.Address)

		catfeeClient.MateOpenBasicAdd(record.Address, strconv.FormatInt(chatID, 10))
	}

	addresses, _ := userSmartTransactionAddressesRepo.List(context.Background(), strconv.FormatInt(chatID, 10))
	var allButtons []tgbotapi.InlineKeyboardButton
	var builder strings.Builder
	for _, st_address := range addresses {
		fmt.Println(st_address)

		builder.WriteString("<code>" + st_address.Address + "</code>")
		builder.WriteString(global.Translations[_lang]["used"])
		builder.WriteString("-")
		builder.WriteString(strconv.Itoa(st_address.UsedCount))
		builder.WriteString("\n") // 添加分隔符

		label := global.Translations[_lang]["catfee_custody_address_energy"]

		if st_address.Status == "1" {
			label = "✅ " + label
		} else {
			label = "□" + label // 或者用空格
		}
		allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(tools.TruncateString(st_address.Address), "noop"), tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("custody_address_check_%d_%s", st_address.ID, st_address.Status)))
	}

	allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["catfee_add_address"], "catfee_add_address"), tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["catfee_remove_address"], "catfee_remove_address"), tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["back_homepage"], "back_bundle_package_ST"))

	fmt.Printf("按钮数量 %d\n", len(allButtons))
	// 调用函数，按每行 2 个排列
	keyboard := LayoutButtonsInRowsOfTwo(allButtons)

	originStr := global.Translations[_lang]["catfee_custody_address_list_head"]

	userRepo := repositories.NewUserRepository(db)

	user, _ := userRepo.GetByUserID(chatID)

	totalTimes := user.StTimes
	usedTimes := user.UsedStTimes
	restTimes := totalTimes - usedTimes

	targetStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(originStr, "{total_times}", strconv.FormatInt(totalTimes, 10)), "{used_times}", strconv.FormatInt(usedTimes, 10)), "{rest_times}", strconv.FormatInt(restTimes, 10))

	//🔋托管地址【2】
	//➖➖➖➖➖➖➖➖➖➖
	//TSwA...ZGCCTV  已用 - 6
	//TXLE...3n2222  已用 - 8
	//
	//🔹【转有U地址】 消耗  65K能量  扣1笔
	//🔹【转无U地址】 消耗131K能量  扣2笔

	//➖➖➖➖➖➖➖➖➖➖
	//TSwA...ZGCCTV  已用 - 6
	//TXLE...3n2222  已用 - 8
	custodyOriginStr := global.Translations[_lang]["catfee_custody_address_count"]
	custodyTargetStr := strings.ReplaceAll(custodyOriginStr, "{custody_address_count}", strconv.Itoa(len(addresses)))

	msg := tgbotapi.NewMessage(chatID, custodyTargetStr+"\n"+targetStr+"\n"+builder.String()+global.Translations[_lang]["catfee_custody_address_energy_rule"])
	// 3. 创建键盘标记
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	msg.ReplyMarkup = inlineKeyboard
	msg.ParseMode = "HTML"
	//bot.Send(msg)
	// 编辑消息，更新按钮
	editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, inlineKeyboard)
	bot.Send(editMsg)

}

// LayoutButtonsInRowsOfTwo 将 N 个 InlineKeyboardButton 按每行 2 个进行排列
// 返回二维切片，可用于 NewInlineKeyboardMarkup
func LayoutButtonsInRowsOfTwo(buttons []tgbotapi.InlineKeyboardButton) [][]tgbotapi.InlineKeyboardButton {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < len(buttons); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// 添加第一个按钮
		row = append(row, buttons[i])

		// 如果还有下一个按钮，添加第二个
		if i+1 < len(buttons) {
			row = append(row, buttons[i+1])
		}

		// 将这一行加入最终键盘
		keyboard = append(keyboard, row)
	}

	return keyboard
}
