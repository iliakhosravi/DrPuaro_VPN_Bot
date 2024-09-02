package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/panel"
	bp "techybat.org/go-vpn/widgets/buttonpage"
	"techybat.org/go-vpn/widgets/form"
)

type PackEditHandler func(ctx context.Context, b *bot.Bot, update *tmodels.Update, pack models.Pack)

func AddPackController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	catKeyboard, typeKeyboard := makeCatKeyboard(db), makeTypeKeyboard()

	fields := []form.Field{
		{
			Name:        "traffic",
			MessageText: "میزان حجم بسته بر حسب گیگ چقدر باشد؟ لطفا صرفا عدد صحیح مثبت وارد نمایید.",
			Validator:   models.PackValidator("traffic"),
		},
		{
			Name:        "period",
			MessageText: "دوره زمانی بسته بر حسب روز چقدر است؟ لطفا صرفا عدد صحیح مثبت وارد نمایید",
			Validator:   models.PackValidator("period"),
		},
		{
			Name:        "price",
			MessageText: "قیمت بسته برحسب تومان چقدر است؟ لطفا صرفا عدد صحیح مثبت وارد نمایید.",
			Validator:   models.PackValidator("price"),
		},
		{
			Name:        "category",
			MessageText: "یکی از دسته بندی های زیر را برای بسته موردنظر انتخاب کنید",
			Type:        form.ButtonField,
			Keyboard:    catKeyboard,
			Validator:   models.PackValidator("category_id"),
		},
		{
			Name:        "type",
			MessageText: "نوع بسته چه باشد؟",
			Type:        form.ButtonField,
			Keyboard:    typeKeyboard,
			Validator:   models.PackValidator("type"),
		},
	}
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	userID := update.CallbackQuery.From.ID
	form := form.CreateForm(cancelBtnText, fields, chatID, userID, packSubmitController, onCancelPack, nil)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        "افزودن بسته (4 مرحله)",
		ReplyMarkup: nil,
	})

	form.Show(ctx, b, update)
}

func packSubmitController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	traffic, _ := strconv.Atoi(form.FindField("traffic").Value)
	period, _ := strconv.Atoi(form.FindField("period").Value)
	price, _ := strconv.Atoi(form.FindField("price").Value)
	categoryID, _ := strconv.ParseUint(form.FindField("category").Value, 10, 0)
	packType := models.PackType(form.FindField("type").Value)

	pack := models.Pack{
		Traffic:    traffic,
		Period:     period,
		Price:      price,
		CategoryID: uint(categoryID),
		Type:       packType,
	}

	if pack.Type == models.SanaeiPack {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: form.ChatID,
			Text:   "درحال دریافت inbound ها از پنل...",
		})
		inBtns := createInboundsBtns(passPack(onPackInboundSubmit, pack))
		inboundsPage := bp.CreateButtonPage(bot.EscapeMarkdown("کدام یک از inbound های زیر به کاربر اختصاص یابد؟\nتوجه کنید که این لیست از پنل سنایی شما استخراج شده است."), inBtns, 5, true)
		inboundsPage.Show(ctx, b, form.ChatID)
		fmt.Println("AFTER SHOW BUTTON PAGE")
		return
	}

	txtMsg := "افزودن بسته با موفقیت انجام شد"

	if err := pack.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, form.ChatID)
}

func onPackInboundSubmit(ctx context.Context, b *bot.Bot, update *tmodels.Update, pack models.Pack) {
	inboundID, _ := strconv.Atoi(update.CallbackQuery.Data)
	pack.InboundID = inboundID

	txtMsg := "بسته با موفقیت ثبت شد"

	if err := pack.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
	}

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, chatID)
}

func onCancelPack(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}
	ShowAdminDialog(ctx, b, update, chatID)
}

func EditPackController(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	var packs []models.Pack
	db.Find(&packs)

	buttons := []dialog.Button{}
	for _, pack := range packs {
		buttons = append(buttons, dialog.Button{
			ID:              fmt.Sprintf("pack%d", pack.ID),
			Text:            pack.String(),
			CallbackHandler: HandleEditPack,
			CallbackData:    fmt.Sprint(pack.ID),
		})
	}

	title := bot.EscapeMarkdown("شما می توانید بسته های ثبت شده زیر را ویرایش کنید.\nبسته موردنظر را انتخاب کنید:")
	buttonPage := bp.CreateButtonPage(title, buttons, 5, true)
	buttonPage.Show(ctx, b, update.CallbackQuery.Message.Message.Chat.ID)

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
}

