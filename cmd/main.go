package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"ushield_bot/internal/service/additional"
	"ushield_bot/internal/service/catfee"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"time"
	"ushield_bot/internal/global"
	trxfee "ushield_bot/internal/infrastructure/3rd"
	"ushield_bot/internal/service"

	"ushield_bot/internal/cache"
	"ushield_bot/internal/domain"
	"ushield_bot/internal/infrastructure/repositories"
	. "ushield_bot/internal/infrastructure/tools"
)

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

// 加载翻译文件
func loadTranslations() {
	global.Mutex.Lock()
	defer global.Mutex.Unlock()

	for _, lang := range global.SupportedLangs {
		filePath := filepath.Join(global.TranslationsDir, lang+".json")
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: Could not load translation file for %s: %v", lang, err)
			continue
		}

		var langTranslations map[string]string
		if err := json.Unmarshal(data, &langTranslations); err != nil {
			log.Printf("Error parsing translation file for %s: %v", lang, err)
			continue
		}

		global.Translations[lang] = langTranslations
	}

	// 确保默认语言存在
	if _, exists := global.Translations[global.DefaultLang]; !exists {
		log.Fatalf("Default language %s not found in translations", global.DefaultLang)
	}
}

// 翻译函数（基础版本）
func T(lang, key string) string {
	global.Mutex.RLock()
	defer global.Mutex.RUnlock()

	if langTranslations, exists := global.Translations[lang]; exists {
		if value, exists := langTranslations[key]; exists {
			return value
		}
	}

	// 回退到默认语言
	if lang != global.DefaultLang {
		if value, exists := global.Translations[global.DefaultLang][key]; exists {
			return value
		}
	}

	// 如果键不存在，返回键本身
	return key
}

