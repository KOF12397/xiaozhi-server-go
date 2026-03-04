package mattermost  
  
import (  
    "context"  
    "fmt"  
    "os"  
    "testing"  
    "time"  
    "xiaozhi-server-go/src/configs"  
    "xiaozhi-server-go/src/core/utils"  
)  
  
// TestClient 测试Mattermost客户端  
func TestClient(t *testing.T) {  
    baseURL := os.Getenv("MATTERMOST_BASE_URL")  
    token := os.Getenv("MATTERMOST_TOKEN")  
    channel := os.Getenv("MATTERMOST_TEST_CHANNEL")  
  
    if baseURL == "" || token == "" || channel == "" {  
        t.Skip("未配置Mattermost测试环境变量，跳过测试")  
    }  
  
    // 使用最小化配置创建logger  
    logger, err := utils.NewLogger(&utils.LogCfg{  
        LogLevel:  "DEBUG",  
        LogFormat: "{time:YYYY-MM-DD HH:mm:ss} - {level} - {message}",  
        LogDir:    "logs",  
        LogFile:   "test.log",  
    })  
    if err != nil {  
        t.Fatalf("创建logger失败: %v", err)  
    }  
  
    config := &configs.MattermostConfig{  
        BaseURL: baseURL,  
        Token:   token,  
        Timeout: 10,  
    }  
      
    client := NewClient(config, logger)  
    if err := client.Initialize(); err != nil {  
        t.Fatalf("初始化客户端失败: %v", err)  
    }  
    defer client.Cleanup()  
  
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)  
    defer cancel()  
      
    err = client.SendMessage(ctx, channel, "测试消息")  
    if err != nil {  
        t.Errorf("发送消息失败: %v", err)  
    }  
}