func HandleEditPack(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	db := database.GetDB()
	chatID := update.CallbackQuery.From.ID
	packID := update.CallbackQuery.Data
	var pack models.Pack
	db.Preload(clause.Associations).Find(&pack, packID)

	txtMsg := pack.FullStr() + "\n\nقصد انجام چه کاری را با این بسته دارید؟"

	nodes := []dialog.Node{
		{
			ID:   "edit-pack",
			Text: bot.EscapeMarkdown(txtMsg),
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "edit",
						Text:            "ویرایش",
						CallbackHandler: onEditPack,
						CallbackData:    packID,
					},
				},
			},
		},
		{
			ID:   "remove-pack",
			Text: bot.EscapeMarkdown("آیا از غیرفعال سازی(حذف) این بسته مطمئن هستید؟"),
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "remove",
						Text:            "بله",
						CallbackHandler: onRemovePack,
						CallbackData:    packID,
					},
					{
						Text:   "خیر",
						NodeID: "edit-pack",
					},
				},
			},
		},
		{
			ID:   "active-pack",
			Text: bot.EscapeMarkdown("آیا از فعال سازی این بسته مطمئن هستید؟"),
			Keyboard: [][]dialog.Button{
				{
					{
						ID:              "active",
						Text:            "بله",
						CallbackHandler: onActivePack,
						CallbackData:    packID,
					},
					{
						Text:   "خیر",
						NodeID: "edit-pack",
					},
				},
			},
		},
	}

	if pack.Status != models.UnactivePack {
		nodes[0].Keyboard[0] = append(nodes[0].Keyboard[0], dialog.Button{
			Text:   "غیرفعال سازی(حذف)",
			NodeID: "remove-pack",
		})
	}

	if pack.Status != models.ActivePack {
		nodes[0].Keyboard[0] = append(nodes[0].Keyboard[0], dialog.Button{
			Text:   "فعال سازی",
			NodeID: "active-pack",
		})
	}

	dg := dialog.New(nodes, dialog.Inline())
	dg.Show(ctx, b, chatID, "edit-pack")
}

func onActivePack(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	packID := update.CallbackQuery.Data
	db := database.GetDB()
	var pack models.Pack
	db.Find(&pack, packID)

	txtMsg := "بسته با موفقیت فعال شد."
	if err := pack.Active(db); err != nil {
		fmt.Println("Unable to active pack. err: ", err)
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})
}

func onRemovePack(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID
	packID := update.CallbackQuery.Data
	db := database.GetDB()
	var pack models.Pack
	db.Find(&pack, packID)

	txtMsg := "بسته با موفقیت غیرفعال شد."
	if err := pack.Deactive(db); err != nil {
		fmt.Println("Unable to remove pack. err: ", err)
		txtMsg = "خطایی پیش آمده"
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   txtMsg,
	})
}

func onEditPack(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	chatID := update.CallbackQuery.Message.Message.Chat.ID

	db := database.GetDB()
	packID, _ := strconv.ParseUint(update.CallbackQuery.Data, 10, 0)
	var pack models.Pack
	db.Preload("Category").Order("created_at desc").Find(&pack, packID)

	catKeyboard, typeKeyboard := makeCatKeyboard(db), makeTypeKeyboard()

	fields := []form.Field{
		{
			Name:        "traffic",
			MessageText: fmt.Sprintf("میزان حجم بسته بر حسب گیگ چقدر باشد؟ لطفا صرفا عدد صحیح مثبت وارد نمایید.\nمقدار فعلی:%v", pack.Traffic),
			Validator:   models.PackValidator("traffic"),
			Value:       fmt.Sprint(pack.Traffic),
			IsSkippable: true,
		},
		{
			Name:        "period",
			MessageText: fmt.Sprintf("دوره زمانی بسته بر حسب روز چقدر است؟ لطفا صرفا عدد صحیح مثبت وارد نمایید\nمقدار فعلی:%v", pack.Period),
			Validator:   models.PackValidator("period"),
			IsSkippable: true,
			Value:       fmt.Sprint(pack.Period),
		},
		{
			Name:        "price",
			MessageText: fmt.Sprintf("قیمت بسته برحسب تومان چقدر است؟ لطفا صرفا عدد صحیح مثبت وارد نمایید.\nمقدار فعلی:%v", pack.Price),
			Validator:   models.PackValidator("price"),
			IsSkippable: true,
			Value:       fmt.Sprint(pack.Price),
		},
		{
			Name:        "category",
			MessageText: fmt.Sprintf("یکی از دسته بندی های زیر را برای بسته موردنظر انتخاب کنید\nمقدار فعلی:%s", pack.Category.Name),
			Type:        form.ButtonField,
			Keyboard:    catKeyboard,
			Validator:   models.PackValidator("category_id"),
			IsSkippable: true,
			Value:       fmt.Sprint(pack.Category.Name),
		},
		{
			Name:        "type",
			MessageText: "نوع بسته چه باشد؟",
			Type:        form.ButtonField,
			Keyboard:    typeKeyboard,
			Value:       string(pack.Type),
			IsSkippable: true,
			Validator:   models.PackValidator("type"),
		},
	}
	form := form.CreateForm("انصراف", fields, chatID, update.CallbackQuery.From.ID, passPack(onEditPackSubmit, pack), onCancelPack, nil)
	form.SkipButtonText = "مقدار فعلی"
	form.SkipMessageText = "مقدار فعلی برای این ورودی قرار گرفت"
	form.Show(ctx, b, update)

	b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      chatID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
	})
}

