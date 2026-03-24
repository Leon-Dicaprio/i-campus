package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 模拟 DeepSeek 的回答内容
const deepSeekIntro = "我是 DeepSeek，由幻方量化下的深度求索公司开发的智能助手。我可以帮你写代码、分析数据和回答问题。"

func main() {
	r := gin.Default()

	// 1. 配置 CORS (允许前端跨域访问)
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	r.Use(cors.New(config))

	// 2. 核心对话接口 (SSE 流式)
	r.POST("/api/chat", func(c *gin.Context) {
		var req struct {
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// 设置 SSE 必要的 Header
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		// 模拟 AI "思考" 和 "生成" 的过程
		c.Stream(func(w io.Writer) bool {
			// 模拟思考延迟
			time.Sleep(500 * time.Millisecond)

			// 构造回复内容 (此处仅为模拟，实际应调用 LLM API)
			responseText := fmt.Sprintf("收到你的问题：“%s”。\n\n%s", req.Message, deepSeekIntro)

			// 将回复打散成字符，模拟打字机效果
			runes := []rune(responseText)
			for _, char := range runes {
				// 构造 SSE 数据格式: data: <content>\n\n
				msg := fmt.Sprintf("data: %s\n\n", string(char))

				if _, err := w.Write([]byte(msg)); err != nil {
					return false // 客户端断开连接
				}

				// 刷新缓冲区，确保前端立即收到
				c.Writer.Flush()

				// 随机延迟，模拟真实的打字速度
				time.Sleep(time.Duration(rand.Intn(50)+20) * time.Millisecond)
			}
			return false // 结束流
		})
	})

	fmt.Println("DeepSeek MVP Backend running on :8080")
	r.Run(":8080")
}
