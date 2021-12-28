package EPUtils

import (
	"regexp"
)

var UrlRegex = regexp.MustCompile(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)?`)

// it will match a package name with a hash too, like org.springframework.security.web.authentication.logout.LogoutFilter@7b1e5e55
// although the purpose was to match an email, it is still a good info, so leave it as it is
var EmailRegex = regexp.MustCompile("[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*")
var IpWithOptionalPortRegex = regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}(:[0-9]+)?`)
var PrivateIpRegexWithOptionalPortRegex = regexp.MustCompile(`^(127\.)|(192\.168\.)|(10\.)|(172\.1[6-9]\.)|(172\.2[0-9]\.)|(172\.3[0-1]\.)|(::1$)|([fF][cCdD])(:[0-9]+)?`)

// things like 3.0.0 or org.sonatype.sisu (possibly package names)
var AtLeastTwoDotsInSentenceRegex = regexp.MustCompile(`(\w+)\.(\w+)\.\w*`)

func FindAllUniquePublicIps(rawString string) []string {
	allIps := IpWithOptionalPortRegex.FindAllString(rawString, -1)
	var allPublicIps []string

	for _, ip := range allIps {
		if len(PrivateIpRegexWithOptionalPortRegex.FindAllString(ip, -1)) == 0 {
			allPublicIps = append(allPublicIps, ip)
		}
	}

	return Unique(allPublicIps)
}

func FindAllUniqueUrls(rawString string) []string {
	allUrls := UrlRegex.FindAllString(rawString, -1)
	var allUniqueUrls []string

	for _, url := range allUrls {
		// because email addr also gets mistakenly considered as url
		if len(EmailRegex.FindAllString(url, -1)) == 0 {
			allUniqueUrls = append(allUniqueUrls, url)
		}
	}

	return Unique(allUniqueUrls)
}
