package bisonbotkit

import (
	"context"
	"sync"

	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
)

// Bot represents a BisonRelay bot instance with configuration, RPC clients,
// and service interfaces for chat and payments.
type Bot struct {
	// Cfg holds the bot's configuration settings
	cfg *config.BotConfig

	LogBackend *logging.LogBackend

	wsc *jsonrpc.WSClient
	ctx context.Context

	wl     map[string]int64
	wlFile string
	wlMtx  sync.Mutex

	gcLog      slog.Logger
	gcChan     chan<- types.GCReceivedMsg
	inviteChan chan<- types.ReceivedGCInvite

	pmLog  slog.Logger
	pmChan chan<- types.ReceivedPM

	postLog  slog.Logger
	postChan chan<- types.ReceivedPost

	postStatusLog  slog.Logger
	postStatusChan chan<- types.ReceivedPostStatus

	tipProgressLog  slog.Logger
	tipProgressChan chan<- types.TipProgressEvent

	tipReceivedLog  slog.Logger
	tipReceivedChan chan<- types.ReceivedTip

	kxLog  slog.Logger
	kxChan chan<- types.KXCompleted

	chatService    types.ChatServiceClient
	gcService      types.GCServiceClient
	paymentService types.PaymentsServiceClient
	postService    types.PostsServiceClient
}

type GCs []*types.ListGCsResponse_GCInfo

func (g GCs) Len() int {
	return len(g)
}

func (g GCs) Less(a, b int) bool {
	// Most members first
	return g[a].NbMembers > g[b].NbMembers
}

func (g GCs) Swap(a, b int) {
	g[a], g[b] = g[b], g[a]
}

func (b *Bot) Close() error {
	return b.wsc.Close()
}
