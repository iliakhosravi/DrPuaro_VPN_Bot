package controllers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"gorm.io/gorm"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/middlewares/auth"
	"techybat.org/go-vpn/models"
	bp "techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

func BuyController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	packId := update.CallbackQuery.Data
	var pack models.Pack
	var packMsg string
	if result := db.Where("status = ?", models.ActivePack).First(&pack, packId); result.RowsAffected == 0 {
		packMsg = "این بسته در دسترس نمی باشد"
	} else {
		packMsg = fmt.Sprintf("شما بسته %s را برای خرید انتخاب کرده اید:", pack.String())
	}

	nodes := []dialog.Node{
		{
			ID:   "buy-options",
			Text: "از چه طریقی می خواهید پرداخت انجام شود؟",
			Keyboard: [][]dialog.Button{
				{
					{
						Text: "کیف پول", NodeID: "wallet",
					},
					{
						ID: "card", Text: "واریز به حساب", CallbackHandler: CardBuyController, CallbackData: packId,
					},
				},
			},
		},
		{
			ID:   "wallet",
			Text: "آیا از پرداخت با کیف پول خود اطمینان دارید؟",
			Keyboard: [][]dialog.Button{
				{
					{ID: "wallet-btn", Text: "بله ✅", CallbackHandler: WalletBuyController, CallbackData: packId},
					{
						Text: "خیر ❌", NodeID: "buy-options",
					},
				},
			},
		},
	}
	dialog := dialog.New(nodes, dialog.Inline())

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        packMsg,
		ReplyMarkup: nil,
	})

	dialog.Show(ctx, b, chatID, "buy-options")
}

func WalletBuyController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	user := ctx.Value(auth.UserKey).(models.User)
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	packId, _ := strconv.Atoi(update.CallbackQuery.Data)
	var pack models.Pack

	db.First(&pack, packId)

	if user.Charge < uint64(pack.Price) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "موجودی شما کافی نیست!",
		})
		return
	}

	var txtMsg string
	order, err := user.BuyPackByCharge(db, &pack)
	if err != nil {
		txtMsg = "خطایی پیش آمده"
	} else {
		txtMsg = fmt.Sprintf("سفارش شما ایجاد و تایید شد.\nاطلاعات سفارش:%s", order.FullStr(db))
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
}

func CardBuyController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	user := ctx.Value(auth.UserKey).(models.User)

	var card models.Card = models.GetActiveCard(db)
	txtMsg := "جهت پرداخت مبلغ ذکر شده را به شماره کارت زیر واریز کرده و سپس تصویری از فیش واریزی را در یک پیام ارسال کنید. پس از این مرحله خرید شما در وضعیت نیاز به تایید قرار گرفته و با تایید نهایی از سوی ادمین کانفیگ به صورت خودکار برای شما ارسال خواهد شد."
	if (card != models.Card{}) {
		txtMsg = fmt.Sprintf("%s\nشماره کارت: %s\nبه نام: %s", txtMsg, card.Number, card.Fullname)
	} else {
		txtMsg = "خطایی پیش آمده"
	}

	packId, _ := strconv.Atoi(update.CallbackQuery.Data)
	var pack models.Pack
	if result := db.First(&pack, packId); result.RowsAffected == 0 {
		txtMsg = "خطایی پیش آمده"
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID

	fields := []form.Field{
		{
			Name:          "receipt",
			MessageText:   txtMsg,
			Type:          form.CustomTextField,
			CustomHandler: onCardReceipt,
		},
	}
	form := form.CreateForm("انصراف از خرید", fields, chatID, user.TelID, passPack(onRetrieveReceipt, pack), onCancelRecipt, nil)
	form.Show(ctx, b, update)

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
}

func onCardReceipt(ctx context.Context, b *bot.Bot, update *tmodels.Update, form form.Form, setter form.FieldSetter) (bool, error) {
	msg, err := b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     os.Getenv("STORAGE_CHANNEL_ID"),
		FromChatID: fmt.Sprintf("%d", update.Message.Chat.ID),
		MessageID:  update.Message.ID,
	})

	if err != nil {
		fmt.Println("unable to forward receipt to storage channel\nerror:", err)
		return false, err
	}

	ok, strErr := setter(fmt.Sprintf("%d", msg.ID))

	return ok, fmt.Errorf(strErr)
}

