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
  
// 测试常量定义（便于维护）  
const (  
    testTimeoutSeconds = 10  
    envVarBaseURL      = "MATTERMOST_BASE_URL"  
    envVarToken        = "MATTERMOST_TOKEN"  
    envVarTestChannel  = "MATTERMOST_TEST_CHANNEL"  
    testMessage        = "测试消息 - 小智服务端Mattermost客户端测试"  
)  
  
// TestClient 测试Mattermost客户端完整流程  
func TestClient(t *testing.T) {  
    // 从环境变量读取测试配置  
    baseURL := os.Getenv(envVarBaseURL)  
    token := os.Getenv(envVarToken)  
    channel := os.Getenv(envVarTestChannel)  
  
    // 检查配置完整性  
    missingEnvs := make([]string, 0)  
    if baseURL == "" {  
        missingEnvs = append(missingEnvs, envVarBaseURL)  
    }  
    if token == "" {  
        missingEnvs = append(missingEnvs, envVarToken)  
    }  
    if channel == "" {  
        missingEnvs = append(missingEnvs, envVarTestChannel)  
    }  
  
    if len(missingEnvs) > 0 {  
        t.Skipf("未配置Mattermost测试环境变量，跳过测试。缺少: %v", missingEnvs)  
        return  
    }  
  
    // 创建配置  
    config := &configs.MattermostConfig{  
        BaseURL: baseURL,  
        Token:   token,  
        Timeout: testTimeoutSeconds,  
    }  
  
    // 创建日志器 - 提供所有必需字段  
    logger, err := utils.NewLogger(&utils.LogCfg{  
        LogLevel:  "DEBUG",  
        LogFormat: "{time:YYYY-MM-DD HH:mm:ss} - {level} - {message}",  
        LogDir:    "logs",  
        LogFile:   "test.log",  
    })  
    if err != nil {  
        t.Fatalf("创建日志器失败: %v", err)  
    }  
  
    // 创建客户端  
    client := NewClient(config, logger)  
    if client == nil {  
        t.Fatal("创建客户端失败")  
    }  
  
    // 初始化客户端  
    err = client.Initialize()  
    if err != nil {  
        t.Errorf("初始化客户端失败: %v", err)  
        return  
    }  
  
    // 测试消息发送  
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)  
    defer cancel()  
  
    err = client.SendMessage(ctx, channel, testMessage)  
    if err != nil {  
        t.Errorf("发送消息失败: %v", err)  
        return  
    }  
  
    t.Log("✅ Mattermost客户端测试通过")  
}  
  
// BenchmarkClient_SendMessage 性能测试  
func BenchmarkClient_SendMessage(b *testing.B) {  
    baseURL := os.Getenv(envVarBaseURL)  
    token := os.Getenv(envVarToken)  
    channel := os.Getenv(envVarTestChannel)  
  
    if baseURL == "" || token == "" || channel == "" {  
        b.Skip("未配置Mattermost测试环境变量，跳过性能测试")  
        return  
    }  
  
    config := &configs.MattermostConfig{  
        BaseURL: baseURL,  
        Token:   token,  
        Timeout: testTimeoutSeconds,  
    }  
  
    // 创建日志器 - 提供所有必需字段  
    logger, err := utils.NewLogger(&utils.LogCfg{  
        LogLevel:  "INFO",  
        LogFormat: "{time:YYYY-MM-DD HH:mm:ss} - {level} - {message}",  
        LogDir:    "logs",  
        LogFile:   "benchmark.log",  
    })  
    if err != nil {  
        b.Fatalf("创建日志器失败: %v", err)  
    }  
  
    client := NewClient(config, logger)  
    _ = client.Initialize()  
    defer client.Cleanup()  
  
    ctx := context.Background()  
    b.ResetTimer() // 重置计时器，排除初始化耗时  
  
    for i := 0; i < b.N; i++ {  
        msg := fmt.Sprintf("基准测试消息 - %d", i)  
        _ = client.SendMessage(ctx, channel, msg)  
    }  
}  
  
// TestClient_WithoutLogger 测试无日志器场景  
func TestClient_WithoutLogger(t *testing.T) {  
    config := &configs.MattermostConfig{  
        BaseURL: "http://test.example.com",  
        Token:   "test-token",  
    }  
    client := NewClient(config, nil)  
    if client == nil {  
        t.Fatal("创建客户端失败（nil logger）")  
    }  
    // 测试初始化不会panic  
    err := client.Initialize()  
    if err != nil {  
        t.Errorf("初始化客户端（nil logger）失败: %v", err)  
    }  
    t.Log("✅ 无日志器客户端测试通过")  
}