func makeTypeKeyboard() [][]tmodels.InlineKeyboardButton {
	return [][]tmodels.InlineKeyboardButton{
		{
			{
				Text:         "خام(بدون پنل)",
				CallbackData: string(models.CustomPack),
			},
			{
				Text:         "سنایی",
				CallbackData: string(models.SanaeiPack),
			},
		},
	}
}

func makeCatKeyboard(db *gorm.DB) [][]tmodels.InlineKeyboardButton {
	var categories []models.Category
	db.Find(&categories)
	var catKeyboard = [][]tmodels.InlineKeyboardButton{}

	for _, category := range categories {
		catKeyboard = append(catKeyboard, []tmodels.InlineKeyboardButton{
			{
				Text:         category.Name,
				CallbackData: strconv.FormatUint(uint64(category.ID), 10),
			},
		})
	}

	return catKeyboard
}

func onEditPackSubmit(ctx context.Context, b *bot.Bot, update *tmodels.Update, pack models.Pack) {
	form := ctx.Value(form.FORM_KEY).(*form.Form)
	traffic, _ := strconv.Atoi(form.FindField("traffic").Value)
	period, _ := strconv.Atoi(form.FindField("period").Value)
	price, _ := strconv.Atoi(form.FindField("price").Value)
	categoryID, err := strconv.ParseUint(form.FindField("category").Value, 10, 0)

	pack.Traffic = traffic
	pack.Period = period
	pack.Price = price

	if err == nil {
		pack.CategoryID = uint(categoryID)
		pack.Category = models.Category{}
	}

	if pack.Type == models.SanaeiPack {
		inBtns := createInboundsBtns(passPack(onPackInboundSubmit, pack))
		inboundsPage := bp.CreateButtonPage("کدام یک از inbound های زیر به کاربر اختصاص یابد؟\nتوجه کنید که این لیست از پنل سنایی شما استخراج شده است.", inBtns, 5, true)
		inboundsPage.Show(ctx, b, form.ChatID)
		return
	}

	txtMsg := "ویرایش دسته بندی با موفقیت انجام شد"
	if err := pack.Store(database.GetDB()); err != nil {
		txtMsg = "خطایی پیش آمده"
		fmt.Println("Error:", err)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: form.ChatID,
		Text:   txtMsg,
	})

	ShowAdminDialog(ctx, b, update, form.ChatID)
}

func passPack(next PackEditHandler, pack models.Pack) bot.HandlerFunc {
	return func(ctx context.Context, bot *bot.Bot, update *tmodels.Update) {
		next(ctx, bot, update, pack)
	}
}

func createInboundsBtns(handler bot.HandlerFunc) []dialog.Button {
	btns := []dialog.Button{}
	p := panel.GetPanel()
	inbounds, _ := p.GetInbounds()
	fmt.Println("THIS IS AFTER GET INBOUNDS: ", inbounds)
	for _, inbound := range inbounds {
		btn := dialog.Button{
			ID:              fmt.Sprint(inbound.ID),
			Text:            inbound.Remark,
			CallbackHandler: handler,
			CallbackData:    fmt.Sprint(inbound.ID),
		}

		btns = append(btns, btn)
	}
	fmt.Println("\n\nTHESE ARE INBOUND_BTNS: ", btns)
	return btns
}
