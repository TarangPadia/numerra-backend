package providers

import "fmt"

func GetDriver(provider string, shop string) (ProviderDriver, error) {
	switch provider {
	case "quickbooks":
		return QuickBooksDriver()
	case "shopify":
		return ShopifyDriver(shop)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
