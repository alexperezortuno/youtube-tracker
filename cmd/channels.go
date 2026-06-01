package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	"github.com/alexperezortuno/youtube-tracker/internal/config"
	"github.com/alexperezortuno/youtube-tracker/internal/logger"
	"github.com/alexperezortuno/youtube-tracker/internal/models"
	"github.com/alexperezortuno/youtube-tracker/internal/storage"
)

var (
	channelID       string
	channelName     string
	channelActive   bool
	channelCategory string
	channelLanguage string
	channelCountry  string
	showAll         bool
)

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage channels",
}

var channelsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a channel",
	Run: func(cmd *cobra.Command, args []string) {
		if channelID == "" {
			logger.Error("channel ID is required")
			os.Exit(1)
		}

		if channelName == "" {
			logger.Error("channel name is required")
			os.Exit(1)
		}

		cfg := config.Load()
		pool, err := pgxpool.New(cmd.Context(), cfg.PostgresURL)
		if err != nil {
			logger.Error("failed to connect to database: %v", err)
			os.Exit(1)
		}
		defer pool.Close()

		dbSource := storage.NewDBSource(pool)

		var category, language, country *string
		if channelCategory != "" {
			category = &channelCategory
		}
		if channelLanguage != "" {
			language = &channelLanguage
		}
		if channelCountry != "" {
			country = &channelCountry
		}

		ch := models.Channel{
			ID:         channelID,
			Name:       channelName,
			Active:     true,
			Category:   category,
			Language:   language,
			Country:    country,
			FollowedAt: time.Now(),
		}

		if err := dbSource.AddChannel(cmd.Context(), ch); err != nil {
			logger.Error("failed to add channel: %v", err)
			os.Exit(1)
		}

		logger.Info("channel added: %s (%s)", channelName, channelID)
	},
}

var channelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List channels",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		pool, err := pgxpool.New(cmd.Context(), cfg.PostgresURL)
		if err != nil {
			logger.Error("failed to connect to database: %v", err)
			os.Exit(1)
		}
		defer pool.Close()

		dbSource := storage.NewDBSource(pool)

		channels, err := dbSource.GetChannels(!showAll)
		if err != nil {
			logger.Error("failed to list channels: %v", err)
			os.Exit(1)
		}

		if len(channels) == 0 {
			fmt.Println("No channels found")
			return
		}

		fmt.Printf("%-15s %-30s %-8s %-10s %-8s %-8s %-20s\n",
			"ID", "Name", "Active", "Category", "Language", "Country", "Followed At")
		fmt.Println("-----------------------------------------------------------------------------------------------")

		for _, ch := range channels {
			cat := ""
			if ch.Category != nil {
				cat = *ch.Category
			}
			lang := ""
			if ch.Language != nil {
				lang = *ch.Language
			}
			country := ""
			if ch.Country != nil {
				country = *ch.Country
			}
			active := "false"
			if ch.Active {
				active = "true"
			}

			fmt.Printf("%-15s %-30s %-8s %-10s %-8s %-8s %-20s\n",
				ch.ID,
				truncate(ch.Name, 30),
				active,
				cat,
				lang,
				country,
				ch.FollowedAt.Format("2006-01-02 15:04"),
			)
		}

		fmt.Printf("\nTotal: %d channels\n", len(channels))
	},
}

var channelsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a channel",
	Run: func(cmd *cobra.Command, args []string) {
		if channelID == "" {
			logger.Error("channel ID is required")
			os.Exit(1)
		}

		cfg := config.Load()
		pool, err := pgxpool.New(cmd.Context(), cfg.PostgresURL)
		if err != nil {
			logger.Error("failed to connect to database: %v", err)
			os.Exit(1)
		}
		defer pool.Close()

		dbSource := storage.NewDBSource(pool)

		if err := dbSource.DeleteChannel(cmd.Context(), channelID); err != nil {
			logger.Error("failed to remove channel: %v", err)
			os.Exit(1)
		}

		logger.Info("channel removed: %s", channelID)
	},
}

var channelsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a channel",
	Run: func(cmd *cobra.Command, args []string) {
		if channelID == "" {
			logger.Error("channel ID is required")
			os.Exit(1)
		}

		cfg := config.Load()
		pool, err := pgxpool.New(cmd.Context(), cfg.PostgresURL)
		if err != nil {
			logger.Error("failed to connect to database: %v", err)
			os.Exit(1)
		}
		defer pool.Close()

		dbSource := storage.NewDBSource(pool)

		ch, err := dbSource.GetChannelByID(cmd.Context(), channelID)
		if err != nil {
			logger.Error("channel not found: %s", channelID)
			os.Exit(1)
		}

		if channelName != "" {
			ch.Name = channelName
		}

		if cmd.Flags().Changed("active") {
			ch.Active = channelActive
		}

		if channelCategory != "" {
			ch.Category = &channelCategory
		}

		if channelLanguage != "" {
			ch.Language = &channelLanguage
		}

		if channelCountry != "" {
			ch.Country = &channelCountry
		}

		if err := dbSource.UpdateChannel(cmd.Context(), *ch); err != nil {
			logger.Error("failed to update channel: %v", err)
			os.Exit(1)
		}

		logger.Info("channel updated: %s (%s)", ch.Name, ch.ID)
	},
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	channelsCmd.AddCommand(channelsAddCmd)
	channelsCmd.AddCommand(channelsListCmd)
	channelsCmd.AddCommand(channelsRemoveCmd)
	channelsCmd.AddCommand(channelsUpdateCmd)

	channelsAddCmd.Flags().StringVarP(&channelID, "id", "i", "", "channel ID")
	channelsAddCmd.Flags().StringVarP(&channelName, "name", "n", "", "channel name")
	channelsAddCmd.Flags().StringVarP(&channelCategory, "category", "c", "", "category")
	channelsAddCmd.Flags().StringVarP(&channelLanguage, "language", "l", "", "language code (e.g., es, en)")
	channelsAddCmd.Flags().StringVarP(&channelCountry, "country", "o", "", "country code (e.g., MX, US)")

	channelsListCmd.Flags().BoolVarP(&showAll, "all", "a", false, "show all channels including inactive")

	channelsRemoveCmd.Flags().StringVarP(&channelID, "id", "i", "", "channel ID to remove")

	channelsUpdateCmd.Flags().StringVarP(&channelID, "id", "i", "", "channel ID to update")
	channelsUpdateCmd.Flags().StringVarP(&channelName, "name", "n", "", "new channel name")
	channelsUpdateCmd.Flags().BoolVarP(&channelActive, "active", "e", false, "set channel active status")
	channelsUpdateCmd.Flags().StringVarP(&channelCategory, "category", "c", "", "category")
	channelsUpdateCmd.Flags().StringVarP(&channelLanguage, "language", "l", "", "language code")
	channelsUpdateCmd.Flags().StringVarP(&channelCountry, "country", "o", "", "country code")

	rootCmd.AddCommand(channelsCmd)
}
