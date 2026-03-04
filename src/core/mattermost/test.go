package mattermost  
  
import (  
    "context"  
    "os"  
    "testing"  
    "time"  
    "xiaozhi-server-go/src/configs"  
)  
  
// TestClient 测试Mattermost客户端  
func TestClient(t *testing.T) {  
    baseURL := os.Getenv("MATTERMOST_BASE_URL")  
    token := os.Getenv("MATTERMOST_TOKEN")  
    channel := os.Getenv("MATTERMOST_TEST_CHANNEL")  
  
    if baseURL == "" || token == "" || channel == "" {  
        t.Skip("未配置Mattermost测试环境变量，跳过测试")  
    }  
  
    config := &configs.MattermostConfig{  
        BaseURL: baseURL,  
        Token:   token,  
        Timeout: 10,  
    }  
      
    client := NewClient(config, nil)  
    if client == nil {  
        t.Fatal("创建客户端失败")  
    }  
      
    err := client.Initialize()  
    if err != nil {  
        t.Errorf("初始化客户端失败: %v", err)  
    }  
      
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)  
    defer cancel()  
      
    err = client.SendMessage(ctx, channel, "测试消息")  
    if err != nil {  
        t.Errorf("发送消息失败: %v", err)  
    }  
      
    client.Cleanup()  
}  
  
// TestClientConfig 测试客户端配置  
func TestClientConfig(t *testing.T) {  
    config := &configs.MattermostConfig{  
        BaseURL: "http://test.example.com",  
        Token:   "test-token",  
        Timeout: 10,  
    }  
      
    client := NewClient(config, nil)  
    if client == nil {  
        t.Fatal("创建客户端失败")  
    }  
      
    if client.baseURL != config.BaseURL {  
        t.Errorf("BaseURL不匹配: 期望 %s, 实际 %s", config.BaseURL, client.baseURL)  
    }  
      
    if client.token != config.Token {  
        t.Errorf("Token不匹配: 期望 %s, 实际 %s", config.Token, client.token)  
    }  
}
