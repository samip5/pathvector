package config

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/natesales/pathvector/internal/probe"
)

type Peer struct {
	Template *string `yaml:"template" description:"Configuration template" default:"-"`

	Description *string `yaml:"description" description:"Peer description" default:"-"`
	Disabled    *bool   `yaml:"disabled" description:"Should the sessions be disabled?" default:"false"`

	// BGP Attributes
	ASN               *int      `yaml:"asn" description:"Local ASN" validate:"required" default:"0"`
	NeighborIPs       *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip" default:"-"`
	Prepends          *int      `yaml:"prepends" description:"Number of times to prepend local AS on export" default:"0"`
	LocalPref         *int      `yaml:"local-pref" description:"BGP local preference" default:"100"`
	Multihop          *bool     `yaml:"multihop" description:"Should BGP multihop be enabled? (255 max hops)" default:"false"`
	Listen            *string   `yaml:"listen" description:"BGP listen address" default:"-"`
	LocalASN          *int      `yaml:"local-asn" description:"Local ASN as defined in the global ASN field" default:"-"`
	LocalPort         *int      `yaml:"local-port" description:"Local TCP port" default:"179"`
	NeighborPort      *int      `yaml:"neighbor-port" description:"Neighbor TCP port" default:"179"`
	Passive           *bool     `yaml:"passive" description:"Should we listen passively?" default:"false"`
	NextHopSelf       *bool     `yaml:"next-hop-self" description:"Should BGP next-hop-self be enabled?" default:"false"`
	BFD               *bool     `yaml:"bfd" description:"Should BFD be enabled?" default:"false"`
	Password          *string   `yaml:"password" description:"BGP MD5 password" default:"-"`
	RSClient          *bool     `yaml:"rs-client" description:"Should this peer be a route server client?" default:"false"`
	RRClient          *bool     `yaml:"rr-client" description:"Should this peer be a route reflector client?" default:"false"`
	RemovePrivateASNs *bool     `yaml:"remove-private-asns" description:"Should private ASNs be removed from path before exporting?" default:"true"`
	MPUnicast46       *bool     `yaml:"mp-unicast-46" description:"Should this peer be configured with multiprotocol IPv4 and IPv6 unicast?" default:"false"`
	AllowLocalAS      *bool     `yaml:"allow-local-as" description:"Should routes originated by the local ASN be accepted?" default:"false"`

	ImportCommunities   *[]string `yaml:"import-communities" description:"List of communities to add to all imported routes" default:"-"`
	ExportCommunities   *[]string `yaml:"export-communities" description:"List of communities to add to all exported routes" default:"-"`
	AnnounceCommunities *[]string `yaml:"announce-communities" description:"Announce all routes matching these communities to the peer" default:"-"`

	// Filtering
	ASSet                   *string `yaml:"as-set" description:"Peer's as-set for filtering" default:"-"`
	ImportLimit4            *int    `yaml:"import-limit4" description:"Maximum number of IPv4 prefixes to import" default:"1000000"`
	ImportLimit6            *int    `yaml:"import-limit6" description:"Maximum number of IPv6 prefixes to import" default:"200000"`
	EnforceFirstAS          *bool   `yaml:"enforce-first-as" description:"Should we only accept routes who's first AS is equal to the configured peer address?" default:"true"`
	EnforcePeerNexthop      *bool   `yaml:"enforce-peer-nexthop" description:"Should we only accept routes with a next hop equal to the configured neighbor address?" default:"true"`
	MaxPrefixTripAction     *string `yaml:"max-prefix-action" description:"What action should be taken when the max prefix limit is tripped?" default:"disable"`
	AllowBlackholeCommunity *bool   `yaml:"allow-blackhole-community" description:"Should this peer be allowed to send routes with the blackhole community?" default:"false"`

	FilterIRR          *bool `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"false"`
	FilterRPKI         *bool `yaml:"filter-rpki" description:"Should RPKI invalids be rejected?" default:"true"`
	FilterMaxPrefix    *bool `yaml:"filter-max-prefix" description:"Should max prefix filtering be applied?" default:"true"`
	FilterBogonRoutes  *bool `yaml:"filter-bogon-routes" description:"Should bogon prefixes be rejected?" default:"true"`
	FilterBogonASNs    *bool `yaml:"filter-bogon-asns" description:"Should paths containing a bogon ASN be rejected?" default:"true"`
	FilterTransitASNs  *bool `yaml:"filter-transit-asns" description:"Should paths containing transit-free ASNs be rejected? (Peerlock Lite)'" default:"false"`
	FilterPrefixLength *bool `yaml:"filter-prefix-length" description:"Should too large/small prefixes (IPv4: len > 24 || len < 8; IPv6: len > 48 || len < 12) be rejected?" default:"true"`

	AutoImportLimits *bool `yaml:"auto-import-limits" description:"Get import limits automatically from PeeringDB?" default:"false"`
	AutoASSet        *bool `yaml:"auto-as-set" description:"Get as-set automatically from PeeringDB?" default:"false"`

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
	OptimizerEnabled      *bool     `yaml:"optimize" description:"Should the optimizer be enabled for this peer?" default:"false"`
	OptimizeInbound       *bool     `yaml:"optimize-inbound" description:"Should the optimizer modify inbound policy?" default:"false"`
	OptimizeOutbound      *bool     `yaml:"optimize-outbound" description:"Should the optimizer modify outbound policy?" default:"false"`

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
}

