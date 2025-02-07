package aws

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
)

// DefaultChainCredentials is a Credentials which will find the first available
// credentials Value from the list of Providers.
//
// This should be used in the default case. Once the type of credentials are
// known switching to the specific Credentials will be more efficient.
var DefaultChainCredentials = credentials.NewChainCredentials(
	[]credentials.Provider{
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
		&credentials.EC2RoleProvider{ExpiryWindow: 5 * time.Minute},
	})

// DefaultMaxRetries is the default number of maximum retries the service client
const DefaultMaxRetries = 3

// DefaultConfig is the default all service configuration will be based off of.
var DefaultConfig = &Config{
	Credentials:                    DefaultChainCredentials,
	Endpoint:                       "",
	Region:                         "",
	DisableSSL:                     false,
	ManualSend:                     false,
	HTTPClient:                     http.DefaultClient,
	LogHTTPBody:                    false,
	LogLevel:                       Off,
	Logger:                         os.Stdout,
	MaxRetries:                     DefaultMaxRetries,
	RetryRules:                     ExponentialRetryRules,
	ShouldRetry:                    ShouldRetry,
	DisableParamValidation:         false,
	DisableComputeChecksums:        false,
	S3ForcePathStyle:               false,
	DomainMode:                     false,
	SignerVersion:                  "V2",
	CrcCheckEnabled:                false,
	DisableRestProtocolURICleaning: true,
}

// A Config provides service configuration
type Config struct {
	Credentials                    *credentials.Credentials
	Endpoint                       string
	Region                         string
	DisableSSL                     bool
	ManualSend                     bool
	HTTPClient                     *http.Client
	LogHTTPBody                    bool
	LogLevel                       uint
	Logger                         io.Writer
	MaxRetries                     int                     // 重试次数
	RetryRules                     func(int) time.Duration // 重试等待时间
	ShouldRetry                    func(error) bool        // 是否需要重试
	DisableParamValidation         bool
	DisableComputeChecksums        bool
	S3ForcePathStyle               bool
	DomainMode                     bool
	SignerVersion                  string
	CrcCheckEnabled                bool // 允许crc64校验，默认为false
	DisableRestProtocolURICleaning bool // 禁用path clean，默认为true
}

// Copy will return a shallow copy of the Config object.
func (c Config) Copy() Config {
	dst := Config{}
	dst.Credentials = c.Credentials
	dst.Endpoint = c.Endpoint
	dst.Region = c.Region
	dst.DisableSSL = c.DisableSSL
	dst.ManualSend = c.ManualSend
	dst.HTTPClient = c.HTTPClient
	dst.LogHTTPBody = c.LogHTTPBody
	dst.LogLevel = c.LogLevel
	dst.Logger = c.Logger
	dst.MaxRetries = c.MaxRetries
	dst.RetryRules = c.RetryRules
	dst.ShouldRetry = c.ShouldRetry
	dst.DisableParamValidation = c.DisableParamValidation
	dst.DisableComputeChecksums = c.DisableComputeChecksums
	dst.S3ForcePathStyle = c.S3ForcePathStyle
	dst.DomainMode = c.DomainMode
	dst.SignerVersion = c.SignerVersion
	dst.CrcCheckEnabled = c.CrcCheckEnabled
	dst.DisableRestProtocolURICleaning = c.DisableRestProtocolURICleaning
	return dst
}