// 带参数的翻译函数（处理占位符）
func TParam(lang, key string, params map[string]string) string {
	text := T(lang, key)

	// 替换占位符
	for key, value := range params {
		placeholder := "{" + key + "}"
		text = strings.ReplaceAll(text, placeholder, value)
	}

	return text
}
func main() {

	logrus.SetFormatter(&logrus.JSONFormatter{})

	if err := initConfig(); err != nil {
		logrus.Fatalf("init configs err: %s", err.Error())
	}

	loadTranslations()

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("load .env file err: %s", err.Error())
	}

	//log.Printf(T("zh", "start"))
	//log.Printf(T("en", "start"))

	// Database connection string
	host := viper.GetString("db.host")
	port := viper.GetString("db.port")
	username := viper.GetString("db.username")
	password := viper.GetString("db.password")
	dbname := viper.GetString("db.dbname")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to the database: " + err.Error())
	}
	TG_BOT_API := os.Getenv("TG_BOT_API")
	bot, err := tgbotapi.NewBotAPI(TG_BOT_API)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	_cookie1 := os.Getenv("COOKIE1")
	_cookie2 := os.Getenv("COOKIE2")
	_cookie3 := os.Getenv("COOKIE3")

	trxfeeUrl := os.Getenv("TRXFEE_BASE_URL")
	trxfeeApiKey := os.Getenv("TRXFEE_APIKEY")
	trxfeeSecret := os.Getenv("TRXFEE_APISECRET")

	fixfloatedUrl := os.Getenv("FIXFLOATED_REF_URL")

	botName := os.Getenv("BOT_NAME")

	log.Printf("Trxfee URL: %s\n", trxfeeUrl)
	log.Printf("trxfeeApiKey: %s\n", trxfeeApiKey)
	log.Printf("trxfeeSecret: %s\n", trxfeeSecret)

	CATFEE_BASE_URL := os.Getenv("CATFEE_BASE_URL")
	CATFEE_APIKEY := os.Getenv("CATFEE_APIKEY")
	CATFEE_APISECRET := os.Getenv("CATFEE_APISECRET")

	log.Printf("CATFEE_BASE_URL: %s\n", CATFEE_BASE_URL)
	log.Printf("CATFEE_APIKEY: %s\n", CATFEE_APIKEY)
	log.Printf("CATFEE_APISECRET: %s\n", CATFEE_APISECRET)

	catfee, _ := trxfee.NewCatfeeService(CATFEE_APIKEY, CATFEE_APISECRET, CATFEE_BASE_URL)
	// 1. 创建字符串数组
	cookies := []string{_cookie1, _cookie2, _cookie3}

	fmt.Printf("cookies: %s\n", cookies)

	_cookie := RandomCookiesString(cookies)

	cache := cache.NewMemoryCache()

	// 设置命令
	_, err = bot.Request(tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "start"},
		tgbotapi.BotCommand{Command: "hide", Description: "hide"},
	))
	if err != nil {
		log.Printf("Error setting commands: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				switch {
				case strings.HasPrefix(update.Message.Command(), "startDispatch"):
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢 功能开发中！想第一时间知道它上线吗？记得关注我们的官方频道：@ushield1 🔔\n\n")
					msg.ParseMode = "HTML"
					bot.Send(msg)

				case strings.HasPrefix(update.Message.Command(), "open_ST"):
					subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "open_ST", "")
					log.Println("subscribeBundleID : " + subscribeBundleID)
					log.Println(subscribeBundleID + "   open_ST command")

					//手动发能

					//trxfee
					userSmartTransactionPackageSubscriptionsRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)

					record, _ := userSmartTransactionPackageSubscriptionsRepo.GetRecordByID(subscribeBundleID)
					if record.ChatID != update.Message.Chat.ID {
						log.Println("不是自己的权利")
						//return
					} else {

						userRepo := repositories.NewUserRepository(db)
						user, _ := userRepo.GetByUserID(update.Message.Chat.ID)

						log.Printf("address is %s\n", record.Address)

						//发送请求给trxfee，开启启动智能托管
						record.Status = 2
						err := userSmartTransactionPackageSubscriptionsRepo.UpdateStatusByID(context.Background(), subscribeBundleID, record.Status)

						if err != nil {

							return

						}
						trxfeeClient := trxfee.NewTrxfeeClient(trxfeeUrl, trxfeeApiKey, trxfeeSecret)
						trxfeeClient.EnableTimesOrder(record.Address)

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅"+global.Translations[user.Lang]["smart_transaction_auto_dispatch_open_successfully"]+"\n")
						msg.ParseMode = "HTML"
						inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
							//tgbotapi.NewInlineKeyboardRow(
							//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
							//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
							//),
							tgbotapi.NewInlineKeyboardRow(
								//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
								tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[user.Lang]["back_homepage"], "back_bundle_package_ST"),
							),
						)
						msg.ReplyMarkup = inlineKeyboard
						bot.Send(msg)

					}

				case strings.HasPrefix(update.Message.Command(), "close_ST"):
					subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "close_ST", "")
					log.Println("subscribeBundleID : " + subscribeBundleID)
					log.Println(subscribeBundleID + "   closeST command")

					//手动发能

					//trxfee
					userSmartTransactionPackageSubscriptionsRepo := repositories.NewUserSmartTransactionPackageSubscriptionsRepository(db)

					record, _ := userSmartTransactionPackageSubscriptionsRepo.GetRecordByID(subscribeBundleID)
					if record.ChatID != update.Message.Chat.ID {
						log.Println("不是自己的权利")
						//return
					} else {

						userRepo := repositories.NewUserRepository(db)
						user, _ := userRepo.GetByUserID(update.Message.Chat.ID)

						log.Printf("address is %s\n", record.Address)

						//发送请求给trxfee，禁止启动智能托管
						record.Status = 1
						err := userSmartTransactionPackageSubscriptionsRepo.UpdateStatusByID(context.Background(), subscribeBundleID, record.Status)

						if err != nil {
							return
						}
						trxfeeClient := trxfee.NewTrxfeeClient(trxfeeUrl, trxfeeApiKey, trxfeeSecret)
						trxfeeClient.EnableTimesOrder(record.Address)

						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅"+global.Translations[user.Lang]["smart_transaction_auto_dispatch_close_successfully"]+"\n")
						msg.ParseMode = "HTML"
						inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
							//tgbotapi.NewInlineKeyboardRow(
							//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["prev"], "next_bundle_package_address_stats"),
							//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["next"], "prev_bundle_package_address_stats"),
							//),
							tgbotapi.NewInlineKeyboardRow(
								//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
								tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[user.Lang]["back_homepage"], "back_bundle_package_ST"),
							),
						)
						msg.ReplyMarkup = inlineKeyboard
						bot.Send(msg)

					}

				case strings.HasPrefix(update.Message.Command(), "dispatchNow"):
					subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "dispatchNow", "")
					log.Println("subscribeBundleID : " + subscribeBundleID)
					log.Println(subscribeBundleID + "   dispatchNow command")

					//手动发能

					//trxfee
					userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

					record, _ := userOperationPackageAddressesRepo.Get(context.Background(), subscribeBundleID)
					if record.ChatID != update.Message.Chat.ID {

						log.Println("不是自己的权利")
						//return
					} else {
						log.Printf("address is %s\n", record.Address)

						userRepo := repositories.NewUserRepository(db)
						user, _ := userRepo.GetByUserID(update.Message.Chat.ID)

						_bundleTimes := user.BundleTimes - 1
						//time.Sleep(100 * time.Millisecond)
						if user.BundleTimes > 0 {
							userRepo.UpdateBundleTimes(_bundleTimes, update.Message.Chat.ID)

							//
							//msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(db, update.Message.Chat.ID)
							//bot.Send(msg2)

							//调用trxfee接口

							var sysOrder domain.UserEnergyOrders
							orderNo, _ := GenerateOrderID(record.Address, 4)
							//fmt.Printf("  OrderNo: %s\n", orderNo)
							sysOrder.OrderNo = orderNo
							sysOrder.TxId = ""
							sysOrder.FromAddress = record.Address
							//sysOrder.ToAddress = item.Address
							sysOrder.Amount = 65000
							sysOrder.ChatId = strconv.FormatInt(update.Message.Chat.ID, 10)
							//
							////添加一条记录
							ueoRepo := repositories.NewUserEnergyOrdersRepo(db)
							errsg := ueoRepo.Create(context.Background(), &sysOrder)

							if errsg == nil {
								trxfeeClient := trxfee.NewTrxfeeClient(trxfeeUrl, trxfeeApiKey, trxfeeSecret)

								fmt.Sprintf("发送（%d）笔能量给（%s），订单号 %s\n", 1, record.Address, orderNo)
								trxfeeClient.Order(orderNo, record.Address, 65_000*1)

								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢【✅"+global.Translations[user.Lang]["UShield_sent_transaction_energy"]+"】\n\n"+
									global.Translations[user.Lang]["to_address"]+record.Address+"\n\n"+
									global.Translations[user.Lang]["remaining_transactions"]+strconv.FormatInt(_bundleTimes, 10)+"\n\n")

								inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
									tgbotapi.NewInlineKeyboardRow(
										tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[user.Lang]["dispatch_again"], "click_bundle_package_address_stats"),
									),
								)
								msg.ReplyMarkup = inlineKeyboard
								msg.ParseMode = "HTML"
								bot.Send(msg)
							}
						} else {
							msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(user.Lang, db, update.Message.Chat.ID)
							bot.Send(msg)
						}
					}
					//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢【✅"+global.Translations[_lang]["UShield_sent_transaction_energy"]+"】\n\n"+
					//	global.Translations[_lang]["to_address"]+record.Address+"\n\n"+
					//	global.Translations[_lang]["remaining_transactions"]+strconv.FormatInt(restTimes, 10)+"\n\n")
					//msg.ParseMode = "HTML"
					//bot.Send(msg)

				case strings.HasPrefix(update.Message.Command(), "stopDispatch"):

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "📢 功能开发中！想第一时间知道它上线吗？记得关注我们的官方频道：@ushield1 🔔\n\n")
					msg.ParseMode = "HTML"
					bot.Send(msg)

					//return

					//subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "stopDispatch", "")
					//log.Println("subscribeBundleID :" + subscribeBundleID)
					//log.Println(subscribeBundleID + "stopDispatch command")
					//userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
					//
					//subscribeBundlePackageID, _ := strconv.ParseInt(subscribeBundleID, 10, 64)
					//
					//userPackageSubscriptionsRepo.UpdateStatus(context.Background(), subscribeBundlePackageID, 2)
					//msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(db, update.Message.Chat.ID)
					//bot.Send(msg)

				case strings.HasPrefix(update.Message.Command(), "dispatchOthers"):
					subscribeBundleID := strings.ReplaceAll(update.Message.Command(), "dispatchOthers", "")
					log.Println("subscribeBundleID :" + subscribeBundleID)
					log.Println(subscribeBundleID + "dispatchOthers command")
					//userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)

					//subscribeBundlePackageID, _ := strconv.ParseInt(subscribeBundleID, 10, 64)
					//userPackageSubscriptionsRepo.UpdateStatus(context.Background(), subscribeBundlePackageID, 2)
					//msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(db, update.Message.Chat.ID)
					//bot.Send(msg)
					//
					userRepo := repositories.NewUserRepository(db)
					record, _ := userRepo.GetByUserID(update.Message.Chat.ID)
					service.DispatchOthers(record.Lang, subscribeBundleID, cache, bot, update.Message.Chat.ID, db)

				case update.Message.Command() == "start":
					log.Printf("1")
					log.Println("text: " + update.Message.Text)
					userRepo := repositories.NewUserRepository(db)
					index := strings.LastIndex(update.Message.Text, " ")
					parentUID := ""
					if index > 0 {

						parentUIDStr := update.Message.Text
						parentUID = parentUIDStr[index+1:]
						fmt.Printf("parentUIDStr: %s\n", parentUID)

						record, err := userRepo.GetByUserIDStr(parentUID)
						if err != nil {
							parentUID = ""
						} else {
							parentUID = record.Associates
						}

					}
					//存用户
					//userRepo := repositories.NewUserRepository(db)
					record, err := userRepo.GetByUserID(update.Message.Chat.ID)
					if err != nil {
						//增加用户
						var user domain.User
						user.Associates = strconv.FormatInt(update.Message.Chat.ID, 10)
						user.Username = update.Message.Chat.UserName
						user.CreatedAt = time.Now()
						user.Lang = "zh"
						user.BotName = botName
						if len(parentUID) > 0 {
							user.ParentUserID = parentUID
						}
						err := userRepo.Create2(context.Background(), &user)

						expiration := 24 * time.Hour // 短时间缓存空值
						cache.Set("LANG_"+strconv.FormatInt(update.Message.Chat.ID, 10), "zh", expiration)
						if err != nil {
							return
						}
					}

					if err == nil {

						record.Username = update.Message.From.UserName

						userRepo.UpdateUserNameByChatID(update.Message.From.UserName, update.Message.Chat.ID)

						log.Printf("UserName: %s", record.Username)
						log.Printf("Associates %s", record.Associates)
						expiration := 24 * time.Hour // 短时间缓存空值
						if len(record.Lang) > 0 {

							cache.Set("LANG_"+strconv.FormatInt(update.Message.Chat.ID, 10), record.Lang, expiration)
						} else {
							cache.Set("LANG_"+strconv.FormatInt(update.Message.Chat.ID, 10), "zh", expiration)
						}
					}

					handleStartCommand(cache, bot, update.Message)
				case update.Message.Command() == "hide":
					log.Printf("2")
					handleHideCommand(cache, bot, update.Message)
				}
			} else {

				log.Printf("3")
				log.Printf("来自于自发的信息[%s] %s", update.Message.From.UserName, update.Message.Text)
				handleRegularMessage(cache, bot, update.Message, db, _cookie, trxfeeUrl, trxfeeApiKey, trxfeeSecret, fixfloatedUrl, catfee)
			}
		} else if update.CallbackQuery != nil {
			log.Printf("4")
			handleCallbackQuery(cache, bot, update.CallbackQuery, db, trxfeeUrl, trxfeeApiKey, trxfeeSecret, catfee)
		}
	}
}

