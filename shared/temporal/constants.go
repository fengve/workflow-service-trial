package temporal

//revive:disable

const (
	CronSchedule_Metering_AWS     = "1 */1 * * *"  // Run in the 1st minute of every one hour.
	CronSchedule_Metering_AZURE   = "10 */1 * * *" // Run in the 10th minute of every one hour.
	CronSchedule_Metering_GCP     = "15 */1 * * *" // Run in the 15th minute of every one hour.
	CronSchedule_Metering_ALIBABA = "20 */1 * * *" // Run in the 20th minute of every one hour.

	CronSchedule_SyncMarketplace_AWS                  = "10 */1 * * *" // Run in the 10th minute of every 1 hour.
	CronSchedule_SyncMarketplace_AZURE                = "30 */2 * * *" // Run in the 30th minute of every 2 hours.
	CronSchedule_SyncMarketplace_GCP                  = "40 */2 * * *" // Run in the 40th minute of every 2 hours.
	CronSchedule_SyncMarketplace_ALIBABA              = "50 */2 * * *" // Run in the 50th minute of every 2 hours.
	CronSchedule_Cosell_RunAutoShare                  = "@daily"       // Run once a day, midnight
	CronSchedule_Cosell_SyncSchemasFromSalesforceToS3 = "@daily"       // Run once a day, midnight
	CronSchedule_Cosell_SyncACEReferralsWithCRM       = "@every 3h"    // Run every 3 hours
	CronSchedule_Cosell_FetchLatestAWSReferrals       = "@every 10m"   // Run every 10 minutes.

	CronSchedule_Sync_LAGO      = "45 */1 * * *" // Run in the 45th minute of every 1 hour.
	CronSchedule_Sync_METRONOME = "35 */1 * * *" // Run in the 35th minute of every 1 hour.
	CronSchedule_Sync_ORB       = "0 11 * * *"   // Run in the 11 AM of every day.

	CronSchedule_BillingEntitlementEngine  = "10 */4 * * *" // Run in the 10th minute of every 4 hours.
	CronSchedule_BillingOfferEngine        = "12 */4 * * *" // Run in the 12th minute of every 4 hours.
	CronSchedule_BillingInvoiceEngine      = "14 */4 * * *" // Run in the 14th minute of every 4 hours.
	CronSchedule_BillingPaymentEngine      = "16 */4 * * *" // Run in the 16th minute of every 4 hours.
	CronSchedule_BillingMeterEngine_Hourly = "30 */1 * * *" // Run in the 30th minute of every one hour.
	CronSchedule_BillingMeterEngine_Daily  = "50 3 * * *"   // Run in the 03:50 AM of every day.

	CronSchedule_DailyUpdateAllEntitlementGrossRevenue = "30 6 * * *"  // Run in the 06:30 AM of every day. 06:30 AM UTC equivalent to 01:30 AM EST, 10:30 PM PST)
	CronSchedule_DailyHeadlessEntitlementsReport       = "0 7 * * *"   // Run in the 07:00 AM of every day. equivalent to 02:00 AM EST, 11:00 PM PST)
	CronSchedule_DailyUnpurchasedAzureOffersReport     = "0 7 * * *"   // Run in the 07:00 AM of every day. equivalent to 02:00 AM EST, 11:00 PM PST)
	CronSchedule_MonthlyUsageMeteringReport            = "0 4 */5 * *" // Run in the 04:00 AM of every 5 days.

	// Billing related workflow IDs.
	WorkflowId_Sync_LAGO      = "Sync_LAGO"
	WorkflowId_Sync_METRONOME = "Sync_METRONOME"
	WorkflowId_Sync_ORB       = "Sync_ORB"

	// Metering related workflow IDs.
	WorkflowId_Metering_AWS                            = "Metering_AWS"
	WorkflowId_Metering_AWS_CHINA                      = "Metering_AWS_CHINA"
	WorkflowId_Metering_AZURE                          = "Metering_AZURE"
	WorkflowId_Metering_GCP                            = "Metering_GCP"
	WorkflowIdTemplate_Metering_AWS_Organization       = "Metering_AWS_orgId/%s"
	WorkflowIdTemplate_Metering_AWS_CHINA_Organization = "Metering_AWS_CHINA_orgId/%s"
	WorkflowIdTemplate_Metering_AZURE_Organization     = "Metering_AZURE_orgId/%s"
	WorkflowIdTemplate_Metering_GCP_Organization       = "Metering_GCP_orgId/%s"
	// Metering Entitlement workflowId template shared by all partners, including AWS, Azure, GCP, Alibaba or Suger (Stripe or Adyen).
	WorkflowIdTemplate_Metering_Entitlement            = "Metering_orgId/%s/entitlementId/%s"
	WorkflowIdTemplate_Batch_Metering_Entitlements     = "Metering_orgId/%s/batch/entitlements/%s"
	WorkflowIdTemplate_TimescaledbMigrate              = "Metering_TimescaledbMigrate"
	WorkflowIdTemplate_TimescaledbMigrate_Organization = "Metering_TimescaledbMigrate_orgId/%s"

	WorkflowIdTemplate_Notification = "Notification_orgID/%s/%s/%s/%s" // orgID/entityType/entityID/action.

	WorkflowIdTemplate_CreatePrivateOffer                         = "CreateOffer_orgId/%s/offerId/%s"
	WorkflowIdTemplate_CreatePrivateOffer_V2                      = "CreateOffer_V2_orgId/%s/offerId/%s"
	WorkflowIdTemplate_CancelPrivateOffer                         = "CancelOffer_orgId/%s/offerId/%s"
	WorkflowIdTemplate_ExtendPrivateOfferExpiryDate               = "ExtendPrivateOfferExpiryDate_orgId/%s/offerId/%s"
	WorkflowIdTemplate_UpdateEntitlementGrossRevenue_Organization = "UpdateEntitlementGrossRevenue_orgId/%s"
	WorkflowIdTemplate_UpdateEntitlementGrossRevenue              = "UpdateEntitlementGrossRevenue_orgId/%s/entitlementId/%s"
	WorkflowIdTemplate_UpdateEntitlementGrossRevenue_Batch        = "UpdateEntitlementGrossRevenue_orgId/%s/batch/entitlements/%s"

	WorkflowId_UpdateAllEntitlementGrossRevenue                  = "UpdateAllEntitlementGrossRevenue"
	WorkflowIdTemplate_HeadlessEntitlementsReport_Organization   = "HeadlessEntitlementsReport_orgId/%s"
	WorkflowId_HeadlessEntitlementsReport                        = "HeadlessEntitlementsReport"
	WorkflowIdTemplate_UnpurchasedAzureOffersReport_Organization = "UnpurchasedAzureOffersReport_orgId/%s"
	WorkflowId_UnpurchasedAzureOffersReport                      = "UnpurchasedAzureOffersReport"
	WorkflowId_UsageMeteringReport                               = "UsageMeteringReport"
	WorkflowIdTemplate_UsageMeteringReport_Organization          = "UsageMeteringReport_orgId/%s_%s"

	WorkflowIdTemplate_SyncMarketplace_AWS_Organization    = "SyncMarketplace_AWS_orgId/%s"
	WorkflowIdTemplate_SyncBasicInfo_AWS_Organization      = "SyncBasicInfo_AWS_orgId/%s"
	WorkflowIdTemplate_SyncMcas_AWS_Organization           = "SyncMcas_AWS_orgId/%s"
	WorkflowIdTemplate_SyncMdfs_AWS_Organization           = "SyncMdfs_AWS_orgId/%s"
	WorkflowIdTemplate_SyncRevenueRecords_AWS_Organization = "SyncRevenueRecords_AWS_orgId/%s"
	WorkflowIdTemplate_PendingCancelEntitlement_AWS        = "PendingCancelEntitlement_AWS_orgId/%s/entitlementId/%s"

	WorkflowIdTemplate_SyncMarketplace_ALIBABA_Organization = "SyncMarketplace_ALIBABA_orgId/%s"

	// For Co-Sell
	WorkflowIdTemplate_AutoShare_Organization = "AutoShare_orgId/%s"

	// For sync schemas from salesforce to s3
	WorkflowIdTemplate_SyncSalesforceSchemas_AWS_Organization = "SyncSalesforceSchema_AWS_orgId/%s"

	WorkflowIdTemplate_FetchLatestAWSReferrals_Organization          = "FetchLatestAWSReferrals_orgId/%s"
	WorkflowIdTemplate_SyncACEReferralsWithCRM_Organization          = "SyncACEReferralsWithCRM_orgId/%s"
	WorkflowIdTemplate_SyncHubspotReferralStateProperty_Organization = "SyncHubspotReferralStateProperty_orgId/%s"

	// For sync Salesforce Suger Connector. E.g.: "SyncSalesforceSugerConnector_orgId/123"
	WorkflowIdTemplate_SyncSalesforceSugerConnector_Organization = "SyncSalesforceSugerConnector_orgId/%s"

	// For Azure Marketplace.
	WorkflowIdTemplate_SyncMarketplace_AZURE_Organization    = "SyncMarketplace_AZURE_orgId/%s"
	WorkflowIdTemplate_SyncCosell_AZURE_Organization         = "SyncCosell_AZURE_orgId/%s"
	WorkflowIdTemplate_SyncCma_AZURE_Organization            = "SyncCma_AZURE_orgId/%s"
	WorkflowIdTemplate_SyncRevenueRecords_AZURE_Organization = "SyncRevenueRecords_AZURE_orgId/%s"
	WorkflowIdTemplate_CancelEntitlement                     = "CancelEntitlement_orgId/%s/entitlementId/%s"
	WorkflowIdTemplate_UpdateEntitlementSeat                 = "UpdateEntitlementSeat_orgId/%s/entitlementId/%s"

	// For GCP Marketplace.
	WorkflowIdTemplate_SyncGCP_Organization                = "SyncGCP_orgId/%s"
	WorkflowIdTemplate_SyncMarketplace_GCP_Organization    = "SyncMarketplace_GCP_orgId/%s"
	WorkflowIdTemplate_SyncReport_GCP_Organization         = "SyncReport_GCP_orgId/%s"
	WorkflowIdTemplate_SyncRevenueRecords_GCP_Organization = "SyncRevenueRecords_GCP_orgId/%s"
	WorkflowIdTemplate_PendingCancelEntitlement_GCP        = "PendingCancelEntitlement_GCP_orgId/%s/entitlementId/%s"

	WorkflowIdTemplate_Sync_METRONOME_Organization = "Sync_METRONOME_orgId/%s/"
	WorkflowIdTemplate_Sync_ORB_Organization       = "Sync_ORB_orgId/%s/"
	WorkflowIdTemplate_Sync_LAGO_Organization      = "Sync_LAGO_orgId/%s/"

	WorkflowIdTemplate_UpdateEntitlement = "UpdateEntitlement_orgId/%s/entitlementId/%s"

	WorkflowIdTemplate_CreateProduct        = "CreateProduct_orgId/%s/productId/%s"
	WorkflowIdTemplate_UpdateProduct        = "UpdateProduct_orgId/%s/productId/%s"
	WorkflowIdTemplate_UpdateProductPricing = "UpdateProductPricing_orgId/%s/productId/%s"
	WorkflowIdTemplate_CreateCppoOutOffer   = "CreateCppoOutOffer_orgId/%s/offerId/%s"
	WorkflowIdTemplate_RestrictCppoOutOffer = "RestrictCppoOutOffer_orgId/%s/offerId/%s"

	// For Workflow Service.
	WorkflowIdTemplate_ScheduleTrigger        = "ScheduleTrigger_orgId/%s/workflowId/%s"
	WorkflowIdTemplate_UnregisterTestWebhooks = "UnregisterTestWebhooks_orgId/%s/workflowId/%s"

	// For Billing engine
	WorkflowId_BillingEntitlementEngine                       = "BillingEntitlementEngine"
	WorkflowIdTemplate_BillingEntitlementEngine_Organization  = "BillingEntitlementEngine_orgId/%s"
	WorkflowId_BillingOfferEngine                             = "BillingOfferEngine"
	WorkflowIdTemplate_BillingOfferEngine_Organization        = "BillingOfferEngine_orgId/%s"
	WorkflowId_BillingInvoiceEngine                           = "BillingInvoiceEngine"
	WorkflowIdTemplate_BillingInvoiceEngine_Organization      = "BillingInvoiceEngine_orgId/%s"
	WorkflowId_BillingMeterEngine_Hourly                      = "BillingMeterEngineHourly"
	WorkflowIdTemplate_BillingMeterEngine_Hourly_Organization = "BillingMeterEngineHourly_orgId/%s"
	WorkflowId_BillingMeterEngine_Daily                       = "BillingMeterEngineDaily"
	WorkflowIdTemplate_BillingMeterEngine_Daily_Organization  = "BillingMeterEngineDaily_orgId/%s"
	WorkflowId_BillingPaymentEngine                           = "BillingPaymentEngine"
	WorkflowId_BillingPaymentEngine_Organization              = "BillingPaymentEngine_orgId/%s"
)

