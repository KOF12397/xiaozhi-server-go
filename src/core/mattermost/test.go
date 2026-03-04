package mattermost  
  
import (  
    "context"  
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
    // 从环境变量读取测试配置（增加日志提示）  
    baseURL := os.Getenv(envVarBaseURL)  
    token := os.Getenv(envVarToken)  
    channel := os.Getenv(envVarTestChannel)  
  
    // 检查配置完整性（更友好的提示）  
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
        t.Skipf("跳过Mattermost测试：缺少环境变量 %v，请配置后重新运行", missingEnvs)  
    }  
  
    // 初始化日志器（增加错误处理）  
    logger, err := utils.NewLogger(&utils.LogCfg{  
        LogLevel:  "DEBUG",  
        LogFormat: "{time:YYYY-MM-DD HH:mm:ss} - {level} - {message}",  
        LogDir:    "logs",  
        LogFile:   "test.log",  
    })  
    if err != nil {  
        t.Fatalf("创建日志器失败: %v", err)  
    }  
  
    // 创建客户端配置  
    config := &configs.MattermostConfig{  
        BaseURL: baseURL,  
        Token:   token,  
        Timeout: testTimeoutSeconds,  
    }  
  
    // 测试客户端创建  
    client := NewClient(config, logger)  
    if client == nil {  
        t.Fatal("创建Mattermost客户端失败")  
    }  
  
    // 测试客户端初始化  
    err = client.Initialize()  
    if err != nil {  
        t.Errorf("Mattermost客户端初始化失败: %v", err)  
        return  
    }  
  
    // 测试消息发送  
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)  
    defer cancel()  
      
    err = client.SendMessage(ctx, channel, testMessage)  
    if err != nil {  
        t.Errorf("Mattermost消息发送失败: %v", err)  
        return  
    }  
  
    // 测试资源清理  
    err = client.Cleanup()  
    if err != nil {  
        t.Errorf("Mattermost客户端清理失败: %v", err)  
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
        b.Skip("缺少环境变量，跳过基准测试")  
    }  
  
    logger, _ := utils.NewLogger(&utils.LogCfg{  
        LogLevel:  "error",  
        LogFormat: "{time:YYYY-MM-DD HH:mm:ss} - {level} - {message}",  
        LogDir:    "logs",  
        LogFile:   "benchmark.log",  
    })  
      
    config := &configs.MattermostConfig{  
        BaseURL: baseURL,  
        Token:   token,  
        Timeout: testTimeoutSeconds,  
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
  
// TestClient_WithoutLogger 测试无日志器场景（兼容优化后的client.go）  
func TestClient_WithoutLogger(t *testing.T) {  
    // 测试传入nil logger是否会panic  
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
