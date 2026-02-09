package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type IPRule struct {
	CIDR   string `json:"cidr"`
	Action string `json:"action"` // ALLOW or BLOCK
}

type FirewallManager struct {
	cfg     *config.Config
	logger  *zap.Logger
	rules   []IPRule
	sync.RWMutex
}

func NewFirewallManager(cfg *config.Config, logger *zap.Logger) *FirewallManager {
	fm := &FirewallManager{
		cfg:    cfg,
		logger: logger,
	}
	
	// Initial fetch
	fm.refreshRules()
	
	// Start background refresh every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			fm.refreshRules()
		}
	}()
	
	return fm
}

func (fm *FirewallManager) refreshRules() {
	fm.logger.Debug("Refreshing IP firewall rules from auth-service")
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/api/v1/security/ip-rules/all", fm.cfg.AuthServiceURL))
	if err != nil {
		fm.logger.Error("Failed to fetch IP rules", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fm.logger.Error("IP rules fetch returned non-200 status", zap.Int("status", resp.StatusCode))
		return
	}

	var data struct {
		Rules []IPRule `json:"rules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fm.logger.Error("Failed to decode IP rules response", zap.Error(err))
		return
	}

	fm.Lock()
	fm.rules = data.Rules
	fm.Unlock()
	
	fm.logger.Info("Successfully refreshed IP firewall rules", zap.Int("count", len(data.Rules)))
}

func (fm *FirewallManager) IsBlocked(ipStr string) bool {
	clientIP := net.ParseIP(ipStr)
	if clientIP == nil {
		return false // Or true for strict security?
	}

	fm.RLock()
	defer fm.RUnlock()

	// Logic:
	// 1. If any BLOCK rule matches, block it immediately.
	// 2. If there are ALLOW rules, and NONE match, it might be blocked (depending on policy).
	// For this implementation: BLOCK rules take precedence.
	
	hasAllowRules := false
	matchAllow := false

	for _, rule := range fm.rules {
		_, ipNet, err := net.ParseCIDR(rule.CIDR)
		if err != nil {
			continue
		}

		if ipNet.Contains(clientIP) {
			if rule.Action == "BLOCK" {
				return true
			}
			if rule.Action == "ALLOW" {
				matchAllow = true
			}
		}
		
		if rule.Action == "ALLOW" {
			hasAllowRules = true
		}
	}

	// If we have ALLOW rules and none matched, block it.
	// (This enables a "Whitelisted IPs only" mode)
	if hasAllowRules && !matchAllow {
		return true
	}

	return false
}
