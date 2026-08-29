package sidereon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// ProductFamily identifies a catalog product family.
type ProductFamily uint32

const (
	// ProductFamilySP3 identifies the product family sp3 case.
	ProductFamilySP3 ProductFamily = ProductFamily(native.ProductFamilySP3)
	// ProductFamilyIONEX identifies the product family ionex case.
	ProductFamilyIONEX ProductFamily = ProductFamily(native.ProductFamilyIONEX)
	// ProductFamilyRINEXClock identifies the product family rinex clock case.
	ProductFamilyRINEXClock ProductFamily = ProductFamily(native.ProductFamilyRINEXClock)
	// ProductFamilyRINEXNavigation identifies the product family rinex navigation case.
	ProductFamilyRINEXNavigation ProductFamily = ProductFamily(native.ProductFamilyRINEXNavigation)
)

// String formats the value for presentation.
func (f ProductFamily) String() string {
	switch f {
	case ProductFamilySP3:
		return "sp3"
	case ProductFamilyIONEX:
		return "ionex"
	case ProductFamilyRINEXClock:
		return "rinex_clock"
	case ProductFamilyRINEXNavigation:
		return "rinex_navigation"
	default:
		return fmt.Sprintf("product_family(%d)", f)
	}
}

// DistributionSource identifies a catalog distributor. HTTPAcquirer supports
// Direct and NASACDDIS; the other values are represented by the C ABI but are
// intentionally not network sources.
type DistributionSource uint32

const (
	// DistributionSourceDirect identifies the distribution source direct case.
	DistributionSourceDirect DistributionSource = DistributionSource(native.DistributionSourceDirect)
	// DistributionSourceNASACDDIS identifies the distribution source nasacddis case.
	DistributionSourceNASACDDIS DistributionSource = DistributionSource(native.DistributionSourceNASACDDIS)
	// DistributionSourceLocalFile identifies the distribution source local file case.
	DistributionSourceLocalFile DistributionSource = DistributionSource(native.DistributionSourceLocalFile)
	// DistributionSourceInMemory identifies the distribution source in memory case.
	DistributionSourceInMemory DistributionSource = DistributionSource(native.DistributionSourceInMemory)
)

// String formats the value for presentation.
func (s DistributionSource) String() string {
	switch s {
	case DistributionSourceDirect:
		return "direct"
	case DistributionSourceNASACDDIS:
		return "nasa_cddis"
	case DistributionSourceLocalFile:
		return "local_file"
	case DistributionSourceInMemory:
		return "in_memory"
	default:
		return fmt.Sprintf("distribution_source(%d)", s)
	}
}

// ArchiveCompression identifies the transport encoding selected by C.
type ArchiveCompression uint32

const (
	// ArchiveCompressionNone identifies the archive compression none case.
	ArchiveCompressionNone ArchiveCompression = ArchiveCompression(native.ArchiveCompressionNone)
	// ArchiveCompressionGZIP identifies the archive compression gzip case.
	ArchiveCompressionGZIP ArchiveCompression = ArchiveCompression(native.ArchiveCompressionGZIP)
	// ArchiveCompressionUnixCompress identifies the archive compression unix compress case.
	ArchiveCompressionUnixCompress ArchiveCompression = ArchiveCompression(native.ArchiveCompressionUnixCompress)
)

// String formats the value for presentation.
func (c ArchiveCompression) String() string {
	switch c {
	case ArchiveCompressionNone:
		return "none"
	case ArchiveCompressionGZIP:
		return "gzip"
	case ArchiveCompressionUnixCompress:
		return "unix_compress"
	default:
		return fmt.Sprintf("archive_compression(%d)", c)
	}
}

// SolutionClass describes the C-catalog solution class.
type SolutionClass uint32

const (
	// SolutionClassFinal identifies the solution class final case.
	SolutionClassFinal SolutionClass = SolutionClass(native.SolutionClassFinal)
	// SolutionClassRapid identifies the solution class rapid case.
	SolutionClassRapid SolutionClass = SolutionClass(native.SolutionClassRapid)
	// SolutionClassUltraRapid identifies the solution class ultra rapid case.
	SolutionClassUltraRapid SolutionClass = SolutionClass(native.SolutionClassUltraRapid)
	// SolutionClassPredicted identifies the solution class predicted case.
	SolutionClassPredicted SolutionClass = SolutionClass(native.SolutionClassPredicted)
	// SolutionClassBroadcast identifies the solution class broadcast case.
	SolutionClassBroadcast SolutionClass = SolutionClass(native.SolutionClassBroadcast)
	// SolutionClassNearRealTime identifies the solution class near real time case.
	SolutionClassNearRealTime SolutionClass = SolutionClass(native.SolutionClassNearRealTime)
)