type VRRPInstance struct {
	State     string   `yaml:"state" description:"VRRP instance state ('primary' or 'backup')" validate:"required"`
	Interface string   `yaml:"interface" description:"Interface to send VRRP packets on" validate:"required"`
	VRID      uint     `yaml:"vrid" description:"RFC3768 VRRP Virtual Router ID (1-255)" validate:"required"`
	Priority  uint     `yaml:"priority" description:"RFC3768 VRRP Priority" validate:"required"`
	VIPs      []string `yaml:"vips" description:"List of virtual IPs" validate:"required,cidr"`

	VIPs4 []string `yaml:"-" description:"-"`
	VIPs6 []string `yaml:"-" description:"-"`
}

type Augments struct {
	Accept4 []string          `yaml:"accept4" description:"List of BIRD protocols to import into the IPv4 table"`
	Accept6 []string          `yaml:"accept6" description:"List of BIRD protocols to import into the IPv6 table"`
	Reject4 []string          `yaml:"reject4" description:"List of BIRD protocols to not import into the IPv4 table"`
	Reject6 []string          `yaml:"reject6" description:"List of BIRD protocols to not import into the IPv6 table"`
	Statics map[string]string `yaml:"statics" description:"List of static routes to include in BIRD"`

	Statics4 map[string]string `yaml:"-" description:"-"`
	Statics6 map[string]string `yaml:"-" description:"-"`
}

type Optimizer struct {
	Targets     []string `yaml:"targets" description:"List of probe targets"`
	PingCount   int      `yaml:"probe-count" description:"Number of pings to send in each run"`
	PingTimeout int      `yaml:"probe-timeout" description:"Number of seconds to wait before considering the ICMP message unanswered"`
	Interval    int      `yaml:"probe-interval" description:"Time to wait between each optimizer run"`
	CacheSize   int      `yaml:"cache-size" description:"Number of probe results to store per peer"` // There will be a total of probeCacheSize*len(targets) results stored (TODO: double check if this is really true)

	Db map[string][]probe.Result `yaml:"-" description:"-"`
}

