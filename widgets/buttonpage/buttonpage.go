package buttonpage

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tmodels "github.com/go-telegram/bot/models"
	"github.com/sinasadeghi83/go-telegram-bot-ui/dialog"
)

type ButtonPage struct {
	Title     string
	Dialog    *dialog.Dialog
	Buttons   []dialog.Button
	PerPage   int
	WithClose bool
}

func CreateButtonPage(title string, buttons []dialog.Button, perPage int, withClose bool) *ButtonPage {
	buttonPage := &ButtonPage{
		Title:     title,
		Buttons:   buttons,
		PerPage:   perPage,
		WithClose: withClose,
	}

	buttonPage.initNodes()

	return buttonPage
}

func (buttonPage *ButtonPage) Show(ctx context.Context, b *bot.Bot, chatID any) {
	buttonPage.Dialog.Show(ctx, b, chatID, "Page1")
}

func (buttonPage *ButtonPage) initNodes() {
	nodes := []dialog.Node{}
	page := 1
	pagesCount := len(buttonPage.Buttons) / buttonPage.PerPage
	if len(buttonPage.Buttons)%buttonPage.PerPage != 0 {
		pagesCount++
	} else if pagesCount == 0 {
		page = 0
	}
	currentNode := dialog.Node{
		ID:       "Page1",
		Text:     bot.EscapeMarkdown(fmt.Sprintf("%s\n(%d/%d)", buttonPage.Title, page, pagesCount)),
		Keyboard: [][]dialog.Button{},
	}

	for _, btn := range buttonPage.Buttons {
		if len(currentNode.Keyboard) == buttonPage.PerPage {
			currentNode.Keyboard = append(currentNode.Keyboard, makeTransitionKeyboard(pagesCount, page, buttonPage.WithClose))
			page++
			nodes = append(nodes, currentNode)
			currentNode = dialog.Node{
				ID:       fmt.Sprintf("Page%d", page),
				Text:     bot.EscapeMarkdown(fmt.Sprintf("%s\n(%d/%d)", buttonPage.Title, page, pagesCount)),
				Keyboard: [][]dialog.Button{},
			}
			fmt.Println("Page ID:", currentNode.ID)
		}
		currentNode.Keyboard = append(currentNode.Keyboard, []dialog.Button{btn})
	}
	currentNode.Keyboard = append(currentNode.Keyboard, makeTransitionKeyboard(pagesCount, page, buttonPage.WithClose))
	nodes = append(nodes, currentNode)

	buttonPage.Dialog = dialog.New(nodes, dialog.Inline())
}

func makeTransitionKeyboard(pagesCount, page int, withClose bool) []dialog.Button {
	row := []dialog.Button{}

	if page > 1 {
		row = append(row, dialog.Button{
			Text:   "◀️",
			NodeID: fmt.Sprintf("Page%d", page-1),
		})
	}

	if withClose {
		row = append(row, dialog.Button{
			ID:              "close",
			Text:            "❌",
			CallbackHandler: onClose,
		})
	}

	if page != pagesCount && pagesCount != 0 {
		row = append(row, dialog.Button{
			Text:   "▶️",
			NodeID: fmt.Sprintf("Page%d", page+1),
		})
	}

	return row
}

func onClose(ctx context.Context, b *bot.Bot, update *tmodels.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
	})

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
}