// String formats the value for presentation.
func (s SolutionClass) String() string {
	switch s {
	case SolutionClassFinal:
		return "final"
	case SolutionClassRapid:
		return "rapid"
	case SolutionClassUltraRapid:
		return "ultra_rapid"
	case SolutionClassPredicted:
		return "predicted"
	case SolutionClassBroadcast:
		return "broadcast"
	case SolutionClassNearRealTime:
		return "near_real_time"
	default:
		return fmt.Sprintf("solution_class(%d)", s)
	}
}

// ProductPublisher identifies the organization that publishes or combines a
// product. Values are the C ABI discriminants.
type ProductPublisher uint32

const (
	// ProductPublisherIGS identifies the product publisher igs case.
	ProductPublisherIGS ProductPublisher = ProductPublisher(native.ProductPublisherIGS)
	// ProductPublisherCODE identifies the product publisher code case.
	ProductPublisherCODE ProductPublisher = ProductPublisher(native.ProductPublisherCODE)
	// ProductPublisherESA identifies the product publisher esa case.
	ProductPublisherESA ProductPublisher = ProductPublisher(native.ProductPublisherESA)
	// ProductPublisherGFZ identifies the product publisher gfz case.
	ProductPublisherGFZ ProductPublisher = ProductPublisher(native.ProductPublisherGFZ)
	// ProductPublisherWHU identifies the product publisher whu case.
	ProductPublisherWHU ProductPublisher = ProductPublisher(native.ProductPublisherWHU)
)

// ProductCampaign identifies the filename campaign.
type ProductCampaign uint32

const (
	// ProductCampaignOperational identifies the product campaign operational case.
	ProductCampaignOperational ProductCampaign = ProductCampaign(native.ProductCampaignOperational)
	// ProductCampaignMultiGNSS identifies the product campaign multi gnss case.
	ProductCampaignMultiGNSS ProductCampaign = ProductCampaign(native.ProductCampaignMultiGNSS)
	// ProductCampaignMultiGNSSExperiment identifies the product campaign multi gnss experiment case.
	ProductCampaignMultiGNSSExperiment ProductCampaign = ProductCampaign(native.ProductCampaignMultiGNSSExperiment)
	// ProductCampaignBroadcast identifies the product campaign broadcast case.
	ProductCampaignBroadcast ProductCampaign = ProductCampaign(native.ProductCampaignBroadcast)
)

// ProductFormat identifies the standard serialization format.
type ProductFormat uint32

const (
	// ProductFormatSP3 identifies the product format sp3 case.
	ProductFormatSP3 ProductFormat = ProductFormat(native.ProductFormatSP3)
	// ProductFormatIONEX identifies the product format ionex case.
	ProductFormatIONEX ProductFormat = ProductFormat(native.ProductFormatIONEX)
	// ProductFormatRINEXClock identifies the product format rinex clock case.
	ProductFormatRINEXClock ProductFormat = ProductFormat(native.ProductFormatRINEXClock)
	// ProductFormatRINEXNavigation identifies the product format rinex navigation case.
	ProductFormatRINEXNavigation ProductFormat = ProductFormat(native.ProductFormatRINEXNavigation)
)

// CatalogRequest is an exact C-catalog query. Date is interpreted as its UTC
// calendar date; Sample and Issue are empty when the catalog should select its
// default or the product line has no issue.
type CatalogRequest struct {
	// Center is the product centre.
	Center string
	// Family is the product family.
	Family ProductFamily
	// Date is the timestamp for this record.
	Date time.Time
	// Sample is the sample value or index.
	Sample string
	// Issue is the issue code.
	Issue string
}

func (r CatalogRequest) dateParts() (int, uint8, uint8) {
	year, month, day := r.Date.UTC().Date()
	return year, uint8(month), uint8(day)
}