func onCancelRecipt(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "خرید شما با موفقیت لغو شد",
	})

	ShowMainDialog(ctx, b, update)
}
func ChargeHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	db := database.GetDB()
	var card models.Card = models.GetActiveCard(db)
	txtMsg := "جهت پرداخت مبلغ ذکر شده را به شماره کارت زیر واریز کرده و سپس تصویری از فیش واریزی را در یک پیام ارسال کنید. پس از این مرحله شارژ شما در وضعیت نیاز به تایید قرار گرفته و با تایید نهایی از سوی ادمین به صورت خودکار اکانت شما شارژ خواهد شد."
	if (card != models.Card{}) {
		txtMsg = fmt.Sprintf("%s\nشماره کارت: %s\nبه نام: %s", txtMsg, card.Number, card.Fullname)
	} else {
		txtMsg = "خطایی پیش آمده"
	}
	fields := []form.Field{
		{
			Name:        "amount",
			MessageText: "میزانی که می خواهید شارژ کنید را به تومان وارد کنید.",
			Validator:   models.MoneyValidator,
		},
		{
			Name:          "receipt",
			Type:          form.CustomTextField,
			MessageText:   txtMsg,
			CustomHandler: onChargeReceipt,
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, onSubmitCharge, onCancelCharge, nil)
	form.Show(ctx, b, update)
}

func onSubmitCharge(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	user := ctx.Value(auth.UserKey).(models.User)
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	amount, _ := strconv.ParseUint(form.FindField("amount").Value, 0, 0)
	msgID, _ := strconv.Atoi(form.FindField("receipt").Value)

	chargeOrder := models.ChargeOrder{
		UserID: user.ID,
		Amount: uint(amount),
		Type:   models.PendingCharge,
	}

	txtMsg := "درخواست شارژ شما با موفقیت ثبت شد. کد رسید: "
	err := db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(&chargeOrder); result.Error != nil {
			return result.Error
		}
		receipt, err := chargeOrder.AddReceipt(tx, msgID)
		if err != nil {
			return err
		}

		txtMsg += fmt.Sprintf("%d", receipt.ID)
		return nil
	})

	if err != nil {
		fmt.Println("Error: unable to save chargeOrder. err: ", err)
		txtMsg = "خطایی پیش آمده"
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   txtMsg,
		ChatID: form.ChatID,
	})
}

func onCancelCharge(ctx context.Context, b *bot.Bot, update *tmodels.Update) {}

func onChargeReceipt(ctx context.Context, b *bot.Bot, update *tmodels.Update, form form.Form, setter form.FieldSetter) (bool, error) {
	msg, err := b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     os.Getenv("STORAGE_CHANNEL_ID"),
		FromChatID: fmt.Sprintf("%d", update.Message.Chat.ID),
		MessageID:  update.Message.ID,
	})

	if err != nil {
		fmt.Println("unable to forward receipt to storage channel\nerror:", err)
		return false, fmt.Errorf("unable to forward receipt to storage channel. error: %s", err)
	}

	ok, strErr := setter(fmt.Sprintf("%d", msg.ID))
	return ok, fmt.Errorf(strErr)
}

func onRetrieveReceipt(ctx context.Context, b *bot.Bot, update *tmodels.Update, pack models.Pack) {
	db := database.GetDB()

	user := ctx.Value(auth.UserKey).(models.User)
	form := ctx.Value(form.FORM_KEY).(*form.Form)

	msgID, _ := strconv.ParseInt(form.FindField("receipt").Value, 0, 0)
	order, err := user.BuyPackByCard(db, &pack, int(msgID))

	var txtMsg string
	if err != nil {
		txtMsg = "خطایی پیش آمده"
	} else {
		txtMsg = fmt.Sprintf("درخواست شما با موفقیت ثبت شد. کد رسید: %d", order.GetReceipt(db).ID)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   txtMsg,
	})

	ShowMainDialog(ctx, b, update)
}

