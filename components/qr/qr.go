package qr

import (
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

func GenerateQRLogo(link, logo_path, save_path string) error {
	qrc, err := qrcode.New(link)
	if err != nil {
		return err
	}

	w, err := standard.New(save_path, standard.WithLogoImageFilePNG(logo_path))
	if err != nil {
		return err
	}

	err = qrc.Save(w)
	return err
}
