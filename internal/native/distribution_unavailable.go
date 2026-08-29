//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

// Product family and distribution values are kept available to the public
// package even when this link-free build cannot execute the native ABI.
type ProductFamily uint32

const (
	ProductFamilySP3             ProductFamily = 0
	ProductFamilyIONEX           ProductFamily = 1
	ProductFamilyRINEXClock      ProductFamily = 2
	ProductFamilyRINEXNavigation ProductFamily = 3
)

type DistributionSource uint32

const (
	DistributionSourceDirect    DistributionSource = 0
	DistributionSourceNASACDDIS DistributionSource = 1
	DistributionSourceLocalFile DistributionSource = 2
	DistributionSourceInMemory  DistributionSource = 3
)

type ArchiveCompression uint32

const (
	ArchiveCompressionNone         ArchiveCompression = 0
	ArchiveCompressionGZIP         ArchiveCompression = 1
	ArchiveCompressionUnixCompress ArchiveCompression = 2
)

type SolutionClass uint32

const (
	SolutionClassFinal        SolutionClass = 0
	SolutionClassRapid        SolutionClass = 1
	SolutionClassUltraRapid   SolutionClass = 2
	SolutionClassPredicted    SolutionClass = 3
	SolutionClassBroadcast    SolutionClass = 4
	SolutionClassNearRealTime SolutionClass = 5
)

type ProductPublisher uint32

const (
	ProductPublisherIGS  ProductPublisher = 0
	ProductPublisherCODE ProductPublisher = 1
	ProductPublisherESA  ProductPublisher = 2
	ProductPublisherGFZ  ProductPublisher = 3
	ProductPublisherWHU  ProductPublisher = 4
)

type ProductCampaign uint32

const (
	ProductCampaignOperational         ProductCampaign = 0
	ProductCampaignMultiGNSS           ProductCampaign = 1
	ProductCampaignMultiGNSSExperiment ProductCampaign = 2
	ProductCampaignBroadcast           ProductCampaign = 3
)

type ProductFormat uint32

const (
	ProductFormatSP3             ProductFormat = 0
	ProductFormatIONEX           ProductFormat = 1
	ProductFormatRINEXClock      ProductFormat = 2
	ProductFormatRINEXNavigation ProductFormat = 3
)

type ProductIdentity struct {
	Family                   ProductFamily
	AnalysisCenter           string
	Publisher                ProductPublisher
	SolutionClass            SolutionClass
	Campaign                 ProductCampaign
	FilenameVersion          uint8
	Year                     int32
	Month                    uint8
	Day                      uint8
	HasIssue                 bool
	Issue                    string
	Span                     string
	Sample                   string
	OfficialFilename         string
	Format                   ProductFormat
	HasFormatVersion         bool
	FormatVersion            string
	HasPredictionHorizonDays bool
	PredictionHorizonDays    uint8
}

type DistributionLocation struct {
	Source          DistributionSource
	HasOriginalURL  bool
	OriginalURL     string
	ArchiveFilename string
	Compression     ArchiveCompression
}

func CatalogDefaultSample(string, ProductFamily, int, uint8, uint8) (string, error) {
	return "", unavailable()
}
func CatalogSupportedSamples(string, ProductFamily, int, uint8, uint8, string) ([]string, error) {
	return nil, unavailable()
}
func CatalogIdentity(string, ProductFamily, int, uint8, uint8, string, string) (ProductIdentity, error) {
	return ProductIdentity{}, unavailable()
}
func CatalogLocation(string, ProductFamily, int, uint8, uint8, string, string, DistributionSource) (DistributionLocation, error) {
	return DistributionLocation{}, unavailable()
}
func CatalogSolutionClass(string, ProductFamily) (SolutionClass, error) {
	return 0, unavailable()
}
func CatalogListingURLs(string, ProductFamily, int, uint8, uint8) ([]byte, error) {
	return nil, unavailable()
}
func CatalogNewestPublished(string, ProductFamily, []byte) ([]byte, error) {
	return nil, unavailable()
}
func CatalogNextIssueDue(string, ProductFamily, int, uint8, uint8, uint8, uint8, uint8) ([]byte, error) {
	return nil, unavailable()
}
func CatalogPredictedIONEXCandidates(int, uint8, uint8, string) ([]byte, error) {
	return nil, unavailable()
}
