package policy

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// WatchedPolicyEngine wraps a PolicyEngine with fsnotify-based hot-reload.
// On file change: reload, validate, atomic swap via sync.RWMutex.
// Zero downtime — reads continue under old policy during reload.
type WatchedPolicyEngine struct {
	*PolicyEngine
	watcher    *fsnotify.Watcher
	policyPath string
	logger     *slog.Logger
	done       chan struct{}
	mu         sync.Mutex // protects start/stop
}

// NewWatchedPolicyEngine loads a policy from a file and starts watching for changes.
func NewWatchedPolicyEngine(policyPath string, logger *slog.Logger) (*WatchedPolicyEngine, error) {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	p, err := LoadPolicyFromFile(policyPath)
	if err != nil {
		return nil, fmt.Errorf("load initial policy: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create file watcher: %w", err)
	}

	wpe := &WatchedPolicyEngine{
		PolicyEngine: NewPolicyEngine(p),
		watcher:      watcher,
		policyPath:   policyPath,
		logger:       logger,
		done:         make(chan struct{}),
	}

	if err := watcher.Add(policyPath); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch policy file: %w", err)
	}

	go wpe.watchLoop()

	logger.Info("policy engine started",
		"path", policyPath,
		"rules", len(p.Rules),
		"issuers", len(p.Issuers),
	)

	return wpe, nil
}

func (wpe *WatchedPolicyEngine) watchLoop() {
	for {
		select {
		case <-wpe.done:
			return
		case event, ok := <-wpe.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				wpe.reload()
			}
		case err, ok := <-wpe.watcher.Errors:
			if !ok {
				return
			}
			wpe.logger.Error("policy watcher error", "error", err)
		}
	}
}

func (wpe *WatchedPolicyEngine) reload() {
	wpe.mu.Lock()
	defer wpe.mu.Unlock()

	newPolicy, err := LoadPolicyFromFile(wpe.policyPath)
	if err != nil {
		wpe.logger.Error("failed to reload policy — keeping current policy",
			"path", wpe.policyPath,
			"error", err,
		)
		return
	}

	wpe.SwapPolicy(newPolicy)

	wpe.logger.Info("policy reloaded successfully",
		"path", wpe.policyPath,
		"rules", len(newPolicy.Rules),
	)
}

// Close stops the file watcher and cleans up resources.
func (wpe *WatchedPolicyEngine) Close() error {
	close(wpe.done)
	return wpe.watcher.Close()
}
