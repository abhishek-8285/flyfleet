package main

import (
	"context"
	"fmt"

	"github.com/flyfleet/integrations-service/internal/connectors"
	"github.com/flyfleet/integrations-service/internal/sync"
)

func main() {
	fmt.Println("=== Starting ERP Sync Engine Verification ===")

	// 1. Initialize Engine
	engine := sync.NewSyncEngine()

	// 2. Register Connectors
	sap := connectors.NewSAPConnector("https://sap-gateway.example.com", "secret-key")
	sfdc := connectors.NewSalesforceConnector("https://na1.salesforce.com", "oauth-token")

	engine.RegisterConnector("SAP_PROD", sap)
	engine.RegisterConnector("SFDC_SALES", sfdc)

	ctx := context.Background()

	// 3. Trigger SAP Sync
	fmt.Println("\n--- Testing SAP Order Sync ---")
	if err := engine.TriggerOrderSync(ctx, "SAP_PROD", "ORD-12345"); err != nil {
		fmt.Printf("SAP Sync Failed: %v\n", err)
	}

	// 4. Trigger Salesforce Sync
	fmt.Println("\n--- Testing Salesforce Order Sync ---")
	if err := engine.TriggerOrderSync(ctx, "SFDC_SALES", "ORD-67890"); err != nil {
		fmt.Printf("Salesforce Sync Failed: %v\n", err)
	}

	fmt.Println("\n=== Verification Completed ===")
}
