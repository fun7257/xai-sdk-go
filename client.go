package xai

import (
	"context"
	"fmt"

	"github.com/fun7257/xai-sdk-go/auth"
	"github.com/fun7257/xai-sdk-go/batch"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/image"
	"github.com/fun7257/xai-sdk-go/internal/conn"
	"github.com/fun7257/xai-sdk-go/models"
	"github.com/fun7257/xai-sdk-go/tokenize"
	"github.com/fun7257/xai-sdk-go/video"
	"google.golang.org/grpc"
)

func init() { conn.SDKVersion = Version }

// Client is the single entry point for the xAI API.
type Client struct {
	api        *grpc.ClientConn
	management *grpc.ClientConn
	ownsAPI    bool
	ownsMgmt   bool

	Auth        *auth.Client
	Batch       *batch.Client
	Chat        *chat.Client
	Collections *collections.Client
	Files       *files.Client
	Image       *image.Client
	Models      *models.Client
	Tokenize    *tokenize.Client
	Video       *video.Client
}

// NewClient constructs a Client from options and/or environment variables.
func NewClient(opts ...Option) (*Client, error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	apiKey, err := conn.ResolveAPIKey(cfg.apiKey, cfg.skipEnv)
	if err != nil {
		return nil, err
	}
	if apiKey == "" {
		return nil, fmt.Errorf("empty xAI API key provided")
	}

	c := &Client{}
	ctx := context.Background()

	var apiCC grpc.ClientConnInterface
	if cfg.apiConn != nil {
		apiCC = cfg.apiConn
	} else {
		api, err := conn.Dial(ctx, cfg.apiHost, apiKey, cfg.insecure, cfg.timeout, cfg.metadata, cfg.dialOpts)
		if err != nil {
			return nil, err
		}
		c.api = api
		c.ownsAPI = true
		apiCC = api
	}

	var mgmtCC grpc.ClientConnInterface
	if cfg.managementConn != nil {
		mgmtCC = cfg.managementConn
	} else {
		mk := conn.ResolveManagementKey(cfg.managementAPIKey, cfg.skipEnv)
		if mk != "" {
			mc, err := conn.Dial(ctx, cfg.managementAPIHost, mk, cfg.insecure, cfg.timeout, cfg.metadata, cfg.dialOpts)
			if err != nil {
				_ = c.Close()
				return nil, err
			}
			c.management = mc
			c.ownsMgmt = true
			mgmtCC = mc
		}
	}

	c.Auth = auth.New(apiCC)
	c.Batch = batch.New(apiCC)
	c.Chat = chat.New(apiCC)
	c.Collections = collections.New(apiCC, mgmtCC)
	c.Files = files.New(apiCC)
	c.Image = image.New(apiCC)
	c.Models = models.New(apiCC)
	c.Tokenize = tokenize.New(apiCC)
	c.Video = video.New(apiCC)
	return c, nil
}

// Close closes connections owned by the client.
func (c *Client) Close() error {
	var first error
	if c.ownsMgmt && c.management != nil {
		if err := c.management.Close(); err != nil && first == nil {
			first = err
		}
	}
	if c.ownsAPI && c.api != nil {
		if err := c.api.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