// 处理 /start 命令 - 显示永久键盘
func handleStartCommand(cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// 创建永久性回复键盘

	_lang, err := cache.Get("LANG_" + strconv.FormatInt(message.Chat.ID, 10))

	if err != nil {
		_lang = "zh"
	}
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(

			//tgbotapi.NewKeyboardButton("⛽"+global.Translations[_lang]["tron_energy_menu"]),

			tgbotapi.NewKeyboardButton("⚡"+global.Translations[_lang]["energy_swap"]),
			tgbotapi.NewKeyboardButton("🖊️"+global.Translations[_lang]["transaction_plans"]),
			//tgbotapi.NewKeyboardButton("🤖"+global.Translations[_lang]["smart_transaction_plans"]),
			tgbotapi.NewKeyboardButton("🤖"+global.Translations[_lang]["catfee_smart_transaction_menu"]),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅"+global.Translations[_lang]["usdt_trx_swap"]),
			tgbotapi.NewKeyboardButton("🔃"+global.Translations[_lang]["coin_swap_coin_menu"]),
			tgbotapi.NewKeyboardButton("🕸"+global.Translations[_lang]["address_trace_menu"]),
			//tgbotapi.NewKeyboardButton("🔍"+global.Translations[_lang]["address_check"]),
			//tgbotapi.NewKeyboardButton("🚨"+global.Translations[_lang]["usdt_freeze_alert"]),
		),
		tgbotapi.NewKeyboardButtonRow(

			tgbotapi.NewKeyboardButton("👤"+global.Translations[_lang]["my_account"]),

			tgbotapi.NewKeyboardButton("🛒"+global.Translations[_lang]["ushield_additional_services_menu"]),
		),
	)

	// 关键设置：确保键盘一直存在
	keyboard.OneTimeKeyboard = false
	keyboard.ResizeKeyboard = true
	keyboard.Selective = false
	originStr := global.Translations[_lang]["welcome_tips"]
	//msg := tgbotapi.NewMessage(message.Chat.ID, "🛡️U盾，做您链上资产的护盾！\n我们不仅关注低价能量，更专注于交易安全！\n让每一笔转账都更安心，让每一次链上交互都值得信任！\n🤖 三大实用功能，助您安全、高效地管理链上资产\n🔋 波场能量闪兑, 节省超过80%!\n🕵️ 地址风险检测, 让每一笔转账都更安心!\n🚨 USDT冻结预警,秒级响应，让您的U永不冻结！\n🎉新用户福利：每日一次免费地址风险查询")
	msg := tgbotapi.NewMessage(message.Chat.ID, originStr)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

// 处理 /hide 命令 - 隐藏键盘
func handleHideCommand(cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	hideKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(message.Chat.ID, "键盘已隐藏，发送 /start 重新显示")
	msg.ReplyMarkup = hideKeyboard
	bot.Send(msg)
}

