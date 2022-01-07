package EPUtils

import (
	"log"
	"strings"
	"testing"
)

// @todo: add actual/expected vars
func TestEmailRegex(t *testing.T) {
	testString := `
		testtestasfewasfaewf@gmail.com
		url.url.url.com
		a1241242.141241241241249012j40129j 12421poj12p9j12pj412poj4 1op2j4po12 j4po12j po12j 09234

		\\\214j092yth0q234ht80q2h340fh24f0h23fwfhi@gmail.comf243oif24iof
		flkawejkfjewkfnklwenfi3o2nfi 3f3nm ifn 3nmm10p92m90fn90 n0n nwalknmvklnsdsfsfasd fdsafsdalkf jlaskdjfklasd jklwelkfklefalkweflkwe jlkawlkj lkj test@hotmail.com test+01@hotmail.com 1231212312312 3213123
	`

	allEmails := EmailRegex.FindAllString(testString, -1)

	if len(allEmails) < 1 {
		t.Errorf("Did not find any emails")
	} else {
		log.Printf("Found these emails: %v\n", strings.Join(allEmails, ","))
	}
}

// @todo: add actual/expected vars
func TestUrlRegex(t *testing.T) {
	testString := `
		testtestasfewasfaewf@gmail.com
		url.url.url.com
		a1241242.141241241241249012j40129j 12421poj12p9j12pj412poj4 1op2j4po12 j4po12j po12j 09234

		https://example.com

		example.com

		\\\214j092yth0q234ht80q2https://2222.comh340fh24f0h23fwfhi@gmail.comf243oif24iof
		flkawejkfje example.com https://qqqqqqq.com  wkfnklwenfi3o2nfi 3f3nm ifn 3nmm10p92m90fn90 n0n nwalknmvklnsdsfsfasd fdsafsdalkf jlaskdjfklasd jklwelkfklefalkweflkwe jlkawlkj lkj test@hotmail.com test+01@hotmail.com 1231212312312 3213123 https://qqqqqqq.com https://qqqqqqq.com
	`

	allUrls := FindAllUniqueUrls(testString)

	if len(allUrls) < 1 {
		t.Errorf("Did not find any urls")
	} else {
		log.Printf("Found these urls: %v\n", strings.Join(allUrls, ","))
	}
}

// only ipv4 for now
func TestFindAllUniquePublicIps(t *testing.T) {
	expectedPublicIps := []string{
		"123.123.123.123:500",
		"123.123.124.14",
		"200.200.200.200:12414",
		"200.200.200.200",
	}
	privateIps := []string{
		"192.168.2.1",
		"192.168.10.2",
		"10.0.0.0",
		"10.255.255.255",
		"172.16.0.0",
		"172.30.222.222",
	}

	rawString := `testtestasfewasfaewf@gmail.com
	url.url.url.com
	a1241242.141241241241249012j40129j 12421poj12p9j12pj412poj4 1op2j4po12 j4po12j po12j 09234

	https://example.com

	example.com

	\\\214j092yth0q234ht80q2https://2222.comh340fh24f0h23fwfhi@gmail.comf243oif24iof
	flkawejkfje example.com https://qqqqqqq.com  wkfnklwenfi3o2nfi 3f3nm ifn 3nmm10p92m90fn90 n0n nwalknmvklnsdsfsfasd fdsafsdalkf jlaskdjfklasd jklwelkfklefalkweflkwe jlkawlkj lkj test@hotmail.com test+01@hotmail.com 1231212312312 3213123


	

	` + expectedPublicIps[0] + `
	999.999.999.999
	asdfasdf999.999.999.999
	100101001101010.100.0.10asdfasdf


	100101001101010.100.0.10
	` + expectedPublicIps[1] + `
		"WEF" 12412094571209571.12512.5125091u012740284091280941.325.1254123.5213.36624624.74,324p4350234850923841234023842343oirh132109h 0912hr9012h sadilkvcdsalkvnsakvl19 1890384019284091 490902 1029209aweljfkweajlk wajlkej
	` + expectedPublicIps[2] + `
		sakldjflawejiofjaewiofjweiofjwejfoiaweawefoiawejfoiawheoifh
	` + expectedPublicIps[3] + "   " + strings.Join(privateIps, "  aaaa ")

	actualPublicIps := FindAllUniquePublicIps(rawString)

	for _, ip := range actualPublicIps {
		if Contains(ip, privateIps) != -1 {
			t.Errorf("%s is not a public ip but was considered to be so", ip)
		} else if Contains(ip, expectedPublicIps) == -1 {
			t.Errorf("%s is a public ip but was not considered to be so", ip)
		}
	}
	log.Printf("Found these public ipv4 urls: %s", strings.Join(actualPublicIps, ","))
}