var (
	WorkflowIdTemplates_Organization = []string{
		WorkflowIdTemplate_Metering_AWS_Organization,
		WorkflowIdTemplate_Metering_AZURE_Organization,
		WorkflowIdTemplate_Metering_GCP_Organization,
		WorkflowIdTemplate_BillingMeterEngine_Hourly_Organization,
		WorkflowIdTemplate_SyncMarketplace_AWS_Organization,
		WorkflowIdTemplate_SyncBasicInfo_AWS_Organization,
		WorkflowIdTemplate_SyncMcas_AWS_Organization,
		WorkflowIdTemplate_SyncMdfs_AWS_Organization,
		WorkflowIdTemplate_SyncRevenueRecords_AWS_Organization,
		WorkflowIdTemplate_SyncMarketplace_AZURE_Organization,
		WorkflowIdTemplate_SyncCosell_AZURE_Organization,
		WorkflowIdTemplate_SyncCma_AZURE_Organization,
		WorkflowIdTemplate_SyncGCP_Organization,
		WorkflowIdTemplate_SyncMarketplace_ALIBABA_Organization,
		WorkflowIdTemplate_SyncMarketplace_GCP_Organization,
		WorkflowIdTemplate_SyncReport_GCP_Organization,
		WorkflowIdTemplate_SyncRevenueRecords_GCP_Organization,
		WorkflowIdTemplate_SyncSalesforceSugerConnector_Organization,
		WorkflowIdTemplate_Sync_METRONOME_Organization,
		WorkflowIdTemplate_Sync_ORB_Organization,
	}
	WorkflowIdTemplates_Product = []string{
		WorkflowIdTemplate_CreateProduct,
		WorkflowIdTemplate_UpdateProduct,
		WorkflowIdTemplate_UpdateProductPricing,
	}
	WorkflowIdTemplates_Offer = []string{
		WorkflowIdTemplate_CreatePrivateOffer,
		WorkflowIdTemplate_CreatePrivateOffer_V2,
		WorkflowIdTemplate_CancelPrivateOffer,
		WorkflowIdTemplate_ExtendPrivateOfferExpiryDate,
		WorkflowIdTemplate_CreateCppoOutOffer,
		WorkflowIdTemplate_RestrictCppoOutOffer,
	}
	WorkflowIdTemplates_Entitlement = []string{
		WorkflowIdTemplate_Metering_Entitlement,
		WorkflowIdTemplate_UpdateEntitlement,
		WorkflowIdTemplate_PendingCancelEntitlement_GCP,
		WorkflowIdTemplate_CancelEntitlement,
		WorkflowIdTemplate_UpdateEntitlementSeat,
		WorkflowIdTemplate_UpdateEntitlementGrossRevenue,
	}
)
