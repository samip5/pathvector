package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-ping/ping"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/natesales/pathvector/internal/util"
)

// Peer stores a single peer config
type Peer struct {
	Template *string `yaml:"template" description:"Configuration template" default:"-"`

	Description *string `yaml:"description" description:"Peer description" default:"-"`
	Disabled    *bool   `yaml:"disabled" description:"Should the sessions be disabled?" default:"false"`

	// BGP Attributes
	ASN                 *int      `yaml:"asn" description:"Local ASN" validate:"required" default:"0"`
	NeighborIPs         *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip" default:"-"`
	Prepends            *int      `yaml:"prepends" description:"Number of times to prepend local AS on export" default:"0"`
	LocalPref           *int      `yaml:"local-pref" description:"BGP local preference" default:"100"`
	Multihop            *bool     `yaml:"multihop" description:"Should BGP multihop be enabled? (255 max hops)" default:"false"`
	Listen4             *string   `yaml:"listen4" description:"IPv4 BGP listen address" default:"-"`
	Listen6             *string   `yaml:"listen6" description:"IPv6 BGP listen address" default:"-"`
	LocalASN            *int      `yaml:"local-asn" description:"Local ASN as defined in the global ASN field" default:"-"`
	LocalPort           *int      `yaml:"local-port" description:"Local TCP port" default:"179"`
	NeighborPort        *int      `yaml:"neighbor-port" description:"Neighbor TCP port" default:"179"`
	Passive             *bool     `yaml:"passive" description:"Should we listen passively?" default:"false"`
	Direct              *bool     `yaml:"direct" description:"Are we an direct session?" default:"false"`
	NextHopSelf         *bool     `yaml:"next-hop-self" description:"Should BGP next-hop-self be enabled?" default:"false"`
	BFD                 *bool     `yaml:"bfd" description:"Should BFD be enabled?" default:"false"`
	Password            *string   `yaml:"password" description:"BGP MD5 password" default:"-"`
	RSClient            *bool     `yaml:"rs-client" description:"Should this peer be a route server client?" default:"false"`
	RRClient            *bool     `yaml:"rr-client" description:"Should this peer be a route reflector client?" default:"false"`
	RemovePrivateASNs   *bool     `yaml:"remove-private-asns" description:"Should private ASNs be removed from path before exporting?" default:"true"`
	MPUnicast46         *bool     `yaml:"mp-unicast-46" description:"Should this peer be configured with multiprotocol IPv4 and IPv6 unicast?" default:"false"`
	AllowLocalAS        *bool     `yaml:"allow-local-as" description:"Should routes originated by the local ASN be accepted?" default:"false"`
	AddPathTx           *bool     `yaml:"add-path-tx" description:"Enable BGP additional paths on export?" default:"false"`
	AddPathRx           *bool     `yaml:"add-path-rx" description:"Enable BGP additional paths on import?" default:"false"`
	ImportNextHop       *string   `yaml:"import-next-hop" description:"Rewrite the BGP next hop before importing routes learned from this peer" default:"-"`
	ExportNextHop       *string   `yaml:"export-next-hop" description:"Rewrite the BGP next hop before announcing routes to this peer" default:"-"`
	Confederation       *int      `yaml:"confederation" description:"BGP confederation (RFC 5065)" default:"-"`
	ConfederationMember *bool     `yaml:"confederation-member" description:"Should this peer be a member of the local confederation?" default:"false"`
	TTLSecurity         *bool     `yaml:"ttl-security" description:"RFC 5082 Generalized TTL Security Mechanism" default:"false"`

	ImportCommunities    *[]string `yaml:"import-communities" description:"List of communities to add to all imported routes" default:"-"`
	ExportCommunities    *[]string `yaml:"export-communities" description:"List of communities to add to all exported routes" default:"-"`
	AnnounceCommunities  *[]string `yaml:"announce-communities" description:"Announce all routes matching these communities to the peer" default:"-"`
	RemoveCommunities    *[]string `yaml:"remove-communities" description:"List of communities to remove before from routes announced by this peer" default:"-"`
	RemoveAllCommunities *int      `yaml:"remove-all-communities" description:"Remove all standard and large communities beginning with this value" default:"-"`

	ASPrefs *map[uint32]uint32 `yaml:"as-prefs" description:"Map of ASN to import local pref (not included in optimizer)" default:"-"`

	// Filtering
	ASSet                   *string `yaml:"as-set" description:"Peer's as-set for filtering" default:"-"`
	ImportLimit4            *int    `yaml:"import-limit4" description:"Maximum number of IPv4 prefixes to import" default:"1000000"`
	ImportLimit6            *int    `yaml:"import-limit6" description:"Maximum number of IPv6 prefixes to import" default:"200000"`
	EnforceFirstAS          *bool   `yaml:"enforce-first-as" description:"Should we only accept routes who's first AS is equal to the configured peer address?" default:"true"`
	EnforcePeerNexthop      *bool   `yaml:"enforce-peer-nexthop" description:"Should we only accept routes with a next hop equal to the configured neighbor address?" default:"true"`
	ForcePeerNexthop        *bool   `yaml:"force-peer-nexthop" description:"Rewrite nexthop to peer address" default:"false"`
	MaxPrefixTripAction     *string `yaml:"max-prefix-action" description:"What action should be taken when the max prefix limit is tripped?" default:"disable"`
	AllowBlackholeCommunity *bool   `yaml:"allow-blackhole-community" description:"Should this peer be allowed to send routes with the blackhole community?" default:"false"`

	FilterIRR                  *bool `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"false"`
	FilterRPKI                 *bool `yaml:"filter-rpki" description:"Should RPKI invalids be rejected?" default:"true"`
	FilterMaxPrefix            *bool `yaml:"filter-max-prefix" description:"Should max prefix filtering be applied?" default:"true"`
	FilterBogonRoutes          *bool `yaml:"filter-bogon-routes" description:"Should bogon prefixes be rejected?" default:"true"`
	FilterBogonASNs            *bool `yaml:"filter-bogon-asns" description:"Should paths containing a bogon ASN be rejected?" default:"true"`
	FilterTransitASNs          *bool `yaml:"filter-transit-asns" description:"Should paths containing transit-free ASNs be rejected? (Peerlock Lite)'" default:"false"`
	FilterPrefixLength         *bool `yaml:"filter-prefix-length" description:"Should too large/small prefixes (IPv4 8 > len > 24 and IPv6 12 > len > 48) be rejected?" default:"true"`
	FilterNeverViaRouteServers *bool `yaml:"filter-never-via-route-servers" description:"Should routes containing an ASN reported in PeeringDB to never be reachable via route servers be filtered?" default:"false"`

	AutoImportLimits *bool `yaml:"auto-import-limits" description:"Get import limits automatically from PeeringDB?" default:"false"`
	AutoASSet        *bool `yaml:"auto-as-set" description:"Get as-set automatically from PeeringDB? If no as-set exists in PeeringDB, a warning will be shown and the peer ASN used instead." default:"false"`

	HonorGracefulShutdown *bool `yaml:"honor-graceful-shutdown" description:"Should RFC8326 graceful shutdown be enabled?" default:"true"`

	Prefixes *[]string `yaml:"prefixes" description:"Prefixes to accept" default:"-"`

	// Export options
	AnnounceDefault    *bool `yaml:"announce-default" description:"Should a default route be exported to this peer?" default:"false"`
	AnnounceOriginated *bool `yaml:"announce-originated" description:"Should locally originated routes be announced to this peer?" default:"true"`

	// Custom daemon configuration
	SessionGlobal  *string `yaml:"session-global" description:"Configuration to add to each session before any defined BGP protocols" default:"-"`
	PreImport      *string `yaml:"pre-import" description:"Configuration to add at the beginning of the import filter" default:"-"`
	PreExport      *string `yaml:"pre-export" description:"Configuration to add at the beginning of the export filter" default:"-"`
	PreImportFinal *string `yaml:"pre-import-final" description:"Configuration to add immediately before the final accept/reject on import" default:"-"`
	PreExportFinal *string `yaml:"pre-export-final" description:"Configuration to add immediately before the final accept/reject on export" default:"-"`

	// Optimizer
	OptimizerProbeSources *[]string `yaml:"probe-sources" description:"Optimizer probe source addresses" default:"-"`
	OptimizeInbound       *bool     `yaml:"optimize-inbound" description:"Should the optimizer modify inbound policy?" default:"false"`

	ProtocolName                *string   `yaml:"-" description:"-" default:"-"`
	Protocols                   *[]string `yaml:"-" description:"-" default:"-"`
	PrefixSet4                  *[]string `yaml:"-" description:"-" default:"-"`
	PrefixSet6                  *[]string `yaml:"-" description:"-" default:"-"`
	ImportStandardCommunities   *[]string `yaml:"-" description:"-" default:"-"`
	ImportLargeCommunities      *[]string `yaml:"-" description:"-" default:"-"`
	ExportStandardCommunities   *[]string `yaml:"-" description:"-" default:"-"`
	ExportLargeCommunities      *[]string `yaml:"-" description:"-" default:"-"`
	AnnounceStandardCommunities *[]string `yaml:"-" description:"-" default:"-"`
	AnnounceLargeCommunities    *[]string `yaml:"-" description:"-" default:"-"`
	RemoveStandardCommunities   *[]string `yaml:"-" description:"-" default:"-"`
	RemoveLargeCommunities      *[]string `yaml:"-" description:"-" default:"-"`
	BooleanOptions              *[]string `yaml:"-" description:"-" default:"-"`
}

