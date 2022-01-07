package EPLookup_addrs

import (
	"regexp"

	EPUtils "github.com/9oelM/elasticpwn/elasticpwn/util"
)

var EXCLUDE_SSL_ORGS = []string{
	"Digicert",
	"DigiCert",
	"Sectigo",
	"SECTIGO",
	"Let's Encrypt",
	"GlobalSign",
	"Amazon",
	"USERTRUST",
	"Internet Security Research Group",
	"Google Trust Services",
	"IdenTrust",
	"GoDaddy.com",
	"The Go Daddy Group",
	"Starfield Technologies",
	"Comodo",
	"COMODO",
	"Acme",
	"ACME",
}

var NOT_REALLY_INTERESTING_DOMAINS = []string{
	"amazonaws.com.",
	"googleusercontent.com.",
	"linode.com.",
	"awsglobalaccelerator.com.",
	"vultr.com.",
	"fios.verizon.net.",
}

var IpWithOptionalPortRegex = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}(:[0-9]+)?`)
var UrlRegex = regexp.MustCompile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

func CreateMaybeVerboseEPLogger(isVerbose bool) func(string) {
	if isVerbose {
		return EPUtils.EPLogger
	} else {
		return func(a string) {}
	}
}
