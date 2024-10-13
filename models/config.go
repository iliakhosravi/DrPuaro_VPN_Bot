package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/google/uuid"
	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/panel"
	"techybat.org/go-vpn/tools/qr"
	"techybat.org/go-vpn/vars"
)

type Config struct {
	BaseModel
	StartDate  time.Time `json:"start_date"`
	OrderID    uint      `json:"order_id"`
	Order      Order     `json:"order"`
	Email      string    `json:"email"`
	SubID      string    `json:"subId"`
	UUID       string    `json:"uuid"`
	CustomLink string    `json:"custom_link"`
}

func (config *Config) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Config{})
}

func (config *Config) Title(db *gorm.DB) string {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	return fmt.Sprintf("%s | %s | %s", vars.Get("BRAND_NAME"), vars.Get("TG_CHANNEL"), c.Order.Pack.Name())
}

func DateValidator(value string) (bool, string) {
	_, err := time.Parse("2006-01-02", value)

	if err != nil {
		fmt.Println("warning: date format wrong ", err)
		return false, "فرمت ورودی نادرست است. لطفا ورودی را دوباره با فرمت درست وارد کنید."
	}
	return err == nil, ""
}

// Returns shortlink, QR Path
func (config *Config) ShortLink(db *gorm.DB) (string, string) {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	shortLink := c.CustomLink
	if c.Order.Pack.Type == SanaeiPack {
		client, _ := c.GetClient()
		shortLinks, _ := panel.GetPanel().ShortLinksConfig(c.Title(db), client)
		shortLink = shortLinks[0]
	}
	qrPath := fmt.Sprintf("./%s/%d.jpg", vars.Get("QR_PATH"), c.ID)
	err := qr.GenerateQRLogo(shortLink, vars.Get("LOGO_PATH"), qrPath)
	if err != nil {
		fmt.Println("Unable to create QR Logo. err: ", err)
	}
	return shortLink, qrPath
}

// Returns subLink, QR Path
func (config *Config) SubLink(db *gorm.DB) (string, string) {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	subLink := c.CustomLink
	if c.Order.Pack.Type == SanaeiPack {
		client, _ := c.GetClient()
		subLink, _ = panel.GetPanel().SubLink(client)
	}
	qrPath := fmt.Sprintf("./%s/%d.jpg", vars.Get("QR_PATH"), c.ID)
	err := qr.GenerateQRLogo(subLink, vars.Get("LOGO_PATH"), qrPath)
	if err != nil {
		fmt.Println("Unable to create QR Logo. err: ", err)
	}
	return subLink, qrPath
}

// Returns panel subLink, QR Path
func (config *Config) PanelSubLink(db *gorm.DB) (string, string) {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	subLink := c.CustomLink
	if c.Order.Pack.Type == SanaeiPack {
		client, _ := c.GetClient()
		subLink, _ = panel.GetPanel().PanelSubLink(client)
	}
	qrPath := fmt.Sprintf("./%s/%d.jpg", vars.Get("QR_PATH"), c.ID)
	err := qr.GenerateQRLogo(subLink, vars.Get("LOGO_PATH"), qrPath)
	if err != nil {
		fmt.Println("Unable to create QR Logo. err: ", err)
	}
	return subLink, qrPath
}

func (config *Config) Link(db *gorm.DB) string {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type == SanaeiPack {
		client, _ := c.GetClient()
		link, _ := panel.GetPanel().PanelSubLink(client)
		return link
	}

	return c.CustomLink
}

func (config *Config) JSONLink(db *gorm.DB) string {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type == SanaeiPack {
		link, _ := url.JoinPath(fmt.Sprintf("http://%s:%s/%s/%s", vars.Get("PANEL_SUB_URL"), vars.Get("PANEL_SUB_PORT"), vars.Get("PANEL_JSON_SUB_PATH"), c.SubID))
		return link
	}

	return c.CustomLink
}

func (config *Config) EndDate(db *gorm.DB) (ptime.Time, error) {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type == SanaeiPack {
		panel := panel.GetPanel()
		client, err := panel.GetClient(config.Email)
		if err != nil {
			fmt.Println("error: unable to retrieve client from panel for endDate", err)
			return ptime.Time{}, err
		}
		return ptime.Unix(client.ExpiryTime/1000, client.ExpiryTime%1000*1000), nil
	}

	return ptime.New(c.StartDate.AddDate(0, 0, c.Order.Pack.Period)), nil
}

