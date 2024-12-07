package postgres

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	acstrings "github.com/zhughes3/go-accelerate/pkg/strings"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort = "5432"
	pgsn        = "postgres://%s:%s@%s:%s/%s?sslmode=%s&application_name=%s"
)

type Config struct {
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`

	// ApplicationName is used to identity the application using the database.
	ApplicationName string `env:"APPLICATION_NAME"`

	// Name is the name of the postgres database to target.
	Name    string `env:"NAME"`
	SSLMode string `env:"SSL_MODE"`

	// ConnectionPoolMax is maximum number of connections in the pool.
	ConnectionPoolMax int `env:"CONNECTION_POOL_MAX"`
	// ConnectionPoolMin is minimum number of connections in the pool.
	ConnectionPoolMin int `env:"CONNECTION_POOL_MIN"`
	// ConnectionRetries is number of maximum number of retries to use when acquiring a new connection.
	ConnectionRetries int `env:"CONNECTION_RETRIES"`
	// ConnectionRetryWaitTime is number of seconds to wait between connection retries.
	ConnectionRetryWaitTime int `env:"CONNECTION_RETRY_WAIT_TIME"`

	// EnableDBLogging enables verbose logging from the database.
	EnableDBLogging bool `env:"ENABLE_DB_LOGGING"`

	// DialTimeout is the number of seconds to wait before timing out the connection request.
	DialTimeout int `env:"DIAL_TIMEOUT"`

	// ConnectionKeepAlive is the number of seconds between sending out TCP "Keep Alives".
	ConnectionKeepAlive int `env:"CONNECTION_KEEP_ALIVE"`

	// ConnectionMaxIdleTime is the duration after which an idle connection will be closed automatically.
	ConnectionMaxIdleTime time.Duration `env:"CONNECTION_MAX_IDLE_TIME"`

	// ConnectionMaxLifetime is the duration since creation after which the connection will be closed automatically.
	ConnectionMaxLifetime time.Duration `env:"CONNECTION_MAX_LIFETIME"`

	// ConnectionMaxLifetimeJitter is the duration after [ConnectionMaxLifetime] to randomly decide to close a connection.
	ConnectionMaxLifetimeJitter time.Duration `env:"CONNECTION_MAX_LIFETIME_JITTER"`

	SecurityString [32]byte `env:"SECURITY_STRING"`

	// TLSCa is the path to the root CA file.
	TLSCa string `env:"TLS_CA"`

	// TLSCertificate is the path to the TLS Certificate.
	TLSCertificate string `env:"TLS_CERTIFICATE"`

	// TLSKey is the path to the TLS Certificate.
	TLSKey string `env:"TLS_KEY"`

	// Subsystem is used when creating metrics. Defaults to database when not set.
	Subsystem string `env:"SUBSYSTEM"`
}

func (c Config) URL() string {
	if acstrings.IsBlank(c.Port) {
		c.Port = defaultPort
	}

	return fmt.Sprintf(pgsn, c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode, c.ApplicationName)
}

func (c Config) ToPostgresConnConfig(logger slog.Logger) (*pgx.ConnConfig, error) {
	cc := newConnConfig()
	cc.Config.Host = c.Host
	cc.Config.Database = c.Name
	cc.Config.User = c.User
	cc.Config.Password = c.Password

	var err error

	cc.Config.Port, err = c.determinePort()
	if err != nil {
		return nil, err
	}

	if c.EnableDBLogging {
		// TODO
	}

	cc.DialFunc = c.determineDialer()

	cc.TLSConfig, err = c.determineTLS()
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func (c Config) NewPoolConfig(connConfig *pgx.ConnConfig) *pgxpool.Config {
	// We need to instantiate the PoolConfig using this function
	cpc, _ := pgxpool.ParseConfig("")
	cpc.ConnConfig = connConfig

	if c.ConnectionPoolMax > 0 {
		cpc.MaxConns = int32(c.ConnectionPoolMax)
	}

	if c.ConnectionPoolMin > 0 {
		cpc.MinConns = int32(c.ConnectionPoolMin)
	}

	if c.ConnectionMaxIdleTime > 0 {
		cpc.MaxConnIdleTime = c.ConnectionMaxIdleTime
	}

	if c.ConnectionMaxLifetime > 0 {
		cpc.MaxConnLifetime = c.ConnectionMaxLifetime
	}

	if c.ConnectionMaxLifetimeJitter > 0 {
		cpc.MaxConnLifetimeJitter = c.ConnectionMaxLifetimeJitter
	}
	return cpc
}

func newConnConfig() *pgx.ConnConfig {
	// It's recommended to use ParseConfig to instantiate a [pgx.ConnConfig]
	// https://github.com/jackc/pgx/issues/588
	cc, _ := pgx.ParseConfig("")
	return cc
}

func (c Config) determinePort() (uint16, error) {
	if acstrings.IsBlank(c.Port) {
		c.Port = defaultPort
	}
	p, err := strconv.ParseUint(c.Port, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(p), nil
}

func (c Config) determineDialer() pgconn.DialFunc {
	if c.DialTimeout == 0 && c.ConnectionKeepAlive == 0 {
		return nil
	}

	dialer := net.Dialer{}

	if c.DialTimeout != 0 {
		dialer.Timeout = time.Duration(c.DialTimeout) * time.Second
	}

	if c.ConnectionKeepAlive != 0 {
		dialer.KeepAlive = time.Duration(c.ConnectionKeepAlive) * time.Second
	}

	return dialer.DialContext
}

func (c Config) determineTLS() (*tls.Config, error) {
	if strings.ToUpper(c.SSLMode) == "DISABLE" {
		return nil, nil
	}

	clientKeyPair, err := tls.LoadX509KeyPair(c.TLSCertificate, c.TLSKey)
	if err != nil {
		return nil, err

	}

	certAuthorityPEM, err := os.ReadFile(c.TLSCa)
	if err != nil {
		return nil, err
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if ok := rootCAs.AppendCertsFromPEM(certAuthorityPEM); !ok {
		return nil, fmt.Errorf("server certificate authority: %q is not a valid PEM file", c.TLSCa)
	}

	tlsCfg := tls.Config{
		Certificates: []tls.Certificate{clientKeyPair},
		RootCAs:      rootCAs,
	}

	if acstrings.IsNotBlank(c.Host) {
		tlsCfg.ServerName = c.Host

	}

	return &tlsCfg, nil
}
