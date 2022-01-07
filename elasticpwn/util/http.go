package EPUtils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type RetryReason int8

const (
	None             RetryReason = 0
	SHOULD_TRY_HTTPS RetryReason = 1
	SHOULD_TRY_HTTP  RetryReason = 2
)
const (
	HTTPS = "https"
	HTTP  = "http"
)

// reuse http client.
// if declared inside the function,
// the memory usage will spike, leading to a forceful exit
var httpClient = http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

func SendFailSafeHTTPRequest(endpoint string, timeoutsecs int, disableRetries bool, headers map[string]string, method string) (string, int, error) {
	var (
		err         error
		response    *http.Response
		retries     int         = 1
		retryReason RetryReason = 0
		cancelFuncs []context.CancelFunc
	)
	for retries > 0 {
		var maybeFixedEndpoint = func(retryMode RetryReason, endpoint string) string {
			switch retryReason {
			case SHOULD_TRY_HTTPS:
				EPLogger(fmt.Sprintf("Retrying with https because http did not work: %v", endpoint))
				firstHttpStringOccurenceRegex := regexp.MustCompile("^(.*?)http(.*)$")
				return firstHttpStringOccurenceRegex.ReplaceAllString(endpoint, HTTPS)
			case SHOULD_TRY_HTTP:
				EPLogger(fmt.Sprintf("Retrying with http because https did not work: %v", endpoint))
				firstHttpsStringOccurenceRegex := regexp.MustCompile("^(.*?)https(.*)$")
				return firstHttpsStringOccurenceRegex.ReplaceAllString(endpoint, HTTPS)
			default:
				return endpoint
			}
		}(retryReason, endpoint)
		var maybeFixedEndpointWithPrefix = func(endpoint string) string {
			if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
				return fmt.Sprintf("http://%s", endpoint)
			}

			return endpoint
		}(maybeFixedEndpoint)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutsecs)*time.Second)
		cancelFuncs = append(cancelFuncs, cancel)
		var req *http.Request

		if method != "GET" && method != "POST" {
			panic(fmt.Sprintf("%v is not an accepted http method", method))
		}
		req, err = http.NewRequestWithContext(ctx, method, maybeFixedEndpointWithPrefix, nil)
		if err != nil {
			EPLogger(fmt.Sprintf("Failed to send request to %v\n", maybeFixedEndpointWithPrefix))

			if disableRetries {
				break
			} else {
				retries -= 1
				continue
			}
		}
		for key, header := range headers {
			req.Header.Set(key, header)
		}
		response, err = httpClient.Do(req)

		if err != nil {
			EPLogger(fmt.Sprintf("Failed to fetch from %s\n", maybeFixedEndpointWithPrefix))
			if !disableRetries {
				EPLogger(fmt.Sprintf("Remaining retries for %s: %d times\n", maybeFixedEndpointWithPrefix, retries))
			}
			var errorMessage = fmt.Sprint(err)
			if strings.Contains(errorMessage, "server gave HTTP response to HTTPS client") {
				EPLogger("Retrying with HTTP instead of HTTPS")
				retryReason = SHOULD_TRY_HTTP
			} else if strings.Contains(errorMessage, "server gave HTTPS response to HTTP client") {
				EPLogger("Retrying with HTTPS instead of HTTP")
				retryReason = SHOULD_TRY_HTTPS
			}
			if disableRetries {
				break
			} else {
				retries -= 1
			}
		} else {
			break
		}
	}

	if response != nil {
		data, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		for _, c := range cancelFuncs {
			c()
		}

		if err != nil {
			return "", response.StatusCode, err
		}

		return string(data), response.StatusCode, nil
	}
	for _, c := range cancelFuncs {
		c()
	}

	return "", -1, err
}
