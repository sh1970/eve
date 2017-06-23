// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

package types

import (
	"github.com/satori/go.uuid"
	"net"
	"time"
)

// Database entry for onboarding certificate.
// Provides a secure binding to the username under which it is registered.
// Can be used to register RemainingUse different devices
type OnboardingCert struct {
	Cert         []byte
	UserName     string
	RegTime      time.Time
	RemainingUse int
	LastUsedTime time.Time
}

// Message payload for json to POST to /rest/self-register
// Certificate is the PEM encoded device certificate.
// The TLS exchange needs to be done using a registered Onboarding Certificate
type RegisterCreate struct {
	PemCert []byte
}

// Device database record. Also used for the json GET response for /rest/device
// TBD separate DeviceConfig i.e. Redirect and below from earlier fields?
type DeviceDb struct {
	DeviceCert        []byte
	DevicePublicKey   []byte
	UserName          string
	RegTime           time.Time
	ReRegisteredCount int
	ReadTime          time.Time
	// Redirect parameters; XXX should we send redirect response code?
	Redirect         bool
	RedirectToServer string
	RedirectRootCert []byte
	// LISP parameters; safe to have DNSname? Or list of IPs and credential
	// strings.
	LispMapServers []LispServerInfo
	LispInstance   uint32
	EID            net.IP
	EIDHashLen     uint8
	ZedServers     ZedServerConfig
	ClientAddr     string // To detect NATs
}

type LispServerInfo struct {
	NameOrIp   string
	Credential string
}

// XXX rename to DeviceHwResources
// XXX also add the measure free memory and storage as DeviceHwStatus
// That new DeviceHwStatus should include the underlay information (set of IPv4, and IPv6 addresses)
type DeviceHwStatus struct {
	// XXX add timestamp? for last update? when sent?
	Manufacturer string // Optionally set in manufacturing
	Model        string // Optionally set in manufacturing
	Serial       string // Optionally set in manufacturing
	Machine      string // From uname -m
	Processor    string // From uname -p
	Platform     string // From uname -i
	Compatible   string // From device-tree's compatible node
	Cpus         uint   // nproc --all
	Memory       uint   // Total memory in Kbyte
	Storage      uint   // Total flash in Kbyte
}

type DeviceSwConfig struct {
	// XXX add timestamp for last update? When sent?
	// XXX add hash for merkel tree
	ApplicationConfig []SwConfig
}

type DeviceSwStatus struct {
	// XXX add timestamp for last update? When sent?
	// XXX add lastReceivedHash and currentHash for merkel tree
	ApplicationStatus []SwStatus
}

// Actual state of sofware on device. Flows from device to ZedCloud.
// Includes all software; applications and Zededa infrastructure
// Need to restucture, since a given EID/DisplayName can have multiple versions.
// Ditto for SwConfig.
type SwStatus struct {
	// XXX add lastReceivedHash and currentHash for merkel tree
	Infra       bool   // Set for Zededa software which does not have an EID
	EID         net.IP // If one assigned. UUID alternative?
	DisplayName string
	Version     string
	Description string // optional
	State       SwState
	Activated   bool
}

// Intended state of sofware on device. Flows from ZedCloud to device.
// Includes all software; applications and Zededa infrastructure
// Note that SwConfig might make private+cert, plus EID, or allow EID generation
// Does that mean we need a cert in SwStatus? Or separate out EID allocation?
// SwConfig would have a Url, DigestAlg, and Digest as well.
type SwConfig struct {
	// XXX add hash for merkel tree
	Infra       bool   // Set for Zededa software which does not have an EID
	EID         net.IP // If one assigned. UUID alternative?
	DisplayName string
	Version     string
	Description string // optional
	State       SwState
	Activate    bool
}

// Enum names from OMA-TS-LWM2M_SwMgmt-V1_0-20151201-C
type SwState uint8

const (
	INITIAL          SwState = iota + 1
	DOWNLOAD_STARTED         // Really download in progress
	DOWNLOADED
	DELIVERED // Package integrity verified
	INSTALLED // Available to be activated
)

// Part of config handed to the device.
// The EIDs in the overlay to which it should connect.
// Think of this as /etc/hosts for the ZedManager - maps from names such as
// "zedcontrol" amd "zedlake0" to EIDs in the management overlay.
type ZedServerConfig struct {
	NameToEidList []NameToEid
}

type NameToEid struct {
	HostName string
	EIDs     []net.IP
}

// UUID plus version
type UUIDandVersion struct {
	UUID    uuid.UUID
	Version string
}

type OverlayNetwork struct {
	IID		uint32
	EID		net.IP
	Signature	string
	// Any additional LISP parameters?
	ACLs		[]ACE
	NameToEidList	[]NameToEid	// Used to populate DNS for the overlay
	// Only used in Status XXX create Status variant
	VifInfo
}

type UnderlayNetwork struct {
	ACLs		[]ACE
	// Only used in Status XXX create Status variant
	VifInfo
}

// Similar support as in draft-ietf-netmod-acl-model
type ACE struct {
	Matches []ACEMatch
	Actions []ACEAction
}

// The Type can be "ip" or "host" (aka domain name) for now. Matches remote.
// For now these are bidirectional.
// The host matching is suffix-matching thus zededa.net matches *.zededa.net.
// Can envision adding "protocol", "fport", "lport", and directionality at least
// Value is always a string.
// There is an implicit reject rule at the end.
// The "eidset" type is special for the overlay. Matches all the EID which
// are part of the NameToEidList.
type ACEMatch struct {
	Type string
	Value string     	
}

type ACEAction struct {
	Drop		bool	// Otherwise accept
	Limit		bool	// Is limiter enabled?
	LimitRate	int	// Packets per unit
	LimitUnit	string	// "s", "m", "h", for second, minute, hour
	LimitBurst	int	// Packets
}

// Indexed by UUID
// If IsZedmanager is set we do not create boN but instead configure the EID
// locally. This will go away once ZedManager runs in a domU like any
// application.
type AppNetworkConfig struct {
	UUIDandVersion      UUIDandVersion
	DisplayName         string
	IsZedmanager        bool
	OverlayNetworkList  []OverlayNetwork
	UnderlayNetworkList []UnderlayNetwork
}

// Indexed by UUID
type AppNetworkStatus struct {
	UUIDandVersion UUIDandVersion
	AppNum         int
	PendingAdd     bool
	PendingModify  bool
	PendingDelete  bool
	UlNum          int // Number of underlay interfaces
	OlNum          int // Number of overlay interfaces
	DisplayName    string
	// Copy from the AppNetworkConfig; used to delete when config is gone.
	IsZedmanager        bool
	OverlayNetworkList  []OverlayNetwork
	UnderlayNetworkList []UnderlayNetwork
}

// Do we want a DeviceNetworkStatus? DeviceNetworkConfig with the underlay
// interfaces?
type DeviceNetworkConfig struct {
	Uplink string // ifname; should have multiple
	// XXX WiFi credentials?? Should already be set?
}

type DeviceNetworkStatus struct {
	Uplink          string // ifname; should have multiple
	// XXX add all the uplink ifaddrs?
	// XXX uplink publicAddr to determine NATed?
}