func (config *Config) RemainedDateStr(db *gorm.DB) (string, error) {
	endDate, err := config.EndDate(db)
	if err != nil {
		return "", nil
	}

	duration := time.Until(endDate.Time())
	hours := int(duration.Hours())
	days := hours / 24
	hours = hours % 24
	mins := int(duration.Minutes()) % 60

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%dD ", days)
	}

	if hours > 0 {
		result += fmt.Sprintf("%dH ", hours)
	}

	if mins > 0 {
		result += fmt.Sprintf("%dM", mins)
	}

	return strings.TrimSpace(result), nil
}

func (config *Config) DepleteSanaei(db *gorm.DB, o Order) error {
	p := panel.GetPanel()
	gb, mb := o.Pack.TrafficGbMb()

	if config.Email == "" {
		config.Email = fmt.Sprintf("U%d_O%d", o.UserID, o.ID)
	}

	if config.UUID == "" {
		config.UUID = uuid.NewString()
	}

	clientForm := panel.ClientForm{
		ID:         config.UUID,
		Email:      config.Email,
		TotalGB:    int64(gb*panel.ONE_GB + mb*panel.ONE_MB),
		ExpiryTime: config.StartDate.AddDate(0, 0, o.Pack.Period).UnixMilli(),
		Enable:     false,
		TgID:       fmt.Sprint(o.User.TelID),
		SubID:      config.SubID,
		LimitIP:    int(o.Pack.LimitIP),
	}

	if _, err := p.StoreClient(o.Pack.InboundID, clientForm); err != nil {
		return err
	}

	config.UUID = clientForm.ID
	res := db.Save(config)
	return res.Error
}

func (config *Config) SetupSanaei(db *gorm.DB, o Order) error {
	p := panel.GetPanel()
	gb, mb := o.Pack.TrafficGbMb()
	if config.SubID == "" {
		config.SubID = bot.RandomString(16)
	}

	if config.UUID == "" {
		config.UUID = uuid.NewString()
	}

	if res := db.Save(config); res.Error != nil {
		return res.Error
	}

	if config.Email == "" {
		config.Email = fmt.Sprintf("u%d_c%d", o.UserID, config.ID)
	}

	clientForm := panel.ClientForm{
		ID:         config.UUID,
		Email:      config.Email,
		TotalGB:    int64(gb*panel.ONE_GB + mb*panel.ONE_MB),
		ExpiryTime: config.StartDate.AddDate(0, 0, o.Pack.Period).UnixMilli(),
		Enable:     true,
		TgID:       fmt.Sprint(o.User.TelID),
		SubID:      config.SubID,
		LimitIP:    int(o.Pack.LimitIP),
	}

	if _, err := p.StoreClient(o.Pack.InboundID, clientForm); err != nil {
		return err
	}
	if _, err := p.ResetClientStats(o.Pack.InboundID, clientForm.Email); err != nil {
		return err
	}

	if res := db.Save(config); res.Error != nil {
		return res.Error
	}

	return nil
}

func (config *Config) RemainedTraffic(db *gorm.DB) (int, error) {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type == SanaeiPack {
		panel := panel.GetPanel()
		client, err := panel.GetClient(config.Email)
		if err != nil {
			fmt.Println("error unable to retrieve client from panel for remained traffic", err)
			return 0.0, err
		}
		return client.RemainedTraffic(), nil
	}
	return 0, fmt.Errorf("no remained traffic for non-panel configs")
}

func (config *Config) GetClient() (panel.Client, error) {
	panel := panel.GetPanel()
	return panel.GetClient(config.Email)
}

func (config *Config) Sync(db *gorm.DB) error {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type != SanaeiPack {
		return nil
	}

	client, err := c.GetClient()
	if err != nil {
		return err
	}

	return c.SyncByClient(db, client)
}

func (c *Config) SyncByClient(db *gorm.DB, client panel.Client) error {
	var err error = nil
	inClient, err := panel.GetPanel().GetInboundClient(&client)
	if err != nil {
		return err
	}
	c.SubID = inClient.SubID
	c.UUID = inClient.ID
	if (client.Enable && c.Order.Type != ActiveOrder) || (!client.Enable && c.Order.Type == ActiveOrder) {
		if client.Enable {
			c.Order.Type = ActiveOrder
		} else {
			c.Order.Type = DepletedOrder
		}
		c.Order.AdminNote = "آخرین تغییر وضعیت بسته توسط سیستم به صورت خودکار انجام شده است."
	}
	err = db.Save(&c.Order).Error

	return err
}