// 处理普通消息（键盘按钮点击）
func handleRegularMessage(cache cache.Cache, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *gorm.DB, _cookie string, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret string, fixfloatedUrl string, catfeeClient *trxfee.CatfeeService) {
	_lang, err := cache.Get("LANG_" + strconv.FormatInt(message.Chat.ID, 10))
	if len(_lang) == 0 || err != nil {
		userRepo := repositories.NewUserRepository(db)
		record, _ := userRepo.GetByUserID(message.Chat.ID)
		expiration := 24 * time.Hour // 短时间缓存空值
		cache.Set("LANG_"+strconv.FormatInt(message.Chat.ID, 10), record.Lang, expiration)
		_lang = record.Lang
	}

	switch message.Text {

	case "🛒" + global.Translations[_lang]["ushield_additional_services_menu"]:
		additional.MenuNavigate(_lang, db, message.Chat.ID, bot)
	case "🤖" + global.Translations[_lang]["catfee_smart_transaction_menu"]:
		catfee.MenuNavigateCatfeeSmartTransactionPlans(_lang, db, message.Chat.ID, bot, "TRX")

	case "🔃" + global.Translations[_lang]["coin_swap_coin_menu"]:
		service.MenuNavigateCoin2CoinSwap(_lang, db, message, bot, fixfloatedUrl)
	case "⛽" + global.Translations[_lang]["tron_energy_menu"]:
		service.MenuNavigateTronEnergy(_lang, db, message, bot)
	case "✅" + global.Translations[_lang]["usdt_trx_swap"]:
		service.MenuNavigateSwapExchange(_lang, db, message, bot)
	case "🕸" + global.Translations[_lang]["address_trace_menu"]:
		service.MenuNavigateAddressTrace(_lang, cache, bot, message.Chat.ID, db)
	case "🔍" + global.Translations[_lang]["address_check"]:
		service.MenuNavigateAddressDetection(_lang, cache, bot, message.Chat.ID, db)
	case "🚨" + global.Translations[_lang]["usdt_freeze_alert"]:
		service.MenuNavigateAddressFreeze(_lang, cache, bot, message.Chat.ID, db)
	case "🖊️" + global.Translations[_lang]["transaction_plans"]:
		service.MenuNavigateBundlePackage(_lang, db, message.Chat.ID, bot, "TRX")

	case "🤖" + global.Translations[_lang]["smart_transaction_plans"]:
		service.MenuNavigateSmartTransactionPlans(_lang, db, message.Chat.ID, bot, "TRX")

	case "⚡" + global.Translations[_lang]["energy_swap"]:
		service.MenuNavigateEnergyExchange(_lang, db, message, bot)
	case "👤" + global.Translations[_lang]["my_account"]:
		service.MenuNavigateHome(_lang, cache, db, message, bot)
	case "🌍" + global.Translations[_lang]["language"]:
		service.MenuNavigateHome2(db, message, bot)
	default:
		status, _ := cache.Get(strconv.FormatInt(message.Chat.ID, 10))

		log.Printf("用户状态status %s", status)
		switch {
		case strings.HasPrefix(status, "user_backup_notify"):

			if service.ExtractBackup(message, bot, db) {
				return
			}
		case strings.HasPrefix(status, "start_freeze_risk"):

			if !IsValidAddress(message.Text) && !IsValidEthereumAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			dictRepo := repositories.NewSysDictionariesRepo(db)

			server_trx_price, _ := dictRepo.GetDictionaryDetail("server_trx_price")

			server_usdt_price, _ := dictRepo.GetDictionaryDetail("server_usdt_price")

			originStr := global.Translations[_lang]["enable_freeze_alerts_tips"]

			targetStr := strings.ReplaceAll(strings.ReplaceAll(originStr, "{server_usdt_price}", server_usdt_price), "{server_trx_price}", server_trx_price)

			msg := tgbotapi.NewMessage(message.Chat.ID, global.Translations[_lang]["enable_freeze_alerts_tips_suffix"]+"\n"+global.Translations[_lang]["address"]+"："+message.Text+"\n\n"+targetStr)
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✅"+global.Translations[_lang]["confirm_freeze_alerts"], "confirm_freeze_risk_"+message.Text),
					tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[_lang]["cancel_freeze_alerts"], "back_risk_home"),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			expiration := 1 * time.Minute // 短时间缓存空值
			//设置用户状态
			cache.Set(strconv.FormatInt(message.Chat.ID, 10), "start_freeze_risk_status", expiration)

		case strings.HasPrefix(status, "address_list_trace"):

		case strings.HasPrefix(status, "address_manager_remove"):
			if IsValidAddress(message.Text) || IsValidEthereumAddress(message.Text) {
				userRepo := repositories.NewUserAddressMonitorRepo(db)
				err := userRepo.Remove(context.Background(), message.Chat.ID, message.Text)
				if err != nil {
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)

				service.ADDRESS_MANAGER(_lang, cache, bot, message.Chat.ID, db)

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		case strings.HasPrefix(status, "dispatch_others"):
			if IsValidAddress(message.Text) {
				//time.Sleep(100 * time.Millisecond)
				//subscribeBundleID := strings.ReplaceAll(status, "DISPATCHOTHERS_", "")
				//trxfee
				//userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
				//record, _ := userPackageSubscriptionsRepo.Query(context.Background(), subscribeBundleID)
				userRepo := repositories.NewUserRepository(db)
				user, _ := userRepo.GetByUserID(message.Chat.ID)

				_bundleTimes := user.BundleTimes - 1
				//time.Sleep(100 * time.Millisecond)
				if user.BundleTimes > 0 && _bundleTimes >= 0 {
					userRepo.UpdateBundleTimes(_bundleTimes, message.Chat.ID)

					//if restTimes >= 0 {
					//	userPackageSubscriptionsRepo.UpdateTimes(context.Background(), record.Id, restTimes)

					//
					//msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(db, message.Chat.ID)
					//bot.Send(msg2)

					//调用trxfee接口

					var sysOrder domain.UserEnergyOrders
					orderNo, _ := GenerateOrderID(message.Text, 4)
					//fmt.Printf("  OrderNo: %s\n", orderNo)
					sysOrder.OrderNo = orderNo
					sysOrder.TxId = ""
					sysOrder.FromAddress = message.Text
					//sysOrder.ToAddress = item.Address
					sysOrder.Amount = 65000
					sysOrder.ChatId = strconv.FormatInt(message.Chat.ID, 10)
					//
					////添加一条记录
					ueoRepo := repositories.NewUserEnergyOrdersRepo(db)
					errsg := ueoRepo.Create(context.Background(), &sysOrder)

					if errsg == nil {
						trxfeeClient := trxfee.NewTrxfeeClient(_trxfeeUrl, _trxfeeApiKey, _trxfeeSecret)

						fmt.Sprintf("发送（%d）笔能量给（%s），订单号 %s\n", 1, message.Text, orderNo)
						trxfeeClient.Order(orderNo, message.Text, 65_000*1)

						msg := tgbotapi.NewMessage(message.Chat.ID, "📢【✅"+global.Translations[_lang]["UShield_sent_transaction_energy"]+"】\n\n"+
							global.Translations[_lang]["to_address"]+message.Text+"\n\n"+
							global.Translations[_lang]["remaining_transactions"]+strconv.FormatInt(_bundleTimes, 10)+"\n\n")

						inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[_lang]["dispatch_again"], "click_bundle_package_address_stats"),
							),
						)
						msg.ReplyMarkup = inlineKeyboard
						msg.ParseMode = "HTML"
						bot.Send(msg)
						expiration := 1 * time.Minute // 短时间缓存空值

						//设置用户状态
						cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_dispatch_others_", expiration)
					}
				} else {
					service.MenuNavigateBundlePackage(_lang, db, message.Chat.ID, bot, "TRX")
				}

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		case strings.HasPrefix(status, "DISPATCHOTHERS_"):
			if IsValidAddress(message.Text) {
				//time.Sleep(100 * time.Millisecond)
				subscribeBundleID := strings.ReplaceAll(status, "DISPATCHOTHERS_", "")
				//trxfee
				userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
				record, _ := userPackageSubscriptionsRepo.Query(context.Background(), subscribeBundleID)

				restTimes := record.Times - 1

				if restTimes >= 0 {
					userPackageSubscriptionsRepo.UpdateTimes(context.Background(), record.Id, restTimes)

					//
					msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, db, message.Chat.ID)
					bot.Send(msg2)

					//调用trxfee接口

					var sysOrder domain.UserEnergyOrders
					orderNo, _ := GenerateOrderID(message.Text, 4)
					//fmt.Printf("  OrderNo: %s\n", orderNo)
					sysOrder.OrderNo = orderNo
					sysOrder.TxId = ""
					sysOrder.FromAddress = message.Text
					//sysOrder.ToAddress = item.Address
					sysOrder.Amount = 65000
					sysOrder.ChatId = strconv.FormatInt(message.Chat.ID, 10)
					//
					////添加一条记录
					ueoRepo := repositories.NewUserEnergyOrdersRepo(db)
					errsg := ueoRepo.Create(context.Background(), &sysOrder)

					if errsg == nil && restTimes >= 0 {
						trxfeeClient := trxfee.NewTrxfeeClient(_trxfeeUrl, _trxfeeApiKey, _trxfeeSecret)

						fmt.Sprintf("发送（%d）笔能量给（%s），订单号 %s\n", 1, message.Text, orderNo)
						trxfeeClient.Order(orderNo, message.Text, 65_000*1)

						msg := tgbotapi.NewMessage(message.Chat.ID, "📢【✅"+global.Translations[_lang]["UShield_sent_transaction_energy"]+"】\n\n"+
							global.Translations[_lang]["to_address"]+message.Text+"\n\n"+
							global.Translations[_lang]["remaining_transactions"]+strconv.FormatInt(restTimes, 10)+"\n\n")

						inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[_lang]["dispatch_again"], "click_bundle_package_address_stats"),
							),
						)
						msg.ReplyMarkup = inlineKeyboard
						msg.ParseMode = "HTML"
						bot.Send(msg)
						expiration := 1 * time.Minute // 短时间缓存空值

						//设置用户状态
						cache.Set(strconv.FormatInt(message.Chat.ID, 10), "null_dispatch_others_", expiration)
					}
				}

			} else {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
		case strings.HasPrefix(status, "address_manager_add"):
			service.ExtractAddressManager(_lang, message, db, bot)

			service.ADDRESS_MANAGER(_lang, cache, bot, message.Chat.ID, db)

		case strings.HasPrefix(status, "bundle_"):
			fmt.Printf(">>>>>>>>>>>>>>>>>>>>bundle: %s", status)

			if service.ExtractBundleService(_lang, message, bot, db, status) {
				return
			}

		case strings.HasPrefix(status, "address_trace_add"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			userRepo := repositories.NewUserAddressTraceRepo(db)

			model, err := userRepo.Find(context.Background(), message.Chat.ID, message.Text)

			if err != nil {
			}
			if model.Id > 0 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_trace_add_repeat_tips"]+"</b>"+"\n")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

					tgbotapi.NewInlineKeyboardRow(
						//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
						tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				bot.Send(msg)

				return
			}

			total, err := userRepo.Count(context.Background(), message.Chat.ID)
			if err != nil {
			}
			if total >= 4 {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_trace_add_max_tips"]+"</b>"+"\n")
				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

					tgbotapi.NewInlineKeyboardRow(
						//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
						tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				bot.Send(msg)

				return
			}
			var record domain.UserAddressTrace
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
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

				tgbotapi.NewInlineKeyboardRow(
					//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"
			bot.Send(msg)

		case strings.HasPrefix(status, "address_trace_delete"):
			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["address_wrong_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}
			userRepo := repositories.NewUserAddressTraceRepo(db)
			err := userRepo.Remove(context.Background(), message.Chat.ID, message.Text)
			if err != nil {
				return
			}
			msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(

				tgbotapi.NewInlineKeyboardRow(
					//tgbotapi.NewInlineKeyboardButtonData("解绑地址", "free_monitor_address"),
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_user_address_trace"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard

			msg.ParseMode = "HTML"
			bot.Send(msg)

		case strings.HasPrefix(status, "usdt_risk_monitor"):
			//fmt.Printf("bundle: %s", status)

			if !IsValidAddress(message.Text) {
				msg := tgbotapi.NewMessage(message.Chat.ID, "💬"+"<b>"+global.Translations[_lang]["invalid_address_tips"]+"</b>"+"\n")
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}

			msg := tgbotapi.NewMessage(message.Chat.ID, "")

			//msg.ReplyMarkup = inlineKeyboard
			msg.ParseMode = "HTML"

			bot.Send(msg)

		case strings.HasPrefix(status, "click_bundle_package_address_manager_remove"):
			if service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_REMOVE(_lang, cache, bot, message, db) {
				return
			}

		case strings.HasPrefix(status, "click_bundle_package_address_manager_add"):
			if service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_ADD(_lang, cache, bot, message, db) {
				return
			}

		case strings.HasPrefix(status, "apply_bundle_package_"):
			if service.APPLY_BUNDLE_PACKAGE(_lang, cache, bot, message, db, status) {
				return
			}
		case strings.HasPrefix(status, "apply_ST_bundle_package_"):
			trxfeeClient := trxfee.NewTrxfeeClient(_trxfeeUrl, _trxfeeApiKey, _trxfeeSecret)

			if service.APPLY_ST_BUNDLE_PACKAGE(trxfeeClient, _lang, cache, bot, message, db, status) {
				return
			}

		case strings.HasPrefix(status, "click_backup_account"):

			log.Printf("进入click_backup_account状态：%s\n", message.Text)
			if strings.Contains(message.Text, "@") {
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌ "+global.Translations[_lang]["backup_account_tips"])
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}
			userName := strings.ReplaceAll(message.Text, "@", "")

			log.Printf("备份用户：%s\n", userName)
			userRepo := repositories.NewUserRepository(db)
			user, err := userRepo.GetByUsername(userName)

			if err != nil {
				log.Printf("访问失败 %s\n", err)
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌"+global.Translations[_lang]["backup_account_tips2"])
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			if user.Id == 0 {
				log.Printf("无该用户 %s\n", userName)
				msg := tgbotapi.NewMessage(message.Chat.ID, "❌"+global.Translations[_lang]["backup_account_tips2"])
				msg.ParseMode = "HTML"
				bot.Send(msg)
				return
			}

			user.BackupChatID = userName

			err2 := userRepo.UpdateBackupChat(context.Background(), userName, message.Chat.ID)
			if err2 == nil {
				msg := tgbotapi.NewMessage(message.Chat.ID, "✅ "+global.Translations[_lang]["backup_account_tips3"]+message.Text)
				msg.ParseMode = "HTML"
				bot.Send(msg)
				//return true
			}

			service.MenuNavigateHome(_lang, cache, db, message, bot)

		case strings.HasPrefix(status, "usdt_risk_query"):
			//fmt.Printf("bundle: %s", status)
			service.ExtractSlowMistRiskQuery(_lang, cache, message, db, _cookie, bot)

		case strings.HasPrefix(status, "catfee_add_address"):
			catfee.CustodyAddressAdd(_lang, cache, db, bot, message)

		case strings.HasPrefix(status, "catfee_remove_address"):
			catfee.CustodyAddressRemove(_lang, cache, db, bot, message, catfeeClient)

		}
	}

}

// 处理内联键盘回调
func handleCallbackQuery(cache cache.Cache, bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, db *gorm.DB, _trxfeeUrl, _trxfeeApiKey, _trxfeeSecret string, catfeeClient *trxfee.CatfeeService) {
	// 先应答回调

	log.Println("已选择: " + callbackQuery.Data)
	callback := tgbotapi.NewCallback(callbackQuery.ID, "已选择: "+callbackQuery.Data)
	if _, err := bot.Request(callback); err != nil {
		log.Printf("Error answering callback: %v", err)
	}
	_lang, err := cache.Get("LANG_" + strconv.FormatInt(callbackQuery.Message.Chat.ID, 10))

	if err != nil {
		_lang = "zh"
	}
	// 根据回调数据执行不同操作
	var responseText string
	switch {

	//case "🖊️" + global.Translations[_lang]["transaction_plans"]:
	//	service.MenuNavigateBundlePackage(_lang, db, message.Chat.ID, bot, "TRX")
	//
	//case "🤖" + global.Translations[_lang]["smart_transaction_plans"]:
	//	service.MenuNavigateSmartTransactionPlans(_lang, db, message.Chat.ID, bot, "TRX")
	//
	//case "⚡" + global.Translations[_lang]["energy_swap"]:
	//	service.MenuNavigateEnergyExchange(_lang, db, message, bot)
	case callbackQuery.Data == "click_energy_swap":
		service.MenuNavigateEnergyExchange(_lang, db, callbackQuery.Message, bot)
	case callbackQuery.Data == "click_transaction_plan":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_smart_transaction_plan":
		service.MenuNavigateSmartTransactionPlans(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_language":
		service.MenuNavigateHome2(db, callbackQuery.Message, bot)
	case callbackQuery.Data == "dispatch_Now_Others":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["enter_address"]+"\n\n")
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
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "dispatch_others", expiration)

	case callbackQuery.Data == "back_address_detection_home":

		service.MenuNavigateAddressDetection(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case strings.HasPrefix(callbackQuery.Data, "dispatch_others_"):
		bundleAddress := strings.ReplaceAll(callbackQuery.Data, "dispatch_others_", "")

		bundleID := strings.Split(bundleAddress, "_")[0]
		address := strings.Split(bundleAddress, "_")[1]

		fmt.Printf("bundleID %s\n", bundleID)
		fmt.Printf("address %s\n", address)

		//trxfee
		userPackageSubscriptionsRepo := repositories.NewUserPackageSubscriptionsRepository(db)
		record, _ := userPackageSubscriptionsRepo.Query(context.Background(), bundleID)

		restTimes := record.Times - 1

		//time.Sleep(100 * time.Millisecond)
		if restTimes >= 0 {
			userPackageSubscriptionsRepo.UpdateTimes(context.Background(), record.Id, restTimes)

			//
			msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, db, callbackQuery.Message.Chat.ID)
			bot.Send(msg2)

			//手动发能

			var sysOrder domain.UserEnergyOrders
			orderNo, _ := GenerateOrderID(address, 4)
			//fmt.Printf("  OrderNo: %s\n", orderNo)
			sysOrder.OrderNo = orderNo
			sysOrder.TxId = ""
			sysOrder.FromAddress = address
			//sysOrder.ToAddress = item.Address
			sysOrder.Amount = 65000
			sysOrder.ChatId = strconv.FormatInt(callbackQuery.Message.Chat.ID, 10)
			//
			////添加一条记录
			ueoRepo := repositories.NewUserEnergyOrdersRepo(db)
			errsg := ueoRepo.Create(context.Background(), &sysOrder)

			if errsg == nil {
				trxfeeClient := trxfee.NewTrxfeeClient(_trxfeeUrl, _trxfeeApiKey, _trxfeeSecret)

				fmt.Sprintf("发送（%d）笔能量给（%s），订单号 %s\n", 1, address, orderNo)
				trxfeeClient.Order(orderNo, address, 65_000*1)

				msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "📢【✅"+global.Translations[_lang]["UShield_sent_transaction_energy"]+"】\n\n"+
					global.Translations[_lang]["to_address"]+address+"\n\n"+
					global.Translations[_lang]["remaining_transactions"]+strconv.FormatInt(restTimes, 10)+"\n\n")

				inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("⚡️"+global.Translations[_lang]["dispatch_again"], "click_bundle_package_address_stats"),
					),
				)
				msg.ReplyMarkup = inlineKeyboard
				msg.ParseMode = "HTML"
				bot.Send(msg)

				expiration := 1 * time.Minute // 短时间缓存空值

				//设置用户状态
				cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "null_dispatch_others_", expiration)
			}
		}

		//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "📢【✅"+global.Translations[_lang]["UShield_sent_transaction_energy"]+"】\n\n"+
		//	global.Translations[_lang]["to_address"]+address+"\n\n"+
		//	global.Translations[_lang]["remaining_transactions"]+strconv.FormatInt(restTimes, 10)+"\n\n")
		//msg.ParseMode = "HTML"
		//bot.Send(msg)

	case strings.HasPrefix(callbackQuery.Data, "confirm_freeze_risk_"):
		address := strings.ReplaceAll(callbackQuery.Data, "confirm_freeze_risk_", "")

		fmt.Printf("address : %s\n", address)
		sysDictionariesRepo := repositories.NewSysDictionariesRepo(db)
		server_trx_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_trx_price")
		server_usdt_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_usdt_price")
		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
		if !CompareStringsWithFloat(user.TronAmount, server_trx_price, 1) && !CompareStringsWithFloat(user.Amount, server_usdt_price, 1) {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "⚠️ "+global.Translations[_lang]["freeze_alert_service_insufficient_balance"]+"\n\n")
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
				),
			)

			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			return
		}
		fmt.Println("余额充足")
		var COST_FROM_TRX bool
		var COST_FROM_USDT bool
		if CompareStringsWithFloat(user.TronAmount, server_trx_price, 1) || CompareStringsWithFloat(user.Amount, server_usdt_price, 1) {

			if CompareStringsWithFloat(user.TronAmount, server_trx_price, float64(1)) {
				rest, _ := SubtractStringNumbers(user.TronAmount, server_trx_price, float64(1))

				user.TronAmount = rest
				userRepo.Update2(context.Background(), &user)
				fmt.Printf("rest: %s", rest)
				COST_FROM_TRX = true
				//扣usdt
			} else if CompareStringsWithFloat(user.Amount, server_usdt_price, float64(1)) {
				rest, _ := SubtractStringNumbers(user.Amount, server_usdt_price, float64(1))
				fmt.Printf("rest: %s", rest)
				user.Amount = rest
				userRepo.Update2(context.Background(), &user)
				COST_FROM_USDT = true
			}

			//添加记录
			userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)

			var event domain.UserAddressMonitorEvent
			event.ChatID = callbackQuery.Message.Chat.ID
			event.Status = 1
			event.Address = address

			if len(address) == 42 {
				event.Network = "Ethereum"
			}
			if len(address) == 34 {
				event.Network = "Tron"
			}

			event.Days = 1
			if COST_FROM_TRX {
				event.Amount = server_trx_price + " TRX"
			}
			if COST_FROM_USDT {
				event.Amount = server_usdt_price + " USDT"
			}
			userAddressEventRepo.Create(context.Background(), &event)

			//后台跟踪起来
			//user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID,
				"✅"+global.Translations[_lang]["enable_freeze_alerts_success"]+"\n"+
					global.Translations[_lang]["address"]+"："+address+"\n"+
					global.Translations[_lang]["network"]+event.Network)
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["alert_monitoring_list"], "address_list_trace"),
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)

		}

	case strings.HasPrefix(callbackQuery.Data, "set_bundle_package_default_"):
		target := strings.ReplaceAll(callbackQuery.Data, "set_bundle_package_default_", "")
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

		errsg := userOperationPackageAddressesRepo.Update(context.Background(), callbackQuery.Message.Chat.ID, target)
		if errsg != nil {
			log.Printf("errsg: %s", errsg)
			return
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅"+"<b>"+"设置默认地址成功 "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case strings.HasPrefix(callbackQuery.Data, "remove_bundle_package_"):
		target := strings.ReplaceAll(callbackQuery.Data, "remove_bundle_package_", "")
		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

		var record domain.UserOperationPackageAddresses
		record.Status = 0
		record.Address = target
		record.ChatID = callbackQuery.Message.Chat.ID

		errsg := userOperationPackageAddressesRepo.Remove(context.Background(), callbackQuery.Message.Chat.ID, target)
		if errsg != nil {
			log.Printf("errsg: %s", errsg)
			return
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "✅"+"<b>"+global.Translations[_lang]["address_deleted_success"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)
		//service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(cache, bot, callbackQuery.Message.Chat.ID, db)
		msg2 := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(_lang, db, callbackQuery.Message.Chat.ID)
		bot.Send(msg2)
	case strings.HasPrefix(callbackQuery.Data, "close_freeze_risk_"):
		target := strings.ReplaceAll(callbackQuery.Data, "close_freeze_risk_", "")

		log.Println("target:", target)
		userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)
		event, _ := userAddressEventRepo.Find(context.Background(), target)

		restDays := fmt.Sprintf("%d", 30-event.Days)

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["confirm_stop_monitoring_address"]+"\n"+
			global.Translations[_lang]["address"]+"："+event.Address+"\n"+
			strings.ReplaceAll(global.Translations[_lang]["confirm_stop_monitoring_address_tips"], "{days}", restDays))
		msg.ParseMode = "HTML"
		// 当点击"按钮 1"时显示内联键盘
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅"+global.Translations[_lang]["confirm_stop_monitoring_address_yes"], "close_risk_"+target),
				tgbotapi.NewInlineKeyboardButtonData("❌"+global.Translations[_lang]["cancel_freeze_alerts"], "back_risk_home"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

	case strings.HasPrefix(callbackQuery.Data, "close_risk_"):
		target := strings.ReplaceAll(callbackQuery.Data, "close_risk_", "")
		log.Println("target:", target)
		userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)
		err := userAddressEventRepo.Close(context.Background(), target)
		if err == nil {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["confirm_stop_monitoring_address_success_tips"])
			msg.ParseMode = "HTML"
			// 当点击"按钮 1"时显示内联键盘
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["alert_monitoring_list"], "address_list_trace"),
					tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"),
				),
			)
			msg.ReplyMarkup = inlineKeyboard

			bot.Send(msg)
		}
	case strings.HasPrefix(callbackQuery.Data, "apply_bundle_package_"):

		target := strings.ReplaceAll(callbackQuery.Data, "apply_bundle_package_", "")
		service.APPLY_BUNDLE_PACKAGE_ADDRESS(_lang, target, cache, bot, callbackQuery.Message, db)

	case strings.HasPrefix(callbackQuery.Data, "config_bundle_package_address_"):

		target := strings.ReplaceAll(callbackQuery.Data, "config_bundle_package_address_", "")
		service.CONFIG_BUNDLE_PACKAGE_ADDRESS(_lang, target, cache, bot, callbackQuery.Message, db)
	case callbackQuery.Data == "click_backup_account":

		//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "👥欢迎使用第二通知人服务"+"\n"+
		//	"为确保实时接收预警信息，您可绑定一个第二通知人TG帐号。"+"\n"+
		//	"绑定前请确保第二通知人已与本机器人互动，绑定后该账号将同步接收预警信息，第二通知人替换请重复绑定步骤，系统将自动替换。请输入的第二通知人TG帐号@用户名 👇")
		//
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["secondary_contact_tips"])
		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_home"),
				//tgbotapi.NewInlineKeyboardButtonData("第二紧急通知", ""),
			),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "click_backup_account", expiration)

	case callbackQuery.Data == "back_user_address_trace":
		service.MenuNavigateAddressTrace(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "back_risk_home":
		service.MenuNavigateAddressFreeze(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "click_switch_trx":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_switch_usdt":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "USDT")

	case callbackQuery.Data == "click_switch_trx_ST":
		service.MenuNavigateSTBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_switch_usdt_ST":
		service.MenuNavigateSTBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "USDT")

	case callbackQuery.Data == "back_bundle_package_ST":
		service.MenuNavigateSTBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")

	case callbackQuery.Data == "back_bundle_package":
		service.MenuNavigateBundlePackage(_lang, db, callbackQuery.Message.Chat.ID, bot, "TRX")
	case callbackQuery.Data == "click_bundle_package_address_manager_config":
		service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGER_CONFIG(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "click_bundle_package_address_manager_remove":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["energy_address_remove_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)

	case callbackQuery.Data == "click_bundle_package_address_manager_add":

		userOperationPackageAddressesRepo := repositories.NewUserOperationPackageAddressesRepo(db)

		list, _ := userOperationPackageAddressesRepo.Query(context.Background(), callbackQuery.Message.Chat.ID)
		if len(list) >= 4 {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "<b>"+global.Translations[_lang]["energy_address_limit_tips"]+"</b>"+"\n")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "<b>"+global.Translations[_lang]["energy_address_limit"]+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
		//笔数套餐地址列表
	case callbackQuery.Data == "click_bundle_package_address_stats":
		msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS2(_lang, db, callbackQuery.Message.Chat.ID)
		bot.Send(msg)

	case callbackQuery.Data == "click_bundle_package_address_stats_ST":
		//msg := service.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS_ST(_lang, db, callbackQuery.Message.Chat.ID)
		//bot.Send(msg)
		catfee.CLICK_BUNDLE_PACKAGE_ADDRESS_STATS_ST(_lang, cache, db, callbackQuery.Message.Chat.ID, bot)

	case strings.HasPrefix(callbackQuery.Data, "custody_address_check_"):
		{
			catfee.CheckOption(_lang, db, callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, callbackQuery.Data, bot, catfeeClient)
		}
	case callbackQuery.Data == "next_bundle_package_address_stats":
		if service.NEXT_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "prev_bundle_package_address_stats":
		state, done := service.PREV_BUNDLE_PACKAGE_ADDRESS_STATS(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "click_bundle_package_address_management":
		service.CLICK_BUNDLE_PACKAGE_ADDRESS_MANAGEMENT(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)
	case callbackQuery.Data == "address_list_trace":
		service.ADDRESS_LIST_TRACE(_lang, cache, bot, callbackQuery, db)
	case callbackQuery.Data == "back_home":
		service.MenuNavigateHome(_lang, cache, db, callbackQuery.Message, bot)
	case callbackQuery.Data == "click_business_cooperation":
		service.ClickBusinessCooperation(_lang, callbackQuery, bot)
	case callbackQuery.Data == "click_offical_channel":
		service.ClickOfficalChannel(_lang, callbackQuery, bot)
	case callbackQuery.Data == "click_callcenter":
		service.ClickCallCenter(_lang, callbackQuery, bot)
	case callbackQuery.Data == "click_my_recepit":
		service.CLICK_MY_RECEPIT(_lang, db, callbackQuery, bot)
	case callbackQuery.Data == "address_freeze_risk_records":
		msg := service.ExtractAddressRiskQuery(_lang, db, callbackQuery)
		bot.Send(msg)
	case callbackQuery.Data == "user_detection_cost_records":
		msg := service.ExtractAddressDetection(_lang, cache, db, callbackQuery)
		bot.Send(msg)
	case callbackQuery.Data == "click_bundle_package_cost_records":
		msg := service.ExtractBundlePackage(_lang, db, callbackQuery)
		bot.Send(msg)

	case callbackQuery.Data == "click_bundle_package_cost_records_ST":
		//msg := service.ExtractBundlePackageST(_lang, db, callbackQuery)
		//bot.Send(msg)
		msg := catfee.ExtractBundlePackageST(_lang, db, callbackQuery)
		bot.Send(msg)

	case callbackQuery.Data == "prev_st_bundle_package_page":
		state, done := catfee.EXTRACT_PREV_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "next_st_bundle_package_page":
		if catfee.EXTRACT_NEXT_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot) {
			return
		}

	case callbackQuery.Data == "catfee_add_address":
		catfee.CustodyAddressCond(_lang, cache, db, bot, callbackQuery)

	case callbackQuery.Data == "catfee_remove_address":
		catfee.CustodyRemoveAddressCond(_lang, cache, db, bot, callbackQuery)

	case callbackQuery.Data == "click_bundle_package_management":
		msg := service.ExtractBundlePackage(_lang, db, callbackQuery)
		bot.Send(msg)
	case callbackQuery.Data == "click_deposit_usdt_records":
		service.CLICK_DEPOSIT_USDT_RECORDS(_lang, db, callbackQuery, bot)
	case callbackQuery.Data == "click_deposit_trx_records":
		service.CLICK_DEPOSIT_TRX_RECORDS(_lang, db, callbackQuery, bot)
	case callbackQuery.Data == "next_address_detection_page":
		if service.EXTRACT_NEXT_ADDRESS_DETECTION_PAGE(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "prev_address_detection_page":
		state, done := service.EXTRACT_PREV_ADDRESS_DETECTION_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)
	case callbackQuery.Data == "prev_deposit_usdt_page":
		state, done := service.EXTRACT_PREV_DEPOSIT_USDT_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)
	case callbackQuery.Data == "prev_deposit_trx_page":
		state, done := service.EXTRACT_PREV_DEPOSIT_TRX_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)
	case callbackQuery.Data == "prev_address_risk_page":
		state, done := service.EXTRACT_PREV_ADDRESS_RISK_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "next_address_risk_page":
		if service.ExtraNextAddressRiskPage(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "next_deposit_usdt_page":
		if service.ExtraNextDepositUSDTPage(_lang, callbackQuery, db, bot) {
			return
		}
	case callbackQuery.Data == "next_deposit_trx_page":
		if service.ExtracNextDepositTrxPage(_lang, callbackQuery, db, bot) {
			return
		}

	case callbackQuery.Data == "prev_bundle_package_page":
		state, done := service.EXTRACT_PREV_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot)
		if done {
			return
		}
		fmt.Printf("state: %v\n", state)

	case callbackQuery.Data == "next_bundle_package_page":
		if service.EXTRACT_NEXT_BUNDLE_PACKAGE_PAGE(_lang, callbackQuery, db, bot) {
			return
		}

	case callbackQuery.Data == "click_QA":
		service.ExtraQA(_lang, cache, bot, callbackQuery)

	case callbackQuery.Data == "user_backup_notify":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+"<b>"+"请输入需添加的第二紧急通知用户电报ID: "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "start_freeze_risk_1":
		//查看余额
		service.START_FREEZE_RISK_1(_lang, cache, db, callbackQuery, bot)

	case callbackQuery.Data == "click_my_service":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "🛡 当前服务状态：\n\n🔋 能量闪兑\n\n- 剩余笔数：12\n- 自动补能：关闭 /开启\n\n➡️ /闪兑\n\n➡️ /笔数套餐\n\n➡️ /手动发能（1笔）\n\n➡️ /开启/关闭自动发能\n\n📍 地址风险检测\n\n- 今日免费次数：已用完\n\n➡️ /地址风险检测\n\n🚨 USDT冻结预警\n\n- 地址1：TX8kY...5a9rP（剩余12天）✅\n- 地址2：TEw9Q...iS6Ht（剩余28天）✅")
		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["alert_monitoring_list"], "address_list_trace"),
				//	tgbotapi.NewInlineKeyboardButtonData("地址管理", "address_manager"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "usdt_risk_monitor", expiration)

	case callbackQuery.Data == "stop_freeze_risk_1":

		//删除event表里面
		userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)

		userAddressEventRepo.RemoveAll(context.Background(), callbackQuery.Message.Chat.ID)

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "已经暂停所有监控")
		msg.ParseMode = "HTML"

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "reset", expiration)

	case callbackQuery.Data == "start_freeze_risk_0":

		sysDictionariesRepo := repositories.NewSysDictionariesRepo(db)

		server_trx_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_trx_price")

		server_usdt_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_usdt_price")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "欢迎使用U盾 USDT冻结预警服务\n"+
			"🛡️ U盾，做您链上资产的护盾！\n"+
			"地址一旦被链上风控冻，资产将难以追回，损失巨大！\n"+
			"每天都有数百个 USDT 钱包地址被冻结锁定，风险就在身边！\n"+
			"✅ 适用于经常收付款 / 被制裁地址感染/与诈骗地址交互\n"+
			"✅ 支持TRON/ETH网络的USDT 钱包地址\n"+
			"📌 服务价格（每地址）：\n • "+server_trx_price+" TRX / 30天\n • "+
			" 或 "+server_usdt_price+" USDT / 30天\n"+
			"🎯 服务开启后U盾将24 小时不间断保护您的资产安全。\n"+
			"⏰ 系统将在冻结前启动预警机制，持续 10 分钟每分钟推送提醒，通知您及时转移资产。\n"+
			"📩 所有预警信息将通过 Telegram 实时推送")
		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["enable_freeze_alert"], "start_freeze_risk"),
				//tgbotapi.NewInlineKeyboardButtonData("地址管理", "address_manager"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["alert_monitoring_list"], "address_list_trace"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["freeze_alert_deduction_record"], "address_freeze_risk_records"),
			),
			//tgbotapi.NewInlineKeyboardRow(
			//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["freeze_alert_deduction_record"], "address_freeze_risk_records"),
			//	//tgbotapi.NewInlineKeyboardButtonData("第二紧急通知", ""),
			//),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "usdt_risk_monitor", expiration)
	case callbackQuery.Data == "stop_freeze_risk":

		log.Println("========================================stop_freeze_risk==================================================")

		userAddressEventRepo := repositories.NewUserAddressMonitorEventRepo(db)
		addresses, _ := userAddressEventRepo.Query(context.Background(), callbackQuery.Message.Chat.ID)

		//msg.ParseMode = "HTML"

		var allButtons []tgbotapi.InlineKeyboardButton
		var extraButtons []tgbotapi.InlineKeyboardButton
		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, item := range addresses {
			allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData(TruncateString(item.Address), "close_freeze_risk_"+fmt.Sprintf("%d", item.Id)))
		}

		extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔙️"+global.Translations[_lang]["back_homepage"], "back_risk_home"))

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

		//msg.ReplyMarkup = inlineKeyboard
		//
		//bot.Send(msg)
		//
		//expiration := 1 * time.Minute // 短时间缓存空值
		//
		////设置用户状态
		//cache.Set(strconv.FormatInt(_chatID, 10), "start_freeze_risk", expiration)
		//
		//msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "📡 是否确认停止该服务？")
		//msg.ParseMode = "HTML"
		//
		//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//	tgbotapi.NewInlineKeyboardRow(
		//		tgbotapi.NewInlineKeyboardButtonData("✅ 确认停止", "stop_freeze_risk_1"),
		//		tgbotapi.NewInlineKeyboardButtonData("❌ 取消操作", "start_freeze_risk_0"),
		//	),
		//tgbotapi.NewInlineKeyboardRow(
		//	tgbotapi.NewInlineKeyboardButtonData("地址", ""),
		//),
		//)
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["monitoring_address_list"]+"\n\n")
		//地址绑定

		msg.ParseMode = "HTML"

		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		//expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		//cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "stop_freeze_risk", expiration)

	case callbackQuery.Data == "start_freeze_risk":

		sysDictionariesRepo := repositories.NewSysDictionariesRepo(db)
		server_trx_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_trx_price")
		server_usdt_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_usdt_price")
		userRepo := repositories.NewUserRepository(db)
		user, _ := userRepo.GetByUserID(callbackQuery.Message.Chat.ID)
		if !CompareStringsWithFloat(user.TronAmount, server_trx_price, 1) && !CompareStringsWithFloat(user.Amount, server_usdt_price, 1) {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "⚠️ "+global.Translations[_lang]["freeze_alert_service_insufficient_balance"]+"\n\n")
			msg.ParseMode = "HTML"
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💵"+global.Translations[_lang]["deposit"], "deposit_amount"),
				),
			)

			msg.ReplyMarkup = inlineKeyboard
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["enter_address_for_alert"])
		msg.ParseMode = "HTML"
		bot.Send(msg)
		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "start_freeze_risk", expiration)

	case callbackQuery.Data == "address_manager_return":

		sysDictionariesRepo := repositories.NewSysDictionariesRepo(db)

		server_trx_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_trx_price")

		server_usdt_price, _ := sysDictionariesRepo.GetDictionaryDetail("server_usdt_price")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "欢迎使用U盾 USDT冻结预警服务\n"+
			"🛡️ U盾，做您链上资产的护盾！\n"+
			"地址一旦被链上风控冻，资产将难以追回，损失巨大！\n"+
			"每天都有数百个 USDT 钱包地址被冻结锁定，风险就在身边！\n"+
			"✅ 适用于经常收付款 / 被制裁地址感染/与诈骗地址交互\n"+
			"✅ 支持TRON/ETH网络的USDT 钱包地址\n"+
			"📌 服务价格（每地址）：\n • "+server_trx_price+" TRX / 30天\n • "+
			" 或 "+server_usdt_price+" USDT / 30天\n"+
			"🎯 服务开启后U盾将24 小时不间断保护您的资产安全。\n"+
			"⏰ 系统将在冻结前启动预警机制，持续 10 分钟每分钟推送提醒，通知您及时转移资产。\n"+
			"📩 所有预警信息将通过 Telegram 实时推送")
		msg.ParseMode = "HTML"

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["enable_freeze_alert"], "start_freeze_risk"),
				//	tgbotapi.NewInlineKeyboardButtonData("地址管理", "address_manager"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["alert_monitoring_list"], "address_list_trace"),
				tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["freeze_alert_deduction_record"], "address_freeze_risk_records"),
			),
			//tgbotapi.NewInlineKeyboardRow(
			//	tgbotapi.NewInlineKeyboardButtonData(global.Translations[_lang]["freeze_alert_deduction_record"], "address_freeze_risk_records"),
			//	//tgbotapi.NewInlineKeyboardButtonData("第二紧急通知", ""),
			//),
		)
		msg.ReplyMarkup = inlineKeyboard

		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), "usdt_risk_monitor", expiration)

	case callbackQuery.Data == "address_manager_add":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+"<b>"+"请输入需添加的地址: "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "address_manager_remove":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "💬"+"<b>"+"请输入需删除的地址: "+"</b>"+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "address_manager":
		service.ADDRESS_MANAGER(_lang, cache, bot, callbackQuery.Message.Chat.ID, db)

	case strings.HasPrefix(callbackQuery.Data, "set_lang_"):
		lang := strings.ReplaceAll(callbackQuery.Data, "set_lang_", "")
		expiration := 24 * time.Hour // 短时间缓存空值
		cache.Set("LANG_"+strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), lang, expiration)
		//数据库设置用户的默认选项语言

		userRepo := repositories.NewUserRepository(db)
		userRepo.UpdateLang(lang, callbackQuery.Message.Chat.ID)

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[lang]["set_lang"]+"\n")

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[lang]["back_home"], "back_home"),
			),
		)
		msg.ReplyMarkup = inlineKeyboard
		msg.ParseMode = "HTML"
		bot.Send(msg)

		handleStartCommand(cache, bot, callbackQuery.Message)

	case strings.HasPrefix(callbackQuery.Data, "ST_bundle_"):
		//service.ST_BUNDLE_CHECK(_lang, cache, bot, callbackQuery, db)
		catfee.ST_BUNDLE_CHECK(_lang, cache, bot, callbackQuery, db)

	case strings.HasPrefix(callbackQuery.Data, "bundle_"):
		service.BUNDLE_CHECK2(_lang, cache, bot, callbackQuery, db)
		//调用trxfee接口进行笔数扣款
	case strings.HasPrefix(callbackQuery.Data, "deposit_usdt"):
		service.DepositPrevUSDTOrder(_lang, cache, bot, callbackQuery, db)
		//responseText = "你选择了选项 A"
	case strings.HasPrefix(callbackQuery.Data, "deposit_trx"):
		service.DepositPrevOrder(_lang, cache, bot, callbackQuery, db)
	case callbackQuery.Data == "cancel_order":
		service.DepositCancelOrder(_lang, cache, bot, callbackQuery, db)
	case callbackQuery.Data == "cancel_catfee_order":
		service.DepositCancelOrder(_lang, cache, bot, callbackQuery, db)
	case callbackQuery.Data == "address_trace_add":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["address_trace_add_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "address_trace_delete":
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["address_trace_delete_tips"]+"\n")
		msg.ParseMode = "HTML"
		bot.Send(msg)

		expiration := 1 * time.Minute // 短时间缓存空值

		//设置用户状态
		cache.Set(strconv.FormatInt(callbackQuery.Message.Chat.ID, 10), callbackQuery.Data, expiration)
	case callbackQuery.Data == "deposit_amount":

		service.DEPOSIT_AMOUNT(_lang, db, callbackQuery, bot)

	case callbackQuery.Data == "forward_deposit_usdt":

		fmt.Printf("\nforward_deposit_usdt\n")
		usdtSubscriptionsRepo := repositories.NewUserUsdtSubscriptionsRepository(db)

		usdtlist, err := usdtSubscriptionsRepo.ListAll(context.Background())

		if err != nil {

		}
		var allButtons []tgbotapi.InlineKeyboardButton
		var extraButtons []tgbotapi.InlineKeyboardButton
		var keyboard [][]tgbotapi.InlineKeyboardButton
		for _, usdtRecord := range usdtlist {
			allButtons = append(allButtons, tgbotapi.NewInlineKeyboardButtonData("💰"+usdtRecord.Name, "deposit_usdt_"+usdtRecord.Amount))
		}

		extraButtons = append(extraButtons, tgbotapi.NewInlineKeyboardButtonData("🔁"+global.Translations[_lang]["switch_to_trx_deposit"], "deposit_amount"), tgbotapi.NewInlineKeyboardButtonData("🔙"+global.Translations[_lang]["back_home"], "back_home"))

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

	case callbackQuery.Data == "click_visa":

		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_visa_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case callbackQuery.Data == "click_sim":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_sim_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)

	case callbackQuery.Data == "click_energy_financing":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_energy_financing_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case callbackQuery.Data == "click_sns":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_sns_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	case callbackQuery.Data == "click_ecs":
		dictRepo := repositories.NewSysDictionariesRepo(db)
		ushield_additional_services_contact, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_contact")
		ushield_additional_services_wallet, _ := dictRepo.GetDictionaryDetail("ushield_additional_services_wallet")

		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, global.Translations[_lang]["ushield_additional_services_ecs_desc"]+"\n"+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_contact"], "{ushield_additional_services_contact}", ushield_additional_services_contact)+strings.ReplaceAll(global.Translations[_lang]["ushield_additional_services_wallet"], "{ushield_additional_services_wallet}", ushield_additional_services_wallet))
		msg.ParseMode = "HTML"
		bot.Send(msg)
	default:
		//responseText = "未知选项"
	}

	// 发送新消息作为响应
	bot.Send(tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, responseText))
}
