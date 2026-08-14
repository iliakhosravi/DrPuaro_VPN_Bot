package renew

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/database"
	"techybat.org/go-vpn/marz"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/panel"
	"techybat.org/go-vpn/sui"
)

var (
	expiringUntil string
	newEndDate    string
)

var RenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew expiring configs and orders",
	Long: `Renew configs and orders that are expiring until a given date.
	
Examples:
  govpn renew --expiring-until 2026-02-28 --new-end-date 2026-03-31
  govpn renew -e 2026-02-28 -n 2026-03-31`,
	Run: func(cmd *cobra.Command, args []string) {
		if expiringUntil == "" || newEndDate == "" {
			log.Fatal("Both --expiring-until and --new-end-date flags are required")
		}

		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		database.Setup()
		db := database.GetDB()

		if err := RenewConfigs(db, expiringUntil, newEndDate); err != nil {
			log.Fatalf("Error renewing configs: %v", err)
		}
	},
}

func init() {
	RenewCmd.Flags().StringVarP(&expiringUntil, "expiring-until", "e", "", "Renew configs expiring until this date (format: YYYY-MM-DD)")
	RenewCmd.Flags().StringVarP(&newEndDate, "new-end-date", "n", "", "New end date for renewed configs (format: YYYY-MM-DD)")
	RenewCmd.MarkFlagRequired("expiring-until")
	RenewCmd.MarkFlagRequired("new-end-date")
}

func RenewConfigs(db *gorm.DB, expiringUntilStr, newEndDateStr string) error {
	// Parse dates
	expiringUntilDate, err := time.Parse("2006-01-02", expiringUntilStr)
	if err != nil {
		return fmt.Errorf("invalid expiring-until date format: %w", err)
	}

	newEndDateParsed, err := time.Parse("2006-01-02", newEndDateStr)
	if err != nil {
		return fmt.Errorf("invalid new-end-date format: %w", err)
	}

	// Find all active configs
	var configs []models.Config
	err = db.Preload(clause.Associations).
		Preload("Order.Pack").
		Joins("JOIN orders ON orders.id = configs.order_id").
		Find(&configs).Error

	if err != nil {
		return fmt.Errorf("error fetching configs: %w", err)
	}

	fmt.Printf("Found %d configs. Checking expiration dates...\n", len(configs))

	renewed := 0
	skipped := 0
	failed := 0

	for _, config := range configs {
		endDate, err := config.EndDate(db)
		if err != nil {
			fmt.Printf("Warning: Could not get end date for config %d: %v\n", config.ID, err)
			failed++
			continue
		}

		// Check if config expires before or on the expiringUntil date
		if endDate.Time().Before(expiringUntilDate) || endDate.Time().After(newEndDateParsed) {
			skipped++
			continue
		}

		fmt.Printf("Renewing config %d (order %d) - Current end date: %s\n",
			config.ID, config.OrderID, endDate.Time().Format("2006-01-02"))

		// Renew the config
		if err := renewConfig(db, &config, newEndDateParsed); err != nil {
			fmt.Printf("Error renewing config %d: %v\n", config.ID, err)
			failed++
		} else {
			renewed++
		}
	}

	fmt.Printf("\nRenewal Summary:\n")
	fmt.Printf("  Renewed: %d\n", renewed)
	fmt.Printf("  Skipped: %d\n", skipped)
	fmt.Printf("  Failed:  %d\n", failed)
	fmt.Printf("  Total:   %d\n", len(configs))

	return nil
}

func renewConfig(db *gorm.DB, config *models.Config, newEndDate time.Time) error {
	var c models.Config
	db.Preload(clause.Associations).Preload("Order.Pack").Find(&c, config.ID)

	newExpiryTime := newEndDate.UnixMilli()

	// Update based on pack type
	if c.Order.Pack.Type == models.SanaeiPack {
		p := panel.GetPanel()
		_, err := p.GetClient(c.Email)
		if err != nil {
			return fmt.Errorf("unable to get client from panel: %w", err)
		}

		// Prepare the client form for update
		gb, mb := c.Order.Pack.TrafficGbMb()
		clientForm := panel.ClientForm{
			ID:         c.UUID,
			Email:      c.Email,
			TotalGB:    int64(gb*panel.ONE_GB + mb*panel.ONE_MB),
			ExpiryTime: newExpiryTime,
			Enable:     true,
			TgID:       fmt.Sprint(c.Order.User.TelID),
			SubID:      c.SubID,
			LimitIP:    int(c.Order.Pack.LimitIP),
		}

		if _, err := p.UpdateClient(c.Order.Pack.InboundID, clientForm); err != nil {
			return fmt.Errorf("unable to update client in panel: %w", err)
		}

		fmt.Printf("  ✓ Updated Sanaei config %d (email: %s)\n", c.ID, c.Email)
	} else if c.Order.Pack.Type == models.SUIPack {
		s := sui.GetSui()
		clientID, _ := strconv.Atoi(c.SubID)
		client, err := s.GetClientByID(clientID)
		if err != nil {
			return fmt.Errorf("unable to get client from S-UI: %w", err)
		}

		client.Enable = true
		client.Expiry = newEndDate.Unix()

		if _, err := s.UpdateClient(*client); err != nil {
			return fmt.Errorf("unable to update client in S-UI: %w", err)
		}

		fmt.Printf("  ✓ Updated S-UI config %d (client ID: %s)\n", c.ID, c.SubID)
	} else if c.Order.Pack.Type == models.MarzPack {
		mz := marz.GetMarz()
		user, err := mz.GetUser(c.Email)
		if err != nil {
			return fmt.Errorf("unable to get user from Marzneshin: %w", err)
		}

		expireDate := newEndDate.Format("2006-01-02T15:04:05.999999")
		user.ExpireStrategy = marz.ExpireStrategyFixedDate
		user.ExpireDate = &expireDate
		user.Enabled = true

		if _, err := mz.UpdateUser(c.Email, *user); err != nil {
			return fmt.Errorf("unable to update user in Marzneshin: %w", err)
		}

		fmt.Printf("  ✓ Updated Marzneshin config %d (username: %s)\n", c.ID, c.Email)
	} else {
		// For custom configs, update the start date to reflect the new period
		daysToAdd := int(time.Until(newEndDate).Hours() / 24)
		c.StartDate = time.Now().AddDate(0, 0, -c.Order.Pack.Period+daysToAdd)

		if result := db.Save(&c); result.RowsAffected == 0 {
			return fmt.Errorf("unable to update custom config in database: %v", result.Error)
		}

		fmt.Printf("  ✓ Updated custom config %d\n", c.ID)
	}

	if err := c.Sync(db); err != nil {
		return fmt.Errorf("unable to sync config: %w", err)
	}

	return nil
}
