package config

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	RootAddress string `yaml:"root_address"`
	UserAgent   string `yaml:"user_agent"`

	MinRTT time.Duration `yaml:"min_rtt"`

	Nick string `yaml:"nick"`
	Pass string `yaml:"pass"`

	TelegramApiKey string `yaml:"telegram_api_key"`
	TelegramChatId int64  `yaml:"telegram_chat_id"`

	BecomeOffline      bool   `yaml:"become_offline"`
	BecomeOfflineEvery string `yaml:"become_offline_every"`
	BecomeOfflineFor   string `yaml:"become_offline_for"`

	RandomizeWait    bool   `yaml:"randomize_wait"`
	RandomizeWaitVal string `yaml:"randomize_wait_val"`
}

func NewConfig(path string) (*Config, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(contents, &c); err != nil {
		return nil, err
	}

	err = validateConfig(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func validateConfig(c *Config) error {

	// Emptiness checks //

	if c.RootAddress == "" {
		return errors.New("empty 'root_address' field value")
	}

	if c.UserAgent == "" {
		return errors.New("empty 'user_agent' field value")
	}

	if c.Nick == "" {
		return errors.New("empty 'nick' field value")
	}

	if c.Pass == "" {
		return errors.New("empty 'pass' field value")
	}

	if c.MinRTT == 0 {
		return errors.New("empty 'min_rtt_time' field value")
	}

	if c.TelegramApiKey == "" {
		return errors.New("empty 'telegram_api_key' field value")
	}

	if c.TelegramChatId == 0 {
		return errors.New("empty 'telegram_chat_id' field value")
	}

	if c.BecomeOffline {
		if c.BecomeOfflineEvery == "" {
			return errors.New("empty 'become_offline_every' field value")
		}
		if c.BecomeOfflineFor == "" {
			return errors.New("empty 'become_offline_for' field value")
		}
	}

	if c.RandomizeWait {
		if c.RandomizeWaitVal == "" {
			return errors.New("empty 'randomize_wait_val' field value")
		}
	}

	// Value checks //

	if !strings.Contains(c.RootAddress, "http") {
		return errors.New("invalid 'root_address' field value")
	}

	if c.MinRTT < 1*time.Millisecond {
		return errors.New("invalid 'min_rtt_time' field value")
	}

	if c.BecomeOffline {
		if err := checkIntervalInput(c.BecomeOfflineEvery, "become_offline_every"); err != nil {
			return err
		}
		if err := checkIntervalInput(c.BecomeOfflineFor, "become_offline_for"); err != nil {
			return err
		}
	}

	if c.RandomizeWait {
		if err := checkIntervalInput(c.RandomizeWaitVal, "randomize_wait_val"); err != nil {
			return err
		}
	}

	return nil
}

func checkIntervalInput(input, field string) error {
	pairs := strings.SplitN(input, ",", 2)
	if len(pairs) != 2 {
		return errors.New("invalid '" + field + "' field value")
	}
	int1, err1 := time.ParseDuration(pairs[0])
	int2, err2 := time.ParseDuration(pairs[1])
	if err1 != nil {
		return errors.New("invalid '" + field + "' field value")
	}
	if err2 != nil {
		return errors.New("invalid '" + field + "' field value")
	}
	if int1 > int2 {
		return errors.New("invalid '" + field + "' field value - first value is higher than the second")
	}
	return nil
}