func (r CatalogRequest) nativeParts() (string, native.ProductFamily, int, uint8, uint8, string, string) {
	year, month, day := r.dateParts()
	return r.Center, native.ProductFamily(r.Family), year, month, day, r.Sample, r.Issue
}

// ProductIdentity is the exact identity returned by the C catalog. Its Date
// is midnight UTC and OfficialFilename excludes transport compression.
type ProductIdentity struct {
	// Family is the product family.
	Family ProductFamily
	// AnalysisCenter identifies the analysis centre that produced the product.
	AnalysisCenter string
	// Publisher identifies the organization that publishes or combines the product.
	Publisher ProductPublisher
	// SolutionClass identifies whether the product is final, rapid, predicted, or another C-defined class.
	SolutionClass SolutionClass
	// Campaign identifies the filename campaign associated with the product.
	Campaign ProductCampaign
	// FilenameVersion is the numeric filename-version discriminator.
	FilenameVersion uint8
	// Date is the timestamp for this record.
	Date time.Time
	// HasIssue reports whether the has issue field is present.
	HasIssue bool
	// Issue is the issue code.
	Issue string
	// Span is the product's coverage span token from the catalog filename.
	Span string
	// Sample is the sample value or index.
	Sample string
	// OfficialFilename is the canonical product filename without transport compression.
	OfficialFilename string
	// Format is the product format.
	Format ProductFormat
	// HasFormatVersion reports whether the has format version field is present.
	HasFormatVersion bool
	// FormatVersion is the optional version token of the product serialization format.
	FormatVersion string
	// HasPredictionHorizonDays reports whether the has prediction horizon days field is present.
	HasPredictionHorizonDays bool
	// PredictionHorizonDays is the prediction horizon days in days.
	PredictionHorizonDays uint8
}

// DistributionLocation is the C-approved URL and archive metadata for an
// exact identity.
type DistributionLocation struct {
	// Source is the source classification.
	Source DistributionSource
	// OriginalURL is the source URL from which the product was obtained.
	OriginalURL string
	// ArchiveFilename is the archive member or downloaded filename when one exists.
	ArchiveFilename string
	// Compression is the archive compression.
	Compression ArchiveCompression
}

// PublishedProduct is the C-selected newest product in a parsed listing.
// ObservedAt preserves the archive's modification text; nil means the listing
// did not provide one.
type PublishedProduct struct {
	// Date is the timestamp for this record.
	Date time.Time
	// Issue is the issue code.
	Issue string
	// Filename is the published product filename.
	Filename string
	// ObservedAt contains a detached copy; nil means this field is absent.
	ObservedAt *string
}

// PredictedIONEXCandidate is an ordered C-catalog candidate for one map date.
type PredictedIONEXCandidate struct {
	// Center is the product centre.
	Center string
	// Date is the timestamp for this record.
	Date time.Time
	// Sample is the sample value or index.
	Sample string
	// Issue is the issue code.
	Issue string
	// Filename is the candidate product filename.
	Filename string
	// URL is the candidate product URL.
	URL string
}

// CoverageInterval is a half-open UTC interval returned by the C schedule.
type CoverageInterval struct {
	// From is the timestamp for this record.
	From time.Time
	// Until is the timestamp for this record.
	Until time.Time
}

// NominalIssue describes the next nominal catalog issue and its C-supplied
// observed/predicted coverage intervals.
type NominalIssue struct {
	// Identity identifies the product family, analysis center, publisher, solution class, campaign, date, issue, format, and prediction metadata.
	Identity ProductIdentity
	// DueAt is the timestamp for this record.
	DueAt time.Time
	// Observed contains a detached copy; nil means this field is absent.
	Observed *CoverageInterval
	// Predicted contains a detached copy; nil means this field is absent.
	Predicted *CoverageInterval
}

type catalogIdentityJSON struct {
	Family                string  `json:"family"`
	AnalysisCenter        string  `json:"analysis_center"`
	Publisher             string  `json:"publisher"`
	SolutionClass         string  `json:"solution_class"`
	Campaign              string  `json:"campaign"`
	FilenameVersion       uint8   `json:"filename_version"`
	Date                  string  `json:"date"`
	Issue                 string  `json:"issue"`
	Span                  string  `json:"span"`
	Sample                string  `json:"sample"`
	OfficialFilename      string  `json:"official_filename"`
	Format                string  `json:"format"`
	FormatVersion         *string `json:"format_version"`
	PredictionHorizonDays *uint8  `json:"prediction_horizon_days"`
}

