package EPLookup_addrs

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	EPUtils "github.com/9oelm/elasticpwn/core/util"
)

/*
 Common name != Subject name

 https://stackoverflow.com/questions/5935369/how-do-common-names-cn-and-subject-alternative-names-san-work-together/29600674

 Important: origin needs correct port number to access HTTP(S) service.

 Example: getSslCertificateInfo("https://google.com:443")

 Default timeout: 10 secs
*/
func GetSslCertificateInfo(origin string) (maybeValidUrls string, maybeValidOrgs string) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := tls.DialWithDialer(
		&net.Dialer{
			Timeout: time.Second * 5,
		},
		"tcp",
		origin,
		tlsConf,
	)
	if err != nil {
		return "", ""
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	for _, cert := range certs {
		// also will match URL inside a wildcard domain, like *.example.com -> .example.com
		maybeValidUrl := UrlRegex.FindString(cert.Subject.CommonName)
		// prevent dups
		if maybeValidUrl != "" && !strings.Contains(maybeValidUrls, maybeValidUrl) {
			maybeValidUrls += strings.TrimPrefix(maybeValidUrl, ".") + ","
		}
		// prevent dups
		for _, org := range EPUtils.Unique(cert.Subject.Organization) {
			if EPUtils.Contains(strings.TrimSpace(org), EXCLUDE_SSL_ORGS) == -1 {
				maybeValidOrgs += strings.TrimSpace(org) + ","
			}
		}
	}
	return strings.TrimSuffix(maybeValidUrls, ","), strings.TrimSuffix(maybeValidOrgs, ",")
}

func GetIpInfo(ipWithMaybePortNum string) (string, string, string, string) {
	ipSplit := strings.Split(ipWithMaybePortNum, ":")
	ip := ipSplit[0]

	wg := sync.WaitGroup{}
	subjectUrlsChan := make(chan string)
	organizationsChan := make(chan string)
	cloudHostingProvidersChan := make(chan string)
	cnameChan := make(chan string)
	wg.Add(1)
	go func(cloudHostingProvidersChan chan string) {
		defer wg.Done()
		validDomains, err := net.LookupAddr(ip)
		if err == nil {
			for _, domain := range validDomains {
				if EPUtils.ContainsEndsWith(domain, NOT_REALLY_INTERESTING_DOMAINS) == -1 {
					result := fmt.Sprintf("%s,%s\n", ipWithMaybePortNum, domain)
					cloudHostingProvidersChan <- result
					return
				} else {
					cloudHostingProvidersChan <- ""
				}
			}
		} else {
			// silently ignore error, otherwise there will be too many error outputs
			cloudHostingProvidersChan <- ""
		}
	}(cloudHostingProvidersChan)
	wg.Add(1)
	go func(organizationsChan chan string, subjectUrlsChan chan string) {
		defer wg.Done()
		subjectUrls, organizations := GetSslCertificateInfo(ipWithMaybePortNum)
		if subjectUrls != "" || organizations != "" {
			subjectUrlsChan <- subjectUrls
			organizationsChan <- organizations
		} else {
			subjectUrlsChan <- ""
			organizationsChan <- ""
		}
	}(organizationsChan, subjectUrlsChan)
	wg.Add(1)
	go func(cnameChan chan string) {
		defer wg.Done()
		cname, err := net.LookupCNAME(ip)

		if cname != "" || err == nil {
			fmt.Println(cname)
			cnameChan <- cname
		} else {
			cnameChan <- ""
		}
	}(cnameChan)

	cloudHostingProvider, subjectUrls, organizations, cname := <-cloudHostingProvidersChan, <-subjectUrlsChan, <-organizationsChan, <-cnameChan
	wg.Wait()

	return cloudHostingProvider, subjectUrls, organizations, cname
}

func parseFlagsAndInit() (inputFilePath *string, outputFilePath *string, numThreads *int, logger func(string)) {
	flagSet := flag.NewFlagSet("lookup-addrs", flag.ContinueOnError)
	inputFilePath = flagSet.String("inputFilePath", "./urls.txt", "Path to the text file that contains list of URLs")
	outputFilePath = flagSet.String("outputFilePath", "./out.csv", "Path to output CSV file")
	numThreads = flagSet.Int("threads", 20, "Number of threads to use")
	isVerbose := flagSet.Bool("verbose", true, "Verbosity")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Println("Failed to parse flags.")
		fmt.Println(`Example: 
./lookup-addrs -threads=30 -inputFilePath=input.txt -outputFilePath=output.csv -verbose=false`)
		fmt.Println("Note that the boolean flag should be fed as -isVerbose=false. -isVerbose false won't get it to work.")
		panic(err)
	}
	logger = CreateMaybeVerboseEPLogger(*isVerbose)

	if flagSet.Parsed() {
		logger(fmt.Sprintf("Received inputFilePath: %s, outputFilePath: %s, verbose: %v, threads: %v", *inputFilePath, *outputFilePath, *isVerbose, *numThreads))
	} else {
		panic("Flags were not parsed")
	}

	return
}

func main() {
	inputFilePath, outputFilePath, numThreads, logger := parseFlagsAndInit()
	urls := EPUtils.ReadUrlsFromFile(*inputFilePath)
	allIps := IpWithOptionalPortRegex.FindAllString(urls, -1)

	if allIps == nil {
		logger(fmt.Sprintf("No valid IPs found from file %s. Check again.", *inputFilePath))
		return
	}

	concurrentGoroutines := make(chan struct{}, *numThreads)
	f, err := os.Create(*outputFilePath)
	EPUtils.ExitOnError(err)
	defer f.Close()

	var wg sync.WaitGroup
	for _, ipWithMaybePortNum := range allIps {
		wg.Add(1)
		go func(ipWithMaybePortNum string, f *os.File) {
			defer wg.Done()

			concurrentGoroutines <- struct{}{}

			cloudHostingProvider, subjectUrls, organization, cname := GetIpInfo(ipWithMaybePortNum)
			result := strings.ReplaceAll(fmt.Sprintf("%s,%s,%s,%s,%s", ipWithMaybePortNum, cloudHostingProvider, subjectUrls, organization, cname), "\n", "")
			if cloudHostingProvider != "" || subjectUrls != "" || organization != "" || cname != "" {
				EPUtils.EPLogger(result)
			}
			_, err = fmt.Fprintf(f, string(result)+"\n")
			<-concurrentGoroutines
		}(ipWithMaybePortNum, f)
	}
	wg.Wait()
	logger("Finished lookup of addresses")
}
