package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
			response := map[string]interface{}{
				"echo":   string(message),
				"status": "success",
				"time":   time.Now().UnixNano(),
			}
			respBytes, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, respBytes)
		}
	}
}

var _ = Describe("WebSocket E2E Tests", Ordered, func() {
	var conn *websocket.Conn

	BeforeAll(func() {
		// 创建测试服务器
		server = httptest.NewServer(http.HandlerFunc(webSocketHandler))
		wsURL = "ws://127.0.0.1:8090/ws"

		DeferCleanup(func() {
			if server != nil {
				server.Close()
			}
		})
	})

	BeforeEach(func() {
		// 建立 WebSocket 连接
		var err error
		conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
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
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			Expect(err).NotTo(HaveOccurred())

			// 接收响应
			messageType, message, err := conn.ReadMessage()
			Expect(err).NotTo(HaveOccurred())
			Expect(messageType).To(Equal(websocket.TextMessage))

			var response map[string]interface{}
			err = json.Unmarshal(message, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["echo"]).To(Equal("ping"))
		})
	})

	Describe("消息通信测试", func() {
		var testMessages = []struct {
			name    string
			payload string
		}{
			{"简单文本", "Hello World"},
			{"JSON 数据", `{"action": "test", "value": 123}`},
			{"空消息", ""},
			{"长消息", "This is a very long message " + string(make([]byte, 1024))},
		}

		for _, tc := range testMessages {
			It(fmt.Sprintf("应该正确处理消息: %s", tc.name), func() {
				By("发送消息")
				err := conn.WriteMessage(websocket.TextMessage, []byte(tc.payload))
				Expect(err).NotTo(HaveOccurred())

				By("接收响应")
				_, message, err := conn.ReadMessage()
				Expect(err).NotTo(HaveOccurred())

				var response map[string]interface{}
				err = json.Unmarshal(message, &response)
				Expect(err).NotTo(HaveOccurred())

				By("验证响应")
				Expect(response["echo"]).To(Equal(tc.payload))
				Expect(response["status"]).To(Equal("success"))
				Expect(response["time"]).To(BeNumerically(">", 0))
			})
		}
	})

	Describe("并发连接测试", func() {
		It("应该支持多个并发客户端", func(ctx context.Context) {
			const clientCount = 10
			var clients []*websocket.Conn

			DeferCleanup(func() {
				for _, client := range clients {
					if client != nil {
						client.Close()
					}
				}
			})

			By("创建多个客户端连接")
			for i := 0; i < clientCount; i++ {
				client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				Expect(err).NotTo(HaveOccurred())
				clients = append(clients, client)

				// 每个客户端发送唯一消息
				message := fmt.Sprintf("client-%d", i)
				err = client.WriteMessage(websocket.TextMessage, []byte(message))
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
				Expect(response["echo"]).To(Equal(expectedMessage))
			}
		}, SpecTimeout(10*time.Second))
	})
})