// Merge merges the newcfg attribute values into this Config. Each attribute
// will be merged into this config if the newcfg attribute's value is non-zero.
// Due to this, newcfg attributes with zero values cannot be merged in. For
// example bool attributes cannot be cleared using Merge, and must be explicitly
// set on the Config structure.
func (c Config) Merge(newcfg *Config) *Config {
	if newcfg == nil {
		return &c
	}

	cfg := Config{}

	if newcfg.Credentials != nil {
		cfg.Credentials = newcfg.Credentials
	} else {
		cfg.Credentials = c.Credentials
	}

	if newcfg.Endpoint != "" {
		cfg.Endpoint = newcfg.Endpoint
	} else {
		cfg.Endpoint = c.Endpoint
	}

	if newcfg.Region != "" {
		cfg.Region = newcfg.Region
	} else {
		cfg.Region = c.Region
	}

	if newcfg.DisableSSL {
		cfg.DisableSSL = newcfg.DisableSSL
	} else {
		cfg.DisableSSL = c.DisableSSL
	}

	if newcfg.ManualSend {
		cfg.ManualSend = newcfg.ManualSend
	} else {
		cfg.ManualSend = c.ManualSend
	}

	if newcfg.HTTPClient != nil {
		cfg.HTTPClient = newcfg.HTTPClient
	} else {
		cfg.HTTPClient = c.HTTPClient
	}
	defaultHTTPRedirect(cfg.HTTPClient)

	if newcfg.LogHTTPBody {
		cfg.LogHTTPBody = newcfg.LogHTTPBody
	} else {
		cfg.LogHTTPBody = c.LogHTTPBody
	}

	if newcfg.LogLevel != 0 {
		cfg.LogLevel = newcfg.LogLevel
	} else {
		cfg.LogLevel = c.LogLevel
	}

	if newcfg.Logger != nil {
		cfg.Logger = newcfg.Logger
	} else {
		cfg.Logger = c.Logger
	}

	if newcfg.MaxRetries != 0 {
		cfg.MaxRetries = newcfg.MaxRetries
	} else {
		cfg.MaxRetries = c.MaxRetries
	}

	if newcfg.RetryRules != nil {
		cfg.RetryRules = newcfg.RetryRules
	} else {
		cfg.RetryRules = c.RetryRules
	}

	if newcfg.ShouldRetry != nil {
		cfg.ShouldRetry = newcfg.ShouldRetry
	} else {
		cfg.ShouldRetry = c.ShouldRetry
	}

	if newcfg.DisableParamValidation {
		cfg.DisableParamValidation = newcfg.DisableParamValidation
	} else {
		cfg.DisableParamValidation = c.DisableParamValidation
	}

	if newcfg.DisableComputeChecksums {
		cfg.DisableComputeChecksums = newcfg.DisableComputeChecksums
	} else {
		cfg.DisableComputeChecksums = c.DisableComputeChecksums
	}

	if newcfg.S3ForcePathStyle {
		cfg.S3ForcePathStyle = newcfg.S3ForcePathStyle
	} else {
		cfg.S3ForcePathStyle = c.S3ForcePathStyle
	}

	if newcfg.DomainMode {
		cfg.DomainMode = newcfg.DomainMode
	} else {
		cfg.DomainMode = c.DomainMode
	}
	if newcfg.SignerVersion != "" {
		cfg.SignerVersion = newcfg.SignerVersion
	} else {
		cfg.SignerVersion = c.SignerVersion
	}
	if newcfg.CrcCheckEnabled {
		cfg.CrcCheckEnabled = newcfg.CrcCheckEnabled
	} else {
		cfg.CrcCheckEnabled = c.CrcCheckEnabled
	}
	if newcfg.DisableRestProtocolURICleaning {
		cfg.DisableRestProtocolURICleaning = newcfg.DisableRestProtocolURICleaning
	} else {
		cfg.DisableRestProtocolURICleaning = c.DisableRestProtocolURICleaning
	}

	return &cfg
}

// Define the level of the output log
const (
	Off = iota
	Error
	Warn
	Info
	Debug
)

// LogTag Tag for each level of log
var LogTag = []string{"[error]", "[warn]", "[info]", "[debug]"}

func (c Config) writeLog(level uint, format string, a ...interface{}) {
	if c.LogLevel < level {
		return
	}

	var logBuffer bytes.Buffer
	logBuffer.WriteString(time.Now().Format("2006/01/02 15:04:05"))
	logBuffer.WriteString(" ")
	logBuffer.WriteString(LogTag[level-1])
	logBuffer.WriteString(fmt.Sprintf(format, a...))
	if logBuffer.Bytes()[logBuffer.Len()-1] != '\n' {
		logBuffer.WriteString("\n")
	}
	fmt.Fprintf(c.Logger, "%s", logBuffer.String())
}

func (c Config) LogError(format string, a ...interface{}) {
	if c.LogLevel < Error {
		return
	}
	c.writeLog(Error, format, a...)
}

func (c Config) LogWarn(format string, a ...interface{}) {
	if c.LogLevel < Warn {
		return
	}
	c.writeLog(Warn, format, a...)
}

func (c Config) LogInfo(format string, a ...interface{}) {
	if c.LogLevel < Info {
		return
	}
	c.writeLog(Info, format, a...)
}

func (c Config) LogDebug(format string, a ...interface{}) {
	if c.LogLevel < Debug {
		return
	}
	c.writeLog(Debug, format, a...)
}