// VRRPInstance stores a single VRRP instance
type VRRPInstance struct {
	State     string   `yaml:"state" description:"VRRP instance state ('primary' or 'backup')" validate:"required"`
	Interface string   `yaml:"interface" description:"Interface to send VRRP packets on" validate:"required"`
	VRID      uint     `yaml:"vrid" description:"RFC3768 VRRP Virtual Router ID (1-255)" validate:"required"`
	Priority  uint     `yaml:"priority" description:"RFC3768 VRRP Priority" validate:"required"`
	VIPs      []string `yaml:"vips" description:"List of virtual IPs" validate:"required,cidr"`

	VIPs4 []string `yaml:"-" description:"-"`
	VIPs6 []string `yaml:"-" description:"-"`
}

// BFDInstance stores a single BFD instance
type BFDInstance struct {
	Neighbor   *string `yaml:"neighbor" description:"Neighbor IP address" default:"-"`
	Interface  *string `yaml:"interface" description:"Interface (pattern accepted)" default:"-"`
	Interval   *uint   `yaml:"interval" description:"RX and TX interval" default:"200"`
	Multiplier *uint   `yaml:"multiplier" description:"Number of missed packets for the state to be declared down" default:"10"`

	ProtocolName *string `yaml:"-" description:"-" default:"-"`
}

