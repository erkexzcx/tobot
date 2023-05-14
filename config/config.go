package config

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel string `yaml:"log_level"`
	Telegram struct {
		ApiKey string `yaml:"api_key"`
		ChatId int64  `yaml:"chat_id"`
	} `yaml:"telegram"`
	OpenAI struct {
		ApiKey       string  `yaml:"api_key"`
		Model        string  `yaml:"model"`
		Temperature  float32 `yaml:"temperature"`
		Instructions string  `yaml:"instructions"`
	} `yaml:"openai"`
	CreatePlayers bool      `yaml:"create_players"`
	Settings      Settings  `yaml:"settings"`
	Players       []*Player `yaml:"players"`
}

type Player struct {
	Nick          string   `yaml:"nick"`
	Pass          string   `yaml:"pass"`
	PassPlain     string   `yaml:"pass_plain"`
	ActivitiesDir string   `yaml:"activities_dir"`
	Settings      Settings `yaml:"settings"`
}

type Settings struct {
	RootAddress   *string        `yaml:"root_address,omitempty"`
	UserAgent     *string        `yaml:"user_agent,omitempty"`
	MinRTT        *time.Duration `yaml:"min_rtt,omitempty"`
	BecomeOffline struct {
		Enabled *bool            `yaml:"enabled,omitempty"`
		Every   *[]time.Duration `yaml:"every,omitempty"`
		For     *[]time.Duration `yaml:"for,omitempty"`
	} `yaml:"become_offline"`
	RandomizeWait struct {
		Enabled     *bool            `yaml:"enabled,omitempty"`
		WaitVal     *[]time.Duration `yaml:"wait_val,omitempty"`
		Probability *float64         `yaml:"probability,omitempty"`
	} `yaml:"randomize_wait"`
}

func NewConfig(path string) (*Config, error) {
	// Read config from the file
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse config as YAML
	var c Config
	if err := yaml.Unmarshal(contents, &c); err != nil {
		return nil, err
	}

	// Fill (overriden) player settings with global settings
	for _, p := range c.Players {
		fillPlayerSettings(p, &c.Settings)
	}

	// Validate config
	err = validateConfig(&c)
	if err != nil {
		return nil, err
	}

	// Generate password hashes
	for _, p := range c.Players {
		if p.PassPlain != "" {
			p.Pass = generatePass(p.Nick, p.PassPlain)
		}
	}

	return &c, nil
}

func fillPlayerSettings(p *Player, c *Settings) {
	if p.Settings.RootAddress == nil {
		p.Settings.RootAddress = c.RootAddress
	}
	if p.Settings.UserAgent == nil {
		p.Settings.UserAgent = c.UserAgent
	}
	if p.Settings.MinRTT == nil {
		p.Settings.MinRTT = c.MinRTT
	}
	if p.Settings.BecomeOffline.Enabled == nil {
		p.Settings.BecomeOffline.Enabled = c.BecomeOffline.Enabled
	}
	if p.Settings.BecomeOffline.Every == nil {
		p.Settings.BecomeOffline.Every = c.BecomeOffline.Every
	}
	if p.Settings.BecomeOffline.For == nil {
		p.Settings.BecomeOffline.For = c.BecomeOffline.For
	}
	if p.Settings.RandomizeWait.Enabled == nil {
		p.Settings.RandomizeWait.Enabled = c.RandomizeWait.Enabled
	}
	if p.Settings.RandomizeWait.WaitVal == nil {
		p.Settings.RandomizeWait.WaitVal = c.RandomizeWait.WaitVal
	}
	if p.Settings.RandomizeWait.Probability == nil {
		p.Settings.RandomizeWait.Probability = c.RandomizeWait.Probability
	}
}

func generatePass(username, password string) string {
	passwordPart := md5Hash(md5Hash(password))                             // Double MD5 encode
	passwordPart = passwordPart[0:15]                                      // Cut first 15 characters
	passwordPart = base64.StdEncoding.EncodeToString([]byte(passwordPart)) // Base64 encode

	usernamePart := md5Hash(username) // MD5 encode

	return passwordPart + usernamePart
}

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
