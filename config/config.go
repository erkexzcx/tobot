package config

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	RootAddress string        `yaml:"root_address"`
	MinRTT      time.Duration `yaml:"min_rtt"`
	Telegram    struct {
		ApiKey string `yaml:"api_key"`
		ChatId int64  `yaml:"chat_id"`
	} `yaml:"telegram"`
	Settings Settings  `yaml:"settings"`
	Players  []*Player `yaml:"players"`
}

type Player struct {
	Nick          string    `yaml:"nick"`
	Pass          string    `yaml:"pass"`
	ActivitiesDir string    `yaml:"activities_dir"`
	Settings      *Settings `yaml:"settings"`
}

type Settings struct {
	UserAgent     string         `yaml:"user_agent"`
	BecomeOffline *BecomeOffline `yaml:"become_offline"`
	RandomizeWait *RandomizeWait `yaml:"randomize_wait"`
}

type BecomeOffline struct {
	Enabled string `yaml:"enabled"`
	Every   string `yaml:"every"`
	For     string `yaml:"for"`
}

type RandomizeWait struct {
	Enabled string `yaml:"enabled"`
	WaitVal string `yaml:"wait_val"`
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

	if c.MinRTT == 0 {
		return errors.New("empty 'min_rtt' field value")
	}

	if c.Telegram.ApiKey == "" {
		return errors.New("empty 'telegram->api_key' field value")
	}

	if c.Telegram.ChatId == 0 {
		return errors.New("empty 'telegram->chat_id' field value")
	}

	if c.Settings.UserAgent == "" {
		return errors.New("empty 'settings->user_agent' field value")
	}

	if c.Settings.BecomeOffline != nil {
		val, err := strconv.ParseBool(c.Settings.BecomeOffline.Enabled)
		if c.Settings.BecomeOffline.Enabled != "" && err == nil && val {
			if c.Settings.BecomeOffline.Every == "" {
				return errors.New("empty 'settings->become_offline->every' field value")
			}
			if c.Settings.BecomeOffline.For == "" {
				return errors.New("empty 'settings->become_offline->for' field value")
			}
		}
	}

	if c.Settings.RandomizeWait != nil {
		val, err := strconv.ParseBool(c.Settings.RandomizeWait.Enabled)
		if c.Settings.RandomizeWait.Enabled != "" && err == nil && val {
			if c.Settings.RandomizeWait.WaitVal == "" {
				return errors.New("empty 'settings->randomize_wait->wait_val' field value")
			}
		}
	}

	if len(c.Players) == 0 {
		return errors.New("no players specified")
	}

	for _, p := range c.Players {
		if p.Nick == "" {
			return errors.New("empty 'nick' field value")
		}
		if p.Pass == "" {
			return errors.New("empty 'pass' field value")
		}
		if p.ActivitiesDir == "" {
			return errors.New("empty 'activities_dir' field value")
		}
	}

	// Value checks //

	if !strings.Contains(c.RootAddress, "http") {
		return errors.New("invalid 'root_address' field value")
	}

	if c.MinRTT < 1*time.Millisecond {
		return errors.New("invalid 'min_rtt' field value")
	}

	if c.Settings.BecomeOffline != nil {
		if c.Settings.BecomeOffline.Every != "" {
			if err := checkIntervalInput(c.Settings.BecomeOffline.Every, "settings->become_offline->every"); err != nil {
				return err
			}
		}
		if c.Settings.BecomeOffline.For != "" {
			if err := checkIntervalInput(c.Settings.BecomeOffline.For, "settings->become_offline->for"); err != nil {
				return err
			}
		}
	}

	if c.Settings.RandomizeWait != nil {
		if c.Settings.RandomizeWait.WaitVal != "" {
			if err := checkIntervalInput(c.Settings.BecomeOffline.Every, "settings->randomize_wait->wait_val"); err != nil {
				return err
			}
		}
	}

	// players->activities_dir will be "checked" when we start using them...

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