type Global struct {
	// Runtime
	BirdDirectory            string `yaml:"bird-directory" description:"Directory to store BIRD configs" default:"/etc/bird/"`
	BirdBinary               string `yaml:"bird-binary" description:"Path to bird binary" default:"/usr/sbin/bird"`
	CacheDirectory           string `yaml:"cache-directory" description:"Directory to store runtime configuration cache" default:"/var/run/pathvector/cache/"`
	BirdSocket               string `yaml:"bird-socket" description:"UNIX control socket for BIRD" default:"/run/bird/bird.ctl"`
	BirdSocketConnectTimeout uint   `yaml:"bird-socket-connect-timeout" description:"Timeout in seconds during BIRD socket connection" default:"5"`
	KeepalivedConfig         string `yaml:"keepalived-config" description:"Configuration file for keepalived" default:"/etc/keepalived.conf"`
	WebUIFile                string `yaml:"web-ui-file" description:"File to write web UI to (disabled if empty)"`
	PeeringDbQueryTimeout    uint   `yaml:"peeringdb-query-timeout" description:"PeeringDB query timeout in seconds" default:"10"`
	IRRQueryTimeout          uint   `yaml:"irr-query-timeout" description:"IRR query timeout in seconds" default:"30"`

	ASN              int      `yaml:"asn" description:"Autonomous System Number" validate:"required" default:"0"`
	Prefixes         []string `yaml:"prefixes" description:"List of prefixes to announce"`
	Communities      []string `yaml:"communities" description:"List of RFC1997 BGP communities"`
	LargeCommunities []string `yaml:"large-communities" description:"List of RFC8092 large BGP communities"`

	RouterID      string `yaml:"router-id" description:"Router ID (dotted quad notation)" validate:"required"`
	IRRServer     string `yaml:"irr-server" description:"Internet routing registry server" default:"rr.ntt.net"`
	RTRServer     string `yaml:"rtr-server" description:"RPKI-to-router server" default:"rtr.rpki.cloudflare.com:8282"`
	KeepFiltered  bool   `yaml:"keep-filtered" description:"Should filtered routes be kept in memory?" default:"false"`
	KernelLearn   bool   `yaml:"kernel-learn" description:"Inject Kernel routes into BIRD wich have been injected into the kernel from other deamons e.g FRR OSPFd" default:"false"`
	MergePaths    bool   `yaml:"merge-paths" description:"Should best and equivalent non-best routes be imported to build ECMP routes?" default:"false"`
	Source4       string `yaml:"source4" description:"Source IPv4 address"`
	Source6       string `yaml:"source6" description:"Source IPv6 address"`
	AcceptDefault bool   `yaml:"accept-default" description:"Should default routes be added to the bogon list?" default:"false"`
	KernelTable   int    `yaml:"kernel-table" description:"Kernel table"`

	Peers         map[string]*Peer `yaml:"peers" description:"BGP peer configuration"`
	Templates     map[string]*Peer `yaml:"templates" description:"BGP peer templates"`
	VRRPInstances []VRRPInstance   `yaml:"vrrp" description:"List of VRRP instances"`
	Augments      Augments         `yaml:"augments" description:"Custom configuration options"`
	Optimizer     Optimizer        `yaml:"optimizer" description:"Route optimizer options"`

	RTRServerHost string   `yaml:"-" description:"-"`
	RTRServerPort int      `yaml:"-" description:"-"`
	Prefixes4     []string `yaml:"-" description:"-"`
	Prefixes6     []string `yaml:"-" description:"-"`
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
func Load(configBlob []byte) (*Global, error) {
	var g Global
	// Set global config defaults
	if err := defaults.Set(&g); err != nil {
		return nil, err
	}

	if err := yaml.UnmarshalStrict(configBlob, &g); err != nil {
		return nil, errors.New("YAML unmarshal: " + err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(&g); err != nil {
		return nil, errors.New("Validation: " + err.Error())
	}

	// Check for invalid templates
	for templateName, templateData := range g.Templates {
		if templateData.Template != nil && *templateData.Template != "" {
			return nil, fmt.Errorf("Templates must not have a template field set, but %s does", templateName)
		}
	}

	for peerName, peerData := range g.Peers {
		if peerData.NeighborIPs == nil || len(*peerData.NeighborIPs) < 1 {
			return nil, fmt.Errorf("[%s] has no neighbors defined", peerName)
		}

		// Assign values from template
		if peerData.Template != nil && *peerData.Template != "" {
			template := g.Templates[*peerData.Template]
			if template == nil || &template == nil {
				return nil, fmt.Errorf("Template %s not found", *peerData.Template)
			}
			templateValue := reflect.ValueOf(*template)
			peerValue := reflect.ValueOf(g.Peers[peerName]).Elem()

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
		peerValue := reflect.ValueOf(g.Peers[peerName]).Elem()
		templateValueType := peerValue.Type()
		for i := 0; i < templateValueType.NumField(); i++ {
			fieldName := templateValueType.Field(i).Name
			fieldValue := peerValue.FieldByName(fieldName)
			defaultString := templateValueType.Field(i).Tag.Get("default")
			if defaultString == "" {
				return nil, fmt.Errorf("Code error: field %s has no default value", fieldName)
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
							return nil, fmt.Errorf("can't convert '%s' to uint", defaultString)
						}
						log.Debugf("[%s] setting field %s to value %+v", peerName, fieldName, defaultValueInt)
						fieldValue.Set(reflect.ValueOf(&defaultValueInt))
					case reflect.Bool:
						var err error // explicit declaration used to avoid scope issues of defaultValue
						defaultBool, err := strconv.ParseBool(defaultString)
						if err != nil {
							return nil, fmt.Errorf("Can't parse bool %s", defaultString)
						}
						log.Debugf("[%s] setting field %s to value %+v", peerName, fieldName, defaultBool)
						fieldValue.Set(reflect.ValueOf(&defaultBool))
					case reflect.Struct, reflect.Slice:
						// Ignore structs and slices
					default:
						return nil, fmt.Errorf("Unknown kind %+v for field %s", elemToSwitch, fieldName)
					}
				}
			} else {
				log.Debugf("[%s] skipping field %s with ingored default (-)", peerName, fieldName)
			}
		}
	}

	// Parse origin routes by assembling OriginIPv{4,6} lists by address family
	for _, prefix := range g.Prefixes {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("Invalid origin prefix: " + prefix)
		}

		if pfx.To4() == nil { // If IPv6
			g.Prefixes6 = append(g.Prefixes6, prefix)
		} else { // If IPv4
			g.Prefixes4 = append(g.Prefixes4, prefix)
		}
	}

	// Initialize static maps
	g.Augments.Statics4 = map[string]string{}
	g.Augments.Statics6 = map[string]string{}

	// Parse static routes
	for prefix, nexthop := range g.Augments.Statics {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("Invalid static prefix: " + prefix)
		}
		if net.ParseIP(nexthop) == nil {
			return nil, errors.New("Invalid static nexthop: " + nexthop)
		}

		if pfx.To4() == nil { // If IPv6
			g.Augments.Statics6[prefix] = nexthop
		} else { // If IPv4
			g.Augments.Statics4[prefix] = nexthop
		}
	}

	// Parse VRRP configs
	for _, vrrpInstance := range g.VRRPInstances {
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
	if g.RTRServer != "" {
		rtrServerParts := strings.Split(g.RTRServer, ":")
		if len(rtrServerParts) != 2 {
			return nil, fmt.Errorf("Invalid rtr-server '%s' format should be host:port", rtrServerParts)
		}
		g.RTRServerHost = rtrServerParts[0]
		rtrServerPort, err := strconv.Atoi(rtrServerParts[1])
		if err != nil {
			return nil, fmt.Errorf("Invalid RTR server port %s", rtrServerParts[1])
		}
		g.RTRServerPort = rtrServerPort
	}

	for _, peerData := range g.Peers {
		// Build static prefix filters
		if peerData.Prefixes != nil {
			for _, prefix := range *peerData.Prefixes {
				pfx, _, err := net.ParseCIDR(prefix)
				if err != nil {
					return nil, errors.New("Invalid prefix: " + prefix)
				}

				if pfx.To4() == nil { // If IPv6
					*peerData.PrefixSet4 = append(*peerData.PrefixSet4, prefix)
				} else { // If IPv4
					*peerData.PrefixSet6 = append(*peerData.PrefixSet6, prefix)
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

		// Check for no originated prefixes but announce-originated enabled
		if len(g.Prefixes) < 1 && *peerData.AnnounceOriginated {
			// No locally originated prefixes are defined, so there's nothing to originate
			*peerData.AnnounceOriginated = false
		}
	} // end peer loop

	return &g, nil // nil error
}

func cleanTypeString(in string) string {
	in = strings.Replace(in, "config.", "", -1)
	in = strings.Replace(in, "*", "", -1)
	return in
}

func documentConfigTypes(t reflect.Type) {
	fmt.Println("<!-- Code generated DO NOT EDIT -->")
	childTypesSet := map[reflect.Type]bool{}
	fmt.Println("## " + cleanTypeString(t.String()))
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
			fmt.Printf("| %s | %s | %s | %s | %s |\n", key, cleanTypeString(field.Type.String()), fDefault, validation, description)
		}
	}
	for childType := range childTypesSet {
		documentConfigTypes(childType)
	}
}

// PrintDocs generates configuration docs
func PrintDocs() {
	fmt.Println("# Options")
	documentConfigTypes(reflect.TypeOf(Global{}))
}