// Augments store BIRD specific options
type Augments struct {
	Accept4        []string          `yaml:"accept4" description:"List of BIRD protocols to import into the IPv4 table"`
	Accept6        []string          `yaml:"accept6" description:"List of BIRD protocols to import into the IPv6 table"`
	Reject4        []string          `yaml:"reject4" description:"List of BIRD protocols to not import into the IPv4 table"`
	Reject6        []string          `yaml:"reject6" description:"List of BIRD protocols to not import into the IPv6 table"`
	Statics        map[string]string `yaml:"statics" description:"List of static routes to include in BIRD"`
	SRDCommunities []string          `yaml:"srd-communities" description:"List of communities to filter routes exported to kernel (if list is not empty, all other prefixes will not be exported)"`

	SRDStandardCommunities []string          `yaml:"-" description:"-"`
	SRDLargeCommunities    []string          `yaml:"-" description:"-"`
	Statics4               map[string]string `yaml:"-" description:"-"`
	Statics6               map[string]string `yaml:"-" description:"-"`
}

// ProbeResult stores a single probe result
type ProbeResult struct {
	Time  int64
	Stats ping.Statistics
}

// Optimizer stores route optimizer configuration
type Optimizer struct {
	Targets             []string `yaml:"targets" description:"List of probe targets"`
	LatencyThreshold    uint     `yaml:"latency-threshold" description:"Maximum allowable latency in milliseconds" default:"100"`
	PacketLossThreshold float64  `yaml:"packet-loss-threshold" description:"Maximum allowable packet loss (percent)" default:"0.5"`
	LocalPrefModifier   uint     `yaml:"modifier" description:"Amount to lower local pref by for depreferred peers" default:"20"`

	PingCount   int `yaml:"probe-count" description:"Number of pings to send in each run" default:"5"`
	PingTimeout int `yaml:"probe-timeout" description:"Number of seconds to wait before considering the ICMP message unanswered" default:"1"`
	Interval    int `yaml:"probe-interval" description:"Number of seconds wait between each optimizer run" default:"120"`
	CacheSize   int `yaml:"cache-size" description:"Number of probe results to store per peer" default:"15"`

	ProbeUDPMode bool `yaml:"probe-udp" description:"Use UDP probe (else ICMP)" default:"false"`

	AlertScript string `yaml:"alert-script" description:"Script to call on optimizer event"`

	ExitOnCacheFull bool `yaml:"exit-on-cache-full" description:"Exit optimizer on cache full" default:"false"`

	Db map[string][]ProbeResult `yaml:"-" description:"-"`
}

