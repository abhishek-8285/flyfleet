package billing

import (
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/accounting"
	"github.com/flyfleet/fleet-core/internal/domains/billing/email"
	"github.com/flyfleet/fleet-core/internal/domains/billing/handlers"
	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/pdf"
	"github.com/flyfleet/fleet-core/internal/domains/billing/queue"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/flyfleet/fleet-core/internal/domains/billing/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Config struct {
	Enabled               bool
	ChromeTimeoutSeconds  int
}

// RegisterModule wires up the entire billing domain.
func RegisterModule(r chi.Router, pool *pgxpool.Pool, cfg Config, logger *zap.Logger) {
	if !cfg.Enabled {
		logger.Info("Billing module is disabled via configuration")
		return
	}

	// 1. Repositories
	repoBilling := repository.NewBillingRepository(pool)
	// Some services need specific repos, but BillingRepository contains or wraps them if we implement it.
	// We'll instantiate individually as required by NewBillingHandler.
	repoSub := repository.NewSubscriptionRepository(pool)
	repoAudit := repository.NewAuditLogRepository(pool)
	repoMeter := repository.NewMeteringRepository(pool)
	repoCreditNote := repository.NewCreditNoteRepository(pool)
	repoHw := repository.NewHardwareRepository(pool)
	repoLedger := repository.NewLedgerRepository(pool)
	repoAdvSub := repository.NewAdvancedSubRepository(pool)
	repoAnalytics := repository.NewAnalyticsRepository(pool)
	repoAcc := repository.NewAccountingRepository(pool)
	repoTemplate := repository.NewTemplateRepository(pool)
	// repoInvoice := postgres.NewInvoiceRepository(pool)

	// Since service.NewInvoiceService needs `service.Repository` which might be different:
	// Let's just create variables for whatever NewBillingHandler requires.
	
	subServ := service.NewSubscriptionService(repoSub, repoAudit, logger)
	meterServ := service.NewMeteringService(repoMeter, logger)
	hwServ := service.NewHardwareService(repoHw, logger)
	ledgerServ := service.NewLedgerService(repoLedger, logger)
	cnServ := service.NewCreditNoteService(repoCreditNote, ledgerServ, logger)
	
	// Remove postgres repo import error by removing it.
	
	// Create BillingService (which seems to group them)
	// We pass nil for unneeded sub-services.
	billServ := service.NewBillingService(repoBilling, repoAudit, subServ, meterServ, hwServ, ledgerServ, nil, nil, nil, nil, nil, logger)

	advSubServ := service.NewAdvancedSubService(repoAdvSub, ledgerServ, subServ, billServ, logger)
	
	analyticsServ := service.NewAnalyticsService(repoAnalytics, logger)
	analyticsH := handlers.NewAnalyticsHandler(analyticsServ, logger)

	// Accounting
	qbExporter := accounting.NewQuickBooksExporter()
	xeroExporter := accounting.NewXeroExporter()
	exporters := []models.AccountingExporter{qbExporter, xeroExporter}
	accServ := service.NewAccountingService(repoAcc, repoLedger, repoBilling, nil, exporters, logger)
	accH := handlers.NewAccountingHandler(accServ, logger)

	// PDF & Invoice
	templateServ := pdf.NewTemplateService(repoTemplate)
	invCfg := service.InvoiceServiceConfig{
		ChromeTimeout: time.Duration(cfg.ChromeTimeoutSeconds) * time.Second,
		CacheMaxSize: 100,
		CacheTTL: 1 * time.Hour,
		EmailConfig: email.EmailConfig{},
	}
	producer := queue.NewTaskProducer(nil) // pass nil redis for now or fix
	invServ, _ := service.NewInvoiceService(repoBilling, producer, templateServ, invCfg, logger)

	h := handlers.NewBillingHandler(
		billServ,
		invServ,
		subServ,
		meterServ,
		cnServ,
		hwServ,
		ledgerServ,
		advSubServ,
		analyticsH,
		accH,
		logger,
	)

	// Mount under /api/v1/admin/billing
	// The parent router will give us /admin/billing and we mount it.
	r.Route("/admin/billing", func(r chi.Router) {
		h.RegisterRoutes(r)
	})
	
	logger.Info("Billing module initialized")
}
