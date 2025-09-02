package server

import (
	"net/http"

	"github.com/dmitrii/llm-gateway/api"
	"github.com/dmitrii/llm-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	proxy *proxy.Proxy
}

func NewProxyHandler(proxy *proxy.Proxy) *ProxyHandler {
	return &ProxyHandler{
		proxy: proxy,
	}
}

// FindPets implements all the handlers in the ServerInterface
func (p *ProxyHandler) CreateChatCompletion(c *gin.Context) {
	var req api.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := p.proxy.ChatCompletionsHandler(c, req)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
