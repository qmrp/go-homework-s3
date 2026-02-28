package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/internal/api/response"
	"github.com/qmrp/go-homework-s3/cmd/huayi-im/tests/sdk"
)

var _ = sdk.NewSDK("http://127.0.0.1:8090")

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Suite")
}

var (
	server   *httptest.Server
	wsURL    string
	upgrader = websocket.Upgrader{}
)

// WebSocket 处理器示例
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 业务逻辑：回显消息
		if messageType == websocket.TextMessage {
			var request map[string]interface{}
			err := json.Unmarshal(message, &request)
			Expect(err).NotTo(HaveOccurred(), "应该成功解析 JSON 消息"+string(message))
			_, ok := request["message-type"].(string)
			Expect(ok).To(BeTrue(), "message-type 字段缺失或类型错误")

			respBytes, _ := json.Marshal(request)

			conn.WriteMessage(websocket.TextMessage, respBytes)

		}
	}
}

var _ = Describe("WebSocket E2E Tests", Ordered, func() {
	var conn *websocket.Conn
	var logResp response.UserDetailResponse
	var userClient sdk.UserClient

	BeforeAll(func() {
		// 创建测试服务器
		server = httptest.NewServer(http.HandlerFunc(webSocketHandler))
		// 使用测试服务器的地址，将 http 替换为 ws
		serverURL := server.URL
		wsURL = "ws" + serverURL[4:] // 将 http 替换为 ws

		// 登录操作
		userClient = sdk.GetSDK().Guest()
		var loginErr error
		logResp, loginErr = userClient.Login("zhou")
		log.Println(logResp)
		Expect(loginErr).NotTo(HaveOccurred(), "登录失败")

		DeferCleanup(func() {
			if server != nil {
				server.Close()
			}
		})
	})

	BeforeEach(func() {
		// 建立 WebSocket 连接
		var err error
		conn, _, err = websocket.DefaultDialer.Dial(wsURL+"?sid="+logResp.Sid, nil)
		Expect(err).NotTo(HaveOccurred(), "应该成功连接到 WebSocket 服务器")

		// 设置连接超时
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	})

	AfterEach(func() {
		// 清理连接
		if conn != nil {
			conn.Close()
		}
	})

	Describe("基本连接测试", func() {
		It("应该成功建立连接", func() {
			Expect(conn).NotTo(BeNil())

			// 发送 ping 测试连接
			err := conn.WriteMessage(websocket.TextMessage, []byte(`{"message-type": "ping","from":"zhou", "message-id": "`+time.Now().Format(time.RFC3339Nano)+`"}`))
			Expect(err).NotTo(HaveOccurred())

			// 接收响应
			messageType, _, err := conn.ReadMessage()
			Expect(err).NotTo(HaveOccurred(), "应该成功接收消息")
			Expect(messageType).To(Equal(websocket.TextMessage))
		})
	})

	Describe("消息通信测试", func() {
		It("发送普通单对单消息", func() {
			err := userClient.Message().SendMessage("test-message", nil, &[]string{"qmrp"})
			Expect(err).NotTo(HaveOccurred(), "发送消息失败")
		})
		It("发送topic消息", func() {
			topic := "topic"
			err := userClient.Message().SendMessage("test-topic-message", &topic, &[]string{"qmrp"})
			Expect(err).NotTo(HaveOccurred(), "发送消息失败")
		})
	})

	Describe("topic相关测试", func() {
		topic := "test-topic"
		It("创建topic", func() {
			err := userClient.Topic().CreateTopic(topic)
			Expect(err).NotTo(HaveOccurred(), "创建topic失败")
		})
		It("加入topic", func() {
			err := userClient.Topic().JoinTopic(topic)
			Expect(err).NotTo(HaveOccurred(), "加入topic失败")
		})
		It("发送topic消息", func() {
			err := userClient.Message().SendMessage("手动创建显示加入并发送信息", &topic, &[]string{"qmrp"})
			Expect(err).NotTo(HaveOccurred(), "发送消息失败")
		})
		It("退出topic", func() {
			topic := "test-topic"
			err := userClient.Topic().LeaveTopic(topic)
			Expect(err).NotTo(HaveOccurred(), "退出topic失败")
		})
	})

	Describe("并发连接测试", func() {
		It("应该支持多个并发客户端", func(ctx context.Context) {
			const clientCount = 10
			var clients []*websocket.Conn
			var userClients []sdk.UserClient

			DeferCleanup(func() {
				for _, client := range clients {
					if client != nil {
						client.Close()
					}
				}
				for _, u := range userClients {
					u.Logout()
				}
			})

			By("创建多个客户端连接")
			for i := 0; i < clientCount; i++ {
				By(fmt.Sprintf("创建第 %d 个客户端", i))
				mu := sdk.GetSDK().Guest()
				userClients = append(userClients, mu)

				logResp1, loginErr := mu.Login(fmt.Sprintf("client-%d", i))
				Expect(loginErr).NotTo(HaveOccurred(), fmt.Sprintf("第 %d 个客户端登录失败", i))

				client, _, err := websocket.DefaultDialer.Dial(wsURL+"?sid="+logResp1.Sid, nil)
				Expect(err).NotTo(HaveOccurred())
				clients = append(clients, client)

				// 每个客户端发送唯一消息
				message := fmt.Sprintf("client-%d", i)
				err = client.WriteMessage(websocket.TextMessage, []byte(`{"message-type": "ping","from":"`+message+`", "message-id": "`+time.Now().Format(time.RFC3339Nano)+`"}`))
				Expect(err).NotTo(HaveOccurred())
			}

			By("验证所有客户端都能收到响应")
			for i, client := range clients {
				client.SetReadDeadline(time.Now().Add(3 * time.Second))
				_, message, err := client.ReadMessage()
				Expect(err).NotTo(HaveOccurred())

				var response map[string]interface{}
				err = json.Unmarshal(message, &response)
				Expect(err).NotTo(HaveOccurred())

				expectedMessage := fmt.Sprintf("client-%d", i)
				Expect(response["from"]).To(Equal(expectedMessage))
			}
		}, SpecTimeout(10*time.Second))
	})

	Describe("测试查询相关功能", func() {
		topic := "test-topic"
		It("查询topic列表", func() {
			topics, err := userClient.Topic().GetTopics()
			Expect(err).NotTo(HaveOccurred(), "查询topic列表失败")
			Expect(topics.Total).To(Equal(2))
			// 检查列表是否包含 test-topic
			Expect(topics.List).To(ContainElement(response.TopicResponse{Topic: topic}))
		})
		It("删除topic", func() {
			err := userClient.Topic().DeleteTopic(topic)
			Expect(err).NotTo(HaveOccurred(), "删除topic失败")
		})
		It("查询用户加入的topic列表", func() {
			users, err := userClient.User().GetUsers()
			Expect(err).NotTo(HaveOccurred(), "查询用户列表失败")
			Expect(users.Total).To(Equal(11))
			// 检查列表是否包含 test-topic
			Expect(users.List).To(ContainElement(response.UserResponse{Username: "zhou"}))
		})
		It("最后登出用户", func() {
			err := userClient.Logout()
			Expect(err).NotTo(HaveOccurred(), "登出用户失败")
		})
	})
})
