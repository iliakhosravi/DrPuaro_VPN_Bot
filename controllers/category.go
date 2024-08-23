package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

type CatEditHandler func(ctx context.Context, b *bot.Bot, update *tmodels.Update, cat models.Category)

var cancelBtnText string = "انصراف"

func AddCatController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	fields := []form.Field{
		{
			Name:        "cat_name",
			MessageText: "نام دسته بندی چه باشد؟",
		},
		{
			Name:        "cat_description",
			MessageText: "توضیحات دسته بندی چه باشد؟",
		},
	}
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	form := form.CreateForm(cancelBtnText, fields, chatID, userID, catSubmitController, onCancelCat, nil)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "افزودن دسته بندی (2 مرحله)",
		ReplyMarkup: nil,
	})

	form.Show(ctx, b, update)
}

func catSubmitController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)

	category := models.Category{
		Name:        form.FindField("cat_name").Value,
		Description: form.FindField("cat_description").Value,
	}

	txtMsg := "افزودن دسته بندی با موفقیت انجام شد"
	if err := category.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, form.ChatID)
}

func onCancelCat(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	ShowAdminDialog(ctx, b, update, chatID)
}

func EditCatController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	var cats []models.Category
	db.Order("created_at desc").Find(&cats)

	buttons := []dialog.Button{}
	for _, cat := range cats {
		buttons = append(buttons, dialog.Button{
			ID:              fmt.Sprintf("cat%d", cat.ID),
			Text:            cat.Name,
			CallbackHandler: HandleEditCat,
			CallbackData:    fmt.Sprint(cat.ID),
		})
	}

	title := bot.EscapeMarkdown("شما می توانید دسته بندی های ثبت شده زیر را ویرایش کنید.\nدسته بندی موردنظر را انتخاب کنید:")
	buttonPage := buttonpage.CreateButtonPage(title, buttons, 5, true)
	if _, err := buttonPage.Show(ctx, b, update.CallbackQuery.Message.Message.Chat.ID); err != nil {
		fmt.Println("Error: cannot show edit category buttonpage: ", err)
	}

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
}

func HandleEditCat(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	db := database.GetDB()
	catID := update.CallbackQuery.Data
	var cat models.Category
	db.Find(&cat, catID)

	txtMsg := cat.String() + "\nقصد انجام چه کاری را با این دسته بندی دارید؟"

	nodes := []dialog.Node{
		{
			ID:   "edit-cat",
			Text: bot.EscapeMarkdown(txtMsg),
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "edit",
						Text:            "ویرایش",
						CallbackHandler: onEditCat,
						CallbackData:    catID,
					},
				},
			},
		},
		{
			ID:   "remove-cat",
			Text: bot.EscapeMarkdown("آیا از غیرفعال سازی(حذف) این دسته بندی مطمئن هستید؟"),
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "remove",
						Text:            "بله",
						CallbackHandler: onRemoveCat,
						CallbackData:    catID,
					},
					{
						Text:   "خیر",
						NodeID: "edit-cat",
					},
				},
			},
		},
		{
			ID:   "active-cat",
			Text: bot.EscapeMarkdown("آیا از فعال سازی این دسته بندی مطمئن هستید؟"),
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "active",
						Text:            "بله",
						CallbackHandler: onActiveCat,
						CallbackData:    catID,
					},
					{
						Text:   "خیر",
						NodeID: "edit-cat",
					},
				},
			},
		},
	}

	if cat.Status != models.UnactiveCat {
		nodes[0].Keyboard[0] = append(nodes[0].Keyboard[0], dialog.Button{
			Text:   "غیرفعال سازی(حذف)",
			NodeID: "remove-cat",
		})
	}

	if cat.Status != models.ActiveCat {
		nodes[0].Keyboard[0] = append(nodes[0].Keyboard[0], dialog.Button{
			Text:   "فعال سازی",
			NodeID: "active-cat",
		})
	}

	dg := dialog.New(nodes, dialog.Inline())

	dg.Show(ctx, b, chatID, "edit-cat")
}

func onActiveCat(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	catID := update.CallbackQuery.Data
	db := database.GetDB()
	var cat models.Category
	db.Find(&cat, catID)

	txtMsg := "دسته بندی با موفقیت فعال شد."
	if err := cat.Active(db); err != nil {
		fmt.Println("Unable to remove category. err: ", err)
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})
}

func onRemoveCat(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	catID := update.CallbackQuery.Data
	db := database.GetDB()
	var cat models.Category
	db.Find(&cat, catID)

	txtMsg := "دسته بندی با موفقیت غیرفعال شد."
	if err := cat.Deactive(db); err != nil {
		fmt.Println("Unable to remove category. err: ", err)
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})
}
func onEditCat(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	db := database.GetDB()
	catID, _ := strconv.ParseUint(update.CallbackQuery.Data, 10, 0)
	var cat models.Category
	db.Find(&cat, catID)

	fields := []form.Field{
		{
			Name:        "cat_name",
			MessageText: fmt.Sprintf("نام دسته بندی چه باشد؟\nنام فعلی:%s", cat.Name),
			Value:       cat.Name,
			IsSkippable: true,
		},
		{
			Name:        "cat_description",
			MessageText: fmt.Sprintf("توضیحات دسته بندی چه باشد؟\nتوضیحات فعلی:%s", cat.Description),
			Value:       cat.Description,
			IsSkippable: true,
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, update.CallbackQuery.From.ID, passCategory(onEditCatSubmit, cat), onCancelCat, nil)
	form.SkipButtonText = "مقدار فعلی"
	form.SkipMessageText = "مقدار فعلی برای این ورودی قرار گرفت"
	form.Show(ctx, b, update)

	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      chatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
	})
}

func onEditCatSubmit(ctx context.Context, b *bot.Bot, update *tmodels.Update, cat models.Category) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)

	cat.Name = form.FindField("cat_name").Value
	cat.Description = form.FindField("cat_description").Value

	txtMsg := "ویرایش دسته بندی با موفقیت انجام شد"
	if err := cat.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, form.ChatID)
}

func passCategory(next CatEditHandler, cat models.Category) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
		next(ctx, bot, update, cat)
	}
}
