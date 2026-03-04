package mattermost  
  
import (  
    "context"  
    "os"  
    "testing"  
    "time"  
    "xiaozhi-server-go/src/configs"  
    "xiaozhi-server-go/src/core/utils"  
)  
  
// TestClient 测试Mattermost客户端  
func TestClient(t *testing.T) {  
    // 优先从环境变量读取测试配置  
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
      
    // 创建日志器  
    logger, _ := utils.NewLogger(&utils.LogCfg{  
        LogLevel:  "debug",  
        LogFormat: "text",  
    })  
      
    // 创建客户端  
    client := NewClient(config, logger)  
      
    // 测试初始化  
    if err := client.Initialize(); err != nil {  
        t.Fatalf("初始化失败: %v", err)  
    }  
      
    // 测试发送消息（需要真实的测试环境）  
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)  
    defer cancel()  
    err := client.SendMessage(ctx, channel, "测试消息")  
    if err != nil {  
        t.Logf("发送消息失败（可能是配置问题）: %v", err)  
    } else {  
        t.Log("发送消息成功")  
    }  
      
    // 清理资源  
    if err := client.Cleanup(); err != nil {  
        t.Errorf("清理失败: %v", err)  
    }  
}