type catalogCoverageJSON struct {
	From  string `json:"from"`
	Until string `json:"until"`
}

// ResolveProductIdentity asks C to derive every identity and filename field.
func ResolveProductIdentity(request CatalogRequest) (ProductIdentity, error) {
	center, family, year, month, day, sample, issue := request.nativeParts()
	value, err := native.CatalogIdentity(center, family, year, month, day, sample, issue)
	if err != nil {
		return ProductIdentity{}, publicError(err)
	}
	return publicIdentity(value), nil
}

// ProductIdentityFor is an explicit-argument convenience around
// ResolveProductIdentity.
func ProductIdentityFor(center string, family ProductFamily, date time.Time, sample, issue string) (ProductIdentity, error) {
	return ResolveProductIdentity(CatalogRequest{Center: center, Family: family, Date: date, Sample: sample, Issue: issue})
}

// DefaultSampleForDate returns the C-catalog default sampling token.
func DefaultSampleForDate(center string, family ProductFamily, date time.Time) (string, error) {
	year, month, day := date.UTC().Date()
	value, err := native.CatalogDefaultSample(center, native.ProductFamily(family), year, uint8(month), uint8(day))
	return value, publicError(err)
}

// SupportedSamples returns every C-catalog-supported sample for a date and
// optional issue.
func SupportedSamples(request CatalogRequest) ([]string, error) {
	center, family, year, month, day, _, issue := request.nativeParts()
	values, err := native.CatalogSupportedSamples(center, family, year, month, day, issue)
	return values, publicError(err)
}

// DistributionLocationFor asks C to resolve a source without performing I/O.
func DistributionLocationFor(request CatalogRequest, source DistributionSource) (DistributionLocation, error) {
	center, family, year, month, day, sample, issue := request.nativeParts()
	value, err := native.CatalogLocation(center, family, year, month, day, sample, issue, native.DistributionSource(source))
	if err != nil {
		return DistributionLocation{}, publicError(err)
	}
	return DistributionLocation{Source: DistributionSource(value.Source), OriginalURL: value.OriginalURL, ArchiveFilename: value.ArchiveFilename, Compression: ArchiveCompression(value.Compression)}, nil
}

// ProductSolutionClass returns the C-catalog solution class and fails closed
// for unsupported center/family pairs.
func ProductSolutionClass(center string, family ProductFamily) (SolutionClass, error) {
	value, err := native.CatalogSolutionClass(center, native.ProductFamily(family))
	return SolutionClass(value), publicError(err)
}

// PublicationListingURLs returns the bounded C-approved listing URLs. It does
// not perform polling.
func PublicationListingURLs(request CatalogRequest) ([]string, error) {
	center, family, year, month, day, _, _ := request.nativeParts()
	data, err := native.CatalogListingURLs(center, family, year, month, day)
	if err != nil {
		return nil, publicError(err)
	}
	var urls []string
	if err := json.Unmarshal(data, &urls); err != nil {
		return nil, fmt.Errorf("sidereon: C listing URL JSON: %w", err)
	}
	return urls, nil
}

