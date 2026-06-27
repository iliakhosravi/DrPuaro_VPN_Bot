package msgTool

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	ptime "github.com/yaa110/go-persian-calendar"
	"techybat.org/go-vpn/database"
)

func SendBackup(ctx context.Context, b *bot.Bot, chatID any) {
	fileName := fmt.Sprintf("bot_%s.sql", ptime.Now().Format("yyyy_MM_d_HH_mm"))
	backupPath := fmt.Sprintf("./backup/%s", fileName)
	err := database.DumpMySQL(backupPath)
	if err == nil {
		fileContent, _ := os.ReadFile(backupPath)
		_, err := b.SendDocument(ctx, &bot.SendDocumentParams{
			ChatID:   chatID,
			Document: &models.InputFileUpload{Filename: fileName, Data: bytes.NewReader(fileContent)},
		})
		if err != nil {
			fmt.Println("tel backup err: ", err)
		}
	} else {
		fmt.Println("backup err: ", err)
	}
}
