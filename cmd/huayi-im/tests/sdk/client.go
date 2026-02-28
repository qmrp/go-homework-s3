package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/model"
)

// SDK is the client interface
type SDK interface {
	Guest() UserClient
	// Healthz performs health check
	Healthz() error
}

type UserClient interface {
	SDK

	Login(username string) (response.UserDetailResponse, error)
	Logout() error
	User() UserAPI
	Topic() TopicAPI
	Message() MessageAPI
}

type UserAPI interface {
	// GetAllUsers 获取所有用户
	GetUsers() (response.UserListResponse, error)
	// GetUserDetail 获取用户详细信息
	GetUserDetail(username string) (model.User, error)
}

type TopicAPI interface {
	CreateTopic(topicName string) error
	JoinTopic(topicName string) error
	LeaveTopic(topicName string) error
	DeleteTopic(topicName string) error
	GetTopics() (response.TopicListResponse, error)
}

type MessageAPI interface {
	// SendMessage 发送消息到指定主题
	SendMessage(message string, topicName *string, to *[]string) error
}

var once sync.Once
var globalSDK SDK

// NewSDK creates a new SDK instance and sets it as global singleton.
// It can be only init onece.
func NewSDK(addr string) SDK {
	once.Do(func() {
		initSDK(addr)
	})
	return globalSDK
}

func initSDK(addr string) {
	baseURL, _ := url.Parse(strings.TrimRight(addr, "/"))
	globalSDK = &sdk{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// GetSDK returns the global SDK singleton
func GetSDK() SDK {
	if globalSDK == nil {
		panic("SDK not initialized, call NewSDK first")
	}
	return globalSDK
}

type sdk struct {
	baseURL *url.URL
	client  *http.Client
	name    string
	token   string
}

func (s *sdk) Guest() UserClient {
	return &sdk{
		baseURL: s.baseURL,
		client:  new(http.Client),
		name:    "guest",
		token:   "",
	}
}

func (s *sdk) setWithToken(token string, username string) UserClient {
	s.token = token
	s.name = username
	return s
}

func (s *sdk) Login(username string) (response.UserDetailResponse, error) {
	resp, err := doRequest[response.UserDetailResponse](s, http.MethodPost, "/api/login", map[string]string{
		"username": username,
	}) //{"sid":"Zhou_1770282010201069000","username":"Zhou"}
	if err != nil {
		return response.UserDetailResponse{}, err
	}
	log.Printf("loginResp: %v", *resp)
	s.setWithToken(resp.Sid, resp.Username)
	return *resp, nil
}

func (s *sdk) Logout() error {
	_, err := doRequest[struct{}](s, http.MethodPost, "/api/logout", nil)
	return err
}

func (s *sdk) Healthz() error {
	_, err := doRequest[struct{}](s, http.MethodGet, "/api/healthz", nil)
	return err
}

type userAPI struct {
	sdk *sdk
}

func (s *sdk) User() UserAPI {
	return &userAPI{
		sdk: s,
	}
}

func (u *userAPI) GetUsers() (response.UserListResponse, error) {
	list, err := doRequest[response.UserListResponse](u.sdk, http.MethodGet, "/api/users", nil)
	if err != nil {
		return response.UserListResponse{}, err
	}
	return *list, nil
}

func (u *userAPI) GetUserDetail(username string) (model.User, error) {
	user, err := doRequest[model.User](u.sdk, http.MethodGet, fmt.Sprintf("/api/users/%s", username), nil)
	if err != nil {
		return model.User{}, err
	}
	return *user, nil
}

type topicAPI struct {
	sdk *sdk
}

func (s *sdk) Topic() TopicAPI {
	return &topicAPI{
		sdk: s,
	}
}

func (t *topicAPI) CreateTopic(topicName string) error {
	_, err := doRequest[struct{}](t.sdk, http.MethodPost, "/api/topics", map[string]string{
		"topic": topicName,
	})
	return err
}

func (t *topicAPI) JoinTopic(topicName string) error {
	_, err := doRequest[struct{}](t.sdk, http.MethodPost, fmt.Sprintf("/api/topics/%s/actions/join", topicName), nil)
	return err
}

func (t *topicAPI) LeaveTopic(topicName string) error {
	_, err := doRequest[struct{}](t.sdk, http.MethodPost, fmt.Sprintf("/api/topics/%s/actions/quit", topicName), nil)
	return err
}

func (t *topicAPI) DeleteTopic(topicName string) error {
	_, err := doRequest[struct{}](t.sdk, http.MethodDelete, fmt.Sprintf("/api/topics/%s", topicName), nil)
	return err
}

func (t *topicAPI) GetTopics() (response.TopicListResponse, error) {
	list, err := doRequest[response.TopicListResponse](t.sdk, http.MethodGet, "/api/topics", nil)
	if err != nil {
		return response.TopicListResponse{}, err
	}
	return *list, nil
}

type messageAPI struct {
	sdk *sdk
}

func (s *sdk) Message() MessageAPI {
	return &messageAPI{
		sdk: s,
	}
}

func (m *messageAPI) SendMessage(message string, topicName *string, to *[]string) error {
	if to == nil && topicName == nil {
		return errors.New("to or topic name must be specified")
	}

	msg := map[string]any{
		"message-type": "message",
		"from":         m.sdk.name,
		"content":      message,
		"to":           *to,
	}

	if topicName != nil {
		msg["topic"] = *topicName
	}

	_, err := doRequest[struct{}](m.sdk, http.MethodPost, "/api/messages", msg)
	return err
}

// Error is the error returned by the dashboard.
type Error struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"-"`
	// Error_ is the error details, it's exclusive with Payload.
	Error_ string `json:"error,omitempty"`
}

func doRequest[T any](s *sdk, method, pathStr string, body any) (*T, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	// Construct full URL using url.URL
	// Parse pathStr to properly handle query parameters
	relativeURL, err := url.Parse(pathStr)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}
	fullURL := s.baseURL.ResolveReference(relativeURL)

	req, err := http.NewRequest(method, fullURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var e = &Error{StatusCode: resp.StatusCode}
		if err = json.Unmarshal(respBody, e); err != nil {
			e.Error_ = errors.Wrapf(err, "response body: %s", string(respBody)).Error()
		}
		return nil, errors.New(e.Error_)
	}

	var out T
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return &out, nil
}
