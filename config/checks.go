package config

import (
	"errors"
	"fmt"
	"net/url"
)

// ### This does not validate settings under 'players' section ###
func validateConfig(c *Config) error {

	// Check log level
	if c.LogLevel != "CRITICAL" && c.LogLevel != "WARNING" && c.LogLevel != "INFO" && c.LogLevel != "DEBUG" {
		return errors.New("invalid 'log_level' field value")
	}

	// Check Telegram
	if c.Telegram.ApiKey == "" {
		return errors.New("empty 'telegram->api_key' field value")
	}
	if c.Telegram.ChatId == 0 {
		return errors.New("empty 'telegram->chat_id' field value")
	}

	// Check OpenAI
	if c.OpenAI.ApiKey == "" {
		return errors.New("empty 'openai->api_key' field value")
	}
	if c.OpenAI.Model == "" {
		return errors.New("empty 'openai->model' field value")
	}
	if c.OpenAI.Temperature < 0 || c.OpenAI.Temperature > 1 {
		return errors.New("empty 'openai->model' field value")
	}
	if c.OpenAI.Instructions == "" {
		return errors.New("empty 'openai->instructions' field value")
	}

	// Check global player settings
	if err := checkSettings(&c.Settings, "global settings"); err != nil {
		return err
	}

	// Check players section
	if len(c.Players) == 0 {
		return errors.New("no players specified")
	}

	for _, p := range c.Players {
		// Validate player non-settings fields
		if p.Nick == "" {
			return errors.New("empty 'nick' field value")
		}
		if p.Pass == "" {
			return errors.New("empty 'pass' field value")
		}
		if p.ActivitiesDir == "" {
			return errors.New("empty 'activities_dir' field value")
		}

		// Validate player settings
		if err := checkSettings(&c.Settings, fmt.Sprintf("player '%s'", p.Nick)); err != nil {
			return err
		}
	}

	return nil
}

func checkSettings(s *Settings, where string) error {
	if s.RootAddress == nil {
		return fmt.Errorf("missing '%s' field value (at %s)", "root_address", where)
	}
	if _, err := url.Parse(*s.RootAddress); err != nil {
		return fmt.Errorf("failed to parse '%s' field value (at %s)", "root_address", where)
	}
	if s.UserAgent == nil || *s.UserAgent == "" {
		return fmt.Errorf("failed to parse '%s' field value (at %s)", "user_agent", where)
	}
	if s.MinRTT == nil || *s.MinRTT == 0 {
		return fmt.Errorf("invalid, missing or equal to 0 '%s' field value (at %s)", "min_rtt", where)
	}
	if s.BecomeOffline.Enabled == nil {
		return fmt.Errorf("missing '%s' field value (at %s)", "become_offline->enabled", where)
	}
	if *s.BecomeOffline.Enabled {
		if s.BecomeOffline.Every == nil || len(*s.BecomeOffline.Every) != 2 || (*s.BecomeOffline.Every)[0] > (*s.BecomeOffline.Every)[1] {
			return fmt.Errorf("missing or invalid '%s' field value (at %s)", "become_offline->every", where)
		}
		if s.BecomeOffline.For == nil || len(*s.BecomeOffline.For) != 2 || (*s.BecomeOffline.For)[0] > (*s.BecomeOffline.For)[1] {
			return fmt.Errorf("missing or invalid '%s' field value (at %s)", "become_offline->for", where)
		}
	}
	if s.RandomizeWait.Enabled == nil {
		return fmt.Errorf("missing '%s' field value (at %s)", "randomize_wait->enabled", where)
	}
	if *s.RandomizeWait.Enabled {
		if s.RandomizeWait.WaitVal == nil || len(*s.RandomizeWait.WaitVal) != 2 || (*s.RandomizeWait.WaitVal)[0] > (*s.RandomizeWait.WaitVal)[1] {
			return fmt.Errorf("missing or invalid '%s' field value (at %s)", "randomize_wait->wait_val", where)
		}
		if s.RandomizeWait.Probability == nil || *s.RandomizeWait.Probability > 1 || *s.RandomizeWait.Probability < 0 {
			return fmt.Errorf("missing or invalid '%s' field value (at %s)", "randomize_wait->wait_val", where)
		}
	}
	return nil
}
