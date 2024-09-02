package admin_controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

func ChargeOrdersHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	chargeType := m.ChargeType(update.CallbackQuery.Data)
	btns := createChargeOrderBtns(db, chargeType)
	bp := buttonpage.CreateButtonPage("درخواست شارژ مورد نظر را برای بررسی انتخاب کنید:", btns, 5, true)
	bp.Show(ctx, b, chatID)
}

func ChargeOrderHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	db := database.GetDB()
	orderID := update.CallbackQuery.Data
	var order m.ChargeOrder
	db.Preload(clause.Associations).Find(&order, orderID)

	b.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     chatID,
		FromChatID: os.Getenv("STORAGE_CHANNEL_ID"),
		MessageID:  order.Receipt(db).MessageID,
	})

	txtMsg := fmt.Sprintf("اطلاعات درخواست شارژ\n%s\n\nکاربر درخواست کننده:\n%s", order.FullStr(), order.User.String())

	nodes := []dialog.Node{
		{
			ID:   "charge-desc",
			Text: txtMsg,
			Keyboard: [][]dialog.Button{
				{
					{Text: "تایید ✅", NodeID: "accept-charge"},
					{Text: "رد ❌", NodeID: "dismiss-charge"},
				},
			},
		},
		{
			ID:   "accept-charge",
			Text: "آیا از تایید این درخواست شارژ اطمینان دارید؟",
			Keyboard: [][]dialog.Button{
				{
					{ID: "accept", Text: "بله 🟢", CallbackHandler: ChargeHandler, CallbackData: "accept_" + orderID},
					{Text: "خیر 🟥", NodeID: "charge-desc"},
				},
			},
		},
		{
			ID:   "dismiss-charge",
			Text: "آیا از رد درخواست شارژ اطمینان دارید؟",
			Keyboard: [][]dialog.Button{
				{
					{ID: "dismiss", Text: "بله 🟥", CallbackHandler: ChargeHandler, CallbackData: "dismiss_" + orderID},
					{Text: "خیر 🟢", NodeID: "charge-desc"},
				},
			},
		},
	}

	if order.Type != m.PendingCharge {
		nodes[0].Keyboard = [][]dialog.Button{}
	}

	dialog := dialog.New(nodes, dialog.Inline())
	dialog.Show(ctx, b, chatID, "charge-desc")
}

func ChargeHandler(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.From.ID
	userID := update.CallbackQuery.From.ID
	fields := []form.Field{
		{
			Name:        "admin-note",
			MessageText: "یادداشت خود را برای کاربر بگذارید",
		},
	}
	data := strings.Split(update.CallbackQuery.Data, "_")
	var handler func(ctx context.Context, b *bot.Bot, update *tmodels.Update, orderID string)
	if data[0] == "accept" {
		handler = onAcceptCharge
	} else if data[0] == "dismiss" {
		handler = onDismissCharge
	}
	form := form.CreateForm("انصراف", fields, chatID, userID, passChargeOrderID(handler, data[1]), onCancelCharge, nil)
	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      chatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
	})
	form.Show(ctx, b, update)

}
func onCancelCharge(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
}

func onAcceptCharge(ctx context.Context, b *bot.Bot, update *tmodels.Update, orderID string) {
	db := database.GetDB()
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	chatID := form.ChatID
	var order m.ChargeOrder
	db.Preload(clause.Associations).Find(&order, orderID)
	order.AdminNote = form.FindField("admin-note").Value

	var txtMsg string = "درخواست شارژ با موفقیت تایید شد."
	if err := order.AcceptCharge(db); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})

	userTxtMsg := fmt.Sprintf("درخواست شارژ شما با مشخصات زیر تایید شده است:\n%s", order.FullStr())
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: order.User.TelID,
		Text:   userTxtMsg,
	})
}

func onDismissCharge(ctx context.Context, b *bot.Bot, update *tmodels.Update, orderID string) {
	db := database.GetDB()
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	chatID := form.ChatID
	var order m.ChargeOrder
	db.Preload(clause.Associations).Find(&order, orderID)
	order.AdminNote = form.FindField("admin-note").Value

	var txtMsg string = "درخواست شارژ با موفقیت رد شد."
	if err := order.DismissCharge(db); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})

	userTxtMsg := fmt.Sprintf("درخواست شارژ شما با مشخصات زیر رد شده است:\n%s", order.FullStr())
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: order.User.TelID,
		Text:   userTxtMsg,
	})
}

func createChargeOrderBtns(db *gorm.DB, chargeType m.ChargeType) []dialog.Button {
	btns := []dialog.Button{}
	var orders []m.ChargeOrder
	db.Where("type = ?", chargeType).Preload(clause.Associations).Find(&orders)
	for _, order := range orders {
		btn := dialog.Button{
			ID:              fmt.Sprintf("c-order#%d", order.ID),
			Text:            fmt.Sprintf("%s | %d تومان", order.User.Fullname(), order.Amount),
			CallbackHandler: ChargeOrderHandler,
			CallbackData:    fmt.Sprintf("%d", order.ID),
		}
		btns = append(btns, btn)
	}
	return btns
}

func passChargeOrderID(next func(ctx context.Context, bot *bot.Bot, update *tmodels.Update, orderID string), orderID string) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
		next(ctx, bot, update, orderID)
	}
}
