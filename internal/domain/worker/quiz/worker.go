package quiz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrToManyRequests = errors.New("too many requests")
)

var (
	quizUrl = "http://193.168.227.93"
)

type worker struct {
	client      *http.Client
	cookies     []*http.Cookie
	end         bool
	rateLimiter chan struct{}
}

func newWorker(rateLimiter chan struct{}) *worker {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Transport: &http.Transport{
			IdleConnTimeout:     30 * time.Second,
			MaxIdleConnsPerHost: 100,
		},
		Timeout: 10 * time.Second,
	}
	return &worker{client: client, rateLimiter: rateLimiter}
}

func (w *worker) doWork(ctx context.Context) (string, error) {
	var err error
	// set cookie
	for {
		err := w.setCookies(ctx, quizUrl+"/start")
		if errors.Is(err, ErrToManyRequests) {
			continue
		}
		if err != nil {
			return "", err
		}
		break
	}

	question := 1
	// get first question
	var content []byte
	for {
		urlReq, err := url.JoinPath(quizUrl, "question", strconv.Itoa(question))
		if err != nil {
			return "", err
		}
		resp, err := w.makeRequestWithCookies(ctx, http.MethodGet, urlReq)
		if errors.Is(err, ErrToManyRequests) {
			continue
		}
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		content, err = io.ReadAll(resp.Body)
		if err != nil {
			err := fmt.Errorf("error reading the response body: %s", err)
			return "", err
		}
		break
	}

	for !w.end {
		urlReq, err := url.JoinPath(quizUrl, "question", strconv.Itoa(question))
		if err != nil {
			return "", err
		}
		newContent, err := w.makeFillRequest(ctx, content, urlReq)
		if errors.Is(err, ErrToManyRequests) {
			continue
		}
		if err != nil {
			return "", err
		}
		content = newContent
		question++
	}

	return string(content), err
}

func (w *worker) makeRequestWithCookies(ctx context.Context, method, reqUrl string) (*http.Response, error) {
	// acquire rate limiter token
	w.rateLimiter <- struct{}{}
	defer func() { <-w.rateLimiter }() // release rate limiter token

	req, err := http.NewRequestWithContext(ctx, method, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	for _, cookie := range w.cookies {
		req.AddCookie(cookie)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		time.Sleep(1 * time.Second)
		return nil, ErrToManyRequests
	}
	return resp, err
}

func (w *worker) setCookies(ctx context.Context, reqUrl string) error {
	resp, err := w.makeRequestWithCookies(ctx, http.MethodGet, reqUrl)
	if err != nil {
		return err
	}

	// read cookies from nestedResponse because of redirection
	nestedResponse := resp.Request.Response
	cookies := nestedResponse.Cookies()
	w.cookies = cookies

	return nil
}

func (w *worker) makeFillRequest(ctx context.Context, htmlContent []byte, reqUrl string) ([]byte, error) {
	formData, err := parseAndFillForm(htmlContent)
	if err != nil {
		err := fmt.Errorf("error parsing the HTML form: %s", err)
		return nil, err
	}

	urlParse, err := url.Parse(reqUrl)
	if err != nil {
		err := fmt.Errorf("url.Parse: %s", err)
		return nil, err
	}
	parsedUrl := urlParse.Query()
	for key, value := range formData {
		parsedUrl.Add(key, value)
	}
	urlWithQueries := reqUrl + "?" + parsedUrl.Encode()

	resp, err := w.makeRequestWithCookies(ctx, http.MethodPost, urlWithQueries)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	nextHtml, err := io.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("error reading the response body: %s", err)
		return nil, err
	}

	hasError := w.hasErrorMessage(nextHtml)
	if hasError {
		err := fmt.Errorf("caught fill error")
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		nextQuestionNumber, err := w.findQuestionNumber(nextHtml)
		if err != nil {
			err := fmt.Errorf("error parsing the next question number: %s", err)
			return nil, err
		}
		if nextQuestionNumber == 0 {
			w.end = true
		}

		return nextHtml, err
	} else {
		err := fmt.Errorf("unhandled status code: %d", resp.StatusCode)
		return nil, err
	}
}

func (w *worker) findQuestionNumber(htmlContent []byte) (int, error) {
	re := regexp.MustCompile(`<title>Question (\d+) of \d+</title>`)
	matches := re.FindSubmatch(htmlContent)
	if matches == nil {
		return 0, nil
	}
	currentQuestionNumber := matches[1]
	return strconv.Atoi(string(currentQuestionNumber))
}

func (w *worker) hasErrorMessage(htmlContent []byte) bool {
	return strings.Contains(string(htmlContent), `<h3>error:`)
}