// Config stores the global configuration
type Config struct {
	PeeringDBQueryTimeout uint   `yaml:"peeringdb-query-timeout" description:"PeeringDB query timeout in seconds" default:"10"`
	IRRQueryTimeout       uint   `yaml:"irr-query-timeout" description:"IRR query timeout in seconds" default:"30"`
	BIRDDirectory         string `yaml:"bird-directory" description:"Directory to store BIRD configs" default:"/etc/bird/"`
	BIRDBinary            string `yaml:"bird-binary" description:"Path to BIRD binary" default:"/usr/sbin/bird"`
	BIRDSocket            string `yaml:"bird-socket" description:"UNIX control socket for BIRD" default:"/run/bird/bird.ctl"`
	CacheDirectory        string `yaml:"cache-directory" description:"Directory to store runtime configuration cache" default:"/var/run/pathvector/cache/"`
	KeepalivedConfig      string `yaml:"keepalived-config" description:"Configuration file for keepalived" default:"/etc/keepalived.conf"`
	WebUIFile             string `yaml:"web-ui-file" description:"File to write web UI to (disabled if empty)" default:""`
	LogFile               string `yaml:"log-file" description:"Log file location" default:"syslog"`

	PortalHost string `yaml:"portal-host" description:"Peering portal host (disabled if empty)" default:""`
	PortalKey  string `yaml:"portal-key" description:"Peering portal API key" default:""`
	Hostname   string `yaml:"hostname" description:"Router hostname (default system hostname)" default:""`

	ASN              int      `yaml:"asn" description:"Autonomous System Number" validate:"required" default:"0"`
	Prefixes         []string `yaml:"prefixes" description:"List of prefixes to announce"`
	Communities      []string `yaml:"communities" description:"List of RFC1997 BGP communities"`
	LargeCommunities []string `yaml:"large-communities" description:"List of RFC8092 large BGP communities"`

	RouterID      string `yaml:"router-id" description:"Router ID (dotted quad notation)" validate:"required"`
	IRRServer     string `yaml:"irr-server" description:"Internet routing registry server" default:"rr.ntt.net"`
	RTRServer     string `yaml:"rtr-server" description:"RPKI-to-router server" default:"rtr.rpki.cloudflare.com:8282"`
	BGPQArgs      string `yaml:"bgpq-args" description:"Additional command line arguments to pass to bgpq4" default:""`
	KeepFiltered  bool   `yaml:"keep-filtered" description:"Should filtered routes be kept in memory?" default:"false"`
	KernelLearn   bool   `yaml:"kernel-learn" description:"Should routes from the kernel be learned into BIRD?" default:"false"`
	KernelExport  bool   `yaml:"kernel-export" description:"Export routes to kernel routing table" default:"true"`
	MergePaths    bool   `yaml:"merge-paths" description:"Should best and equivalent non-best routes be imported to build ECMP routes?" default:"false"`
	Source4       string `yaml:"source4" description:"Source IPv4 address"`
	Source6       string `yaml:"source6" description:"Source IPv6 address"`
	DefaultRoute  bool   `yaml:"default-route" description:"Add a default route" default:"true"`
	AcceptDefault bool   `yaml:"accept-default" description:"Should default routes be added to the bogon list?" default:"false"`
	KernelTable   int    `yaml:"kernel-table" description:"Kernel table"`
	RPKIEnable    bool   `yaml:"rpki-enable" description:"Enable RPKI RTR session" default:"true"`

	Peers         map[string]*Peer         `yaml:"peers" description:"BGP peer configuration"`
	Templates     map[string]*Peer         `yaml:"templates" description:"BGP peer templates"`
	VRRPInstances map[string]*VRRPInstance `yaml:"vrrp" description:"List of VRRP instances"`
	BFDInstances  map[string]*BFDInstance  `yaml:"bfd" description:"BFD instances"`
	Augments      Augments                 `yaml:"augments" description:"Custom configuration options"`
	Optimizer     Optimizer                `yaml:"optimizer" description:"Route optimizer options"`

	RTRServerHost string   `yaml:"-" description:"-"`
	RTRServerPort int      `yaml:"-" description:"-"`
	Prefixes4     []string `yaml:"-" description:"-"`
	Prefixes6     []string `yaml:"-" description:"-"`
	QueryNVRS     bool     `yaml:"-" description:"-"`
	NVRSASNs      []uint32 `yaml:"-" description:"-"`
}

