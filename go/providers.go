package sdk

// Provider type constants - stable string identifiers accepted by the
// admin storage API for Storage.providerType / CreateStorageRequest.providerType.
const (
	ProviderTypeS3              = "s3"
	ProviderTypeBackblaze       = "backblaze"
	ProviderTypeCloudflare      = "cloudflare"
	ProviderTypeDigitalOcean    = "digitalocean"
	ProviderTypeIBMCloud        = "ibmcloud"
	ProviderTypeImpossibleCloud = "impossiblecloud"
	ProviderTypeLyve            = "lyve"
	ProviderTypeWasabi          = "wasabi"
	ProviderTypeS3Compatible    = "s3compatible"
	ProviderTypeAzure           = "azure"
	ProviderTypeMountOS         = "mountOS"
)

// AzureCredentialFields documents how generic credential fields map to Azure
// Blob Storage concepts when constructing a CreateStorageRequest with
// ProviderType=ProviderTypeAzure.
//
//	Endpoint  -> https://<account>.blob.core.windows.net  (or Azurite URL)
//	Bucket    -> container name
//	AccessKey -> storage account name
//	SecretKey -> base64-encoded account key
//	Region    -> informational only (Azure infers from endpoint)
type AzureCredentialFields struct{}