func VerifyBuyController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	var orders []models.Order
	db.Where(models.Order{Type: models.PendingOrder}).Preload("Pack").Find(&orders)

	orderBtns := []dialog.Button{}

	for _, order := range orders {
		orderBtns = append(orderBtns, dialog.Button{
			ID:              fmt.Sprintf("Order%d", order.ID),
			Text:            order.Pack.String(),
			CallbackHandler: onWatchOrder,
			CallbackData:    strconv.FormatUint(uint64(order.ID), 10),
		})
	}

	buttonPage := bp.CreateButtonPage(bot.EscapeMarkdown("لیست درخواست های ارسالی:"), orderBtns, 5, true)

	buttonPage.Show(ctx, b, update.CallbackQuery.Message.Message.Chat.ID)
}

func onWatchOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()

	orderID, _ := strconv.ParseUint(update.CallbackQuery.Data, 10, 0)
	var order models.Order
	db.Where(orderID).Preload("User").Preload("Pack").Preload("Pack.Category").Find(&order)

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	_, err := b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     chatID,
		FromChatID: os.Getenv("STORAGE_CHANNEL_ID"),
		MessageID:  order.GetReceipt(db).MessageID,
	})

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "خطایی پیش آمده یا درخواست ارسالی دردسترس نمی باشد",
		})
	}

	txtMsg := fmt.Sprintf("شماره رسید:%d\nنام کاربر:%s\nآیدی کاربر:%d\nنام کاربری:%s\n\nبسته درخواستی:%s\nگروه بسته درخواستی:%s\n",
		order.GetReceipt(db).ID,
		order.User.Fullname(),
		order.User.TelID,
		order.User.Username,
		order.Pack.String(),
		order.Pack.Category.Name,
	)

	fields := []form.Field{
		{
			Name:        "order",
			MessageText: txtMsg,
			Type:        form.ButtonField,
			Keyboard: [][]tmodels.InlineKeyboardButton{
				{
					{
						Text:         "تایید✅",
						CallbackData: "true_" + update.CallbackQuery.Data,
					},
					{
						Text:         "رد🚫",
						CallbackData: "false_" + update.CallbackQuery.Data,
					},
				},
			},
		},

		{
			Name:        "carry_msg",
			MessageText: "در جواب چه پیامی به کاربر ارسال شود؟",
			Type:        form.TextField,
		},
	}
	form := form.CreateForm("انصراف از ادامه پروسه", fields, chatID, update.CallbackQuery.From.ID, onSubmitOrder, onCancelOrder, nil)
	form.Description = "شروع"
	form.Show(ctx, b, update)
}

func onSubmitOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	db := database.GetDB()
	orderField := form.FindField("order").Value
	splitedOrderField := strings.Split(orderField, "_")
	ok, strOrderID := splitedOrderField[0], splitedOrderField[1]
	orderID, _ := strconv.ParseUint(strOrderID, 10, 0)
	var order models.Order
	db.Preload("User").Preload("Pack").First(&order, orderID)
	order.AdminNote = form.FindField("carry_msg").Value
	var txtMsg, orderResult string
	if ok == "true" {
		err := order.Verify(db, order.AdminNote)
		if err != nil {
			txtMsg = "خطایی پیش آمده"
			fmt.Println("Unable to verify order. err: ", err)
		} else {
			txtMsg = fmt.Sprintf("سفارش کاربر تایید شد. کد درخواست: %d", order.ID)
			orderResult = "<b>تایید شده✅</b>"
		}

	} else {
		err := order.Dismiss(db, order.AdminNote)
		if err != nil {
			txtMsg = "خطایی پیش آمده"
		} else {
			txtMsg = fmt.Sprintf("سفارش کاربر رد شد. کد درخواست: %d", orderID)
			orderResult = "<b>رد شده❌</b>"
		}
	}

	carryMsg := fmt.Sprintf("سفارش شما بررسی شد.\nبسته انتخابی:%s\nشماره سفارش: %d\nنتیجه:%s\nتوضیحات ادمین:%s", order.Pack.String(), orderID, orderResult, order.AdminNote)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    order.User.TelID,
		Text:      carryMsg,
		ParseMode: tmodels.ParseModeHTML,
	})
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})
	ShowAdminDialog(ctx, b, update, form.ChatID)

}

func onCancelOrder(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	ShowAdminDialog(ctx, b, update, chatID)
}