// categorizeCommunity checks if the community is in standard or large form, or an empty string if invalid
func categorizeCommunity(input string) string {
	// Test if it fits the criteria for a standard community
	standardSplit := strings.Split(input, ",")
	if len(standardSplit) == 2 {
		firstPart, err := strconv.Atoi(standardSplit[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(standardSplit[1])
		if err != nil {
			return ""
		}

		if firstPart < 0 || firstPart > 65535 {
			return ""
		}
		if secondPart < 0 || secondPart > 65535 {
			return ""
		}
		return "standard"
	}

	// Test if it fits the criteria for a large community
	largeSplit := strings.Split(input, ":")
	if len(largeSplit) == 3 {
		firstPart, err := strconv.Atoi(largeSplit[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(largeSplit[1])
		if err != nil {
			return ""
		}
		thirdPart, err := strconv.Atoi(largeSplit[2])
		if err != nil {
			return ""
		}

		if firstPart < 0 || firstPart > 4294967295 {
			return ""
		}
		if secondPart < 0 || secondPart > 4294967295 {
			return ""
		}
		if thirdPart < 0 || thirdPart > 4294967295 {
			return ""
		}
		return "large"
	}

	return ""
}

// Load loads a configuration file from a YAML file
func Load(configBlob []byte) (*Config, error) {
	var c Config
	// Set global config defaults
	if err := defaults.Set(&c); err != nil {
		log.Fatal(err)
	}

	if err := yaml.UnmarshalStrict(configBlob, &c); err != nil {
		return nil, errors.New("YAML unmarshal: " + err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(&c); err != nil {
		return nil, errors.New("Validation: " + err.Error())
	}

	// Check for invalid templates
	for templateName, templateData := range c.Templates {
		if templateData.Template != nil && *templateData.Template != "" {
			log.Fatalf("Templates must not have a template field set, but %s does", templateName)
		}
	}

	// Set hostname if empty
	if c.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("Hostname is not defined and unable to get system hostname: %s", err)
		}
		c.Hostname = hostname
	}

	for peerName, peerData := range c.Peers {
		// Set sanitized peer name
		peerData.ProtocolName = util.Sanitize(peerName)

		// If any peer has NVRS filtering enabled, mark it for querying.
		if peerData.FilterNeverViaRouteServers != nil {
			c.QueryNVRS = true
		}

		if peerData.NeighborIPs == nil || len(*peerData.NeighborIPs) < 1 {
			log.Fatalf("[%s] has no neighbors defined", peerName)
		}

		peerData.BooleanOptions = &[]string{}

		// Assign values from template
		if peerData.Template != nil && *peerData.Template != "" {
			template := c.Templates[*peerData.Template]
			if template == nil {
				log.Fatalf("Template %s not found", *peerData.Template)
			}
			templateValue := reflect.ValueOf(*template)
			peerValue := reflect.ValueOf(c.Peers[peerName]).Elem()

			templateValueType := templateValue.Type()
			for i := 0; i < templateValueType.NumField(); i++ {
				fieldName := templateValueType.Field(i).Name
				peerFieldValue := peerValue.FieldByName(fieldName)
				if fieldName != "Template" { // Ignore the template field
					pVal := reflect.Indirect(peerFieldValue)
					peerHasValueConfigured := pVal.IsValid()
					tValue := templateValue.Field(i)
					templateHasValueConfigured := !tValue.IsNil()
					if templateHasValueConfigured && !peerHasValueConfigured {
						// Use the template's value
						peerFieldValue.Set(templateValue.Field(i))
					}

					log.Debugf("[%s] field: %s template's value: %+v kind: %T templateHasValueConfigured: %v", peerName, fieldName, reflect.Indirect(tValue), tValue.Kind().String(), templateHasValueConfigured)
				}
			}
		} // end peer template processor

		// Set default values
		peerValue := reflect.ValueOf(c.Peers[peerName]).Elem()
		templateValueType := peerValue.Type()
		for i := 0; i < templateValueType.NumField(); i++ {
			fieldName := templateValueType.Field(i).Name
			fieldValue := peerValue.FieldByName(fieldName)
			defaultString := templateValueType.Field(i).Tag.Get("default")
			if defaultString == "" {
				log.Fatalf("Code error: field %s has no default value", fieldName)
			} else if defaultString != "-" {
				log.Debugf("[%s] (before defaulting, after templating) field %s value %+v", peerName, fieldName, reflect.Indirect(fieldValue))
				if fieldValue.IsNil() {
					elemToSwitch := templateValueType.Field(i).Type.Elem().Kind()
					switch elemToSwitch {
					case reflect.String:
						log.Debugf("[%s] setting field %s to value %+v", peerName, fieldName, defaultString)
						fieldValue.Set(reflect.ValueOf(&defaultString))
					case reflect.Int:
						defaultValueInt, err := strconv.Atoi(defaultString)
						if err != nil {
							log.Fatalf("Can't convert '%s' to uint", defaultString)
						}
						log.Debugf("[%s] setting field %s to value %+v", peerName, fieldName, defaultValueInt)
						fieldValue.Set(reflect.ValueOf(&defaultValueInt))
					case reflect.Bool:
						var err error // explicit declaration used to avoid scope issues of defaultValue
						defaultBool, err := strconv.ParseBool(defaultString)
						if err != nil {
							log.Fatalf("Can't parse bool %s", defaultString)
						}
						log.Debugf("[%s] setting field %s to value %+v", peerName, fieldName, defaultBool)
						fieldValue.Set(reflect.ValueOf(&defaultBool))
					case reflect.Struct, reflect.Slice:
						// Ignore structs and slices
					default:
						log.Fatalf("Unknown kind %+v for field %s", elemToSwitch, fieldName)
					}
				} else {
					// Add boolean values to the peer's config
					if templateValueType.Field(i).Type.Elem().Kind() == reflect.Bool {
						*peerData.BooleanOptions = append(*peerData.BooleanOptions, templateValueType.Field(i).Tag.Get("yaml"))
					}
				}
			} else {
				log.Debugf("[%s] skipping field %s with ignored default (-)", peerName, fieldName)
			}
		}
	}

	// Parse origin routes by assembling OriginIPv{4,6} lists by address family
	for _, prefix := range c.Prefixes {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("Invalid origin prefix: " + prefix)
		}

		if pfx.To4() == nil { // If IPv6
			c.Prefixes6 = append(c.Prefixes6, prefix)
		} else { // If IPv4
			c.Prefixes4 = append(c.Prefixes4, prefix)
		}
	}

	// Initialize static maps
	c.Augments.Statics4 = map[string]string{}
	c.Augments.Statics6 = map[string]string{}

	// Categorize communities
	if c.Augments.SRDCommunities != nil {
		for _, community := range c.Augments.SRDCommunities {
			communityType := categorizeCommunity(community)
			if communityType == "standard" {
				if c.Augments.SRDStandardCommunities == nil {
					c.Augments.SRDStandardCommunities = []string{}
				}
				c.Augments.SRDStandardCommunities = append(c.Augments.SRDStandardCommunities, community)
			} else if communityType == "large" {
				if c.Augments.SRDLargeCommunities == nil {
					c.Augments.SRDLargeCommunities = []string{}
				}
				c.Augments.SRDLargeCommunities = append(c.Augments.SRDLargeCommunities, strings.ReplaceAll(community, ":", ","))
			} else {
				return nil, errors.New("Invalid SRD community: " + community)
			}
		}
	}

	// Parse static routes
	for prefix, nexthop := range c.Augments.Statics {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("Invalid static prefix: " + prefix)
		}
		if net.ParseIP(nexthop) == nil {
			return nil, errors.New("Invalid static nexthop: " + nexthop)
		}

		if pfx.To4() == nil { // If IPv6
			c.Augments.Statics6[prefix] = nexthop
		} else { // If IPv4
			c.Augments.Statics4[prefix] = nexthop
		}
	}

	// Parse BFD configs
	for instanceName, bfdInstance := range c.BFDInstances {
		if net.ParseIP(*bfdInstance.Neighbor) == nil {
			return nil, fmt.Errorf("invalid BFD neighbor %s", *bfdInstance.Neighbor)
		}
		bfdInstance.ProtocolName = util.Sanitize(instanceName)
	}

	// Parse VRRP configs
	for _, vrrpInstance := range c.VRRPInstances {
		// Sort VIPs by address family
		for _, vip := range vrrpInstance.VIPs {
			ip, _, err := net.ParseCIDR(vip)
			if err != nil {
				return nil, errors.New("Invalid VIP: " + vip)
			}

			if ip.To4() == nil { // If IPv6
				vrrpInstance.VIPs6 = append(vrrpInstance.VIPs6, vip)
			} else { // If IPv4
				vrrpInstance.VIPs4 = append(vrrpInstance.VIPs4, vip)
			}
		}

		// Validate vrrpInstance
		if vrrpInstance.State == "primary" {
			vrrpInstance.State = "MASTER"
		} else if vrrpInstance.State == "backup" {
			vrrpInstance.State = "BACKUP"
		} else {
			return nil, errors.New("VRRP state must be 'primary' or 'backup', unexpected " + vrrpInstance.State)
		}
	}

	// Parse RTR server
	if c.RTRServer != "" {
		rtrServerParts := strings.Split(c.RTRServer, ":")
		if len(rtrServerParts) != 2 {
			log.Fatalf("Invalid rtr-server '%s' format should be host:port", rtrServerParts)
		}
		c.RTRServerHost = rtrServerParts[0]
		rtrServerPort, err := strconv.Atoi(rtrServerParts[1])
		if err != nil {
			log.Fatalf("Invalid RTR server port %s", rtrServerParts[1])
		}
		c.RTRServerPort = rtrServerPort
	}

	for _, peerData := range c.Peers {
		// Build static prefix filters
		if peerData.Prefixes != nil {
			for _, prefix := range *peerData.Prefixes {
				pfx, _, err := net.ParseCIDR(prefix)
				if err != nil {
					return nil, errors.New("Invalid prefix: " + prefix)
				}

				if pfx.To4() == nil { // If IPv6
					if peerData.PrefixSet6 == nil {
						peerData.PrefixSet6 = &[]string{}
					}
					pfxSet6 := append(*peerData.PrefixSet6, prefix)
					peerData.PrefixSet6 = &pfxSet6
				} else { // If IPv4
					if peerData.PrefixSet4 == nil {
						peerData.PrefixSet4 = &[]string{}
					}
					pfxSet4 := append(*peerData.PrefixSet4, prefix)
					peerData.PrefixSet4 = &pfxSet4
				}
			}
		}

		// Categorize communities
		if peerData.ImportCommunities != nil {
			for _, community := range *peerData.ImportCommunities {
				communityType := categorizeCommunity(community)
				if communityType == "standard" {
					if peerData.ImportStandardCommunities == nil {
						peerData.ImportStandardCommunities = &[]string{}
					}
					*peerData.ImportStandardCommunities = append(*peerData.ImportStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.ImportLargeCommunities == nil {
						peerData.ImportLargeCommunities = &[]string{}
					}
					*peerData.ImportLargeCommunities = append(*peerData.ImportLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid import community: " + community)
				}
			}
		}

		if peerData.ExportCommunities != nil {
			for _, community := range *peerData.ExportCommunities {
				communityType := categorizeCommunity(community)
				if communityType == "standard" {
					if peerData.ExportStandardCommunities == nil {
						peerData.ExportStandardCommunities = &[]string{}
					}
					*peerData.ExportStandardCommunities = append(*peerData.ExportStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.ExportLargeCommunities == nil {
						peerData.ExportLargeCommunities = &[]string{}
					}
					*peerData.ExportLargeCommunities = append(*peerData.ExportLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid export community: " + community)
				}
			}
		}
		if peerData.AnnounceCommunities != nil {
			for _, community := range *peerData.AnnounceCommunities {
				communityType := categorizeCommunity(community)

				if communityType == "standard" {
					if peerData.AnnounceStandardCommunities == nil {
						peerData.AnnounceStandardCommunities = &[]string{}
					}
					*peerData.AnnounceStandardCommunities = append(*peerData.AnnounceStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.AnnounceLargeCommunities == nil {
						peerData.AnnounceLargeCommunities = &[]string{}
					}
					*peerData.AnnounceLargeCommunities = append(*peerData.AnnounceLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid announce community: " + community)
				}
			}
		}
		if peerData.RemoveCommunities != nil {
			for _, community := range *peerData.RemoveCommunities {
				communityType := categorizeCommunity(community)

				if communityType == "standard" {
					if peerData.RemoveStandardCommunities == nil {
						peerData.RemoveStandardCommunities = &[]string{}
					}
					*peerData.RemoveStandardCommunities = append(*peerData.RemoveStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.RemoveLargeCommunities == nil {
						peerData.RemoveLargeCommunities = &[]string{}
					}
					*peerData.RemoveLargeCommunities = append(*peerData.RemoveLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid remove community: " + community)
				}
			}
		}

		// Check for no originated prefixes but announce-originated enabled
		if len(c.Prefixes) < 1 && *peerData.AnnounceOriginated {
			// No locally originated prefixes are defined, so there's nothing to originate
			*peerData.AnnounceOriginated = false
		}
	} // end peer loop

	return &c, nil // nil error
}

func sanitizeConfigName(s string) string {
	out := s
	out = strings.ReplaceAll(out, "*", "")
	out = strings.ReplaceAll(out, "config.", "")
	return out
}

func documentConfigTypes(t reflect.Type) {
	childTypesSet := map[reflect.Type]bool{}
	fmt.Println("## " + sanitizeConfigName(t.String()))
	fmt.Println("| Option | Type | Default | Validation | Description |")
	fmt.Println("|--------|------|---------|------------|-------------|")
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		description := field.Tag.Get("description")
		key := field.Tag.Get("yaml")
		validation := field.Tag.Get("validate")
		fDefault := field.Tag.Get("default")
		if fDefault == "-" {
			fDefault = ""
		}

		if description == "" {
			log.Fatalf("Code error: %s doesn't have a description", field.Name)
		} else if description != "-" { // Ignore descriptions that are -
			if strings.Contains(field.Type.String(), "config.") { // If the type is a config struct
				if field.Type.Kind() == reflect.Map || field.Type.Kind() == reflect.Slice { // Extract the element if the type is a map or slice and add to set (reflect.Type to bool map)
					childTypesSet[field.Type.Elem()] = true
				} else {
					childTypesSet[field.Type] = true
				}
			}
			fmt.Printf("| %s | %s | %s | %s | %s |\n", key, sanitizeConfigName(field.Type.String()), fDefault, validation, description)
		}
	}
	fmt.Println()
	for childType := range childTypesSet {
		documentConfigTypes(childType)
	}
}

// DocumentConfig prints a YAML file with autogenerated configuration documentation
func DocumentConfig() {
	documentConfigTypes(reflect.TypeOf(Config{}))
}
