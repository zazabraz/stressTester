package quiz

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"stress-tester/internal/domain/worker"
	"strings"
	"time"
)

type quiz struct {
	log            slog.Logger
	client         *http.Client
	cookies        []*http.Cookie
	end            bool
	questionNumber int
}

var (
	quizUrl = "http://193.168.227.93"
)

func (q *quiz) setCookies(reqUrl string) error {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return err
	}
	resp, err := q.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//read cookies from nestedResponse because of redirection
	nestedResponse := resp.Request.Response
	cookies := nestedResponse.Cookies()
	q.cookies = cookies

	return nil
}

func New(log slog.Logger) worker.Worker {
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
	return &quiz{client: client, log: log, questionNumber: 1, end: false}
}

func (q *quiz) DoWork(ctx context.Context) error {
	err := q.setCookies(quizUrl + "/start")
	if err != nil {
		return err
	}

	for !q.end {
		urlReq, err := url.JoinPath(quizUrl, "question", strconv.Itoa(q.questionNumber))
		if err != nil {
			return err
		}
		err = q.makeFillRequest(urlReq)
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (q *quiz) makeRequestWithCookies(method, reqUrl string) (*http.Response, error) {
	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	for _, cookie := range q.cookies {
		req.AddCookie(cookie)
	}
	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (q *quiz) makeFillRequest(reqUrl string) error {
	resp, err := q.makeRequestWithCookies(http.MethodPost, reqUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		q.log.Error(fmt.Sprintf("Error reading the response body: %s", err))
		return err
	}

	formData, err := parseAndFillForm(htmlContent)
	if err != nil {
		q.log.Error(fmt.Sprintf("Error parsing the HTML form: %s", err))
		return err
	}

	urlParse, err := url.Parse(reqUrl)
	if err != nil {
		err := fmt.Errorf("url.Parse: %s", err)
		q.log.Error(err.Error())
		return err
	}
	qq := urlParse.Query()
	for key, value := range formData {
		qq.Add(key, value)
	}
	urlWithQueries := reqUrl + "?" + qq.Encode()

	resp, err = q.makeRequestWithCookies(http.MethodPost, urlWithQueries)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	htmlContent2, err := io.ReadAll(resp.Body)
	if err != nil {
		q.log.Error(fmt.Sprintf("Error reading the response body: %s", err))
		return err
	}

	hasError := hasErrorMessage(htmlContent2)
	if hasError {
		return fmt.Errorf("caught fill error")
	}

	q.log.Info("resp.StatusCode", resp.StatusCode)
	switch resp.StatusCode {
	case http.StatusOK:
		nextQuestionNumber, err := nextNumber(htmlContent2)
		if err != nil {
			return err
		}
		if nextQuestionNumber == 0 {
			q.end = true
		}
		q.questionNumber = nextQuestionNumber

		return nil
	case http.StatusTooManyRequests:
		time.Sleep(1 * time.Second)
		err := q.makeFillRequest(reqUrl)
		if err != nil {
			return err
		}
	}

	return nil
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