// NewestPublishedProduct parses one listing body through C and returns its
// C-selected newest product, or nil when the readable listing has no match.
func NewestPublishedProduct(center string, family ProductFamily, body []byte) (*PublishedProduct, error) {
	data, err := native.CatalogNewestPublished(center, native.ProductFamily(family), body)
	if err != nil {
		return nil, publicError(err)
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil, nil
	}
	var value struct {
		Date       string  `json:"date"`
		Issue      string  `json:"issue"`
		Filename   string  `json:"filename"`
		ObservedAt *string `json:"observed_at"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("sidereon: C published-product JSON: %w", err)
	}
	date, err := parseCatalogDate(value.Date)
	if err != nil {
		return nil, err
	}
	return &PublishedProduct{Date: date, Issue: value.Issue, Filename: value.Filename, ObservedAt: value.ObservedAt}, nil
}

// NextIssueDue asks C for the next nominal issue at or after when. Fractional
// seconds are rounded up to the next whole UTC second because the C route's
// public input is an integer second.
func NextIssueDue(center string, family ProductFamily, when time.Time) (NominalIssue, error) {
	when = when.UTC()
	if when.Nanosecond() != 0 {
		when = when.Add(time.Second - time.Duration(when.Nanosecond()))
	}
	data, err := native.CatalogNextIssueDue(center, native.ProductFamily(family), when.Year(), uint8(when.Month()), uint8(when.Day()), uint8(when.Hour()), uint8(when.Minute()), uint8(when.Second()))
	if err != nil {
		return NominalIssue{}, publicError(err)
	}
	var value struct {
		Identity catalogIdentityJSON `json:"identity"`
		DueAt    string              `json:"due_at"`
		Covers   struct {
			Observed  *catalogCoverageJSON `json:"observed"`
			Predicted *catalogCoverageJSON `json:"predicted"`
		} `json:"covers"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return NominalIssue{}, fmt.Errorf("sidereon: C nominal-issue JSON: %w", err)
	}
	identity, err := identityFromJSON(value.Identity)
	if err != nil {
		return NominalIssue{}, err
	}
	due, err := time.Parse(time.RFC3339, value.DueAt)
	if err != nil {
		return NominalIssue{}, fmt.Errorf("sidereon: C due time: %w", err)
	}
	result := NominalIssue{Identity: identity, DueAt: due.UTC()}
	if value.Covers.Observed != nil {
		result.Observed, err = parseCoverage(value.Covers.Observed.From, value.Covers.Observed.Until)
		if err != nil {
			return NominalIssue{}, err
		}
	}
	if value.Covers.Predicted != nil {
		result.Predicted, err = parseCoverage(value.Covers.Predicted.From, value.Covers.Predicted.Until)
		if err != nil {
			return NominalIssue{}, err
		}
	}
	return result, nil
}

// PredictedIONEXLineCandidates returns C's ordered same-map-date candidates.
func PredictedIONEXLineCandidates(date time.Time, sample string) ([]PredictedIONEXCandidate, error) {
	year, month, day := date.UTC().Date()
	data, err := native.CatalogPredictedIONEXCandidates(year, uint8(month), uint8(day), sample)
	if err != nil {
		return nil, publicError(err)
	}
	var values []struct {
		Center   string `json:"center"`
		Date     string `json:"date"`
		Sample   string `json:"sample"`
		Issue    string `json:"issue"`
		Filename string `json:"filename"`
		URL      string `json:"url"`
	}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("sidereon: C IONEX-candidate JSON: %w", err)
	}
	result := make([]PredictedIONEXCandidate, len(values))
	for i, value := range values {
		candidateDate, err := parseCatalogDate(value.Date)
		if err != nil {
			return nil, err
		}
		result[i] = PredictedIONEXCandidate{Center: value.Center, Date: candidateDate, Sample: value.Sample, Issue: value.Issue, Filename: value.Filename, URL: value.URL}
	}
	return result, nil
}

func publicIdentity(value native.ProductIdentity) ProductIdentity {
	date := time.Date(int(value.Year), time.Month(value.Month), int(value.Day), 0, 0, 0, 0, time.UTC)
	return ProductIdentity{Family: ProductFamily(value.Family), AnalysisCenter: value.AnalysisCenter, Publisher: ProductPublisher(value.Publisher), SolutionClass: SolutionClass(value.SolutionClass), Campaign: ProductCampaign(value.Campaign), FilenameVersion: value.FilenameVersion, Date: date, HasIssue: value.HasIssue, Issue: value.Issue, Span: value.Span, Sample: value.Sample, OfficialFilename: value.OfficialFilename, Format: ProductFormat(value.Format), HasFormatVersion: value.HasFormatVersion, FormatVersion: value.FormatVersion, HasPredictionHorizonDays: value.HasPredictionHorizonDays, PredictionHorizonDays: value.PredictionHorizonDays}
}

func parseCatalogDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("sidereon: C catalog date %q: %w", value, err)
	}
	return date.UTC(), nil
}

func parseCoverage(from, until string) (*CoverageInterval, error) {
	start, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, fmt.Errorf("sidereon: C coverage start: %w", err)
	}
	end, err := time.Parse(time.RFC3339, until)
	if err != nil {
		return nil, fmt.Errorf("sidereon: C coverage end: %w", err)
	}
	return &CoverageInterval{From: start.UTC(), Until: end.UTC()}, nil
}

