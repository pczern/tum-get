package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const (
	authURL    = "https://www.moodle.tum.de/Shibboleth.sso/Login?providerId=https://tumidp.lrz.de/idp/shibboleth&target=https://www.moodle.tum.de/auth/shibboleth/index.php"
	idpBaseURL = "https://login.tum.de"
)

var (
	baseHeaders = map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	}

	additionalHeaders = map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Accept":       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
	}
)

func Auth(username string, password string) (*http.Client, error) {
	fmt.Println("Authenticating")
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println("Error creating cookie jar:", err)
		return nil, err
	}

	requests := 0
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			requests++

			if requests == 1 || requests == 2 || requests >= 5 {
				// fmt.Println("Not Redirected to:", req.URL.Path)
				return http.ErrUseLastResponse
			} else {
				// fmt.Println("Redirected to:", req.URL.Path)
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	for k, v := range baseHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		fmt.Println("Error: status code is not 302", resp.StatusCode)
		return nil, err
	}

	// Get SAML URL from Location header
	samlURL := resp.Header.Get("Location")
	if samlURL == "" {
		fmt.Println("Error: could not find SAML URL")
		return nil, err
	}

	req, err = http.NewRequest("GET", samlURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	for k, v := range baseHeaders {
		req.Header.Set(k, v)
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Get SSO URL from Location header
	ssoURL := resp.Header.Get("Location")
	if ssoURL == "" {
		fmt.Println("Error: could not find SSO URL")
		return nil, err
	}

	ssoPostUrl := idpBaseURL + ssoURL

	req, err = http.NewRequest("POST", ssoPostUrl, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	for k, v := range additionalHeaders {
		req.Header.Set(k, v)
	}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse HTML response
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return nil, err
	}

	// Find CSRF token
	var csrfToken string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var isCsrfToken bool
			var value string
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "csrf_token" {
					isCsrfToken = true
				}
				if a.Key == "value" {
					value = a.Val
				}
			}
			if isCsrfToken {
				csrfToken = value
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if csrfToken == "" {
		fmt.Println("Error: could not find CSRF token")
		return nil, err
	}

	// Make POST request to SSO URL
	data := url.Values{}
	data.Set("csrf_token", csrfToken)
	data.Set("j_username", username)
	data.Set("j_password", password)
	data.Set("donotcache", "1")
	data.Set("_eventId_proceed", "")

	req, err = http.NewRequest("POST", ssoPostUrl, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	for k, v := range additionalHeaders {
		req.Header.Set(k, v)
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: status code is not 200 but ", resp.StatusCode)
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	bodyString := string(bodyBytes)

	doc1, err := goquery.NewDocumentFromReader(strings.NewReader(bodyString))
	if err != nil {
		fmt.Println("Error loading HTTP response body. ", err)
	}

	actionURL, ssoData := findSSOData(doc1)

	data = url.Values{}
	for key, value := range ssoData {
		data.Set(key, value)
	}

	req, err = http.NewRequest("POST", actionURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println("Error creating HTTP request. ", err)
	}

	for key, value := range additionalHeaders {
		req.Header.Add(key, value)
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request. ", err)
	}

	defer resp.Body.Close()

	return client, nil

}

func findSSOData(doc *goquery.Document) (string, map[string]string) {
	form := doc.Find("form").First()
	actionURL, _ := form.Attr("action")

	formDiv := form.Find("div").First()
	if formDiv.Length() == 0 {
		return "", nil
	}

	ssoHeaders := make(map[string]string)
	formDiv.Find("input").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		ssoHeaders[name] = value
	})

	return actionURL, ssoHeaders
}
