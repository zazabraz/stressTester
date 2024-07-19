package quiz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	log            slog.Logger
	client         *http.Client
	cookies        []*http.Cookie
	end            bool
	questionNumber int
}

func newWorker(log slog.Logger) *worker {
	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 100,
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	return &worker{log: log, client: client, questionNumber: 1}
}

func (w *worker) setCookies(reqUrl string) error {
	resp, err := w.makeRequestWithCookies(http.MethodGet, reqUrl)
	if err != nil {
		return err
	}

	//read cookies from nestedResponse because of redirection
	nestedResponse := resp.Request.Response
	cookies := nestedResponse.Cookies()
	w.cookies = cookies

	return nil
}

func (w *worker) doWork(ctx context.Context) (string, error) {
	var err error
	//set cookie
	for {
		err := w.setCookies(quizUrl + "/start")
		if errors.Is(ErrToManyRequests, err) {
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			return "", err
		}
		break
	}

	//get first question
	var content []byte
	for {
		urlReq, err := url.JoinPath(quizUrl, "question", strconv.Itoa(w.questionNumber))
		if err != nil {
			return "", err
		}
		resp, err := w.makeRequestWithCookies(http.MethodGet, urlReq)
		if errors.Is(ErrToManyRequests, err) {
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		content, err = io.ReadAll(resp.Body)
		if err != nil {
			w.log.Error(fmt.Sprintf("error reading the response body: %s", err))
			return "", err
		}
		break
	}

	for !w.end {
		urlReq, err := url.JoinPath(quizUrl, "question", strconv.Itoa(w.questionNumber))
		if err != nil {
			return "", err
		}
		newContent, err := w.makeFillRequest(content, urlReq)
		if errors.Is(ErrToManyRequests, err) {
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			return "", err
		}
		content = newContent
	}

	return string(content), err
}

func (w *worker) makeRequestWithCookies(method, reqUrl string) (*http.Response, error) {
	req, err := http.NewRequest(method, reqUrl, nil)
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
		return nil, ErrToManyRequests
	}
	return resp, err
}

func (w *worker) makeFillRequest(htmlContent []byte, reqUrl string) ([]byte, error) {
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

	resp, err := w.makeRequestWithCookies(http.MethodPost, urlWithQueries)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	nextHtml, err := io.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("error reading the response body: %s", err)
		return nil, err
	}

	hasError := hasErrorMessage(nextHtml)
	if hasError {
		err := fmt.Errorf("caught fill error")
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		nextQuestionNumber, err := nextNumber(nextHtml)
		if err != nil {
			err := fmt.Errorf("error parsing the next question number: %s", err)
			return nil, err
		}
		if nextQuestionNumber == 0 {
			w.end = true
		}
		w.questionNumber = nextQuestionNumber

		return nextHtml, err
	} else {
		err := fmt.Errorf("unhandled ststus code: %d", resp.StatusCode)
		return nil, err
	}
}

func nextNumber(htmlContent []byte) (int, error) {
	re := regexp.MustCompile(`<title>Question (\d+) of \d+</title>`)
	matches := re.FindSubmatch(htmlContent)
	if matches == nil {
		return 0, nil
	}
	currentQuestionNumber := matches[1]
	return strconv.Atoi(string(currentQuestionNumber))
}

func hasErrorMessage(htmlContent []byte) bool {
	return strings.Contains(string(htmlContent), `<h3>error:`)
}