func identityFromJSON(value catalogIdentityJSON) (ProductIdentity, error) {
	date, err := parseCatalogDate(value.Date)
	if err != nil {
		return ProductIdentity{}, err
	}
	family, err := familyFromCode(value.Family)
	if err != nil {
		return ProductIdentity{}, err
	}
	publisher, err := publisherFromCode(value.Publisher)
	if err != nil {
		return ProductIdentity{}, err
	}
	solution, err := solutionFromCode(value.SolutionClass)
	if err != nil {
		return ProductIdentity{}, err
	}
	campaign, err := campaignFromCode(value.Campaign)
	if err != nil {
		return ProductIdentity{}, err
	}
	format, err := formatFromCode(value.Format)
	if err != nil {
		return ProductIdentity{}, err
	}
	result := ProductIdentity{Family: family, AnalysisCenter: value.AnalysisCenter, Publisher: publisher, SolutionClass: solution, Campaign: campaign, FilenameVersion: value.FilenameVersion, Date: date, HasIssue: value.Issue != "", Issue: value.Issue, Span: value.Span, Sample: value.Sample, OfficialFilename: value.OfficialFilename, Format: format}
	if value.FormatVersion != nil {
		result.HasFormatVersion = true
		result.FormatVersion = *value.FormatVersion
	}
	if value.PredictionHorizonDays != nil {
		result.HasPredictionHorizonDays = true
		result.PredictionHorizonDays = *value.PredictionHorizonDays
	}
	return result, nil
}

func familyFromCode(value string) (ProductFamily, error) {
	switch value {
	case "sp3":
		return ProductFamilySP3, nil
	case "ionex":
		return ProductFamilyIONEX, nil
	case "clk", "rinex_clock":
		return ProductFamilyRINEXClock, nil
	case "nav", "rinex_navigation":
		return ProductFamilyRINEXNavigation, nil
	}
	return 0, fmt.Errorf("sidereon: unknown C product family %q", value)
}
func publisherFromCode(value string) (ProductPublisher, error) {
	switch value {
	case "igs", "IGS":
		return ProductPublisherIGS, nil
	case "code", "COD":
		return ProductPublisherCODE, nil
	case "esa", "ESA":
		return ProductPublisherESA, nil
	case "gfz", "GFZ":
		return ProductPublisherGFZ, nil
	case "whu", "WUM":
		return ProductPublisherWHU, nil
	}
	return 0, fmt.Errorf("sidereon: unknown C publisher %q", value)
}
func solutionFromCode(value string) (SolutionClass, error) {
	switch value {
	case "final":
		return SolutionClassFinal, nil
	case "rapid":
		return SolutionClassRapid, nil
	case "ultra_rapid":
		return SolutionClassUltraRapid, nil
	case "predicted":
		return SolutionClassPredicted, nil
	case "broadcast":
		return SolutionClassBroadcast, nil
	case "near_real_time":
		return SolutionClassNearRealTime, nil
	}
	return 0, fmt.Errorf("sidereon: unknown C solution class %q", value)
}
func campaignFromCode(value string) (ProductCampaign, error) {
	switch value {
	case "operational", "OPS":
		return ProductCampaignOperational, nil
	case "multi_gnss", "MGN":
		return ProductCampaignMultiGNSS, nil
	case "multi_gnss_experiment", "MGX":
		return ProductCampaignMultiGNSSExperiment, nil
	case "broadcast", "BRD":
		return ProductCampaignBroadcast, nil
	}
	return 0, fmt.Errorf("sidereon: unknown C product campaign %q", value)
}
func formatFromCode(value string) (ProductFormat, error) {
	switch value {
	case "sp3", "SP3":
		return ProductFormatSP3, nil
	case "ionex", "IONEX":
		return ProductFormatIONEX, nil
	case "rinex_clock", "clk", "RINEX_CLK":
		return ProductFormatRINEXClock, nil
	case "rinex_navigation", "nav", "RINEX_NAV":
		return ProductFormatRINEXNavigation, nil
	}
	return 0, fmt.Errorf("sidereon: unknown C product format %q", value)
}
