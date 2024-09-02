package models

import (
	"fmt"
	"net/url"
	"os"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/panel"
)

type Config struct {
	BaseModel
	StartDate time.Time `json:"start_date"`
	OrderID   uint      `json:"order_id"`
	Order     Order     `json:"order"`
	Email     string    `json:"email"`
	SubID     string    `json:"subId"`
	UUID      string    `json:"uuid"`
}

func (config *Config) Migrate(db *gorm.DB) {
	db.AutoMigrate(&Config{})
}

func DateValidator(value string) (bool, string) {
	_, err := time.Parse("2006-01-02", value)

	if err != nil {
		fmt.Println("warning: date format wrong ", err)
		return false, "فرمت ورودی نادرست است. لطفا ورودی را دوباره با فرمت درست وارد کنید."
	}
	return err == nil, ""
}

func (config *Config) Link(db *gorm.DB) string {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type == SanaeiPack {
		link, _ := url.JoinPath(fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("PANEL_SUB_URL"), os.Getenv("PANEL_SUB_PORT"), os.Getenv("PANEL_SUB_PATH"), c.SubID))
		return link
	}

	return c.Order.AdminNote
}

func (config *Config) JSONLink(db *gorm.DB) string {
	var c Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)
	if c.Order.Pack.Type == SanaeiPack {
		link, _ := url.JoinPath(fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("PANEL_SUB_URL"), os.Getenv("PANEL_SUB_PORT"), os.Getenv("PANEL_JSON_SUB_PATH"), c.SubID))
		return link
	}

	return c.Order.AdminNote
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

func (config *Config) RemainedTraffic(db *gorm.DB) (float32, error) {
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
	if (client.Enable && c.Order.Type != ActiveOrder) || (!client.Enable && c.Order.Type == ActiveOrder) {
		if client.Enable {
			c.Order.Type = ActiveOrder
		} else {
			c.Order.Type = DepletedOrder
		}
		c.Order.AdminNote = "آخرین تغییر وضعیت بسته توسط سیستم به صورت خودکار انجام شده است."
		err = db.Save(&c.Order).Error
	}
	return err
}
